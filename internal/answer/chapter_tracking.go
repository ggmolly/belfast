package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func ChapterTracking(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13101
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13102, err
	}
	if payload.GetFleet() == nil {
		response := protobuf.SC_13102{Result: proto.Uint32(1)}
		return client.SendMessage(13102, &response)
	}
	if client.Commander.CommanderItemsMap == nil && client.Commander.MiscItemsMap == nil {
		if err := client.Commander.Load(); err != nil {
			return 0, 13102, err
		}
	}
	template, err := loadChapterTemplate(payload.GetId(), payload.GetLoopFlag())
	if err != nil {
		return 0, 13102, err
	}
	if template == nil {
		response := protobuf.SC_13102{Result: proto.Uint32(1)}
		return client.SendMessage(13102, &response)
	}
	rate, err := calculateOperationItemCostRate(payload.GetOperationItem())
	if err != nil {
		return 0, 13102, err
	}
	baseOil := template.Oil
	oilCost := uint32(float64(baseOil) * rate)
	if !client.Commander.HasEnoughResource(2, oilCost) {
		response := protobuf.SC_13102{Result: proto.Uint32(1)}
		return client.SendMessage(13102, &response)
	}
	if payload.GetOperationItem() != 0 && !client.Commander.HasEnoughItem(payload.GetOperationItem(), 1) {
		response := protobuf.SC_13102{Result: proto.Uint32(1)}
		return client.SendMessage(13102, &response)
	}
	if oilCost > 0 {
		if err := client.Commander.ConsumeResource(2, oilCost); err != nil {
			return 0, 13102, err
		}
	}
	if payload.GetOperationItem() != 0 {
		if err := client.Commander.ConsumeItem(payload.GetOperationItem(), 1); err != nil {
			return 0, 13102, err
		}
	}
	operationBuffID, err := findOperationBuffID(payload.GetOperationItem())
	if err != nil {
		return 0, 13102, err
	}
	current, _, err := buildCurrentChapterInfo(template, &payload, operationBuffID)
	if err != nil {
		return 0, 13102, err
	}
	stateBytes, err := proto.Marshal(current)
	if err != nil {
		return 0, 13102, err
	}
	state := orm.ChapterState{
		CommanderID: client.Commander.CommanderID,
		ChapterID:   payload.GetId(),
		State:       stateBytes,
	}
	if err := orm.UpsertChapterState(orm.GormDB, &state); err != nil {
		return 0, 13102, err
	}
	if err := ensureChapterProgress(client.Commander.CommanderID, payload.GetId()); err != nil {
		return 0, 13102, err
	}
	response := protobuf.SC_13102{
		Result:         proto.Uint32(0),
		CurrentChapter: current,
	}
	return client.SendMessage(13102, &response)
}

func ChapterTrackingKR(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13101_KR
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13102, err
	}
	if client.Commander.CommanderItemsMap == nil && client.Commander.MiscItemsMap == nil {
		if err := client.Commander.Load(); err != nil {
			return 0, 13102, err
		}
	}
	template, err := loadChapterTemplate(payload.GetId(), payload.GetLoopFlag())
	if err != nil {
		return 0, 13102, err
	}
	if template == nil {
		response := protobuf.SC_13102{Result: proto.Uint32(1)}
		return client.SendMessage(13102, &response)
	}
	rate, err := calculateOperationItemCostRate(payload.GetOperationItem())
	if err != nil {
		return 0, 13102, err
	}
	oilCost := uint32(float64(template.Oil) * rate)
	if !client.Commander.HasEnoughResource(2, oilCost) {
		response := protobuf.SC_13102{Result: proto.Uint32(1)}
		return client.SendMessage(13102, &response)
	}
	if payload.GetOperationItem() != 0 && !client.Commander.HasEnoughItem(payload.GetOperationItem(), 1) {
		response := protobuf.SC_13102{Result: proto.Uint32(1)}
		return client.SendMessage(13102, &response)
	}
	if oilCost > 0 {
		if err := client.Commander.ConsumeResource(2, oilCost); err != nil {
			return 0, 13102, err
		}
	}
	if payload.GetOperationItem() != 0 {
		if err := client.Commander.ConsumeItem(payload.GetOperationItem(), 1); err != nil {
			return 0, 13102, err
		}
	}
	operationBuffID, err := findOperationBuffID(payload.GetOperationItem())
	if err != nil {
		return 0, 13102, err
	}
	current, _, err := buildCurrentChapterInfoKR(template, &payload, operationBuffID)
	if err != nil {
		return 0, 13102, err
	}
	stateBytes, err := proto.Marshal(current)
	if err != nil {
		return 0, 13102, err
	}
	state := orm.ChapterState{
		CommanderID: client.Commander.CommanderID,
		ChapterID:   payload.GetId(),
		State:       stateBytes,
	}
	if err := orm.UpsertChapterState(orm.GormDB, &state); err != nil {
		return 0, 13102, err
	}
	if err := ensureChapterProgress(client.Commander.CommanderID, payload.GetId()); err != nil {
		return 0, 13102, err
	}
	response := protobuf.SC_13102{
		Result:         proto.Uint32(0),
		CurrentChapter: current,
	}
	return client.SendMessage(13102, &response)
}

func ensureChapterProgress(commanderID uint32, chapterID uint32) error {
	if _, err := orm.GetChapterProgress(orm.GormDB, commanderID, chapterID); err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	progress := orm.ChapterProgress{
		CommanderID: commanderID,
		ChapterID:   chapterID,
	}
	return orm.UpsertChapterProgress(orm.GormDB, &progress)
}
