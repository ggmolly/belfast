package orm

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ggmolly/belfast/internal/db"
)

func loadCommanderWithDetailsSQLC(id uint32) (Commander, error) {
	ctx := context.Background()

	row, err := db.DefaultStore.Queries.GetCommanderByID(ctx, int64(id))
	err = mapSQLCNotFound(err)
	if err != nil {
		return Commander{}, err
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

	ships, err := db.DefaultStore.Queries.ListOwnedShipsWithShipByOwnerID(ctx, int64(id))
	if err != nil {
		return Commander{}, err
	}
	commander.Ships = make([]OwnedShip, 0, len(ships))
	shipIndexByID := make(map[uint32]int, len(ships))
	for _, s := range ships {
		owned := OwnedShip{
			OwnerID:             uint32(s.OwnerID),
			ShipID:              uint32(s.ShipID),
			ID:                  uint32(s.ID),
			Level:               uint32(s.Level),
			Exp:                 uint32(s.Exp),
			SurplusExp:          uint32(s.SurplusExp),
			MaxLevel:            uint32(s.MaxLevel),
			Intimacy:            uint32(s.Intimacy),
			IsLocked:            s.IsLocked,
			Propose:             s.Propose,
			CommonFlag:          s.CommonFlag,
			BlueprintFlag:       s.BlueprintFlag,
			Proficiency:         s.Proficiency,
			ActivityNPC:         uint32(s.ActivityNpc),
			CustomName:          s.CustomName,
			ChangeNameTimestamp: s.ChangeNameTimestamp.Time,
			CreateTime:          s.CreateTime.Time,
			Energy:              uint32(s.Energy),
			State:               uint32(s.State),
			StateInfo1:          uint32(s.StateInfo1),
			StateInfo2:          uint32(s.StateInfo2),
			StateInfo3:          uint32(s.StateInfo3),
			StateInfo4:          uint32(s.StateInfo4),
			SkinID:              uint32(s.SkinID),
			IsSecretary:         s.IsSecretary,
			SecretaryPosition:   pgInt8PtrToUint32Ptr(toPgInt8(s.SecretaryPosition)),
			SecretaryPhantomID:  uint32(s.SecretaryPhantomID),
			Ship: Ship{
				TemplateID:  uint32(s.ShipTemplateID),
				Name:        s.ShipName,
				EnglishName: s.ShipEnglishName,
				RarityID:    uint32(s.ShipRarityID),
				Star:        uint32(s.ShipStar),
				Type:        uint32(s.ShipType),
				Nationality: uint32(s.ShipNationality),
				BuildTime:   uint32(s.ShipBuildTime),
				PoolID:      pgInt8PtrToUint32Ptr(toPgInt8(s.ShipPoolID)),
			},
		}
		shipIndexByID[owned.ID] = len(commander.Ships)
		commander.Ships = append(commander.Ships, owned)
	}

	equipments, err := db.DefaultStore.Queries.ListOwnedShipEquipmentsByOwnerID(ctx, int64(id))
	if err != nil {
		return Commander{}, err
	}
	for _, e := range equipments {
		idx, ok := shipIndexByID[uint32(e.ShipID)]
		if !ok {
			continue
		}
		commander.Ships[idx].Equipments = append(commander.Ships[idx].Equipments, OwnedShipEquipment{
			OwnerID: uint32(e.OwnerID),
			ShipID:  uint32(e.ShipID),
			Pos:     uint32(e.Pos),
			EquipID: uint32(e.EquipID),
			SkinID:  uint32(e.SkinID),
		})
	}

	strengths, err := db.DefaultStore.Queries.ListOwnedShipStrengthsByOwnerID(ctx, int64(id))
	if err != nil {
		return Commander{}, err
	}
	for _, st := range strengths {
		idx, ok := shipIndexByID[uint32(st.ShipID)]
		if !ok {
			continue
		}
		commander.Ships[idx].Strengths = append(commander.Ships[idx].Strengths, OwnedShipStrength{
			OwnerID:    uint32(st.OwnerID),
			ShipID:     uint32(st.ShipID),
			StrengthID: uint32(st.StrengthID),
			Exp:        uint32(st.Exp),
		})
	}

	transformRows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT owner_id, ship_id, transform_id, level
FROM owned_ship_transforms
WHERE owner_id = $1
`, int64(id))
	if err != nil {
		return Commander{}, err
	}
	for transformRows.Next() {
		var transform OwnedShipTransform
		if err := transformRows.Scan(&transform.OwnerID, &transform.ShipID, &transform.TransformID, &transform.Level); err != nil {
			transformRows.Close()
			return Commander{}, err
		}
		idx, ok := shipIndexByID[transform.ShipID]
		if !ok {
			continue
		}
		commander.Ships[idx].Transforms = append(commander.Ships[idx].Transforms, transform)
	}
	if err := transformRows.Err(); err != nil {
		transformRows.Close()
		return Commander{}, err
	}
	transformRows.Close()

	items, err := db.DefaultStore.Queries.ListCommanderItemsWithItemByCommanderID(ctx, int64(id))
	if err != nil {
		return Commander{}, err
	}
	commander.Items = make([]CommanderItem, 0, len(items))
	for _, it := range items {
		commander.Items = append(commander.Items, CommanderItem{
			CommanderID: uint32(it.CommanderID),
			ItemID:      uint32(it.ItemID),
			Count:       uint32(it.Count),
			Item: Item{
				ID:          uint32(it.ItemIDFull),
				Name:        it.ItemName,
				Rarity:      int(it.ItemRarity),
				ShopID:      int(it.ItemShopID),
				Type:        int(it.ItemType),
				VirtualType: int(it.ItemVirtualType),
			},
		})
	}

	miscItems, err := db.DefaultStore.Queries.ListCommanderMiscItemsWithItemByCommanderID(ctx, int64(id))
	if err != nil {
		return Commander{}, err
	}
	commander.MiscItems = make([]CommanderMiscItem, 0, len(miscItems))
	for _, it := range miscItems {
		commander.MiscItems = append(commander.MiscItems, CommanderMiscItem{
			CommanderID: uint32(it.CommanderID),
			ItemID:      uint32(it.ItemID),
			Data:        uint32(it.Data),
			Item: Item{
				ID:          uint32(it.ItemIDFull),
				Name:        it.ItemName,
				Rarity:      int(it.ItemRarity),
				ShopID:      int(it.ItemShopID),
				Type:        int(it.ItemType),
				VirtualType: int(it.ItemVirtualType),
			},
		})
	}

	resources, err := db.DefaultStore.Queries.ListOwnedResourcesWithResourceByCommanderID(ctx, int64(id))
	if err != nil {
		return Commander{}, err
	}
	commander.OwnedResources = make([]OwnedResource, 0, len(resources))
	for _, r := range resources {
		commander.OwnedResources = append(commander.OwnedResources, OwnedResource{
			CommanderID: uint32(r.CommanderID),
			ResourceID:  uint32(r.ResourceID),
			Amount:      uint32(r.Amount),
			Resource: Resource{
				ID:     uint32(r.ResourceIDFull),
				ItemID: uint32(r.ResourceItemID),
				Name:   r.ResourceName,
			},
		})
	}

	builds, err := db.DefaultStore.Queries.ListBuildsWithShipByBuilderID(ctx, int64(id))
	if err != nil {
		return Commander{}, err
	}
	commander.Builds = make([]Build, 0, len(builds))
	for _, b := range builds {
		commander.Builds = append(commander.Builds, Build{
			ID:         uint32(b.ID),
			BuilderID:  uint32(b.BuilderID),
			ShipID:     uint32(b.ShipID),
			PoolID:     uint32(b.PoolID),
			FinishesAt: b.FinishesAt.Time,
			Ship: Ship{
				TemplateID:  uint32(b.ShipTemplateID),
				Name:        b.ShipName,
				EnglishName: b.ShipEnglishName,
				RarityID:    uint32(b.ShipRarityID),
				Star:        uint32(b.ShipStar),
				Type:        uint32(b.ShipType),
				Nationality: uint32(b.ShipNationality),
				BuildTime:   uint32(b.ShipBuildTime),
				PoolID:      pgInt8PtrToUint32Ptr(toPgInt8(b.ShipPoolID)),
			},
		})
	}

	mails, err := db.DefaultStore.Queries.ListMailsByReceiverID(ctx, int64(id))
	if err != nil {
		return Commander{}, err
	}
	commander.Mails = make([]Mail, 0, len(mails))
	mailIndexByID := make(map[uint32]int, len(mails))
	mailIDs := make([]int64, 0, len(mails))
	for _, m := range mails {
		mail := Mail{
			ID:                   uint32(m.ID),
			ReceiverID:           uint32(m.ReceiverID),
			Read:                 m.Read,
			Date:                 m.Date.Time,
			Title:                m.Title,
			Body:                 m.Body,
			AttachmentsCollected: m.AttachmentsCollected,
			IsImportant:          m.IsImportant,
			CustomSender:         pgTextPtr(m.CustomSender),
			IsArchived:           m.IsArchived,
			CreatedAt:            m.CreatedAt.Time,
		}
		mailIndexByID[mail.ID] = len(commander.Mails)
		commander.Mails = append(commander.Mails, mail)
		mailIDs = append(mailIDs, m.ID)
	}
	if len(mailIDs) > 0 {
		attachments, err := db.DefaultStore.Queries.ListMailAttachmentsByMailIDs(ctx, mailIDs)
		if err != nil {
			return Commander{}, err
		}
		for _, a := range attachments {
			idx, ok := mailIndexByID[uint32(a.MailID)]
			if !ok {
				continue
			}
			commander.Mails[idx].Attachments = append(commander.Mails[idx].Attachments, MailAttachment{
				ID:       uint32(a.ID),
				MailID:   uint32(a.MailID),
				Type:     uint32(a.Type),
				ItemID:   uint32(a.ItemID),
				Quantity: uint32(a.Quantity),
			})
		}
	}

	skins, err := db.DefaultStore.Queries.ListOwnedSkinsByCommanderID(ctx, int64(id))
	if err != nil {
		return Commander{}, err
	}
	commander.OwnedSkins = make([]OwnedSkin, 0, len(skins))
	for _, s := range skins {
		commander.OwnedSkins = append(commander.OwnedSkins, OwnedSkin{
			CommanderID: uint32(s.CommanderID),
			SkinID:      uint32(s.SkinID),
			ExpiresAt:   pgTimestamptzPtr(s.ExpiresAt),
		})
	}

	equipBag, err := db.DefaultStore.Queries.ListOwnedEquipmentsByCommanderID(ctx, int64(id))
	if err != nil {
		return Commander{}, err
	}
	commander.OwnedEquipments = make([]OwnedEquipment, 0, len(equipBag))
	for _, e := range equipBag {
		commander.OwnedEquipments = append(commander.OwnedEquipments, OwnedEquipment{
			CommanderID: uint32(e.CommanderID),
			EquipmentID: uint32(e.EquipmentID),
			Count:       uint32(e.Count),
		})
	}

	spweapons, err := db.DefaultStore.Queries.ListOwnedSpWeaponsByOwnerID(ctx, int64(id))
	if err != nil {
		return Commander{}, err
	}
	commander.OwnedSpWeapons = make([]OwnedSpWeapon, 0, len(spweapons))
	for _, w := range spweapons {
		commander.OwnedSpWeapons = append(commander.OwnedSpWeapons, OwnedSpWeapon{
			OwnerID:        uint32(w.OwnerID),
			ID:             uint32(w.ID),
			TemplateID:     uint32(w.TemplateID),
			Attr1:          uint32(w.Attr1),
			Attr2:          uint32(w.Attr2),
			AttrTemp1:      uint32(w.AttrTemp1),
			AttrTemp2:      uint32(w.AttrTemp2),
			Effect:         uint32(w.Effect),
			Pt:             uint32(w.Pt),
			EquippedShipID: uint32(w.EquippedShipID),
		})
	}

	fleets, err := db.DefaultStore.Queries.ListFleetsByCommanderID(ctx, int64(id))
	if err != nil {
		return Commander{}, err
	}
	commander.Fleets = make([]Fleet, 0, len(fleets))
	for _, f := range fleets {
		var shipList Int64List
		if err := shipList.Scan(extractJSONBytes(f.ShipList)); err != nil {
			return Commander{}, fmt.Errorf("fleet ship_list decode: %w", err)
		}
		var meowList Int64List
		if err := meowList.Scan(extractJSONBytes(f.MeowfficerList)); err != nil {
			return Commander{}, fmt.Errorf("fleet meowfficer_list decode: %w", err)
		}
		commander.Fleets = append(commander.Fleets, Fleet{
			ID:             uint32(f.ID),
			GameID:         uint32(f.GameID),
			CommanderID:    uint32(f.CommanderID),
			Name:           f.Name,
			ShipList:       shipList,
			MeowfficerList: meowList,
		})
	}

	return commander, nil
}

func toPgInt8(value any) pgtype.Int8 {
	// sqlc can generate nullable integers as pgtype.Int8, *int64, or sql.NullInt64
	// depending on column nullability and query shape. We normalize into pgtype.Int8
	// so existing pointer helpers can be used.
	switch v := value.(type) {
	case pgtype.Int8:
		return v
	case *int64:
		if v == nil {
			return pgtype.Int8{}
		}
		return pgtype.Int8{Int64: *v, Valid: true}
	case int64:
		return pgtype.Int8{Int64: v, Valid: true}
	default:
		return pgtype.Int8{}
	}
}

func extractJSONBytes(value any) any {
	// Int64List.Scan supports string and []byte.
	switch v := value.(type) {
	case []byte:
		return v
	case string:
		return v
	default:
		return value
	}
}
