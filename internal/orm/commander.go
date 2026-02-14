package orm

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/logger"
)

type Commander struct {
	CommanderID             uint32     `gorm:"primary_key"`
	AccountID               uint32     `gorm:"not_null"`
	Level                   int        `gorm:"default:1;not_null"`
	Exp                     int        `gorm:"default:0;not_null"`
	Name                    string     `gorm:"size:30;not_null;uniqueIndex"`
	LastLogin               time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	GuideIndex              uint32     `gorm:"default:0;not_null"`
	NewGuideIndex           uint32     `gorm:"default:0;not_null"`
	NameChangeCooldown      time.Time  `gorm:"type:timestamp;default:'1970-01-01 00:00:00';not_null"`
	RoomID                  uint32     `gorm:"default:0;not_null"`
	ExchangeCount           uint32     `gorm:"default:0;not_null"` // Number of times the commander has built ships, can be exchanged for UR ships
	DrawCount1              uint32     `gorm:"default:0;not_null"`
	DrawCount10             uint32     `gorm:"default:0;not_null"`
	SupportRequisitionCount uint32     `gorm:"default:0;not_null"`
	SupportRequisitionMonth uint32     `gorm:"default:0;not_null"`
	CollectAttackCount      uint32     `gorm:"default:0;not_null"`
	AccPayLv                uint32     `gorm:"default:0;not_null"`
	LivingAreaCoverID       uint32     `gorm:"default:0;not_null"`
	SelectedIconFrameID     uint32     `gorm:"default:0;not_null"`
	SelectedChatFrameID     uint32     `gorm:"default:0;not_null"`
	SelectedBattleUIID      uint32     `gorm:"default:0;not_null"`
	DisplayIconID           uint32     `gorm:"default:0;not_null"`
	DisplaySkinID           uint32     `gorm:"default:0;not_null"`
	DisplayIconThemeID      uint32     `gorm:"default:0;not_null"`
	Manifesto               string     `gorm:"size:200;default:'';not_null"`
	DormName                string     `gorm:"size:50;default:'';not_null"`
	RandomShipMode          uint32     `gorm:"default:0;not_null"`
	RandomFlagShipEnabled   bool       `gorm:"default:false;not_null"`
	DeletedAt               *time.Time `gorm:"index"`

	Punishments      []Punishment        `gorm:"foreignKey:PunishedID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Ships            []OwnedShip         `gorm:"foreignKey:OwnerID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Items            []CommanderItem     `gorm:"foreignKey:CommanderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	MiscItems        []CommanderMiscItem `gorm:"foreignKey:CommanderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	OwnedResources   []OwnedResource     `gorm:"foreignKey:CommanderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Builds           []Build             `gorm:"foreignKey:BuilderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Mails            []Mail              `gorm:"foreignKey:ReceiverID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Compensations    []Compensation      `gorm:"-:migration"`
	OwnedSkins       []OwnedSkin         `gorm:"foreignKey:CommanderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	OwnedEquipments  []OwnedEquipment    `gorm:"foreignKey:CommanderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	OwnedSpWeapons   []OwnedSpWeapon     `gorm:"foreignKey:OwnerID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Secretaries      []*OwnedShip        `gorm:"-"`
	Fleets           []Fleet             `gorm:"foreignKey:CommanderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	EventCollections []EventCollection   `gorm:"foreignKey:CommanderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	// These maps will be populated by the Load() method
	OwnedShipsMap     map[uint32]*OwnedShip         `gorm:"-"`
	OwnedResourcesMap map[uint32]*OwnedResource     `gorm:"-"`
	CommanderItemsMap map[uint32]*CommanderItem     `gorm:"-"`
	MiscItemsMap      map[uint32]*CommanderMiscItem `gorm:"-"`
	BuildsMap         map[uint32]*Build             `gorm:"-"`
	OwnedSkinsMap     map[uint32]*OwnedSkin         `gorm:"-"`
	OwnedEquipmentMap map[uint32]*OwnedEquipment    `gorm:"-"`
	OwnedSpWeaponsMap map[uint32]*OwnedSpWeapon     `gorm:"-"`
	MailsMap          map[uint32]*Mail              `gorm:"-"`
	CompensationsMap  map[uint32]*Compensation      `gorm:"-"`
	FleetsMap         map[uint32]*Fleet             `gorm:"-"`
}

func (c *Commander) HasEnoughGold(n uint32) bool {
	return c.HasEnoughResource(1, n)
}

func (c *Commander) HasEnoughCube(n uint32) bool {
	return c.HasEnoughItem(20001, n)
}

func (c *Commander) HasEnoughItem(itemId uint32, n uint32) bool {
	if item, ok := c.CommanderItemsMap[itemId]; ok {
		return item.Count >= n
	} else if miscItem, ok := c.MiscItemsMap[itemId]; ok {
		return miscItem.Data >= n
	} else {
		return false
	}
}

func (c *Commander) HasEnoughResource(resourceId uint32, n uint32) bool {
	DealiasResource(&resourceId)
	if resource, ok := c.OwnedResourcesMap[resourceId]; ok {
		return resource.Amount >= n
	} else {
		return false
	}
}

