package routes

import (
	"github.com/bettercallmolly/belfast/logger"
	"github.com/bettercallmolly/belfast/misc"
	"github.com/bettercallmolly/belfast/orm"
	"github.com/bettercallmolly/belfast/web/dto"
	"github.com/bettercallmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

func AnnounceServer(c *fiber.Ctx) error {
	serverAnnounce := c.Locals("dtoServerData").(dto.DtoAnnounceServer)

	newServer := orm.Server{
		Name:      serverAnnounce.Name,
		IP:        serverAnnounce.Ip,
		Port:      serverAnnounce.Port,
		StateID:   &serverAnnounce.State,
		ProxyIP:   serverAnnounce.ProxyIp,
		ProxyPort: serverAnnounce.ProxyPort,
	}
	err := orm.GormDB.Create(&newServer)

	if err.Error != nil {
		logger.LogEvent("Database", "Servers", err.Error.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, "A database error occurred, check the logs.", c)
	}

	c.Locals("include_layout", false)
	return RenderServerTemplate(c)
}

func DeleteServer(c *fiber.Ctx) error {
	serverDelete := c.Locals("dtoServerId").(dto.DtoServerId)

	err := orm.GormDB.Where("id = ?", serverDelete.Id).Delete(&orm.Server{})

	if err.Error != nil {
		logger.LogEvent("Database", "Servers", err.Error.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, "A database error occurred, check the logs.", c)
	}

	c.Locals("include_layout", false)
	return RenderServerTemplate(c)
}

func UpdateServer(c *fiber.Ctx) error {
	serverUpdate := c.Locals("dtoServerData").(dto.DtoAnnounceServer)
	serverId := c.Locals("dtoServerId").(dto.DtoServerId)

	err := orm.GormDB.Where("id = ?", serverId.Id).Updates(&orm.Server{
		IP:        serverUpdate.Ip,
		Port:      serverUpdate.Port,
		StateID:   &serverUpdate.State,
		ProxyIP:   serverUpdate.ProxyIp,
		ProxyPort: serverUpdate.ProxyPort,
		Name:      serverUpdate.Name,
	})
	if err.Error != nil {
		logger.LogEvent("Database", "Servers", err.Error.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, "A database error occurred, check the logs.", c)
	}

	c.Locals("include_layout", false)
	return RenderServer(c)
}

func EditServer(c *fiber.Ctx) error {
	serverEdit := c.Locals("dtoServerId").(dto.DtoServerId)

	var server orm.Server
	err := orm.GormDB.Preload("State").Where("id = ?", serverEdit.Id).First(&server)

	if err.Error != nil {
		logger.LogEvent("Database", "Servers", err.Error.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, "A database error occurred, check the logs.", c)
	}

	return c.Render("components/editable_server_row", fiber.Map{
		"ID":    server.ID,
		"IP":    server.IP,
		"Port":  server.Port,
		"State": server.State.ID,
		"Name":  server.Name,
	})
}

func RenderServer(c *fiber.Ctx) error {
	serverId := c.Locals("dtoServerId").(dto.DtoServerId)
	var server orm.Server
	err := orm.GormDB.Preload("State").Where("id = ?", serverId.Id).First(&server)

	if err.Error != nil {
		logger.LogEvent("Database", "Servers", err.Error.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, "A database error occurred, check the logs.", c)
	}

	return c.Render("components/server_row", fiber.Map{
		"ID":    server.ID,
		"IP":    server.IP,
		"Port":  server.Port,
		"State": server.State,
		"Name":  server.Name,
	})
}

func RenderServers(c *fiber.Ctx) error {
	c.Locals("include_layout", true)
	return RenderServerTemplate(c)
}

func RenderServerTemplate(c *fiber.Ctx) error {
	var servers []orm.Server

	// join the servers table with the server_states table to get the color that's in server_states
	err := orm.GormDB.
		Preload("State").
		Order("id ASC").
		Find(&servers)

	if err.Error != nil {
		logger.LogEvent("Database", "Servers", err.Error.Error(), logger.LOG_LEVEL_ERROR)
		return c.Status(500).SendString("A database error occurred, check the logs.")
	}

	if c.Locals("include_layout") != nil {
		return c.Render("servers", fiber.Map{
			"title":   "Belfast - Servers",
			"commits": misc.GetCommits(),
			"page":    "servers",
			"servers": servers,
		})
	}
	return c.Render("servers", fiber.Map{
		"title":   "Belfast - Servers",
		"commits": misc.GetCommits(),
		"page":    "servers",
		"servers": servers,
	}, "layouts/main")
}
