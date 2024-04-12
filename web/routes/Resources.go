package routes

import (
	"github.com/ggmolly/belfast/orm"
	"github.com/ggmolly/belfast/web/dto"
	"github.com/gofiber/fiber/v2"
)

type ownedResourceWithAttrs struct {
	orm.OwnedResource
	Name   string `gorm:"column:name"`
	ItemID uint32 `gorm:"column:item_id"`
}

type ResourceEditAction func(dto.DtoResourceId, dto.DtoResourceEdit, uint32) error

var (
	resourceEditActions = map[string]ResourceEditAction{
		"save":   saveResource,
		"delete": deleteResource,
		"new":    newResource,
	}
)

func saveResource(resourceId dto.DtoResourceId, data dto.DtoResourceEdit, commanderId uint32) error {
	return orm.GormDB.
		Table("owned_resources").
		Where("commander_id = ? and resource_id = ?", commanderId, resourceId.ResourceId).
		Update("amount", data.Amount).Error
}

func deleteResource(resourceId dto.DtoResourceId, data dto.DtoResourceEdit, commanderId uint32) error {
	return orm.GormDB.
		Table("owned_resources").
		Where("commander_id = ? and resource_id = ?", commanderId, resourceId.ResourceId).
		Delete(&orm.OwnedResource{}).Error
}

func newResource(resourceId dto.DtoResourceId, data dto.DtoResourceEdit, commanderId uint32) error {
	var newResource orm.OwnedResource
	newResource.CommanderID = commanderId
	newResource.ResourceID = data.ResourceId
	newResource.Amount = data.Amount
	return orm.GormDB.Create(&newResource).Error
}

func UpdateResource(c *fiber.Ctx) error {
	resourceId := c.Locals("dtoResourceId").(dto.DtoResourceId)
	resourceEdit := c.Locals("dtoResourceEdit").(dto.DtoResourceEdit)
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId

	fn, ok := resourceEditActions[resourceEdit.Action]
	if !ok {
		return c.Status(400).SendString("invalid action")
	}
	err := fn(resourceId, resourceEdit, commanderId)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return RenderResourcesContent(c)
}

func NewResourcesModal(c *fiber.Ctx) error {
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId

	var resources []orm.Resource
	err := orm.GormDB.
		Select("name, id").
		Order("name asc").
		Find(&resources).Error

	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Render("components/resources/new_modal", fiber.Map{
		"AvailableResources": resources,
		"commander_id":       commanderId,
	})
}

func ResourcesModal(c *fiber.Ctx) error {
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId
	resourceId := c.Locals("dtoResourceId").(dto.DtoResourceId).ResourceId

	var resource ownedResourceWithAttrs
	err := orm.GormDB.
		Table("owned_resources").
		Joins("inner join resources on resources.id = owned_resources.resource_id").
		Select("owned_resources.*, resources.*").
		Where("commander_id = ? and owned_resources.resource_id = ?", commanderId, resourceId).
		Scan(&resource).Error
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Render("components/resources/modal", fiber.Map{
		"resource":     resource,
		"commander_id": commanderId,
	})
}

func RenderResourcesContent(c *fiber.Ctx) error {
	commanderId := c.Locals("dtoCommanderId").(dto.DtoCommanderId).CommanderId
	var resources []ownedResourceWithAttrs

	err := orm.GormDB.
		Table("owned_resources").
		Joins("inner join resources on resources.id = owned_resources.resource_id").
		Select("owned_resources.*, resources.*").
		Where("commander_id = ?", commanderId).
		Scan(&resources).Error

	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Render("components/resources_content", fiber.Map{
		"resources":    resources,
		"commander_id": commanderId,
	})
}

func RenderResourcesTemplate(c *fiber.Ctx) error {
	var commanders []orm.Commander
	err := orm.GormDB.
		Select("name, commander_id").
		Order("name asc").
		Find(&commanders).Error

	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	fiberMap := fiber.Map{
		"title":      "Belfast - Resources",
		"page":       "resources",
		"commanders": commanders,
		"endpoint":   "/api/v1/resources",
	}

	if c.Locals("no_layout") == true {
		return c.Render("resources", fiberMap)
	}

	return c.Render("resources", fiberMap, "layouts/main")
}