func (c *Commander) CreateBuild(poolId uint32, runningBuilds *int) (*Build, uint32, error) {
	ship, err := GetRandomPoolShip(poolId)
	if err != nil {
		return nil, 0, err
	}
	newBuild := Build{
		BuilderID:  c.CommanderID,
		ShipID:     ship.TemplateID,
		PoolID:     poolId,
		FinishesAt: time.Now().Add(time.Second * time.Duration(ship.BuildTime)),
	}
	if err := newBuild.Create(); err != nil {
		return nil, 0, err
	}
	*runningBuilds++ // the game requires us to send a sequential build id

	// Add the build to the commander's list of BuildsMap
	c.Builds = append(c.Builds, newBuild)
	c.BuildsMap[newBuild.ID] = &newBuild

	return &newBuild, ship.BuildTime, nil
}

func (c *Commander) AddShip(shipId uint32) (*OwnedShip, error) {
	ctx := context.Background()
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	// Validate ship exists.
	var templateID int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT template_id FROM ships WHERE template_id = $1`, int64(shipId)).Scan(&templateID); err != nil {
		return nil, db.MapNotFound(err)
	}

	var newShip OwnedShip
	err := db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
INSERT INTO owned_ships (owner_id, ship_id)
VALUES ($1, $2)
RETURNING id, level, exp, surplus_exp, max_level, intimacy, is_locked, propose, common_flag, blueprint_flag, proficiency, activity_npc, custom_name, change_name_timestamp, create_time, energy, state, state_info1, state_info2, state_info3, state_info4, skin_id, is_secretary, secretary_position, secretary_phantom_id
`, int64(c.CommanderID), int64(shipId))
		var id int64
		var secretaryPos *int64
		if err := row.Scan(
			&id,
			&newShip.Level,
			&newShip.Exp,
			&newShip.SurplusExp,
			&newShip.MaxLevel,
			&newShip.Intimacy,
			&newShip.IsLocked,
			&newShip.Propose,
			&newShip.CommonFlag,
			&newShip.BlueprintFlag,
			&newShip.Proficiency,
			&newShip.ActivityNPC,
			&newShip.CustomName,
			&newShip.ChangeNameTimestamp,
			&newShip.CreateTime,
			&newShip.Energy,
			&newShip.State,
			&newShip.StateInfo1,
			&newShip.StateInfo2,
			&newShip.StateInfo3,
			&newShip.StateInfo4,
			&newShip.SkinID,
			&newShip.IsSecretary,
			&secretaryPos,
			&newShip.SecretaryPhantomID,
		); err != nil {
			return err
		}
		newShip.ID = uint32(id)
		newShip.OwnerID = c.CommanderID
		newShip.ShipID = shipId
		if secretaryPos != nil {
			v := uint32(*secretaryPos)
			newShip.SecretaryPosition = &v
		}
		return createDefaultShipEquipments(ctx, tx, c.CommanderID, newShip.ID)
	})
	if err != nil {
		return nil, err
	}

	c.Ships = append(c.Ships, newShip)
	if c.OwnedShipsMap == nil {
		c.OwnedShipsMap = make(map[uint32]*OwnedShip)
	}
	c.OwnedShipsMap[newShip.ID] = &c.Ships[len(c.Ships)-1]
	return &c.Ships[len(c.Ships)-1], nil
}

func (c *Commander) AddShipTx(ctx context.Context, tx pgx.Tx, shipId uint32) (*OwnedShip, error) {
	// Validate ship exists.
	var templateID int64
	if err := tx.QueryRow(ctx, `SELECT template_id FROM ships WHERE template_id = $1`, int64(shipId)).Scan(&templateID); err != nil {
		return nil, db.MapNotFound(err)
	}
	var newShip OwnedShip
	row := tx.QueryRow(ctx, `
INSERT INTO owned_ships (owner_id, ship_id)
VALUES ($1, $2)
RETURNING id, create_time, change_name_timestamp
`, int64(c.CommanderID), int64(shipId))
	var id int64
	if err := row.Scan(&id, &newShip.CreateTime, &newShip.ChangeNameTimestamp); err != nil {
		return nil, err
	}
	newShip.ID = uint32(id)
	newShip.OwnerID = c.CommanderID
	newShip.ShipID = shipId
	if err := createDefaultShipEquipments(ctx, tx, c.CommanderID, newShip.ID); err != nil {
		return nil, err
	}
	c.Ships = append(c.Ships, newShip)
	if c.OwnedShipsMap == nil {
		c.OwnedShipsMap = make(map[uint32]*OwnedShip)
	}
	c.OwnedShipsMap[newShip.ID] = &c.Ships[len(c.Ships)-1]
	return &c.Ships[len(c.Ships)-1], nil
}

func createDefaultShipEquipments(ctx context.Context, tx pgx.Tx, ownerID uint32, ownedShipID uint32) error {
	// TODO(M6): Implement ship equipment defaults via config_entries.
	for pos := uint32(1); pos <= 3; pos++ {
		if _, err := tx.Exec(ctx, `
INSERT INTO owned_ship_equipments (owner_id, ship_id, pos, equip_id, skin_id)
VALUES ($1, $2, $3, 0, 0)
ON CONFLICT (owner_id, ship_id, pos)
DO NOTHING
`, int64(ownerID), int64(ownedShipID), int64(pos)); err != nil {
			return err
		}
	}
	return nil
}

