package web

import (
	"time"

	"github.com/ggmolly/belfast/logger"
	"github.com/ggmolly/belfast/misc"
	"github.com/ggmolly/belfast/web/dto"
	"github.com/ggmolly/belfast/web/routes"
	"github.com/ggmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func StartWeb() {
	engine := html.New("./web/views", ".html")
	{
		engine.AddFunc("TabColor", utils.TabColor)
		engine.AddFunc("TimeSpan", utils.TimeSpan)
		engine.AddFunc("TimeFormat", utils.TimeFormat)
		engine.AddFunc("TimeLeft", utils.TimeLeft)
		engine.AddFunc("SecondsLeft", utils.SecondsLeft)
		engine.AddFunc("HumanReadableSize", utils.HumanReadableSize)
		engine.AddFunc("AccountStatusBadge", utils.AccountStatusBadge)
		engine.AddFunc("GetGitHash", misc.GetGitHash)
		engine.AddFunc("TrimString", utils.TrimString)
		engine.AddFunc("RepeatString", utils.RepeatString)
		engine.AddFunc("ISOTimestamp", utils.ISOTimestamp)
		engine.Reload(true)
	}

	app := fiber.New(fiber.Config{
		ServerHeader:          "belfast/1.0",
		ReadTimeout:           time.Second * 10,
		WriteTimeout:          time.Second * 10,
		AppName:               "belfast",
		Views:                 engine,
		DisableStartupMessage: true,
	})

	// Pages
	{
		app.Static("/static", "./web/static")
		app.Get("/", routes.Index)
		app.Get("/servers", routes.RenderServerTemplate)
		app.Get("/debug", routes.RenderDebugTemplate)
		app.Get("/players", routes.RenderPlayersTemplate)
		app.Get("/dock", routes.RenderDockTemplate)
		app.Get("/depot", routes.RenderDepotTemplate)
		app.Get("/resources", routes.RenderResourcesTemplate)
		app.Get("/builds", routes.RenderBuildsTemplate)
	}
	api := app.Group("/api/v1")
	serverGroup := api.Group("/servers")
	// Server Group
	{
		// Returns all servers
		serverGroup.Get(
			"/",
			routes.RenderServers,
		)
		// Return only the server with the given ID
		serverGroup.Get("/:server_id",
			dto.ServerIdMiddleware,
			routes.RenderServer,
		)
		// Return an editable server with the given ID
		serverGroup.Get("/edit/:server_id",
			dto.ServerIdMiddleware,
			routes.EditServer,
		)
		// Delete a server
		serverGroup.Delete("/:server_id",
			dto.ServerIdMiddleware,
			routes.DeleteServer,
		)
		// Announce a new server
		serverGroup.Post("/",
			dto.ServerDataMiddleware,
			routes.AnnounceServer,
		)
		// Update a server
		serverGroup.Patch("/:server_id",
			dto.ServerDataMiddleware,
			dto.ServerIdMiddleware,
			routes.UpdateServer,
		)
	}
	framesGroup := api.Group("/frames")
	{
		// Return all frames of a certain type
		framesGroup.Get("/",
			dto.FrameIdMiddleware,
			routes.RenderFrames,
		)
		// Delete all frames
		framesGroup.Delete("/",
			routes.DeleteFrames,
		)
		// Return HTML code mockup for JSON data
		framesGroup.Get("/dissect",
			dto.FrameIdMiddleware,
			routes.DissectFrame,
		)
	}
	playersGroup := api.Group("/players")
	{
		// Return an editable server with the given ID
		playersGroup.Get("/edit/:player_id",
			dto.PlayerIdMiddleware,
			routes.EditPlayer,
		)

		// Return a player
		playersGroup.Get("/:player_id",
			dto.PlayerIdMiddleware,
			routes.RenderPlayer,
		)

		// Edit a player
		playersGroup.Patch("/:player_id",
			dto.PlayerIdMiddleware,
			dto.QuickPlayerEditMiddleware,
			routes.UpdatePlayer,
		)
	}
	dockGroup := api.Group("/dock")
	{
		// Return the modal for editing a ship
		dockGroup.Get("/edit/:ship_id",
			dto.CommanderIdMiddleware,
			dto.ShipIdMiddleware,
			routes.EditShipModal,
		)

		// Receive modifications
		dockGroup.Patch("/edit/:ship_id",
			dto.CommanderIdMiddleware,
			dto.ShipIdMiddleware,
			dto.ShipEditMiddleware,
			routes.UpdateShip,
		)

		// New ship modal
		dockGroup.Get("/new",
			dto.CommanderIdMiddleware,
			routes.NewShipModal,
		)

		// Create new ship
		dockGroup.Post("/new",
			func(c *fiber.Ctx) error {
				c.Locals("dtoShipId", dto.DtoShipId{Id: -1}) // -1 is a placeholder for the new ship
				return c.Next()
			},
			dto.CommanderIdMiddleware,
			dto.ShipEditMiddleware,
			routes.UpdateShip,
		)

		// Get skin for template
		dockGroup.Get("/skins",
			dto.TemplateIdMiddleware,
			routes.GetSkins,
		)

		// Return dock content
		dockGroup.Get("/",
			dto.CommanderIdMiddleware,
			routes.RenderDockContent,
		)
	}
	depotGroup := api.Group("/depot")
	{
		// Return dock content
		depotGroup.Get("/",
			dto.CommanderIdMiddleware,
			routes.RenderDepotContent,
		)

		// Return the modal for editing an item
		depotGroup.Get("/edit/:item_id",
			dto.CommanderIdMiddleware,
			dto.ItemIdMiddleware,
			routes.EditItemModal,
		)

		// Receive modifications
		depotGroup.Patch("/edit/:item_id",
			dto.CommanderIdMiddleware,
			dto.ItemIdMiddleware,
			dto.ItemEditMiddleware,
			routes.UpdateItem,
		)

		// New item modal
		depotGroup.Get("/new",
			dto.CommanderIdMiddleware,
			routes.NewItemModal,
		)

		// Create new item
		depotGroup.Post("/new",
			dto.CommanderIdMiddleware,
			dto.ItemEditMiddleware,
			func(c *fiber.Ctx) error {
				c.Locals("dtoItemId", dto.DtoItemId{Id: -1}) // -1 is a placeholder for the new item
				return c.Next()
			},
			routes.UpdateItem,
		)
	}
	resourcesGroup := api.Group("/resources")
	{
		// Return dock content
		resourcesGroup.Get("/",
			dto.CommanderIdMiddleware,
			routes.RenderResourcesContent,
		)

		// Return the modal for editing a resource
		resourcesGroup.Get("/edit/:resource_id",
			dto.CommanderIdMiddleware,
			dto.ResourcesIdMiddleware,
			routes.ResourcesModal,
		)

		// Receive modifications
		resourcesGroup.Patch("/edit/:resource_id",
			dto.CommanderIdMiddleware,
			dto.ResourcesIdMiddleware,
			dto.ResourcesEditMiddleware,
			routes.UpdateResource,
		)

		// New resources modal
		resourcesGroup.Get("/new",
			dto.CommanderIdMiddleware,
			routes.NewResourcesModal,
		)

		// Create new resources
		resourcesGroup.Post("/new",
			dto.CommanderIdMiddleware,
			dto.ResourcesEditMiddleware,
			func(c *fiber.Ctx) error {
				c.Locals("dtoResourceId", dto.DtoResourceId{ResourceId: -1}) // -1 is a placeholder for the new item
				return c.Next()
			},
			routes.UpdateResource,
		)
	}
	buildsGroup := api.Group("/builds")
	{
		// Return build content
		buildsGroup.Get("/",
			dto.CommanderIdMiddleware,
			routes.RenderBuildsContent,
		)

		// Return the modal for editing a build
		buildsGroup.Get("/edit/:build_id",
			dto.CommanderIdMiddleware,
			dto.BuildIdMiddleware,
			routes.EditBuildModal,
		)

		// Receive modifications
		buildsGroup.Patch("/edit/:build_id",
			dto.CommanderIdMiddleware,
			dto.BuildIdMiddleware,
			dto.BuildEditMiddleware,
			routes.UpdateBuild,
		)

		// New build modal
		buildsGroup.Get("/new",
			dto.CommanderIdMiddleware,
			routes.NewBuildModal,
		)

		// Create new build
		buildsGroup.Post("/new",
			dto.CommanderIdMiddleware,
			dto.BuildEditMiddleware,
			func(c *fiber.Ctx) error {
				c.Locals("dtoBuildId", dto.DtoBuildId{Id: -1}) // -1 is a placeholder for the new item
				return c.Next()
			},
			routes.UpdateBuild,
		)
	}
	logger.LogEvent("Web", "Server", "Starting web server", logger.LOG_LEVEL_INFO)
	if err := app.Listen("127.0.0.1:8000"); err == nil {
		logger.LogEvent("Web", "Server", "Listening on 127.0.0.1:8000", logger.LOG_LEVEL_INFO)
	} else {
		logger.LogEvent("Web", "Server", "Failed to start web server", logger.LOG_LEVEL_ERROR)
	}
}
