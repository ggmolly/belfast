package orm

import (
	"context"
	"errors"
	"fmt"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

var ErrCommanderNameExists = errors.New("commander name already exists")

func CreateYostarusMap(arg2 uint32, accountID uint32) error {
	if db.DefaultStore == nil {
		return fmt.Errorf("database is not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateYostarusMap(ctx, gen.CreateYostarusMapParams{Arg2: int64(arg2), AccountID: int64(accountID)})
}

func GetYostarusMapByArg2(arg2 uint32) (*YostarusMap, error) {
	if db.DefaultStore == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetYostarusMapByArg2(ctx, int64(arg2))
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	entry := YostarusMap{Arg2: uint32(row.Arg2), AccountID: uint32(row.AccountID)}
	return &entry, nil
}

func UpsertDeviceAuthMap(deviceID string, arg2 uint32, accountID uint32) error {
	if db.DefaultStore == nil {
		return fmt.Errorf("database is not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertDeviceAuthMap(ctx, gen.UpsertDeviceAuthMapParams{DeviceID: deviceID, Arg2: int64(arg2), AccountID: int64(accountID)})
}

func GetDeviceAuthMapByDeviceID(deviceID string) (*DeviceAuthMap, error) {
	if db.DefaultStore == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetDeviceAuthMapByDeviceID(ctx, deviceID)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	entry := DeviceAuthMap{
		DeviceID:  row.DeviceID,
		Arg2:      uint32(row.Arg2),
		AccountID: uint32(row.AccountID),
		UpdatedAt: row.UpdatedAt.Time,
	}
	return &entry, nil
}

func GetCommanderByAccountID(accountID uint32) (*Commander, error) {
	if db.DefaultStore == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetCommanderByAccountID(ctx, int64(accountID))
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	commander := Commander{
		CommanderID:             uint32(row.CommanderID),
		AccountID:               uint32(row.AccountID),
		Level:                   int(row.Level),
		Exp:                     int(row.Exp),
		Name:                    row.Name,
		LastLogin:               row.LastLogin.Time,
		GuideIndex:              uint32(row.GuideIndex),
		NewGuideIndex:           uint32(row.NewGuideIndex),
		NameChangeCooldown:      row.NameChangeCooldown.Time,
		RoomID:                  uint32(row.RoomID),
		ExchangeCount:           uint32(row.ExchangeCount),
		DrawCount1:              uint32(row.DrawCount1),
		DrawCount10:             uint32(row.DrawCount10),
		SupportRequisitionCount: uint32(row.SupportRequisitionCount),
		SupportRequisitionMonth: uint32(row.SupportRequisitionMonth),
		CollectAttackCount:      uint32(row.CollectAttackCount),
		AccPayLv:                uint32(row.AccPayLv),
		LivingAreaCoverID:       uint32(row.LivingAreaCoverID),
		SelectedIconFrameID:     uint32(row.SelectedIconFrameID),
		SelectedChatFrameID:     uint32(row.SelectedChatFrameID),
		SelectedBattleUIID:      uint32(row.SelectedBattleUiID),
		DisplayIconID:           uint32(row.DisplayIconID),
		DisplaySkinID:           uint32(row.DisplaySkinID),
		DisplayIconThemeID:      uint32(row.DisplayIconThemeID),
		Manifesto:               row.Manifesto,
		DormName:                row.DormName,
		RandomShipMode:          uint32(row.RandomShipMode),
		RandomFlagShipEnabled:   row.RandomFlagShipEnabled,
	}
	return &commander, nil
}

func CreateCommanderAccountRoot(accountID uint32, nickname string, guideIndex uint32, newGuideIndex uint32) error {
	if db.DefaultStore == nil {
		return fmt.Errorf("database is not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateCommander(ctx, gen.CreateCommanderParams{
		CommanderID:   int64(accountID),
		AccountID:     int64(accountID),
		Name:          nickname,
		GuideIndex:    int64(guideIndex),
		NewGuideIndex: int64(newGuideIndex),
	})
}

func CreateDefaultStarterInventory(accountID uint32) error {
	if db.DefaultStore == nil {
		return fmt.Errorf("database is not initialized")
	}
	ctx := context.Background()
	if err := db.DefaultStore.Queries.UpsertCommanderItemSet(ctx, gen.UpsertCommanderItemSetParams{CommanderID: int64(accountID), ItemID: 20001, Count: 1}); err != nil {
		return err
	}
	if err := db.DefaultStore.Queries.UpsertCommanderItemSet(ctx, gen.UpsertCommanderItemSetParams{CommanderID: int64(accountID), ItemID: 15003, Count: 10}); err != nil {
		return err
	}
	if err := db.DefaultStore.Queries.UpsertOwnedResourceSet(ctx, gen.UpsertOwnedResourceSetParams{CommanderID: int64(accountID), ResourceID: 1, Amount: 3000}); err != nil {
		return err
	}
	if err := db.DefaultStore.Queries.UpsertOwnedResourceSet(ctx, gen.UpsertOwnedResourceSetParams{CommanderID: int64(accountID), ResourceID: 2, Amount: 500}); err != nil {
		return err
	}
	if err := db.DefaultStore.Queries.UpsertOwnedResourceSet(ctx, gen.UpsertOwnedResourceSetParams{CommanderID: int64(accountID), ResourceID: 4, Amount: 0}); err != nil {
		return err
	}
	return nil
}

func CheckCommanderNameAvailability(name string) error {
	if db.DefaultStore == nil {
		return fmt.Errorf("database is not initialized")
	}
	ctx := context.Background()
	count, err := db.DefaultStore.Queries.CountCommandersByName(ctx, name)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrCommanderNameExists
	}
	return nil
}
