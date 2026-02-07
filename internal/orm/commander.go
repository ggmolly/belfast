package orm

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/ggmolly/belfast/internal/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Commander struct {
	CommanderID             uint32         `gorm:"primary_key"`
	AccountID               uint32         `gorm:"not_null"`
	Level                   int            `gorm:"default:1;not_null"`
	Exp                     int            `gorm:"default:0;not_null"`
	Name                    string         `gorm:"size:30;not_null;uniqueIndex"`
	LastLogin               time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	GuideIndex              uint32         `gorm:"default:0;not_null"`
	NewGuideIndex           uint32         `gorm:"default:0;not_null"`
	NameChangeCooldown      time.Time      `gorm:"type:timestamp;default:'1970-01-01 00:00:00';not_null"`
	RoomID                  uint32         `gorm:"default:0;not_null"`
	ExchangeCount           uint32         `gorm:"default:0;not_null"` // Number of times the commander has built ships, can be exchanged for UR ships
	DrawCount1              uint32         `gorm:"default:0;not_null"`
	DrawCount10             uint32         `gorm:"default:0;not_null"`
	SupportRequisitionCount uint32         `gorm:"default:0;not_null"`
	SupportRequisitionMonth uint32         `gorm:"default:0;not_null"`
	CollectAttackCount      uint32         `gorm:"default:0;not_null"`
	AccPayLv                uint32         `gorm:"default:0;not_null"`
	LivingAreaCoverID       uint32         `gorm:"default:0;not_null"`
	SelectedIconFrameID     uint32         `gorm:"default:0;not_null"`
	SelectedChatFrameID     uint32         `gorm:"default:0;not_null"`
	SelectedBattleUIID      uint32         `gorm:"default:0;not_null"`
	DisplayIconID           uint32         `gorm:"default:0;not_null"`
	DisplaySkinID           uint32         `gorm:"default:0;not_null"`
	DisplayIconThemeID      uint32         `gorm:"default:0;not_null"`
	Manifesto               string         `gorm:"size:200;default:'';not_null"`
	DormName                string         `gorm:"size:50;default:'';not_null"`
	RandomShipMode          uint32         `gorm:"default:0;not_null"`
	RandomFlagShipEnabled   bool           `gorm:"default:false;not_null"`
	DeletedAt               gorm.DeletedAt `gorm:"index"`

	Punishments      []Punishment        `gorm:"foreignKey:PunishedID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Ships            []OwnedShip         `gorm:"foreignKey:OwnerID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Items            []CommanderItem     `gorm:"foreignKey:CommanderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	MiscItems        []CommanderMiscItem `gorm:"foreignKey:CommanderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	OwnedResources   []OwnedResource     `gorm:"foreignKey:CommanderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Builds           []Build             `gorm:"foreignKey:BuilderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Mails            []Mail              `gorm:"foreignKey:ReceiverID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Compensations    []Compensation      `gorm:"foreignKey:CommanderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
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

func (c *Commander) BeforeSave(tx *gorm.DB) error {
	if c.Level > 120 {
		c.Level = 120
	}
	return nil
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
	err = GormDB.Create(&newBuild).Error
	if err != nil {
		return nil, 0, err
	}
	*runningBuilds++ // the game requires us to send a sequential build id

	// Add the build to the commander's list of BuildsMap
	c.Builds = append(c.Builds, newBuild)
	c.BuildsMap[newBuild.ID] = &newBuild

	return &newBuild, ship.BuildTime, nil
}

func (c *Commander) AddShip(shipId uint32) (*OwnedShip, error) {
	var ship Ship
	err := GormDB.Where("template_id = ?", shipId).First(&ship).Error
	if err != nil {
		return nil, err
	}
	newShip := OwnedShip{
		ShipID:  ship.TemplateID,
		OwnerID: c.CommanderID,
	}
	tx := GormDB.Begin()
	if err := tx.Create(&newShip).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := createDefaultShipEquipments(tx, c.CommanderID, newShip.ID, ship.TemplateID); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	// Add the ship to the commander's list of owned ships
	c.Ships = append(c.Ships, newShip)
	c.OwnedShipsMap[newShip.ID] = &newShip
	return &newShip, nil
}

func (c *Commander) AddShipTx(tx *gorm.DB, shipId uint32) (*OwnedShip, error) {
	var ship Ship
	if err := tx.Where("template_id = ?", shipId).First(&ship).Error; err != nil {
		return nil, err
	}
	newShip := OwnedShip{
		ShipID:  ship.TemplateID,
		OwnerID: c.CommanderID,
	}
	if err := tx.Create(&newShip).Error; err != nil {
		return nil, err
	}
	if err := createDefaultShipEquipments(tx, c.CommanderID, newShip.ID, ship.TemplateID); err != nil {
		return nil, err
	}
	c.Ships = append(c.Ships, newShip)
	c.OwnedShipsMap[newShip.ID] = &newShip
	return &newShip, nil
}

func createDefaultShipEquipments(tx *gorm.DB, ownerID uint32, ownedShipID uint32, templateID uint32) error {
	config, err := GetShipEquipConfigTx(tx, templateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return createDefaultShipEquipmentsWithoutConfig(tx, ownerID, ownedShipID)
		}
		return err
	}
	return createDefaultShipEquipmentsFromConfig(tx, ownerID, ownedShipID, config)
}

