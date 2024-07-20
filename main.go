package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/akamensky/argparse"
	"github.com/ggmolly/belfast/answer"
	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/consts"
	"github.com/ggmolly/belfast/debug"
	"github.com/ggmolly/belfast/logger"
	"github.com/ggmolly/belfast/misc"
	"github.com/ggmolly/belfast/orm"
	"github.com/ggmolly/belfast/packets"
	"github.com/ggmolly/belfast/protobuf"
	"github.com/ggmolly/belfast/web"
	"github.com/joho/godotenv"
	"github.com/mattn/go-tty"
	"google.golang.org/protobuf/proto"
)

var validRegions = map[string]interface{}{
	"CN": nil,
	"EN": nil,
	"JP": nil,
	"KR": nil,
	"TW": nil,
}

func main() {
	parser := argparse.NewParser("belfast", "Azur Lane server emulator")
	reseed := parser.Flag("s", "reseed", &argparse.Options{
		Required: false,
		Help:     "Forces the reseed of the database with the latest data",
		Default:  false,
	})
	adb := parser.Flag("a", "adb", &argparse.Options{
		Required: false,
		Help:     "Parse ADB logs for debugging purposes (experimental -- tested on Linux only)",
		Default:  false,
	})
	flushLogcat := parser.Flag("f", "flush-logcat", &argparse.Options{
		Required: false,
		Help:     "Flush the logcat buffer upon starting the ADB watcher",
		Default:  false,
	})
	if err := parser.Parse(os.Args); err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}
	if *reseed {
		logger.LogEvent("Reseed", "Forced", "Forcing reseed of the database...", logger.LOG_LEVEL_INFO)
		misc.UpdateAllData(os.Getenv("AL_REGION"))
	}
	server := connection.NewServer("0.0.0.0", 80, packets.Dispatch)

	// Open TTY for adb controls
	tty, err := tty.Open()
	if err != nil {
		log.Println("failed to open tty:", err)
		log.Println("adb background routine will be disabled.")
		return
	}
	// wait for SIGINT
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt)
	go func() {
		<-sigChannel
		fmt.Printf("\r") // trick to avoid ^C in the terminal, could use low-level RawMode() but why bother
		// disconnect all clients from the server
		server.DisconnectAll(consts.DR_CONNECTION_TO_SERVER_LOST)
		os.Exit(0)
	}()
	// Prepare web server
	go func() {
		web.StartWeb()
	}()

	// Prepare adb background task
	if *adb {
		go debug.ADBRoutine(tty, *flushLogcat)
	}
	if err := server.Run(); err != nil {
		logger.LogEvent("Server", "Run", fmt.Sprintf("%v", err), logger.LOG_LEVEL_ERROR)
		tty.Close()
		os.Exit(1)
	}
}

