package answer

import (
	"errors"
	"sort"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GetChapterDropShipList(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13109
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13110, err
	}
	chapterID := payload.GetId()
	if chapterID == 0 {
		return 0, 13110, errors.New("missing chapter id")
	}
	// Existence check only (unlock/access validation is out of scope).
	template, err := loadChapterTemplate(chapterID, 0)
	if err != nil {
		return 0, 13110, err
	}
	if template == nil {
		return 0, 13110, errors.New("chapter not found")
	}

	drops, err := orm.GetChapterDrops(orm.GormDB, client.Commander.CommanderID, chapterID)
	if err != nil {
		return 0, 13110, err
	}
	unique := make(map[uint32]struct{}, len(drops))
	for _, drop := range drops {
		unique[drop.ShipID] = struct{}{}
	}
	shipIDs := make([]uint32, 0, len(unique))
	for shipID := range unique {
		shipIDs = append(shipIDs, shipID)
	}
	sort.Slice(shipIDs, func(i, j int) bool { return shipIDs[i] < shipIDs[j] })

	response := protobuf.SC_13110{DropShipList: shipIDs}
	return client.SendMessage(13110, &response)
}