func createDefaultShipEquipmentsWithoutConfig(tx *gorm.DB, ownerID uint32, ownedShipID uint32) error {
	for pos := uint32(1); pos <= 3; pos++ {
		entry := OwnedShipEquipment{
			OwnerID: ownerID,
			ShipID:  ownedShipID,
			Pos:     pos,
			EquipID: 0,
			SkinID:  0,
		}
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}
	}
	return nil
}

func createDefaultShipEquipmentsFromConfig(tx *gorm.DB, ownerID uint32, ownedShipID uint32, config *ShipEquipConfig) error {
	slotCount := config.SlotCount()
	for pos := uint32(1); pos <= slotCount; pos++ {
		entry := OwnedShipEquipment{
			OwnerID: ownerID,
			ShipID:  ownedShipID,
			Pos:     pos,
			EquipID: config.DefaultEquipID(pos),
			SkinID:  0,
		}
		if err := tx.Create(&entry).Error; err != nil {
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
		res := GormDB.Model(&CommanderItem{}).
			Where("commander_id = ? AND item_id = ? AND count >= ?", c.CommanderID, itemId, count).
			UpdateColumn("count", gorm.Expr("count - ?", count))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("not enough items")
		}
		item.Count -= count
		return nil
	} else if miscItem, ok := c.MiscItemsMap[itemId]; ok {
		if miscItem.Data >= count {
			miscItem.Data -= count
			return GormDB.Save(&miscItem).Error
		}
	}
	return fmt.Errorf("not enough items")
}

func (c *Commander) ConsumeItemTx(tx *gorm.DB, itemId uint32, count uint32) error {
	if item, ok := c.CommanderItemsMap[itemId]; ok {
		if item.Count < count {
			return fmt.Errorf("not enough items")
		}
		res := tx.Model(&CommanderItem{}).
			Where("commander_id = ? AND item_id = ? AND count >= ?", c.CommanderID, itemId, count).
			UpdateColumn("count", gorm.Expr("count - ?", count))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("not enough items")
		}
		item.Count -= count
		return nil
	} else if miscItem, ok := c.MiscItemsMap[itemId]; ok {
		if miscItem.Data >= count {
			miscItem.Data -= count
			return tx.Save(&miscItem).Error
		}
	}
	return fmt.Errorf("not enough items")
}

func (c *Commander) SaveTx(tx *gorm.DB) error {
	return tx.Save(c).Error
}

func (c *Commander) ConsumeResource(resourceId uint32, count uint32) error {
	DealiasResource(&resourceId)
	// check if the commander has enough of the resource
	if resource, ok := c.OwnedResourcesMap[resourceId]; ok {
		if resource.Amount >= count {
			resource.Amount -= count
			return GormDB.Save(&resource).Error
		}
	}
	return fmt.Errorf("not enough resources")
}

