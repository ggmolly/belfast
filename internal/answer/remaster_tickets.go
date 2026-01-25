package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func RemasterTickets(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13503
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13504, err
	}
	state, err := orm.GetOrCreateRemasterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 13504, err
	}
	if orm.ApplyRemasterDailyReset(state, time.Now()) {
		if err := orm.GormDB.Save(state).Error; err != nil {
			return 0, 13504, err
		}
	}
	response := protobuf.SC_13504{Result: proto.Uint32(1)}
	if payload.GetType() != 0 {
		return client.SendMessage(13504, &response)
	}
	daily, err := loadGamesetValue("reactivity_ticket_daily")
	if err != nil {
		return 0, 13504, err
	}
	maxTickets, err := loadGamesetValue("reactivity_ticket_max")
	if err != nil {
		return 0, 13504, err
	}
	if state.DailyCount > 0 {
		return client.SendMessage(13504, &response)
	}
	if state.TicketCount >= maxTickets {
		return client.SendMessage(13504, &response)
	}
	grant := daily
	if remaining := maxTickets - state.TicketCount; remaining < grant {
		grant = remaining
	}
	if grant == 0 {
		return client.SendMessage(13504, &response)
	}
	state.TicketCount += grant
	state.DailyCount = daily
	if err := orm.GormDB.Save(state).Error; err != nil {
		return 0, 13504, err
	}
	response.Result = proto.Uint32(0)
	return client.SendMessage(13504, &response)
}
