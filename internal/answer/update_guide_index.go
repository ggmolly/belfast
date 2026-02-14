package answer

import (
	"context"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
)

func UpdateGuideIndex(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11016
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11018, err
	}
	response := protobuf.SC_11018{
		Result:   proto.Uint32(0),
		DropList: []*protobuf.DROPINFO{},
	}
	if payload.GetType() == 1 {
		client.Commander.NewGuideIndex = payload.GetGuideIndex()
	} else {
		client.Commander.GuideIndex = payload.GetGuideIndex()
	}
	ctx := context.Background()
	if err := orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		return client.Commander.SaveTx(ctx, tx)
	}); err != nil {
		response.Result = proto.Uint32(1)
	}
	return client.SendMessage(11018, &response)
}