func init() {
	// Set log format to have the file name and line number
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	err := godotenv.Load()
	if err != nil {
		logger.LogEvent("Environment", "Load", err.Error(), logger.LOG_LEVEL_ERROR)
	}
	// Check if the region is valid
	if _, ok := validRegions[os.Getenv("AL_REGION")]; !ok {
		logger.LogEvent("Environment", "Invalid", fmt.Sprintf("AL_REGION is not a valid region ('%s' was supplied)", os.Getenv("AL_REGION")), logger.LOG_LEVEL_ERROR)
		os.Exit(1)
	}
	if orm.InitDatabase() { // if first run, populate the database
		misc.UpdateAllData(os.Getenv("AL_REGION"))
	}
	packets.RegisterPacketHandler(10800, []packets.PacketHandler{answer.Forge_SC10801})
	packets.RegisterPacketHandler(8239, []packets.PacketHandler{answer.Forge_SC8239})
	packets.RegisterPacketHandler(10020, []packets.PacketHandler{answer.Forge_SC10021})
	packets.RegisterLocalizedPacketHandler(10802, packets.LocalizedHandler{
		CN: &[]packets.PacketHandler{answer.Forge_SC10803_CN_JP_KR_TW},
		TW: &[]packets.PacketHandler{answer.Forge_SC10803_CN_JP_KR_TW},
		JP: &[]packets.PacketHandler{answer.Forge_SC10803_CN_JP_KR_TW},
		KR: &[]packets.PacketHandler{answer.Forge_SC10803_CN_JP_KR_TW},
	})
	packets.RegisterPacketHandler(10018, []packets.PacketHandler{answer.Forge_SC10019})
	packets.RegisterPacketHandler(10022, []packets.PacketHandler{answer.JoinServer})
	packets.RegisterPacketHandler(10026, []packets.PacketHandler{answer.PlayerExist})
	packets.RegisterPacketHandler(11001, []packets.PacketHandler{
		answer.LastLogin,                 // SC_11000
		answer.PlayerInfo,                // SC_11003
		answer.PlayerBuffs,               // SC_11015
		answer.GetMetaProgress,           // SC_63315
		answer.LastOnlineInfo,            // SC_11752
		answer.ResourcesInfo,             // SC_22001
		answer.EventData,                 // SC_26120
		answer.Meowfficers,               // SC_25001
		answer.CommanderCollection,       // SC_17001
		answer.OngoingBuilds,             // SC_12024
		answer.PlayerDock,                // SC_12001
		answer.CommanderDock,             // SC_12010
		answer.CommanderFleet,            // SC_12101
		answer.CommanderOwnedSkins,       // SC_12201
		answer.UNK_63000,                 // SC_63000
		answer.ShipyardData,              // SC_63100
		answer.TechnologyNationProxy,     // SC_64000
		answer.CommanderStoryProgress,    // SC_13001
		answer.UNK_13002,                 // SC_13002
		answer.CommanderCommissionsFleet, // SC_13201
		answer.ShopData,                  // SC_16200
		answer.UNK_33114,                 // SC_33114
		answer.EquipedSpecialWeapons,     // SC_14001
		answer.EquippedWeaponSkin,        // SC_14101
		answer.OwnedItems,                // SC_15001
		answer.CommanderMissions,         // SC_20001
		answer.WeeklyMissions,            // SC_20101
		answer.DormData,                  // SC_19001
		answer.FleetEnergyRecoverTime,    // SC_12031
		answer.GameMailbox,               // SC_30001
		answer.CommanderFriendList,       // SC_50000
		// answer.JuustagramData,            // SC_11700
		answer.Activities,          // SC_11200
		answer.PermanentActivites,  // SC_11210
		answer.GameNotices,         // SC_11300
		answer.SendPlayerShipCount, // SC_11002 -> Will trigger a scene change in the client
	})
	packets.RegisterPacketHandler(25026, []packets.PacketHandler{answer.UNK_25027})
	packets.RegisterPacketHandler(34501, []packets.PacketHandler{answer.UNK_34502})
	packets.RegisterPacketHandler(63317, []packets.PacketHandler{answer.MetaCharacterTacticsInfoRequestCommandResponse})
	packets.RegisterPacketHandler(34001, []packets.PacketHandler{answer.GetMetaShipsPointsResponse})
	packets.RegisterPacketHandler(18001, []packets.PacketHandler{answer.ExerciseEnemies})
	packets.RegisterPacketHandler(60037, []packets.PacketHandler{answer.CommanderGuildData})
	packets.RegisterPacketHandler(62100, []packets.PacketHandler{answer.CommanderGuildTechnologies})
	packets.RegisterPacketHandler(26101, []packets.PacketHandler{answer.UNK_26102})
	packets.RegisterPacketHandler(24020, []packets.PacketHandler{answer.UNK_24021})
	packets.RegisterPacketHandler(11603, []packets.PacketHandler{answer.FetchSecondaryPasswordCommandResponse})
	packets.RegisterPacketHandler(17203, []packets.PacketHandler{answer.UNK_17204})
	packets.RegisterPacketHandler(16104, []packets.PacketHandler{answer.UNK_16105})
	packets.RegisterPacketHandler(60100, []packets.PacketHandler{answer.CommanderGuildChat})
	packets.RegisterPacketHandler(60102, []packets.PacketHandler{answer.GuildGetUserInfoCommand})
	packets.RegisterPacketHandler(61009, []packets.PacketHandler{answer.GetMyAssaultFleetCommandResponse})
	packets.RegisterPacketHandler(61011, []packets.PacketHandler{answer.GuildGetAssaultFleetCommandResponse})
	packets.RegisterPacketHandler(61005, []packets.PacketHandler{answer.GuildGetActivationEventCommandResponse})
	packets.RegisterPacketHandler(60003, []packets.PacketHandler{answer.GetGuildRequestsCommandResponse})
	packets.RegisterPacketHandler(13505, []packets.PacketHandler{answer.UNK_13506})
	packets.RegisterPacketHandler(11202, []packets.PacketHandler{answer.GiveItem})
	packets.RegisterPacketHandler(11751, []packets.PacketHandler{answer.LastOnlineInfo})
	packets.RegisterPacketHandler(10100, []packets.PacketHandler{answer.SendHeartbeat})
	packets.RegisterPacketHandler(11013, []packets.PacketHandler{answer.GiveResources})
	packets.RegisterPacketHandler(33000, []packets.PacketHandler{answer.UNK_33001})
	packets.RegisterPacketHandler(10994, []packets.PacketHandler{answer.CheaterMark})

	// Build
	packets.RegisterPacketHandler(12002, []packets.PacketHandler{answer.ShipBuild})
	packets.RegisterPacketHandler(12008, []packets.PacketHandler{answer.BuildQuickFinish})
	packets.RegisterPacketHandler(12043, []packets.PacketHandler{answer.BuildFinish})
	packets.RegisterPacketHandler(12025, []packets.PacketHandler{answer.UNK_12026})
	packets.RegisterPacketHandler(12045, []packets.PacketHandler{answer.ConfirmShip})

	// Exchange ships
	packets.RegisterPacketHandler(12047, []packets.PacketHandler{answer.ExchangeShip})

	// Mails
	packets.RegisterPacketHandler(30002, []packets.PacketHandler{answer.SendMailList})
	packets.RegisterPacketHandler(30004, []packets.PacketHandler{answer.GetCollectionMailList})
	packets.RegisterPacketHandler(30006, []packets.PacketHandler{answer.HandleMailDealCmd})
	packets.RegisterPacketHandler(30008, []packets.PacketHandler{answer.DeleteArchivedMail})
	// packets.RegisterPacketHandler(30010, []packets.PacketHandler{answer.UpdateMailImpFlag})

	// Shop
	packets.RegisterPacketHandler(16001, []packets.PacketHandler{answer.ShoppingCommandAnswer})
	packets.RegisterPacketHandler(11501, []packets.PacketHandler{answer.ChargeCommandAnswer})

	// Retire
	packets.RegisterPacketHandler(12004, []packets.PacketHandler{answer.RetireShip})

	// Chat
	packets.RegisterPacketHandler(11401, []packets.PacketHandler{answer.ChatRoomChange})
	packets.RegisterPacketHandler(50102, []packets.PacketHandler{answer.ReceiveChatMessage})

	// Propose
	packets.RegisterPacketHandler(12032, []packets.PacketHandler{answer.ProposeShip})

	// Ship interaction quest
	packets.RegisterPacketHandler(20007, []packets.PacketHandler{func(b *[]byte, c *connection.Client) (int, int, error) {
		response := protobuf.SC_20008{
			Result: proto.Uint32(1),
		}
		return c.SendMessage(20008, &response)
	}})

	// Update secretaries
	packets.RegisterPacketHandler(11011, []packets.PacketHandler{answer.UpdateSecretaries})

	// Set ship as favorite
	packets.RegisterPacketHandler(12040, []packets.PacketHandler{answer.SetFavoriteShip})

	// Lock ships
	packets.RegisterPacketHandler(12022, []packets.PacketHandler{answer.ChangeShipLockState})

	// Change selected skin
	packets.RegisterPacketHandler(12202, []packets.PacketHandler{answer.ChangeSelectedSkin})

	// Rename proposed ship
	packets.RegisterPacketHandler(12034, []packets.PacketHandler{answer.RenameProposedShip})

	// Education / Child (aka. TB as secretary)
	packets.RegisterPacketHandler(27000, []packets.PacketHandler{answer.UNK_27001})
	packets.RegisterPacketHandler(27010, []packets.PacketHandler{func(b *[]byte, c *connection.Client) (int, int, error) {
		response := protobuf.SC_27011{}
		return c.SendMessage(27011, &response)
	}})

	// Fleet
	packets.RegisterPacketHandler(12102, []packets.PacketHandler{answer.FleetCommit})
	packets.RegisterPacketHandler(12104, []packets.PacketHandler{answer.FleetRename})

	packets.RegisterPacketHandler(13107, []packets.PacketHandler{func(b *[]byte, c *connection.Client) (int, int, error) {
		response := protobuf.SC_13108{
			Result: proto.Uint32(0),
		}
		return c.SendMessage(13108, &response)
	}})

	// UpdateCommonFlagCommand, unknown what it does
	packets.RegisterPacketHandler(11019, []packets.PacketHandler{answer.UpdateCommonFlagCommand})

	// Ship comments tab
	packets.RegisterPacketHandler(17101, []packets.PacketHandler{answer.GetShipDiscuss}) // Ship discussion (placeholder)
	packets.RegisterPacketHandler(17107, []packets.PacketHandler{answer.UpdateShipLike})

	// ???
	packets.RegisterPacketHandler(15300, []packets.PacketHandler{func(b *[]byte, c *connection.Client) (int, int, error) {
		return 0, 0, nil
	}})

	// ???
	packets.RegisterPacketHandler(12299, []packets.PacketHandler{func(b *[]byte, c *connection.Client) (int, int, error) {
		return 0, 0, nil
	}})

	// track
	packets.RegisterPacketHandler(10993, []packets.PacketHandler{func(b *[]byte, c *connection.Client) (int, int, error) {
		return 0, 0, nil
	}})
}
