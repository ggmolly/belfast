package routes

import (
	"fmt"
	"strings"

	"github.com/bettercallmolly/belfast/logger"
	"github.com/bettercallmolly/belfast/misc"
	"github.com/bettercallmolly/belfast/orm"
	"github.com/bettercallmolly/belfast/web/dto"
	"github.com/bettercallmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

// Returns a <select> with all frames of a given packet id
func RenderFrames(c *fiber.Ctx) error {
	frameId := c.Locals("dtoFrameId").(dto.DtoFrameId).Id
	var frames []orm.Debug
	err := orm.GormDB.Where("packet_id = ?", frameId).Order("logged_at ASC").Select("frame_id, logged_at, packet_id").Find(&frames).Error
	if err != nil {
		logger.LogEvent("Database", "Servers", err.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, "A database error occurred, check the logs.", c)
	}
	return c.Render("components/frame_selector", fiber.Map{
		"frames": frames,
	})
}

// Returns a daisyUI code mockup, with JSON data for the given frame ID
func DissectFrame(c *fiber.Ctx) error {
	frameId := c.Locals("dtoFrameId").(dto.DtoFrameId).Id
	var frame orm.Debug
	err := orm.GormDB.Preload("DebugName").First(&frame, "frame_id = ?", frameId).Error
	if err != nil {
		logger.LogEvent("Database", "Servers", err.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, "A database error occurred, check the logs.", c)
	}

	// Unmarshal the packet
	packet, err := misc.ProtoToJson(frame.PacketID, &frame.Data)
	if err != nil {
		logger.LogEvent("Debug", "Protobuf", err.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, fmt.Sprintf("Failed to unmarshal packet: %v", err.Error()), c)
	}

	return c.Render("components/frame_dissection", fiber.Map{
		"lines":    strings.Split(string(packet), "\n"),
		"filename": fmt.Sprintf("%s_%d.json", frame.DebugName.Name, frame.FrameID),
		"size":     len(string(packet)),
		"name":     frame.DebugName.Name,
		"fields":   misc.GetPacketFields(frame.PacketID),
	})
}

func RenderDebugTemplate(c *fiber.Ctx) error {
	var packets []orm.Debug
	var totalSize int
	if err := orm.GormDB.
		Preload("DebugName").
		Distinct("PacketID").
		Order("debugs.packet_id ASC").
		Find(&packets).
		Error; err != nil {
		logger.LogEvent("Database", "Debug", err.Error(), logger.LOG_LEVEL_ERROR)
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), c)
	}
	orm.GormDB.Model(&orm.Debug{}).Select("SUM(LENGTH(data))").Row().Scan(&totalSize)
	if c.Locals("no_layout") != nil {
		return c.Render("debug", fiber.Map{
			"title":     "Belfast - Debug",
			"page":      "debug",
			"packets":   packets,
			"totalSize": totalSize,
		})
	}
	return c.Render("debug", fiber.Map{
		"title":     "Belfast - Debug",
		"page":      "debug",
		"packets":   packets,
		"totalSize": totalSize,
	}, "layouts/main")
}

func DeleteFrames(c *fiber.Ctx) error {
	if err := orm.GormDB.Where("1 = 1").Delete(&orm.Debug{}).Error; err != nil {
		c.Status(500)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), c)
	}
	c.Locals("no_layout", true)
	return RenderDebugTemplate(c)
}