func (c *Commander) ConsumeItem(itemId uint32, count uint32) error {
	if item, ok := c.CommanderItemsMap[itemId]; ok {
		if item.Count < count {
			return fmt.Errorf("not enough items")
		}
		ctx := context.Background()
		if db.DefaultStore == nil {
			return errors.New("db not initialized")
		}
		res, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE commander_items
SET count = count - $3
WHERE commander_id = $1
  AND item_id = $2
  AND count >= $3
`, int64(c.CommanderID), int64(itemId), int64(count))
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("not enough items")
		}
		item.Count -= count
		return nil
	} else if miscItem, ok := c.MiscItemsMap[itemId]; ok {
		if miscItem.Data < count {
			return fmt.Errorf("not enough items")
		}
		ctx := context.Background()
		res, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE commander_misc_items
SET data = data - $3
WHERE commander_id = $1
  AND item_id = $2
  AND data >= $3
`, int64(c.CommanderID), int64(itemId), int64(count))
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("not enough items")
		}
		miscItem.Data -= count
		return nil
	}
	return fmt.Errorf("not enough items")
}

func (c *Commander) ConsumeItemTx(ctx context.Context, tx pgx.Tx, itemId uint32, count uint32) error {
	if item, ok := c.CommanderItemsMap[itemId]; ok {
		if item.Count < count {
			return fmt.Errorf("not enough items")
		}
		res, err := tx.Exec(ctx, `
UPDATE commander_items
SET count = count - $3
WHERE commander_id = $1
  AND item_id = $2
  AND count >= $3
`, int64(c.CommanderID), int64(itemId), int64(count))
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("not enough items")
		}
		item.Count -= count
		return nil
	} else if miscItem, ok := c.MiscItemsMap[itemId]; ok {
		if miscItem.Data < count {
			return fmt.Errorf("not enough items")
		}
		res, err := tx.Exec(ctx, `
UPDATE commander_misc_items
SET data = data - $3
WHERE commander_id = $1
  AND item_id = $2
  AND data >= $3
`, int64(c.CommanderID), int64(itemId), int64(count))
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("not enough items")
		}
		miscItem.Data -= count
		return nil
	}
	return fmt.Errorf("not enough items")
}

func (c *Commander) SaveTx(ctx context.Context, tx pgx.Tx) error {
	if c.Level > 120 {
		c.Level = 120
	}
	_, err := tx.Exec(ctx, `
UPDATE commanders
SET
  account_id = $2,
  level = $3,
  exp = $4,
  name = $5,
  last_login = $6,
  guide_index = $7,
  new_guide_index = $8,
  name_change_cooldown = $9,
  room_id = $10,
  exchange_count = $11,
  draw_count1 = $12,
  draw_count10 = $13,
  support_requisition_count = $14,
  support_requisition_month = $15,
  collect_attack_count = $16,
  acc_pay_lv = $17,
  living_area_cover_id = $18,
  selected_icon_frame_id = $19,
  selected_chat_frame_id = $20,
  selected_battle_ui_id = $21,
  display_icon_id = $22,
  display_skin_id = $23,
  display_icon_theme_id = $24,
  manifesto = $25,
  dorm_name = $26,
  random_ship_mode = $27,
  random_flag_ship_enabled = $28
WHERE commander_id = $1
`,
		int64(c.CommanderID),
		int64(c.AccountID),
		c.Level,
		c.Exp,
		c.Name,
		c.LastLogin,
		int64(c.GuideIndex),
		int64(c.NewGuideIndex),
		c.NameChangeCooldown,
		int64(c.RoomID),
		int64(c.ExchangeCount),
		int64(c.DrawCount1),
		int64(c.DrawCount10),
		int64(c.SupportRequisitionCount),
		int64(c.SupportRequisitionMonth),
		int64(c.CollectAttackCount),
		int64(c.AccPayLv),
		int64(c.LivingAreaCoverID),
		int64(c.SelectedIconFrameID),
		int64(c.SelectedChatFrameID),
		int64(c.SelectedBattleUIID),
		int64(c.DisplayIconID),
		int64(c.DisplaySkinID),
		int64(c.DisplayIconThemeID),
		c.Manifesto,
		c.DormName,
		int64(c.RandomShipMode),
		c.RandomFlagShipEnabled,
	)
	return err
}

