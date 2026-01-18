package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	juusagramPlaceholderShipGroup = 960007
	juusagramPlaceholderChatGroup = 1
)

func JuustagramData(buffer *[]byte, client *connection.Client) (int, int, error) {
	// TODO: Populate Juustagram groups from persistence.
	response := protobuf.SC_11711{
		Groups: []*protobuf.JUUS_GROUP{juusGroupPlaceholder()},
	}
	return client.SendMessage(11711, &response)
}

func juusGroupPlaceholder() *protobuf.JUUS_GROUP {
	return &protobuf.JUUS_GROUP{
		Id:           proto.Uint32(juusagramPlaceholderShipGroup),
		SkinId:       proto.Uint32(0),
		Favorite:     proto.Uint32(0),
		CurChatGroup: proto.Uint32(juusagramPlaceholderChatGroup),
		ChatGroupList: []*protobuf.JUUS_CHAT_GROUP{
			{
				Id:        proto.Uint32(juusagramPlaceholderChatGroup),
				OpTime:    proto.Uint32(0),
				ReadFlag:  proto.Uint32(0),
				ReplyList: []*protobuf.KEYVALUE_P11{},
			},
		},
	}
}
