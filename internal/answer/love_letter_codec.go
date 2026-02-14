package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type loveLetterSnapshot struct {
	ConvertedItems   []orm.LoveLetterConvertedItem
	RewardedIDs      []uint32
	Medals           []orm.LoveLetterMedalState
	Letters          []orm.LoveLetterLetterState
	ConvertedLetters []orm.LoveLetterLetterState
}

func decodeCS12400(payload []byte) (uint32, error) {
	req := &protobuf.CS_12400{}
	if err := proto.Unmarshal(payload, req); err != nil {
		return 0, err
	}
	if req.Id == nil {
		return 0, fmt.Errorf("missing required field id")
	}
	return req.GetId(), nil
}

func decodeCS12402(payload []byte) ([]uint32, error) {
	req := &protobuf.CS_12402{}
	if err := proto.Unmarshal(payload, req); err != nil {
		return nil, err
	}
	return append([]uint32{}, req.GetIdList()...), nil
}

func decodeCS12404(payload []byte) ([]orm.LoveLetterConvertedItem, error) {
	req := &protobuf.CS_12404{}
	if err := proto.Unmarshal(payload, req); err != nil {
		return nil, err
	}
	items := make([]orm.LoveLetterConvertedItem, 0, len(req.GetItemList()))
	for _, item := range req.GetItemList() {
		if item == nil || item.ItemId == nil || item.GroupId == nil || item.Year == nil {
			return nil, fmt.Errorf("missing required converted item fields")
		}
		items = append(items, orm.LoveLetterConvertedItem{
			ItemID:  item.GetItemId(),
			GroupID: item.GetGroupId(),
			Year:    item.GetYear(),
		})
	}
	return items, nil
}

func decodeCS12406(payload []byte) (uint32, error) {
	req := &protobuf.CS_12406{}
	if err := proto.Unmarshal(payload, req); err != nil {
		return 0, err
	}
	if req.Type == nil {
		return 0, fmt.Errorf("missing required field type")
	}
	return req.GetType(), nil
}

func decodeCS12408(payload []byte) (uint32, error) {
	req := &protobuf.CS_12408{}
	if err := proto.Unmarshal(payload, req); err != nil {
		return 0, err
	}
	if req.GroupId == nil {
		return 0, fmt.Errorf("missing required field group_id")
	}
	return req.GetGroupId(), nil
}

func decodeCS12410(payload []byte) (uint32, error) {
	req := &protobuf.CS_12410{}
	if err := proto.Unmarshal(payload, req); err != nil {
		return 0, err
	}
	if req.LetterId == nil {
		return 0, fmt.Errorf("missing required field letter_id")
	}
	return req.GetLetterId(), nil
}

func newSC12401(result uint32) *protobuf.SC_12401 {
	return &protobuf.SC_12401{Result: proto.Uint32(result)}
}

func newSC12403(result uint32, drops []*protobuf.DROPINFO) *protobuf.SC_12403 {
	return &protobuf.SC_12403{Result: proto.Uint32(result), DropList: drops}
}

func newSC12405(result uint32) *protobuf.SC_12405 {
	return &protobuf.SC_12405{Result: proto.Uint32(result)}
}

func newSC12407(snapshot loveLetterSnapshot) *protobuf.SC_12407 {
	converted := make([]*protobuf.PT_OLD_LOVER_ITEM, 0, len(snapshot.ConvertedItems))
	for _, item := range snapshot.ConvertedItems {
		converted = append(converted, &protobuf.PT_OLD_LOVER_ITEM{
			ItemId:  proto.Uint32(item.ItemID),
			GroupId: proto.Uint32(item.GroupID),
			Year:    proto.Uint32(item.Year),
		})
	}
	medals := make([]*protobuf.PT_LOVE_LETTER_MEDAL, 0, len(snapshot.Medals))
	for _, medal := range snapshot.Medals {
		medals = append(medals, &protobuf.PT_LOVE_LETTER_MEDAL{
			GroupId: proto.Uint32(medal.GroupID),
			Exp:     proto.Uint32(medal.Exp),
			Level:   proto.Uint32(medal.Level),
		})
	}
	letters := make([]*protobuf.PT_SHIP_LOVE_LETTER, 0, len(snapshot.Letters))
	for _, letter := range snapshot.Letters {
		letters = append(letters, &protobuf.PT_SHIP_LOVE_LETTER{
			GroupId:      proto.Uint32(letter.GroupID),
			LetterIdList: append([]uint32{}, letter.LetterIDList...),
		})
	}
	convertedLetters := make([]*protobuf.PT_SHIP_LOVE_LETTER, 0, len(snapshot.ConvertedLetters))
	for _, letter := range snapshot.ConvertedLetters {
		convertedLetters = append(convertedLetters, &protobuf.PT_SHIP_LOVE_LETTER{
			GroupId:      proto.Uint32(letter.GroupID),
			LetterIdList: append([]uint32{}, letter.LetterIDList...),
		})
	}
	return &protobuf.SC_12407{
		ConvertedList:       converted,
		RewardedList:        append([]uint32{}, snapshot.RewardedIDs...),
		MedalList:           medals,
		LetterList:          letters,
		ConvertedLetterList: convertedLetters,
	}
}

func newSC12409(result uint32) *protobuf.SC_12409 {
	return &protobuf.SC_12409{Ret: proto.Uint32(result)}
}

func newSC12411(content string) *protobuf.SC_12411 {
	return &protobuf.SC_12411{Content: proto.String(content)}
}