func (c *Commander) ConsumeResource(resourceId uint32, count uint32) error {
	DealiasResource(&resourceId)
	// check if the commander has enough of the resource
	if resource, ok := c.OwnedResourcesMap[resourceId]; ok {
		if resource.Amount >= count {
			ctx := context.Background()
			res, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE owned_resources
SET amount = amount - $3
WHERE commander_id = $1
  AND resource_id = $2
  AND amount >= $3
`, int64(c.CommanderID), int64(resourceId), int64(count))
			if err != nil {
				return err
			}
			if res.RowsAffected() == 0 {
				return fmt.Errorf("not enough resources")
			}
			resource.Amount -= count
			return nil
		}
	}
	return fmt.Errorf("not enough resources")
}

func (c *Commander) ConsumeResourceTx(ctx context.Context, tx pgx.Tx, resourceId uint32, count uint32) error {
	DealiasResource(&resourceId)
	if resource, ok := c.OwnedResourcesMap[resourceId]; ok {
		if resource.Amount >= count {
			res, err := tx.Exec(ctx, `
UPDATE owned_resources
SET amount = amount - $3
WHERE commander_id = $1
  AND resource_id = $2
  AND amount >= $3
`, int64(c.CommanderID), int64(resourceId), int64(count))
			if err != nil {
				return err
			}
			if res.RowsAffected() == 0 {
				return fmt.Errorf("not enough resources")
			}
			resource.Amount -= count
			return nil
		}
	}
	return fmt.Errorf("not enough resources")
}

func (c *Commander) SetResource(resourceId uint32, amount uint32) error {
	// check if the commander already has the resource, if so set the amount and save
	if resource, ok := c.OwnedResourcesMap[resourceId]; ok {
		resource.Amount = amount
		ctx := context.Background()
		_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO owned_resources (commander_id, resource_id, amount)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, resource_id)
DO UPDATE SET amount = EXCLUDED.amount
`, int64(c.CommanderID), int64(resourceId), int64(amount))
		return err
	}
	// otherwise create a new resource
	newResource := OwnedResource{
		CommanderID: c.CommanderID,
		ResourceID:  resourceId,
		Amount:      amount,
	}
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO owned_resources (commander_id, resource_id, amount)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, resource_id)
DO UPDATE SET amount = EXCLUDED.amount
`, int64(c.CommanderID), int64(resourceId), int64(amount))
	if err != nil {
		return err
	}
	c.OwnedResources = append(c.OwnedResources, newResource)
	c.OwnedResourcesMap[resourceId] = &c.OwnedResources[len(c.OwnedResources)-1]
	return nil
}

func (c *Commander) SetItem(itemId uint32, amount uint32) error {
	// check if the commander already has the item, if so set the amount and save
	if item, ok := c.CommanderItemsMap[itemId]; ok {
		item.Count = amount
		ctx := context.Background()
		_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO commander_items (commander_id, item_id, count)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, item_id)
DO UPDATE SET count = EXCLUDED.count
`, int64(c.CommanderID), int64(itemId), int64(amount))
		return err
	}
	// otherwise create a new item
	newItem := CommanderItem{
		CommanderID: c.CommanderID,
		ItemID:      itemId,
		Count:       amount,
	}
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO commander_items (commander_id, item_id, count)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, item_id)
DO UPDATE SET count = EXCLUDED.count
`, int64(c.CommanderID), int64(itemId), int64(amount))
	if err != nil {
		return err
	}
	c.Items = append(c.Items, newItem)
	c.CommanderItemsMap[itemId] = &c.Items[len(c.Items)-1]
	return nil
}

func (c *Commander) AddResource(resourceId uint32, amount uint32) error {
	if c.OwnedResourcesMap == nil {
		c.OwnedResourcesMap = make(map[uint32]*OwnedResource)
	}
	DealiasResource(&resourceId)
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO owned_resources (commander_id, resource_id, amount)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, resource_id)
DO UPDATE SET amount = owned_resources.amount + EXCLUDED.amount
`, int64(c.CommanderID), int64(resourceId), int64(amount))
	if err != nil {
		return err
	}
	if resource, ok := c.OwnedResourcesMap[resourceId]; ok {
		resource.Amount += amount
		return nil
	}
	c.OwnedResources = append(c.OwnedResources, OwnedResource{CommanderID: c.CommanderID, ResourceID: resourceId, Amount: amount})
	c.OwnedResourcesMap[resourceId] = &c.OwnedResources[len(c.OwnedResources)-1]
	return nil
}

func (c *Commander) AddResourceTx(ctx context.Context, tx pgx.Tx, resourceId uint32, amount uint32) error {
	DealiasResource(&resourceId)
	if c.OwnedResourcesMap == nil {
		c.OwnedResourcesMap = make(map[uint32]*OwnedResource)
	}
	res, err := tx.Exec(ctx, `
INSERT INTO owned_resources (commander_id, resource_id, amount)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, resource_id)
DO UPDATE SET amount = owned_resources.amount + EXCLUDED.amount
`, int64(c.CommanderID), int64(resourceId), int64(amount))
	_ = res
	if err != nil {
		return err
	}
	if existing, ok := c.OwnedResourcesMap[resourceId]; ok {
		existing.Amount += amount
		return nil
	}
	c.OwnedResources = append(c.OwnedResources, OwnedResource{CommanderID: c.CommanderID, ResourceID: resourceId, Amount: amount})
	c.OwnedResourcesMap[resourceId] = &c.OwnedResources[len(c.OwnedResources)-1]
	return nil
}

func (c *Commander) AddItem(itemId uint32, amount uint32) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO commander_items (commander_id, item_id, count)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, item_id)
DO UPDATE SET count = commander_items.count + EXCLUDED.count
`, int64(c.CommanderID), int64(itemId), int64(amount))
	if err != nil {
		return err
	}
	if item, ok := c.CommanderItemsMap[itemId]; ok {
		item.Count += amount
		return nil
	}
	c.Items = append(c.Items, CommanderItem{CommanderID: c.CommanderID, ItemID: itemId, Count: amount})
	if c.CommanderItemsMap == nil {
		c.CommanderItemsMap = make(map[uint32]*CommanderItem)
	}
	c.CommanderItemsMap[itemId] = &c.Items[len(c.Items)-1]
	return nil
}

func (c *Commander) AddItemTx(ctx context.Context, tx pgx.Tx, itemId uint32, amount uint32) error {
	if c.CommanderItemsMap == nil {
		c.CommanderItemsMap = make(map[uint32]*CommanderItem)
	}
	_, err := tx.Exec(ctx, `
