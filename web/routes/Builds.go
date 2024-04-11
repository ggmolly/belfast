package routes

import (
	"fmt"
	"time"

	"github.com/bettercallmolly/belfast/orm"
	"github.com/bettercallmolly/belfast/web/dto"
	"github.com/bettercallmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

type BuildEditAction func(int, dto.DtoBuildEdit, uint32) error

type buildWithShip struct {
	orm.Ship
	orm.Build
}

var (
	buildActionMap = map[string]BuildEditAction{
		"new":       createBuild,
		"save":      updateBuild,
		"duplicate": duplicateBuild,
		"delete":    deleteBuild,
		"finish":    finishBuild,
	}
)

func finishBuild(buildId int, data dto.DtoBuildEdit, commanderId uint32) error {
	var build orm.Build
	if err := orm.GormDB.
		Where("id = ?", buildId).
		First(&build).Error; err != nil {
		return err
	}
	if err := orm.GormDB.
		Model(&build).
		Updates(map[string]interface{}{
			"finishes_at": time.Now().Add(-time.Second * 1), // remove 1 second to ensure it's finished
		}).Error; err != nil {
		return err
	}
	return nil
}

func createBuild(buildId int, data dto.DtoBuildEdit, commanderId uint32) error {
	if buildId != -1 {
		return fmt.Errorf("invalid action new")
	}
	// parse finishes_at from string
	finishesAt, err := time.Parse("2006-01-02T15:04:05", data.FinishesAt)
	if err != nil {
		finishesAt = time.Now()
	}
	newBuild := orm.Build{
		ShipID:     data.ShipId,
		BuilderID:  commanderId,
		FinishesAt: finishesAt,
	}
	if err := orm.GormDB.Create(&newBuild).Error; err != nil {
		return err
	}
	return nil
}

func updateBuild(buildId int, data dto.DtoBuildEdit, commanderId uint32) error {
	var build orm.Build
	if err := orm.GormDB.
		Where("id = ?", buildId).
		First(&build).Error; err != nil {
		return err
	} else if err := orm.GormDB.
		Model(&build).
		Updates(map[string]interface{}{
			"ship_id":     data.ShipId,
			"finishes_at": data.FinishesAt,
		}).Error; err != nil {
		return err
	}
	return nil
}

func duplicateBuild(buildId int, data dto.DtoBuildEdit, commanderId uint32) error {
	var build orm.Build
	if err := orm.GormDB.
		Where("id = ?", buildId).
		First(&build).Error; err != nil {
		return err
	}

	// Recreate the build with the same ship
	newBuild := orm.Build{
		ShipID:     build.ShipID,
		BuilderID:  commanderId,
		FinishesAt: build.FinishesAt,
	}
	if err := orm.GormDB.Create(&newBuild).Error; err != nil {
		return err
	}
	return nil
}

func deleteBuild(buildId int, data dto.DtoBuildEdit, commanderId uint32) error {
	if err := orm.GormDB.
		Where("id = ?", buildId).
		Delete(&orm.Build{}).Error; err != nil {
		return err
	}
	return nil
}

func UpdateBuild(c *fiber.Ctx) error {
	buildId := c.Locals("dtoBuildId").(dto.DtoBuildId).Id
	buildEdit := c.Locals("dtoBuildEdit").(dto.DtoBuildEdit)
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId

	fn, ok := buildActionMap[buildEdit.Action]
	if !ok {
		c.Status(400)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, "Invalid action '"+buildEdit.Action+"'", c)
	}
	err := fn(buildId, buildEdit, commanderId)
	if err != nil {
		c.Status(500)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), c)
	}
	return RenderBuildsContent(c)
}

func EditBuildModal(c *fiber.Ctx) error {
	buildId := c.Locals("dtoBuildId").(dto.DtoBuildId).Id
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId

	var build buildWithShip

	if err := orm.GormDB.
		Preload("Ship").
		Table("builds").
		Joins("inner join ships on ships.template_id = builds.ship_id").
		Select("builds.*, ships.name, ships.rarity_id").
		Where("id = ?", buildId).
		Order("finishes_at asc").
		Find(&build).Error; err != nil {
		c.Status(500)
		return c.SendString(err.Error())
	}

	var availableTemplates []orm.Ship
	if err := orm.GormDB.
		Order("rarity_id asc, template_id asc").
		Find(&availableTemplates).Error; err != nil {
		c.Status(500)
		return c.SendString(err.Error())
	}
	return c.Render("components/builds/modal", fiber.Map{
		"Rarity":             build.RarityID,
		"Name":               build.Ship.Name,
		"ID":                 build.ID,
		"ShipID":             build.ShipID,
		"FinishesAt":         build.FinishesAt,
		"AvailableTemplates": availableTemplates,
		"SelectedTemplate":   build.Ship,
		"BuilderID":          commanderId,
	})
}

func NewBuildModal(c *fiber.Ctx) error {
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId
	var ships []orm.Ship
	if err := orm.GormDB.
		Select("ships.*").
		Order("rarity_id asc, template_id asc").
		Find(&ships).Error; err != nil {
		c.Status(500)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), c)
	}

	return c.Render("components/builds/new_modal", fiber.Map{
		"AvailableTemplates": ships,
		"NextHour":           time.Now().Add(time.Hour),
		"BuilderID":          commanderId,
	})
}

func GetBuilds(c *fiber.Ctx) error {
	// templateId := c.Locals("dtoTemplateId").(dto.DtoTemplateId).Id
	// templateId /= 10 // remove last digit
	// var skins []orm.Skin
	// if err := orm.GormDB.
	// 	Select("skins.*").
	// 	Where("ship_group = ?", templateId).
	// 	Find(&skins).Error; err != nil {
	// 	c.Status(500)
	// 	return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), c)
	// }

	// return c.Render("components/ship/skins_options", fiber.Map{
	// 	"AvailableSkins": skins,
	// })
	return c.Status(500).SendString("GetBuilds not implemented")
}

func RenderBuildsContent(c *fiber.Ctx) error {
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId
	var builds []buildWithShip

	err := orm.GormDB.
		Preload("Ship").
		Table("builds").
		Joins("inner join ships on ships.template_id = builds.ship_id").
		Select("builds.*, ships.name, ships.rarity_id").
		Where("builder_id = ?", commanderId).
		Order("finishes_at asc").
		Find(&builds).Error

	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Render("components/builds_content", fiber.Map{
		"builds":       builds,
		"commander_id": commanderId,
	})
}

func RenderBuildsTemplate(c *fiber.Ctx) error {
	var commanders []orm.Commander
	err := orm.GormDB.
		Select("name, commander_id").
		Order("name asc").
		Find(&commanders).Error

	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	fiberMap := fiber.Map{
		"title":      "Belfast - Builds",
		"page":       "builds",
		"commanders": commanders,
		"endpoint":   "/api/v1/builds",
	}

	if c.Locals("no_layout") == true {
		return c.Render("builds", fiberMap)
	}

	return c.Render("builds", fiberMap, "layouts/main")
}
