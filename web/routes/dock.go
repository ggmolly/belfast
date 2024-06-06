package routes

import (
	"fmt"
	"time"

	"github.com/ggmolly/belfast/orm"
	"github.com/ggmolly/belfast/web/dto"
	"github.com/ggmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

type ownedShipTemplateAttr struct {
	orm.OwnedShip
	Name     string `gorm:"column:name"`
	RarityID int16  `gorm:"column:rarity_id"`
	Propose  bool   `gorm:"column:propose"`
}

type shipModalContent struct {
	ownedShipTemplateAttr
	AvailableTemplates []orm.Ship // represents other versions of the same ship
}

type ShipEditAction func(int, dto.DtoShipEdit, uint32) error

var (
	actionMap = map[string]ShipEditAction{
		"new":       createShip,
		"save":      updateShip,
		"duplicate": duplicateShip,
		"delete":    deleteShip,
	}
)

func createShip(shipId int, data dto.DtoShipEdit, commanderId uint32) error {
	if shipId != -1 {
		return fmt.Errorf("invalid action new")
	}
	// parse createTime from string
	createTime, err := time.Parse("2006-01-02T15:04:05", data.CreateTime)
	if err != nil {
		createTime = time.Now()
	}
	// set a position for the secretary is enabled
	secretaryPosition := new(uint32)
	if data.IsSecretary {
		*secretaryPosition = 0
	}
	newShip := orm.OwnedShip{
		Ship: orm.Ship{
			TemplateID: data.TemplateId,
		},
		Level:               data.Level,
		MaxLevel:            data.MaxLevel,
		Energy:              data.Energy,
		Intimacy:            data.Intimacy,
		IsLocked:            data.IsLocked,
		IsSecretary:         data.IsSecretary,
		Propose:             data.Propose,
		CommonFlag:          data.CommonFlag,
		Proficiency:         data.Proficiency,
		BlueprintFlag:       data.BluePrintFlag,
		ActivityNPC:         data.ActivityNPC,
		CreateTime:          createTime,
		CustomName:          data.CustomName,
		ChangeNameTimestamp: data.CustomNameTime,
		Commander: orm.Commander{
			CommanderID: commanderId,
		},
		SkinID:            data.SkinId,
		SecretaryPosition: secretaryPosition,
	}
	err = orm.GormDB.Create(&newShip).Error
	if err != nil {
		return fmt.Errorf("failed to create ship: %s", err.Error())
	}
	return nil
}

func updateShip(shipId int, data dto.DtoShipEdit, commanderId uint32) error {
	createTime, err := time.Parse("2006-01-02T15:04:05", data.CreateTime)
	if err != nil {
		createTime = time.Now()
	}
	var ship orm.OwnedShip
	err = orm.GormDB.Where("id = ?", shipId).First(&ship).Error
	if err != nil {
		return fmt.Errorf("failed to update ship: %s", err.Error())
	}
	// If the ship is set as secretary but has no position, set it to 0
	if data.IsSecretary && ship.SecretaryPosition == nil {
		// set a position for the secretary
		var secretaryPosition uint32 = 0
		ship.SecretaryPosition = &secretaryPosition
	} else if !data.IsSecretary { // if the ship is not set as secretary but has a position, remove it
		ship.SecretaryPosition = nil
	}
	editedShip := orm.OwnedShip{
		ID: uint32(shipId),
		Ship: orm.Ship{
			TemplateID: data.TemplateId,
		},
		Level:               data.Level,
		MaxLevel:            data.MaxLevel,
		Energy:              data.Energy,
		Intimacy:            data.Intimacy,
		IsLocked:            data.IsLocked,
		IsSecretary:         data.IsSecretary,
		Propose:             data.Propose,
		CommonFlag:          data.CommonFlag,
		Proficiency:         data.Proficiency,
		BlueprintFlag:       data.BluePrintFlag,
		ActivityNPC:         data.ActivityNPC,
		CreateTime:          createTime,
		CustomName:          data.CustomName,
		ChangeNameTimestamp: data.CustomNameTime,
		Commander: orm.Commander{
			CommanderID: commanderId,
		},
		SkinID:            data.SkinId,
		SecretaryPosition: ship.SecretaryPosition,
	}
	err = orm.GormDB.UpdateColumns(&editedShip).Error
	if err != nil {
		return fmt.Errorf("failed to update ship: %s", err.Error())
	}
	return nil
}

func duplicateShip(shipId int, data dto.DtoShipEdit, commanderId uint32) error {
	var copy orm.OwnedShip
	err := orm.GormDB.Where("id = ?", shipId).First(&copy).Error
	copy.ID = 0 // reset the ID so it will be inserted as a new row
	if err != nil {
		return fmt.Errorf("failed to duplicate ship: %s", err.Error())
	}
	err = orm.GormDB.Create(&copy).Error
	if err != nil {
		return fmt.Errorf("failed to duplicate ship: %s", err.Error())
	}
	return nil
}

func deleteShip(shipId int, data dto.DtoShipEdit, commanderId uint32) error {
	err := orm.GormDB.Where("id = ?", shipId).Delete(&orm.OwnedShip{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete ship: %s", err.Error())
	}
	return nil
}

func UpdateShip(c *fiber.Ctx) error {
	shipId := c.Locals("dtoShipId").(dto.DtoShipId).Id
	shipEdit := c.Locals("dtoShipEdit").(dto.DtoShipEdit)
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId
	shipEdit.Id = uint32(shipId)

	fn, ok := actionMap[shipEdit.Action]
	if !ok {
		c.Status(400)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, "Invalid action '"+shipEdit.CustomName+"'", c)
	}
	err := fn(shipId, shipEdit, commanderId)
	if err != nil {
		c.Status(500)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), c)
	}
	return RenderDockContent(c)
}