INSERT INTO commander_items (commander_id, item_id, count)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, item_id)
DO UPDATE SET count = commander_items.count + EXCLUDED.count
`, int64(c.CommanderID), int64(itemId), int64(amount))
	if err != nil {
		return err
	}
	if existing, ok := c.CommanderItemsMap[itemId]; ok {
		existing.Count += amount
		return nil
	}
	c.Items = append(c.Items, CommanderItem{CommanderID: c.CommanderID, ItemID: itemId, Count: amount})
	c.CommanderItemsMap[itemId] = &c.Items[len(c.Items)-1]
	return nil
}

func (c *Commander) GetItem(itemId uint32) (CommanderItem, error) {
	if item, ok := c.CommanderItemsMap[itemId]; ok {
		return *item, nil
	}
	return CommanderItem{}, fmt.Errorf("item not found")
}

func (c *Commander) GetResource(resourceId uint32) (OwnedResource, error) {
	if resource, ok := c.OwnedResourcesMap[resourceId]; ok {
		return *resource, nil
	}
	return OwnedResource{}, fmt.Errorf("resource not found")
}

// GetItemCount returns the amount of items the commander has, returns 0 if the item is not found
func (c *Commander) GetItemCount(itemId uint32) uint32 {
	if item, ok := c.CommanderItemsMap[itemId]; ok {
		return item.Count
	}
	return 0
}

// GetResourceCount returns the amount of resources the commander has, returns 0 if the resource is not found
func (c *Commander) GetResourceCount(resourceId uint32) uint32 {
	DealiasResource(&resourceId)
	if resource, ok := c.OwnedResourcesMap[resourceId]; ok {
		return resource.Amount
	}
	return 0
}

func (c *Commander) Punish(liftTimestamp *time.Time, permanent bool) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO punishments (punished_id, lift_timestamp, is_permanent)
VALUES ($1, $2, $3)
`, int64(c.CommanderID), liftTimestamp, permanent)
	return err
}

func (c *Commander) RevokeActivePunishment() error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM punishments WHERE punished_id = $1 AND lift_timestamp IS NULL`, int64(c.CommanderID))
	return err
}

// Load loads the commander's data from the database (ships, items, resources, etc)
func (c *Commander) Load() error {
	loaded, err := LoadCommanderWithDetails(c.CommanderID)
	if err != nil {
		return err
	}
	*c = loaded
	punishments, err := ListPunishmentsByCommanderID(c.CommanderID)
	if err != nil {
		return err
	}
	c.Punishments = punishments

	now := time.Now()
	activePunishments := c.Punishments[:0]
	for i := range c.Punishments {
		punishment := c.Punishments[i]
		if punishment.IsPermanent || punishment.LiftTimestamp == nil || punishment.LiftTimestamp.After(now) {
			activePunishments = append(activePunishments, punishment)
		}
	}
	c.Punishments = activePunishments
	if len(c.Punishments) > 1 {
		sort.Slice(c.Punishments, func(i, j int) bool {
			return c.Punishments[i].ID > c.Punishments[j].ID
		})
	}

	// load ships
	c.OwnedShipsMap = make(map[uint32]*OwnedShip)
	for i, ship := range c.Ships {
		c.OwnedShipsMap[ship.ID] = &c.Ships[i]
	}

	// load equipment bag
	c.rebuildOwnedEquipmentMap()

	// load spweapons
	c.OwnedSpWeaponsMap = make(map[uint32]*OwnedSpWeapon)
	for i, entry := range c.OwnedSpWeapons {
		c.OwnedSpWeaponsMap[entry.ID] = &c.OwnedSpWeapons[i]
	}

	// load resources
	c.OwnedResourcesMap = make(map[uint32]*OwnedResource)
	for i, resource := range c.OwnedResources {
		c.OwnedResourcesMap[resource.ResourceID] = &c.OwnedResources[i]
	}

	// load items
	c.CommanderItemsMap = make(map[uint32]*CommanderItem)
	for i, item := range c.Items {
		c.CommanderItemsMap[item.ItemID] = &c.Items[i]
	}

	// load misc items
	c.MiscItemsMap = make(map[uint32]*CommanderMiscItem)
	for i, item := range c.MiscItems {
		c.MiscItemsMap[item.ItemID] = &c.MiscItems[i]
	}

	// load BuildsMap
	c.BuildsMap = make(map[uint32]*Build)
	for i, build := range c.Builds {
		c.BuildsMap[build.ID] = &c.Builds[i]
	}

	// load skins
	c.OwnedSkinsMap = make(map[uint32]*OwnedSkin)
	for i, skin := range c.OwnedSkins {
		c.OwnedSkinsMap[skin.SkinID] = &c.OwnedSkins[i]
	}

	// load MailsMap
	c.MailsMap = make(map[uint32]*Mail)
	for i, mail := range c.Mails {
		c.MailsMap[mail.ID] = &c.Mails[i]
	}

	// load CompensationsMap
	c.CompensationsMap = make(map[uint32]*Compensation)
	for i, compensation := range c.Compensations {
		c.CompensationsMap[compensation.ID] = &c.Compensations[i]
	}

	// load FleetsMap
	c.FleetsMap = make(map[uint32]*Fleet)
	for i, fleet := range c.Fleets {
		c.FleetsMap[fleet.GameID] = &c.Fleets[i]
	}
	return nil
}

// Commit saves the commander's data to the database (ships, items, resources, etc)
func (c *Commander) Commit() error {
	ctx := context.Background()
	return db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		return c.SaveTx(ctx, tx)
	})
}

// Get a range of builds (special weird query, probably to save battery on phones)
func (c *Commander) GetBuildRange(minPos, maxPos uint32) ([]Build, error) {
	ctx := context.Background()
	limit := int64(maxPos - minPos + 1)
	offset := int64(minPos)
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, builder_id, ship_id, pool_id, finishes_at
FROM builds
WHERE builder_id = $1
ORDER BY id ASC
OFFSET $2
LIMIT $3
`, int64(c.CommanderID), offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var builds []Build
	for rows.Next() {
		var b Build
		var id, builderID, shipID, poolID int64
		if err := rows.Scan(&id, &builderID, &shipID, &poolID, &b.FinishesAt); err != nil {
			return nil, err
		}
		b.ID = uint32(id)
		b.BuilderID = uint32(builderID)
		b.ShipID = uint32(shipID)
		b.PoolID = uint32(poolID)
		builds = append(builds, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return builds, nil
}

// Bump last login
func (c *Commander) BumpLastLogin() error {
	now := time.Now().UTC()
	c.LastLogin = now
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `UPDATE commanders SET last_login = $2 WHERE commander_id = $1`, int64(c.CommanderID), now)
	return err
}

