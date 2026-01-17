package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
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
		Id:            proto.Uint32(0),
		SkinId:        proto.Uint32(0),
		Favorite:      proto.Uint32(0),
		CurChatGroup:  proto.Uint32(0),
		ChatGroupList: []*protobuf.JUUS_CHAT_GROUP{},
	}
}
