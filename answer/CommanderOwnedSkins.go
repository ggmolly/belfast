package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func CommanderOwnedSkins(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_12201
	response.SkinList = make([]*protobuf.IDTIMEINFO, len(client.Commander.OwnedSkins))
	for i, skin := range client.Commander.OwnedSkins {
		var expiryTimestamp uint32
		if skin.ExpiresAt != nil {
			expiryTimestamp = uint32(skin.ExpiresAt.Unix())
		}
		response.SkinList[i] = &protobuf.IDTIMEINFO{
			Id:   proto.Uint32(skin.SkinID),
			Time: proto.Uint32(expiryTimestamp),
		}
	}
	// TODO: find out what these are
	response.ForbiddenSkinList = []uint32{}
	response.ForbiddenSkinType = []uint32{}
	return client.SendMessage(12201, &response)
}