func (c *Commander) GetSecretaries() []*OwnedShip {
	if len(c.Secretaries) > 0 {
		return c.Secretaries
	}
	// filter out the ships that are not secretaries
	for i, ship := range c.Ships {
		if ship.IsSecretary {
			c.Secretaries = append(c.Secretaries, &c.Ships[i])
		}
	}
	// Sort for PlayerInfo packet (SC_11003)
	sort.Slice(c.Secretaries, func(i, j int) bool {
		if c.Secretaries[i].SecretaryPosition == nil {
			return false
		}
		return *c.Secretaries[i].SecretaryPosition < *c.Secretaries[j].SecretaryPosition
	})
	return c.Secretaries
}

func (c *Commander) GiveSkin(skinId uint32) error {
	if c.OwnedSkinsMap != nil {
		if _, ok := c.OwnedSkinsMap[skinId]; ok {
			return nil
		}
	}
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO owned_skins (commander_id, skin_id, expires_at)
VALUES ($1, $2, NULL)
ON CONFLICT (commander_id, skin_id)
DO NOTHING
`, int64(c.CommanderID), int64(skinId))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return nil
	}
	newSkin := OwnedSkin{CommanderID: c.CommanderID, SkinID: skinId, ExpiresAt: nil}
	c.OwnedSkins = append(c.OwnedSkins, newSkin)
	if c.OwnedSkinsMap == nil {
		c.OwnedSkinsMap = make(map[uint32]*OwnedSkin)
	}
	c.OwnedSkinsMap[skinId] = &c.OwnedSkins[len(c.OwnedSkins)-1]
	return nil
}

func (c *Commander) GiveSkinTx(ctx context.Context, tx pgx.Tx, skinId uint32) error {
	if c.OwnedSkinsMap != nil {
		if _, ok := c.OwnedSkinsMap[skinId]; ok {
			return nil
		}
	}
	res, err := tx.Exec(ctx, `
INSERT INTO owned_skins (commander_id, skin_id, expires_at)
VALUES ($1, $2, NULL)
ON CONFLICT (commander_id, skin_id)
DO NOTHING
`, int64(c.CommanderID), int64(skinId))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return nil
	}
	newSkin := OwnedSkin{CommanderID: c.CommanderID, SkinID: skinId, ExpiresAt: nil}
	c.OwnedSkins = append(c.OwnedSkins, newSkin)
	if c.OwnedSkinsMap != nil {
		c.OwnedSkinsMap[skinId] = &c.OwnedSkins[len(c.OwnedSkins)-1]
	}
	return nil
}

func (c *Commander) GiveSkinWithExpiry(skinId uint32, expiresAt *time.Time) error {
	if c.OwnedSkinsMap != nil {
		if owned, ok := c.OwnedSkinsMap[skinId]; ok {
			if expiresAt != nil {
				if owned.ExpiresAt == nil || expiresAt.After(*owned.ExpiresAt) {
					owned.ExpiresAt = expiresAt
					ctx := context.Background()
					_, err := db.DefaultStore.Pool.Exec(ctx, `UPDATE owned_skins SET expires_at = $3 WHERE commander_id = $1 AND skin_id = $2`, int64(c.CommanderID), int64(skinId), expiresAt)
					return err
				}
			}
			return nil
		}
	}
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO owned_skins (commander_id, skin_id, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, skin_id)
DO UPDATE SET expires_at = CASE
  WHEN owned_skins.expires_at IS NULL THEN EXCLUDED.expires_at
  WHEN EXCLUDED.expires_at IS NULL THEN owned_skins.expires_at
  ELSE GREATEST(owned_skins.expires_at, EXCLUDED.expires_at)
END
`, int64(c.CommanderID), int64(skinId), expiresAt)
	if err != nil {
		return err
	}
	newSkin := OwnedSkin{CommanderID: c.CommanderID, SkinID: skinId, ExpiresAt: expiresAt}
	c.OwnedSkins = append(c.OwnedSkins, newSkin)
	if c.OwnedSkinsMap != nil {
		c.OwnedSkinsMap[skinId] = &newSkin
	}
	return nil
}

