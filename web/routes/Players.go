package routes

import (
	"github.com/bettercallmolly/belfast/logger"
	"github.com/bettercallmolly/belfast/misc"
	"github.com/bettercallmolly/belfast/orm"
	"github.com/bettercallmolly/belfast/web/dto"
	"github.com/bettercallmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func getCommanderById(id int) (orm.Commander, error) {
	var commander orm.Commander
	err := orm.GormDB.
		Preload("Punishments", func(db *gorm.DB) *gorm.DB {
			return db.
				Where("lift_timestamp > CURRENT_TIMESTAMP OR is_permanent = true").
				Order("is_permanent DESC, lift_timestamp DESC")
		}).
		Where("commander_id = ?", id).
		First(&commander).Error
	return commander, err
}

func EditPlayer(c *fiber.Ctx) error {
	playerEdit := c.Locals("dtoPlayerId").(dto.DtoPlayerId)

	player, err := getCommanderById(playerEdit.Id)

	if err != nil {
		logger.LogEvent("Database", "Players", err.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, "A database error occurred, check the logs.", c)
	}

	return c.Render("components/editable_player_row", fiber.Map{
		"ID":          player.CommanderID,
		"Name":        player.Name,
		"Level":       player.Level,
		"Exp":         player.Exp,
		"Punishments": player.Punishments,
	})
}

func RenderPlayer(c *fiber.Ctx) error {
	playerId := c.Locals("dtoPlayerId").(dto.DtoPlayerId)

	player, err := getCommanderById(playerId.Id)

	if err != nil {
		logger.LogEvent("Database", "Players", err.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, "A database error occurred, check the logs.", c)
	}

	return c.Render("components/player_row", fiber.Map{
		"CommanderID": player.CommanderID,
		"Name":        player.Name,
		"Level":       player.Level,
		"Exp":         player.Exp,
		"Punishments": player.Punishments,
	})
}

func UpdatePlayer(c *fiber.Ctx) error {
	playerId := c.Locals("dtoPlayerId").(dto.DtoPlayerId)
	quickPlayerEdit := c.Locals("dtoQuickPlayerEdit").(dto.DtoQuickPlayerEdit)

	var dbPlayer orm.Commander
	err := orm.GormDB.First(&dbPlayer, playerId.Id).Error
	if err != nil {
		logger.LogEvent("Database", "Players", err.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, "A database error occurred, check the logs.", c)
	}
	dbPlayer.Name = quickPlayerEdit.Name
	dbPlayer.Level = quickPlayerEdit.Level
	err = orm.GormDB.Save(&dbPlayer).Error
	if err != nil {
		logger.LogEvent("Database", "Players", err.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, "A database error occurred, check the logs.", c)
	}

	player, err := getCommanderById(playerId.Id)
	if err != nil {
		logger.LogEvent("Database", "Players", err.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, "A database error occurred, check the logs.", c)
	}
	return c.Render("components/player_row", fiber.Map{
		"CommanderID": player.CommanderID,
		"Name":        player.Name,
		"Level":       player.Level,
		"Exp":         player.Exp,
		"Punishments": player.Punishments,
	})
}

func RenderPlayersTemplate(c *fiber.Ctx) error {
	var commanders []orm.Commander
	if err := orm.GormDB.
		Preload("Punishments", func(db *gorm.DB) *gorm.DB {
			return db.
				Where("lift_timestamp > CURRENT_TIMESTAMP OR is_permanent = true").
				Order("is_permanent DESC, lift_timestamp DESC")
		}).Order("commander_id").
		Find(&commanders).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Render("players", fiber.Map{
		"title":   "Belfast - Players",
		"commits": misc.GetCommits(),
		"page":    "players",
		"players": commanders,
	}, "layouts/main")
}
