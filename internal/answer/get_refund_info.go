package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GetRefundInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	// No billing support yet; return an empty refund list intentionally.
	response := protobuf.SC_11024{
		Result:   proto.Uint32(0),
		ShopInfo: []*protobuf.REFUND_SHOPINFO{},
	}

	return client.SendMessage(11024, &response)
}