func (c *Commander) ConsumeResourceTx(tx *gorm.DB, resourceId uint32, count uint32) error {
	DealiasResource(&resourceId)
	if resource, ok := c.OwnedResourcesMap[resourceId]; ok {
		if resource.Amount >= count {
			resource.Amount -= count
			return tx.Save(&resource).Error
		}
	}
	return fmt.Errorf("not enough resources")
}

func (c *Commander) SetResource(resourceId uint32, amount uint32) error {
	// check if the commander already has the resource, if so set the amount and save
	if resource, ok := c.OwnedResourcesMap[resourceId]; ok {
		resource.Amount = amount
		return GormDB.Save(resource).Error
	}
	// otherwise create a new resource
	newResource := OwnedResource{
		CommanderID: c.CommanderID,
		ResourceID:  resourceId,
		Amount:      amount,
	}
	err := GormDB.Create(&newResource).Error
	if err != nil {
		// append the new resource to the commander's list of resources
		c.OwnedResources = append(c.OwnedResources, newResource)
		c.OwnedResourcesMap[resourceId] = &newResource
	}
	return err
}

func (c *Commander) SetItem(itemId uint32, amount uint32) error {
	// check if the commander already has the item, if so set the amount and save
	if item, ok := c.CommanderItemsMap[itemId]; ok {
		item.Count = amount
		return GormDB.Save(item).Error
	}
	// otherwise create a new item
	newItem := CommanderItem{
		CommanderID: c.CommanderID,
		ItemID:      itemId,
		Count:       amount,
	}
	err := GormDB.Create(&newItem).Error
	if err != nil {
		// append the new item to the commander's list of items
		c.Items = append(c.Items, newItem)
		c.CommanderItemsMap[itemId] = &newItem
	}
	return err
}

func (c *Commander) AddResource(resourceId uint32, amount uint32) error {
	if c.OwnedResourcesMap == nil {
		c.OwnedResourcesMap = make(map[uint32]*OwnedResource)
	}
	// check if the commander already has the resource, if so increment the amount and save
	if resource, ok := c.OwnedResourcesMap[resourceId]; ok {
		resource.Amount += amount
		return GormDB.Save(resource).Error
	}
	// otherwise create or increment the resource
	newResource := OwnedResource{
		CommanderID: c.CommanderID,
		ResourceID:  resourceId,
		Amount:      amount,
	}
	if err := GormDB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commander_id"}, {Name: "resource_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"amount": gorm.Expr("amount + ?", amount),
		}),
	}).Create(&newResource).Error; err != nil {
		return err
	}
	var stored OwnedResource
	if err := GormDB.Where("commander_id = ? AND resource_id = ?", c.CommanderID, resourceId).First(&stored).Error; err != nil {
		return err
	}
	for i := range c.OwnedResources {
		if c.OwnedResources[i].ResourceID == resourceId {
			c.OwnedResources[i] = stored
			c.OwnedResourcesMap[resourceId] = &c.OwnedResources[i]
			return nil
		}
	}
	c.OwnedResources = append(c.OwnedResources, stored)
	c.OwnedResourcesMap[resourceId] = &c.OwnedResources[len(c.OwnedResources)-1]
	return nil
}

func (c *Commander) AddResourceTx(tx *gorm.DB, resourceId uint32, amount uint32) error {
	DealiasResource(&resourceId)
	if c.OwnedResourcesMap == nil {
		c.OwnedResourcesMap = make(map[uint32]*OwnedResource)
	}

	entry := OwnedResource{CommanderID: c.CommanderID, ResourceID: resourceId, Amount: amount}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commander_id"}, {Name: "resource_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"amount": gorm.Expr("amount + ?", amount),
		}),
	}).Create(&entry).Error; err != nil {
		return err
	}
	if existing, ok := c.OwnedResourcesMap[resourceId]; ok {
		existing.Amount += amount
		return nil
	}

	var stored OwnedResource
	if err := tx.Where("commander_id = ? AND resource_id = ?", c.CommanderID, resourceId).First(&stored).Error; err != nil {
		return err
	}
	for i := range c.OwnedResources {
		if c.OwnedResources[i].ResourceID == resourceId {
			c.OwnedResources[i] = stored
			c.OwnedResourcesMap[resourceId] = &c.OwnedResources[i]
			return nil
		}
	}
	c.OwnedResources = append(c.OwnedResources, stored)
	c.OwnedResourcesMap[resourceId] = &c.OwnedResources[len(c.OwnedResources)-1]
	return nil
}

