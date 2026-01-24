package answer

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type JuustagramDiscussOption struct {
	DiscussID  uint32
	Index      uint32
	Text       string
	NpcReplyID uint32
}

type juustagramOptionMap map[uint32]map[uint32]JuustagramDiscussOption

func BuildJuustagramMessage(commanderID uint32, messageID uint32, now uint32) (*protobuf.INS_MESSAGE, error) {
	template, err := orm.GetJuustagramTemplate(messageID)
	if err != nil {
		return nil, err
	}
	state, err := orm.GetOrCreateJuustagramMessageState(commanderID, messageID, now)
	if err != nil {
		return nil, err
	}
	text, err := resolveJuustagramText(template.MessagePersist)
	if err != nil {
		return nil, err
	}
	messageTime := parseJuustagramTime(template.TimePersist, now)

	options, optionMap, err := loadJuustagramDiscussOptions(messageID)
	if err != nil {
		return nil, err
	}
	playerDiscuss, opReplyIDs, err := buildJuustagramPlayerDiscuss(commanderID, messageID, options, optionMap, now)
	if err != nil {
		return nil, err
	}
	return buildJuustagramMessagePayload(template, state, text, messageTime, playerDiscuss, opReplyIDs, now)
}

func ListJuustagramDiscussOptions(messageID uint32) ([]JuustagramDiscussOption, error) {
	options, _, err := loadJuustagramDiscussOptions(messageID)
	return options, err
}

func loadJuustagramDiscussOptions(messageID uint32) ([]JuustagramDiscussOption, juustagramOptionMap, error) {
	prefix := fmt.Sprintf("ins_op_%d_", messageID)
	entries, err := orm.ListJuustagramLanguageByPrefix(prefix)
	if err != nil {
		return nil, nil, err
	}
	replies, err := orm.ListJuustagramOpReplies(messageID)
	if err != nil {
		return nil, nil, err
	}
	replyMap := make(map[uint32]map[uint32]uint32)
	for _, reply := range replies {
		discussID, index, ok := parseJuustagramOpKey(reply.MessagePersist, fmt.Sprintf("op_reply_%d_", messageID))
		if !ok {
			continue
		}
		if replyMap[discussID] == nil {
			replyMap[discussID] = make(map[uint32]uint32)
		}
		replyMap[discussID][index] = reply.ID
	}

	options := make([]JuustagramDiscussOption, 0, len(entries))
	for _, entry := range entries {
		discussID, index, ok := parseJuustagramOpKey(entry.Key, prefix)
		if !ok {
			continue
		}
		option := JuustagramDiscussOption{
			DiscussID: discussID,
			Index:     index,
			Text:      entry.Value,
		}
		if replyMap[discussID] != nil {
			option.NpcReplyID = replyMap[discussID][index]
		}
		options = append(options, option)
	}
	sort.Slice(options, func(i, j int) bool {
		if options[i].DiscussID == options[j].DiscussID {
			return options[i].Index < options[j].Index
		}
		return options[i].DiscussID < options[j].DiscussID
	})
	optionMap := make(juustagramOptionMap)
	for _, option := range options {
		if optionMap[option.DiscussID] == nil {
			optionMap[option.DiscussID] = make(map[uint32]JuustagramDiscussOption)
		}
		optionMap[option.DiscussID][option.Index] = option
	}
	return options, optionMap, nil
}