func EditShipModal(c *fiber.Ctx) error {
	shipId := c.Locals("dtoShipId").(dto.DtoShipId).Id
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId

	var ship shipModalContent

	if err := orm.GormDB.
		Table("owned_ships").
		Joins("inner join ships on ships.template_id = owned_ships.ship_id").
		Select("owned_ships.*, ships.name, ships.rarity_id").
		Where("id = ?", shipId).
		Find(&ship.ownedShipTemplateAttr).Error; err != nil {
		c.Status(500)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), c)
	}

	// get all available templates
	if err := orm.GormDB.
		Select("ships.*").
		Where("name = ?", ship.Name).
		Find(&ship.AvailableTemplates).Error; err != nil {
		c.Status(500)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), c)
	}

	// get all available skins
	var availableSkins []orm.Skin
	strTemplateId := ship.ShipID / 10
	if err := orm.GormDB.
		Select("skins.*").
		Where("ship_group = ?", strTemplateId).
		Find(&availableSkins).Error; err != nil {
		c.Status(500)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), c)
	}

	var selectedTemplate orm.Ship
	for _, template := range ship.AvailableTemplates {
		if template.TemplateID == ship.ShipID {
			selectedTemplate = template
			break
		}
	}

	var selectedSkin orm.Skin
	for _, skin := range availableSkins {
		if skin.ID == uint32(ship.SkinID) {
			selectedSkin = skin
			break
		}
	}

	return c.Render("components/ship/modal", fiber.Map{
		"Rarity":             ship.RarityID,
		"Name":               ship.Name,
		"ID":                 ship.ID,
		"Level":              ship.Level,
		"MaxLevel":           ship.MaxLevel,
		"Energy":             ship.Energy,
		"Intimacy":           ship.Intimacy,
		"IsLocked":           ship.IsLocked,
		"IsSecretary":        ship.IsSecretary,
		"Propose":            ship.Propose,
		"CommonFlag":         ship.CommonFlag,
		"BluePrintFlag":      ship.BlueprintFlag,
		"Proficiency":        ship.Proficiency,
		"CustomName":         ship.CustomName,
		"CustomNameTime":     ship.ChangeNameTimestamp,
		"CreateTime":         ship.CreateTime,
		"ActivityNPC":        ship.ActivityNPC,
		"AvailableTemplates": ship.AvailableTemplates,
		"AvailableSkins":     availableSkins,
		"SelectedTemplate":   selectedTemplate,
		"SelectedSkin":       selectedSkin,
		"OwnerID":            commanderId,
	})
}

func NewShipModal(c *fiber.Ctx) error {
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId
	var ships []orm.Ship
	if err := orm.GormDB.
		Select("ships.*").
		Order("rarity_id asc, template_id asc").
		Find(&ships).Error; err != nil {
		c.Status(500)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), c)
	}

	return c.Render("components/ship/new_modal", fiber.Map{
		"AvailableTemplates": ships,
		"CreateTime":         time.Now(),
		"CommanderId":        commanderId,
	})
}

func GetSkins(c *fiber.Ctx) error {
	templateId := c.Locals("dtoTemplateId").(dto.DtoTemplateId).Id
	templateId /= 10 // remove last digit
	var skins []orm.Skin
	if err := orm.GormDB.
		Select("skins.*").
		Where("ship_group = ?", templateId).
		Find(&skins).Error; err != nil {
		c.Status(500)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), c)
	}

	return c.Render("components/ship/skins_options", fiber.Map{
		"AvailableSkins": skins,
	})
}

func RenderDockContent(c *fiber.Ctx) error {
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId
	var ships []ownedShipTemplateAttr

	err := orm.GormDB.
		Table("owned_ships").
		Joins("inner join ships on ships.template_id = owned_ships.ship_id").
		Select("owned_ships.*, ships.name, ships.rarity_id").
		Where("owner_id = ?", commanderId).
		Where("owned_ships.deleted_at is null").
		Order("owned_ships.is_secretary desc, ships.rarity_id desc").
		Scan(&ships).Error

	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Render("components/dock_content", fiber.Map{
		"ships":   ships,
		"dock_id": commanderId,
	})
}

func RenderDockTemplate(c *fiber.Ctx) error {
	var commanders []orm.Commander
	err := orm.GormDB.
		Select("name, commander_id").
		Order("name asc").
		Find(&commanders).Error

	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	fiberMap := fiber.Map{
		"title":      "Belfast - Docks",
		"page":       "dock",
		"commanders": commanders,
		"endpoint":   "/api/v1/dock",
	}

	if c.Locals("no_layout") == true {
		return c.Render("dock", fiberMap)
	}

	return c.Render("dock", fiberMap, "layouts/main")
}