func (c *Commander) AddItem(itemId uint32, amount uint32) error {
	// check if the commander already has the item, if so increment the amount and save
	if item, ok := c.CommanderItemsMap[itemId]; ok {
		item.Count += amount
		return GormDB.Save(item).Error
	}
	// otherwise create or increment the item
	newItem := CommanderItem{
		CommanderID: c.CommanderID,
		ItemID:      itemId,
		Count:       amount,
	}
	if err := GormDB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commander_id"}, {Name: "item_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"count": gorm.Expr("count + ?", amount),
		}),
	}).Create(&newItem).Error; err != nil {
		return err
	}
	var stored CommanderItem
	if err := GormDB.Where("commander_id = ? AND item_id = ?", c.CommanderID, itemId).First(&stored).Error; err != nil {
		return err
	}
	for i := range c.Items {
		if c.Items[i].ItemID == itemId {
			c.Items[i] = stored
			c.CommanderItemsMap[itemId] = &c.Items[i]
			return nil
		}
	}
	c.Items = append(c.Items, stored)
	c.CommanderItemsMap[itemId] = &c.Items[len(c.Items)-1]
	return nil
}

func (c *Commander) AddItemTx(tx *gorm.DB, itemId uint32, amount uint32) error {
	if c.CommanderItemsMap == nil {
		c.CommanderItemsMap = make(map[uint32]*CommanderItem)
	}

	entry := CommanderItem{CommanderID: c.CommanderID, ItemID: itemId, Count: amount}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commander_id"}, {Name: "item_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"count": gorm.Expr("count + ?", amount),
		}),
	}).Create(&entry).Error; err != nil {
		return err
	}
	if existing, ok := c.CommanderItemsMap[itemId]; ok {
		existing.Count += amount
		return nil
	}

	var stored CommanderItem
	if err := tx.Where("commander_id = ? AND item_id = ?", c.CommanderID, itemId).First(&stored).Error; err != nil {
		return err
	}
	for i := range c.Items {
		if c.Items[i].ItemID == itemId {
			c.Items[i] = stored
			c.CommanderItemsMap[itemId] = &c.Items[i]
			return nil
		}
	}
	c.Items = append(c.Items, stored)
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
	punishment := Punishment{
		PunishedID:    c.CommanderID,
		IsPermanent:   permanent,
		LiftTimestamp: liftTimestamp,
	}
	return GormDB.Create(&punishment).Error
}

func (c *Commander) RevokeActivePunishment() error {
	return GormDB.Where("punished_id = ? AND lift_timestamp IS NULL", c.CommanderID).Delete(&Punishment{}).Error
}

// Load loads the commander's data from the database (ships, items, resources, etc)
func (c *Commander) Load() error {
	err := GormDB.
		Preload(clause.Associations).
		Preload("Ships.Ship"). // force preload the ship's data (might be rolled back later for a lazy load instead and replacement of retire switches to map)
		Preload("Ships.Equipments").
		Preload("Ships.Transforms").
		Preload("Mails.Attachments"). // force preload attachments
		Preload("Compensations.Attachments").
		First(c, c.CommanderID).
		Error

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
	return err
}

// Commit saves the commander's data to the database (ships, items, resources, etc)
func (c *Commander) Commit() error {
	return GormDB.Session(&gorm.Session{FullSaveAssociations: true}).Save(c).Error
}

// Get a range of builds (special weird query, probably to save battery on phones)
func (c *Commander) GetBuildRange(minPos, maxPos uint32) ([]Build, error) {
	var builds []Build
	err := GormDB.
		Where("builder_id = ?", c.CommanderID).
		Offset(int(minPos)).
		Limit(int(maxPos - minPos + 1)). // stupid hack to select a range of rows
		Order("id ASC").
		Find(&builds).
		Error
	return builds, err
}

