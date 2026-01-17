package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ShopData(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_16200{
		Month: proto.Uint32(uint32(time.Now().Month())),
	}
	return client.SendMessage(16200, &response)
}