func (c *Commander) CleanMailbox() error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM mails WHERE receiver_id = $1`, int64(c.CommanderID))
	return err
}

func (c *Commander) SendMail(mail *Mail) error {
	ctx := context.Background()
	return db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		return c.SendMailTx(ctx, tx, mail)
	})
}

func (c *Commander) SendMailTx(ctx context.Context, tx pgx.Tx, mail *Mail) error {
	mail.ReceiverID = c.CommanderID
	row := tx.QueryRow(ctx, `
INSERT INTO mails (receiver_id, read, date, title, body, attachments_collected, is_important, custom_sender, is_archived, created_at)
VALUES ($1, $2, now(), $3, $4, $5, $6, $7, $8, now())
RETURNING id, date, created_at
`, int64(mail.ReceiverID), mail.Read, mail.Title, mail.Body, mail.AttachmentsCollected, mail.IsImportant, mail.CustomSender, mail.IsArchived)
	var id int64
	if err := row.Scan(&id, &mail.Date, &mail.CreatedAt); err != nil {
		return err
	}
	mail.ID = uint32(id)
	for i := range mail.Attachments {
		att := &mail.Attachments[i]
		att.MailID = mail.ID
		attRow := tx.QueryRow(ctx, `
INSERT INTO mail_attachments (mail_id, type, item_id, quantity)
VALUES ($1, $2, $3, $4)
RETURNING id
`, int64(att.MailID), int64(att.Type), int64(att.ItemID), int64(att.Quantity))
		var attID int64
		if err := attRow.Scan(&attID); err != nil {
			return err
		}
		att.ID = uint32(attID)
	}
	c.Mails = append(c.Mails, *mail)
	if c.MailsMap == nil {
		c.MailsMap = make(map[uint32]*Mail)
	}
	c.MailsMap[mail.ID] = &c.Mails[len(c.Mails)-1]
	return nil
}

func (c *Commander) DestroyShips(shipIds []uint32) error {
	ctx := context.Background()
	ids := make([]int64, 0, len(shipIds))
	for _, id := range shipIds {
		ids = append(ids, int64(id))
	}
	_, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM owned_ships WHERE owner_id = $1 AND id = ANY($2::bigint[])`, int64(c.CommanderID), ids)
	return err
}

// Retire a list of ships, return the amount of medals gained, and an error if any
// Data from : https://azurlane.koumakan.jp/wiki/Building#Retire
func (c *Commander) RetireShips(shipIds *[]uint32) error {
	var coins uint32            // given when a ship is retired
	var totalMedals uint32      // given when a Rare (or higher) ship is retired
	var specializedCores uint32 // given when a UR ship is retired
	for _, shipId := range *shipIds {
		ship, ok := c.OwnedShipsMap[shipId]
		if !ok {
			return fmt.Errorf("ship #%d not found", shipId)
		}
		// Give coins
		switch ship.Ship.Type {
		case 1: // destroyer
			coins += 12
		case 2: // light cruiser
			coins += 14
		case 3: // heavy cruiser
			coins += 18
		case 18: // large cruiser
			coins += 19
		case 13: // monitor
			coins += 13
		case 4: // battlecruiser
			coins += 22
		case 5: // battleship
			coins += 26
		case 10: // aviation battleship
			coins += 25
		case 6: // light carrier
			coins += 16
		case 7: // aircraft carrier
			coins += 16
		case 8: // submarine
		case 17: // submarine carrier
			coins += 10
		case 12: // repair ship
			coins += 13
		case 19: // munition ship
			coins += 11
		default:
			return fmt.Errorf("unknown ship type: %d", ship.Ship.Type)
		}

		// give medals / specialized cores
		switch ship.Ship.RarityID {
		case 2: // normal
			totalMedals += 0
		case 3: // rare
			totalMedals += 1
		case 4: // elite
			totalMedals += 4
		case 5: // super rare
			totalMedals += 10
		case 6: // ultra rare
			totalMedals += 30
			specializedCores += 500
		default:
			return fmt.Errorf("unknown ship rarity: %d", ship.Ship.RarityID)
		}
	}
	if err := c.AddResource(1, coins); err != nil {
		return err
	}
	if err := c.AddItem(15001, totalMedals); err != nil {
		return err
	}
	if err := c.AddItem(59010, specializedCores); err != nil {
		return err
	}
	logger.LogEvent("RetireShip", "Success", fmt.Sprintf("uid=%d, coins: %d, medals: %d, cores: %d", c.CommanderID, coins, totalMedals, specializedCores), logger.LOG_LEVEL_INFO)
	return c.DestroyShips(*shipIds)
}
func (c *Commander) ProposeShip(shipId uint32) (bool, error) {
	// Check if the ship exists
	ship, ok := c.OwnedShipsMap[shipId]
	if !ok {
		logger.LogEvent("Dock", "Propose", fmt.Sprintf("uid=%d has proposed ship id=%d, but it doesn't exist", c.CommanderID, shipId), logger.LOG_LEVEL_ERROR)
		return false, fmt.Errorf("ship #%d not found", shipId)
	}
	// Check if the ship is already proposed
	if ship.Propose {
		logger.LogEvent("Dock", "Propose", fmt.Sprintf("uid=%d has proposed ship id=%d, but it's already proposed", c.CommanderID, shipId), logger.LOG_LEVEL_ERROR)
		return false, fmt.Errorf("ship #%d already proposed", shipId)
	}
	// Check if the commander has a promise ring (id=15006)
	if !c.HasEnoughItem(15006, 1) {
		logger.LogEvent("Dock", "Propose", fmt.Sprintf("uid=%d has proposed ship id=%d, but doesn't have a promise ring", c.CommanderID, shipId), logger.LOG_LEVEL_ERROR)
		return false, fmt.Errorf("missing promise ring")
	}
	// Consume the promise ring
	if err := c.ConsumeItem(15006, 1); err != nil {
		logger.LogEvent("Dock", "Propose", fmt.Sprintf("uid=%d has proposed ship id=%d, but failed to consume the promise ring: %s", c.CommanderID, shipId, err.Error()), logger.LOG_LEVEL_ERROR)
		return false, err
	}
	// Propose the ship
	if err := ship.ProposeShip(); err != nil {
		logger.LogEvent("Dock", "Propose", fmt.Sprintf("uid=%d has proposed ship id=%d, but it failed: %s", c.CommanderID, shipId, err.Error()), logger.LOG_LEVEL_ERROR)
		return false, err
	}
	logger.LogEvent("Dock", "Propose", fmt.Sprintf("uid=%d has proposed ship id=%d successfully", c.CommanderID, shipId), logger.LOG_LEVEL_INFO)
	return true, nil
}

