package answer

import (
	"errors"
	"sort"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	juusagramPlaceholderShipGroup = 960007
	juusagramPlaceholderChatGroup = 1
)

func JuustagramData(buffer *[]byte, client *connection.Client) (int, int, error) {
	if client.Commander == nil {
		return 0, 11711, errors.New("missing commander")
	}
	groups, err := orm.GetJuustagramGroups(client.Commander.CommanderID)
	if err != nil {
		return 0, 11711, err
	}
	if len(groups) == 0 {
		group, err := orm.CreateJuustagramGroup(client.Commander.CommanderID, juusagramPlaceholderShipGroup, juusagramPlaceholderChatGroup)
		if err != nil {
			return 0, 11711, err
		}
		groups = []orm.JuustagramGroup{*group}
	}
	responseGroups := make([]*protobuf.JUUS_GROUP, 0, len(groups))
	for _, group := range groups {
		responseGroups = append(responseGroups, juusGroupFromModel(group))
	}
	response := protobuf.SC_11711{Groups: responseGroups}
	return client.SendMessage(11711, &response)
}

func juusGroupFromModel(group orm.JuustagramGroup) *protobuf.JUUS_GROUP {
	chatGroups := make([]*protobuf.JUUS_CHAT_GROUP, 0, len(group.ChatGroups))
	sort.Slice(group.ChatGroups, func(i, j int) bool {
		return group.ChatGroups[i].ChatGroupID < group.ChatGroups[j].ChatGroupID
	})
	for _, chatGroup := range group.ChatGroups {
		replies := make([]*protobuf.KEYVALUE_P11, 0, len(chatGroup.ReplyList))
		sort.Slice(chatGroup.ReplyList, func(i, j int) bool {
			return chatGroup.ReplyList[i].Sequence < chatGroup.ReplyList[j].Sequence
		})
		for _, reply := range chatGroup.ReplyList {
			replies = append(replies, &protobuf.KEYVALUE_P11{
				Key:   proto.Uint32(reply.Key),
				Value: proto.Uint32(reply.Value),
			})
		}
		chatGroups = append(chatGroups, &protobuf.JUUS_CHAT_GROUP{
			Id:        proto.Uint32(chatGroup.ChatGroupID),
			OpTime:    proto.Uint32(chatGroup.OpTime),
			ReadFlag:  proto.Uint32(chatGroup.ReadFlag),
			ReplyList: replies,
		})
	}
	return &protobuf.JUUS_GROUP{
		Id:            proto.Uint32(group.GroupID),
		SkinId:        proto.Uint32(group.SkinID),
		Favorite:      proto.Uint32(group.Favorite),
		CurChatGroup:  proto.Uint32(group.CurChatGroup),
		ChatGroupList: chatGroups,
	}
}
