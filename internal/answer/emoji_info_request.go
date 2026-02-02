package answer

import (
	"encoding/json"
	"sort"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type emojiTemplateEntry struct {
	ID      uint32 `json:"id"`
	Achieve uint32 `json:"achieve"`
}

func EmojiInfoRequest(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_11601
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 11601, err
	}

	entries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/emoji_template.json")
	if err != nil {
		return 0, 11601, err
	}

	emojiList := make([]uint32, 0)
	for _, entry := range entries {
		var template emojiTemplateEntry
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return 0, 11601, err
		}
		if template.Achieve == 1 {
			emojiList = append(emojiList, template.ID)
		}
	}
	sort.Slice(emojiList, func(i, j int) bool {
		return emojiList[i] < emojiList[j]
	})

	response := protobuf.SC_11602{EmojiList: emojiList}
	return client.SendMessage(11602, &response)
}
