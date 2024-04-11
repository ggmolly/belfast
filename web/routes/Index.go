package routes

import (
	"github.com/bettercallmolly/belfast/misc"
	"github.com/gofiber/fiber/v2"
)

func Index(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"title":       "Belfast",
		"commits":     misc.GetCommits(),
		"page":        "index",
		"versions":    misc.GetLatestVersions(),
		"hashes":      misc.GetGameHashes(),
		"lastUpdate":  misc.LastCacheUpdate(),
		"lastVersion": misc.LastCacheUpdateVersion(),
	}, "layouts/main")
}