func buildJuustagramPlayerDiscuss(commanderID uint32, messageID uint32, options []JuustagramDiscussOption, optionMap juustagramOptionMap, now uint32) ([]*protobuf.INS_PLAYER, []uint32, error) {
	selections, err := orm.ListJuustagramPlayerDiscuss(commanderID, messageID)
	if err != nil {
		return nil, nil, err
	}
	selectionMap := make(map[uint32]orm.JuustagramPlayerDiscuss)
	for _, selection := range selections {
		selectionMap[selection.DiscussID] = selection
	}
	grouped := make(map[uint32][]JuustagramDiscussOption)
	for _, option := range options {
		grouped[option.DiscussID] = append(grouped[option.DiscussID], option)
	}
	ids := make([]uint32, 0, len(grouped))
	for discussID := range grouped {
		ids = append(ids, discussID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	playerDiscuss := make([]*protobuf.INS_PLAYER, 0, len(ids))
	opReplyIDs := make([]uint32, 0)
	for _, discussID := range ids {
		optionsForDiscuss := grouped[discussID]
		selection, selected := selectionMap[discussID]
		if selected {
			if optionMap[discussID] == nil {
				return nil, nil, errors.New("missing discuss options")
			}
			option, ok := optionMap[discussID][selection.OptionIndex]
			if !ok {
				return nil, nil, errors.New("invalid discuss option")
			}
			text := option.Text
			if text == "" {
				text = selectionOptionText(optionsForDiscuss, selection.OptionIndex)
			}
			npcReplyID := selection.NpcReplyID
			if npcReplyID == 0 {
				npcReplyID = option.NpcReplyID
			}
			playerDiscuss = append(playerDiscuss, &protobuf.INS_PLAYER{
				Id:       proto.Uint32(discussID),
				Time:     proto.Uint32(selectionTime(selection.CommentTime, now)),
				TextList: []string{},
				Text:     proto.String(text),
				NpcReply: proto.Uint32(npcReplyID),
			})
			if npcReplyID != 0 {
				opReplyIDs = append(opReplyIDs, npcReplyID)
			}
			continue
		}
		textList := make([]string, 0, len(optionsForDiscuss))
		for _, option := range optionsForDiscuss {
			textList = append(textList, option.Text)
			if option.NpcReplyID != 0 {
				opReplyIDs = append(opReplyIDs, option.NpcReplyID)
			}
		}
		playerDiscuss = append(playerDiscuss, &protobuf.INS_PLAYER{
			Id:       proto.Uint32(discussID),
			Time:     proto.Uint32(now),
			TextList: textList,
			Text:     proto.String(""),
			NpcReply: proto.Uint32(0),
		})
	}
	return playerDiscuss, uniqueUint32(opReplyIDs), nil
}

func buildJuustagramMessagePayload(template *orm.JuustagramTemplate, state *orm.JuustagramMessageState, messageText string, messageTime uint32, playerDiscuss []*protobuf.INS_PLAYER, opReplyIDs []uint32, now uint32) (*protobuf.INS_MESSAGE, error) {
	npcDiscuss, replyIDs, err := buildJuustagramNpcDiscuss(template.NpcDiscussPersist, now)
	if err != nil {
		return nil, err
	}
	for _, replyID := range opReplyIDs {
		replyIDs = append(replyIDs, replyID)
	}
	npcReply, err := buildJuustagramNpcReply(replyIDs, now)
	if err != nil {
		return nil, err
	}
	return &protobuf.INS_MESSAGE{
		Id:            proto.Uint32(template.ID),
		Time:          proto.Uint32(messageTime),
		Text:          proto.String(messageText),
		Picture:       proto.String(template.PicturePersist),
		PlayerDiscuss: playerDiscuss,
		NpcDiscuss:    npcDiscuss,
		NpcReply:      npcReply,
		Good:          proto.Uint32(state.GoodCount),
		IsGood:        proto.Uint32(state.IsGood),
		IsRead:        proto.Uint32(state.IsRead),
	}, nil
}

func buildJuustagramNpcDiscuss(ids orm.JuustagramUint32List, now uint32) ([]*protobuf.INS_NPC, []uint32, error) {
	if len(ids) == 0 {
		return []*protobuf.INS_NPC{}, []uint32{}, nil
	}
	entries := make([]*protobuf.INS_NPC, 0, len(ids))
	replyIDs := make([]uint32, 0)
	for _, id := range ids {
		template, err := orm.GetJuustagramNpcTemplate(id)
		if err != nil {
			return nil, nil, err
		}
		entry, err := buildJuustagramNpcEntry(template, now)
		if err != nil {
			return nil, nil, err
		}
		entries = append(entries, entry)
		replyIDs = append(replyIDs, template.NpcReplyPersist...)
	}
	return entries, replyIDs, nil
}

func buildJuustagramNpcReply(replyIDs []uint32, now uint32) ([]*protobuf.INS_NPC, error) {
	if len(replyIDs) == 0 {
		return []*protobuf.INS_NPC{}, nil
	}
	queue := uniqueUint32(replyIDs)
	seen := make(map[uint32]bool)
	entries := make([]*protobuf.INS_NPC, 0)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if seen[id] {
			continue
		}
		seen[id] = true
		template, err := orm.GetJuustagramNpcTemplate(id)
		if err != nil {
			return nil, err
		}
		entry, err := buildJuustagramNpcEntry(template, now)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
		queue = append(queue, template.NpcReplyPersist...)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].GetId() < entries[j].GetId()
	})
	return entries, nil
}