// Bump last login
func (c *Commander) BumpLastLogin() error {
	c.LastLogin = time.Now()
	return GormDB.Save(c).Error
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
	newSkin := OwnedSkin{
		CommanderID: c.CommanderID,
		SkinID:      skinId,
	}
	if err := GormDB.Create(&newSkin).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	c.OwnedSkins = append(c.OwnedSkins, newSkin)
	c.OwnedSkinsMap[skinId] = &newSkin
	return nil
}

func (c *Commander) GiveSkinTx(tx *gorm.DB, skinId uint32) error {
	if c.OwnedSkinsMap != nil {
		if _, ok := c.OwnedSkinsMap[skinId]; ok {
			return nil
		}
	}
	newSkin := OwnedSkin{
		CommanderID: c.CommanderID,
		SkinID:      skinId,
		ExpiresAt:   nil,
	}
	if err := tx.Create(&newSkin).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
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
					return GormDB.Save(owned).Error
				}
			}
			return nil
		}
	}
	newSkin := OwnedSkin{
		CommanderID: c.CommanderID,
		SkinID:      skinId,
		ExpiresAt:   expiresAt,
	}
	if err := GormDB.Create(&newSkin).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	c.OwnedSkins = append(c.OwnedSkins, newSkin)
	if c.OwnedSkinsMap != nil {
		c.OwnedSkinsMap[skinId] = &newSkin
	}
	return nil
}

func (c *Commander) CleanMailbox() error {
	return GormDB.Where("receiver_id = ?", c.CommanderID).Delete(&Mail{}).Error
}

func (c *Commander) SendMail(mail *Mail) error {
	return c.SendMailTx(GormDB, mail)
}

func (c *Commander) SendMailTx(tx *gorm.DB, mail *Mail) error {
	mail.ReceiverID = c.CommanderID
	if err := tx.Create(mail).Error; err != nil {
		return err
	}
	c.Mails = append(c.Mails, *mail)
	if c.MailsMap == nil {
		c.MailsMap = make(map[uint32]*Mail)
	}
	c.MailsMap[mail.ID] = &c.Mails[len(c.Mails)-1]
	return nil
}

func (c *Commander) DestroyShips(shipIds []uint32) error {
	return GormDB.Where("owner_id = ? AND id IN ?", c.CommanderID, shipIds).Delete(&OwnedShip{}).Error
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
	return GormDB.Model(c).Update("room_id", roomID).Error
}

// RemoveSecretaries removes all secretaries from the commander
func (c *Commander) RemoveSecretaries() error {
	tx := GormDB.Begin()
	for _, ship := range c.GetSecretaries() {
		ship.IsSecretary = false
		ship.SecretaryPosition = nil
		ship.SecretaryPhantomID = 0
		if err := tx.Save(ship).Error; err != nil {
			return err
		}
	}
	return tx.Commit().Error
}

// UpdateSecretaries changes the commander's secretaries (dirty implementation, but it works)
func (c *Commander) UpdateSecretaries(updates []SecretaryUpdate) error {
	tx := GormDB.Begin() // start a transaction to update all at once
	// remove all secretaries
	for _, ship := range c.GetSecretaries() {
		ship.IsSecretary = false
		ship.SecretaryPosition = nil
		ship.SecretaryPhantomID = 0
		if err := tx.Save(ship).Error; err != nil {
			return err
		}
	}
	// add the new secretaries
	for i, update := range updates {
		ship, ok := c.OwnedShipsMap[update.ShipID]
		if !ok {
			return fmt.Errorf("ship #%d not found", update.ShipID)
		}
		ship.IsSecretary = true
		ship.SecretaryPosition = new(uint32)
		*ship.SecretaryPosition = uint32(i)
		ship.SecretaryPhantomID = update.PhantomID
		if err := tx.Save(ship).Error; err != nil {
			return err
		}
	}
	return tx.Commit().Error
}

// Add n exchange count to the commander, n represents the number of built ships, caps at 400
func (c *Commander) IncrementExchangeCount(n uint32) error {
	c.ExchangeCount += n
	if c.ExchangeCount > 400 {
		c.ExchangeCount = 400
	}
	return GormDB.Save(c).Error
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
	return GormDB.Save(c).Error
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
	return GormDB.Create(&like).Error
}
