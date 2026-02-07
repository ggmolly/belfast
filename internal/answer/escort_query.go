package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/misc"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func EscortQuery(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13301
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13302, err
	}

	if payload.GetType() != 0 && payload.GetType() != 1 {
		logger.LogEvent("Escort", "Query", fmt.Sprintf("unknown query type %d", payload.GetType()), logger.LOG_LEVEL_INFO)
	}

	// Escort config is expected to come from belfast-data imports. If config isn't
	// available yet, fall back to returning persisted state only.
	if _, err := misc.GetEscortConfig(); err != nil {
		logger.LogEvent("Escort", "Config", fmt.Sprintf("failed to load escort config: %v", err), logger.LOG_LEVEL_ERROR)
	}

	escortInfo := []*protobuf.ESCORT_INFO{}
	if client.Commander != nil {
		infos, err := misc.LoadEscortState(client.Commander.AccountID)
		if err != nil {
			return 0, 13302, err
		}
		escortInfo = infos
	}

	response := protobuf.SC_13302{
		EscortInfo: escortInfo,
		DropList:   []*protobuf.DROPINFO{},
	}
	return client.SendMessage(13302, &response)
}
