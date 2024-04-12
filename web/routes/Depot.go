package routes

import (
	"fmt"
	"time"

	"github.com/ggmolly/belfast/orm"
	"github.com/ggmolly/belfast/web/dto"
	"github.com/ggmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

type itemWithName struct {
	orm.CommanderItem
	Name string `gorm:"column:name"`
}

type itemWithCount struct {
	orm.Item
	Count uint32 `gorm:"column:count"`
}

type ItemEditAction func(dto.DtoItemId, dto.DtoItemEdit, uint32) error

var (
	itemActionMap = map[string]ItemEditAction{
		"save":   updateItem,
		"delete": deleteItem,
		"new":    createItem,
	}
)

func createItem(itemId dto.DtoItemId, data dto.DtoItemEdit, commanderId uint32) error {
	if itemId.Id != -1 {
		return fmt.Errorf("invalid action new")
	}
	if data.TemplateId == 44001 { // Special stupid case for Valentine Gift which is a misc item
		var newItem orm.CommanderMiscItem
		newItem.CommanderID = commanderId
		newItem.ItemID = data.TemplateId
		newItem.Data = data.Data
		return orm.GormDB.Create(&newItem).Error
	}
	var newItem orm.CommanderItem
	newItem.CommanderID = commanderId
	newItem.ItemID = data.TemplateId
	newItem.Count = data.Count
	return orm.GormDB.Create(&newItem).Error
}

func deleteItem(itemId dto.DtoItemId, data dto.DtoItemEdit, commanderId uint32) error {
	if itemId.Id == 44001 {
		return orm.GormDB.
			Where("commander_id = ? and item_id = ? and data = ?", commanderId, itemId.Id, itemId.Data).
			Delete(&orm.CommanderMiscItem{}).Error
	}
	return orm.GormDB.
		Where("commander_id = ? and item_id = ?", commanderId, itemId.Id).
		Delete(&orm.CommanderItem{}).Error
}

func updateItem(itemId dto.DtoItemId, data dto.DtoItemEdit, commanderId uint32) error {
	if itemId.Id == 44001 {
		return orm.GormDB.
			Where("commander_id = ? and item_id = ? and data = ?", commanderId, itemId.Id, itemId.Data).
			Assign("data", data.Data).
			FirstOrCreate(&orm.CommanderMiscItem{}).Error
	}

	return orm.GormDB.
		Where("commander_id = ? and item_id = ?", commanderId, itemId.Id).
		Assign("count", data.Count).
		FirstOrCreate(&orm.CommanderItem{}).Error
}

func UpdateItem(c *fiber.Ctx) error {
	itemId := c.Locals("dtoItemId").(dto.DtoItemId)
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId
	itemEdit := c.Locals("dtoItemEdit").(dto.DtoItemEdit)

	fn, ok := itemActionMap[itemEdit.Action]
	if !ok {
		return c.Status(400).SendString("invalid action")
	}
	err := fn(itemId, itemEdit, commanderId)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return RenderDepotContent(c)
}

func NewItemModal(c *fiber.Ctx) error {
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId
	var items []orm.Item
	if err := orm.GormDB.
		Order("rarity asc, id asc").
		Find(&items).Error; err != nil {
		c.Status(500)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), c)
	}
	return c.Render("components/item/new_modal", fiber.Map{
		"AvailableTemplates": items,
		"CreateTime":         time.Now(),
		"CommanderId":        commanderId,
	})
}

func EditItemModal(c *fiber.Ctx) error {
	itemId := c.Locals("dtoItemId").(dto.DtoItemId)
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId

	parameterMap := fiber.Map{
		"ItemID":      itemId.Id,
		"CommanderID": commanderId,
	}

	if itemId.Id == 44001 { // Special stupid case for Valentine Gift which is a misc item
		var item orm.CommanderMiscItem
		err := orm.GormDB.
			Where("commander_id = ? and item_id = ? and data = ?", commanderId, itemId.Id, itemId.Data).
			First(&item).Error
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		parameterMap["Data"] = item.Data
		parameterMap["Count"] = 1
		parameterMap["Name"] = item.Item.Name
		parameterMap["Rarity"] = item.Item.Rarity
		parameterMap["Type"] = 23
	} else {
		var item itemWithCount
		err := orm.GormDB.
			Table("items").
			Select("items.*, commander_items.count").
			Joins("left join commander_items on commander_items.item_id = items.id and commander_items.commander_id = ?", commanderId).
			Where("items.id = ?", itemId.Id).
			First(&item).Error
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		parameterMap["Count"] = item.Count
		parameterMap["Name"] = item.Name
		parameterMap["Rarity"] = item.Rarity
		parameterMap["Type"] = item.Type
	}
	return c.Render("components/item/modal", parameterMap)
}

func RenderDepotContent(c *fiber.Ctx) error {
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId

	var items []orm.CommanderItem
	var miscItems []orm.CommanderMiscItem
	err := orm.GormDB.
		Preload("Item").
		Where("commander_id = ?", commanderId).
		Find(&items).Error
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	err = orm.GormDB.
		Preload("Item").
		Where("commander_id = ?", commanderId).
		Find(&miscItems).Error
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Render("components/depot_content", fiber.Map{
		"items":      items,
		"misc_items": miscItems,
		"depot_id":   commanderId,
	})
}

func RenderDepotTemplate(c *fiber.Ctx) error {
	var commanders []orm.Commander
	err := orm.GormDB.
		Select("name, commander_id").
		Order("name asc").
		Find(&commanders).Error

	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	fiberMap := fiber.Map{
		"title":      "Belfast - Depots",
		"page":       "depot",
		"commanders": commanders,
		"endpoint":   "/api/v1/depot",
	}

	if c.Locals("no_layout") == true {
		return c.Render("depot", fiberMap)
	}

	return c.Render("depot", fiberMap, "layouts/main")
}