func buildJuustagramNpcEntry(template *orm.JuustagramNpcTemplate, now uint32) (*protobuf.INS_NPC, error) {
	text, err := resolveJuustagramText(template.MessagePersist)
	if err != nil {
		return nil, err
	}
	entryTime := parseJuustagramTime(template.TimePersist, now)
	return &protobuf.INS_NPC{
		Id:       proto.Uint32(template.ID),
		Time:     proto.Uint32(entryTime),
		Text:     proto.String(text),
		NpcReply: append([]uint32{}, template.NpcReplyPersist...),
	}, nil
}

func parseJuustagramTime(config orm.JuustagramTimeConfig, fallback uint32) uint32 {
	if len(config) < 2 || len(config[0]) < 3 || len(config[1]) < 3 {
		return fallback
	}
	date := config[0]
	timeParts := config[1]
	if date[0] == 0 || date[1] == 0 || date[2] == 0 {
		return fallback
	}
	parsed := time.Date(date[0], time.Month(date[1]), date[2], timeParts[0], timeParts[1], timeParts[2], 0, time.UTC)
	return uint32(parsed.Unix())
}

func parseJuustagramOpKey(key string, prefix string) (uint32, uint32, bool) {
	if !strings.HasPrefix(key, prefix) {
		return 0, 0, false
	}
	parts := strings.Split(strings.TrimPrefix(key, prefix), "_")
	if len(parts) < 2 {
		return 0, 0, false
	}
	discussID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return 0, 0, false
	}
	index, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return 0, 0, false
	}
	return uint32(discussID), uint32(index), true
}

func selectionTime(value uint32, fallback uint32) uint32 {
	if value == 0 {
		return fallback
	}
	return value
}

func selectionOptionText(options []JuustagramDiscussOption, index uint32) string {
	for _, option := range options {
		if option.Index == index {
			return option.Text
		}
	}
	return ""
}

func resolveJuustagramText(key string) (string, error) {
	if key == "" {
		return "", nil
	}
	text, err := orm.GetJuustagramLanguage(key)
	if err == nil {
		return text, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return "", err
}

func uniqueUint32(values []uint32) []uint32 {
	seen := make(map[uint32]bool)
	unique := make([]uint32, 0, len(values))
	for _, value := range values {
		if value == 0 || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	sort.Slice(unique, func(i, j int) bool { return unique[i] < unique[j] })
	return unique
}

func ensureJuustagramOption(messageID uint32, discussID uint32, index uint32) (JuustagramDiscussOption, error) {
	options, optionMap, err := loadJuustagramDiscussOptions(messageID)
	if err != nil {
		return JuustagramDiscussOption{}, err
	}
	if optionMap[discussID] == nil {
		return JuustagramDiscussOption{}, errors.New("missing discuss options")
	}
	option, ok := optionMap[discussID][index]
	if !ok {
		return JuustagramDiscussOption{}, errors.New("invalid discuss option")
	}
	if option.Text == "" {
		for _, candidate := range options {
			if candidate.DiscussID == discussID && candidate.Index == index {
				option.Text = candidate.Text
				break
			}
		}
	}
	return option, nil
}