// UpdateRoom changes the commander's room id
func (c *Commander) UpdateRoom(roomID uint32) error {
	c.RoomID = roomID
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `UPDATE commanders SET room_id = $2 WHERE commander_id = $1`, int64(c.CommanderID), int64(roomID))
	return err
}

// RemoveSecretaries removes all secretaries from the commander
func (c *Commander) RemoveSecretaries() error {
	ctx := context.Background()
	return db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
UPDATE owned_ships
SET is_secretary = false,
    secretary_position = NULL,
    secretary_phantom_id = 0
WHERE owner_id = $1
  AND deleted_at IS NULL
  AND is_secretary = true
`, int64(c.CommanderID))
		if err != nil {
			return err
		}
		for _, ship := range c.GetSecretaries() {
			ship.IsSecretary = false
			ship.SecretaryPosition = nil
			ship.SecretaryPhantomID = 0
		}
		c.Secretaries = nil
		return nil
	})
}

// UpdateSecretaries changes the commander's secretaries (dirty implementation, but it works)
func (c *Commander) UpdateSecretaries(updates []SecretaryUpdate) error {
	ctx := context.Background()
	return db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		// remove all secretaries
		if _, err := tx.Exec(ctx, `
UPDATE owned_ships
SET is_secretary = false,
    secretary_position = NULL,
    secretary_phantom_id = 0
WHERE owner_id = $1
  AND deleted_at IS NULL
  AND is_secretary = true
`, int64(c.CommanderID)); err != nil {
			return err
		}
		for _, ship := range c.GetSecretaries() {
			ship.IsSecretary = false
			ship.SecretaryPosition = nil
			ship.SecretaryPhantomID = 0
		}
		c.Secretaries = nil

		for i, update := range updates {
			ship, ok := c.OwnedShipsMap[update.ShipID]
			if !ok {
				return fmt.Errorf("ship #%d not found", update.ShipID)
			}
			pos := uint32(i)
			_, err := tx.Exec(ctx, `
UPDATE owned_ships
SET is_secretary = true,
    secretary_position = $3,
    secretary_phantom_id = $4
WHERE owner_id = $1
  AND id = $2
  AND deleted_at IS NULL
`, int64(c.CommanderID), int64(update.ShipID), int64(pos), int64(update.PhantomID))
			if err != nil {
				return err
			}
			ship.IsSecretary = true
			ship.SecretaryPosition = &pos
			ship.SecretaryPhantomID = update.PhantomID
		}
		return nil
	})
}

// Add n exchange count to the commander, n represents the number of built ships, caps at 400
func (c *Commander) IncrementExchangeCount(n uint32) error {
	c.ExchangeCount += n
	if c.ExchangeCount > 400 {
		c.ExchangeCount = 400
	}
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `UPDATE commanders SET exchange_count = $2 WHERE commander_id = $1`, int64(c.CommanderID), int64(c.ExchangeCount))
	return err
}

func (c *Commander) IncrementDrawCount(count uint32) error {
	switch count {
	case 1:
		c.DrawCount1++
	case 10:
		c.DrawCount10++
	default:
		return nil
	}
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `UPDATE commanders SET draw_count1 = $2, draw_count10 = $3 WHERE commander_id = $1`, int64(c.CommanderID), int64(c.DrawCount1), int64(c.DrawCount10))
	return err
}

func SupportRequisitionMonth(now time.Time) uint32 {
	now = now.UTC()
	return uint32(now.Year()*100 + int(now.Month()))
}

func (c *Commander) EnsureSupportRequisitionMonth(now time.Time) bool {
	month := SupportRequisitionMonth(now)
	if c.SupportRequisitionMonth == month {
		return false
	}
	c.SupportRequisitionMonth = month
	c.SupportRequisitionCount = 0
	return true
}

// Likes a ship, inserts a row into the likes table with the ship's group_id
func (c *Commander) Like(groupId uint32) error {
	like := Like{
		GroupID: groupId,
		LikerID: c.CommanderID,
	}
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `INSERT INTO likes (group_id, liker_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, int64(like.GroupID), int64(like.LikerID))
	return err
}
