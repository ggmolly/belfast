package entrypoint

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/akamensky/argparse"
	"github.com/ggmolly/belfast/internal/api"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/debug"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/misc"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/region"
	"github.com/mattn/go-tty"
)

type Options struct {
	CommandName   string
	Description   string
	DefaultConfig string
}

var runtimeOnce sync.Once

func Run(opts Options) {
	initRuntime()
	defaultConfig := opts.DefaultConfig
	if defaultConfig == "" {
		defaultConfig = "server.toml"
	}
	parser := argparse.NewParser(opts.CommandName, opts.Description)
	noAPI := parser.Flag("", "no-api", &argparse.Options{
		Required: false,
		Help:     "Disable the embedded REST API server",
		Default:  false,
	})
	configPath := parser.String("", "config", &argparse.Options{
		Required: false,
		Help:     "Path to TOML config file",
		Default:  defaultConfig,
	})
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
	restartGame := parser.Flag("r", "restart", &argparse.Options{
		Required: false,
		Help:     "Restart the game on ADB watcher start (requires -a)",
		Default:  false,
	})
	if err := parser.Parse(os.Args); err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}
	loadedConfig, err := config.Load(*configPath)
	if err != nil {
		logger.LogEvent("Config", "Load", err.Error(), logger.LOG_LEVEL_ERROR)
		os.Exit(1)
	}
	if err := region.SetCurrent(loadedConfig.Region.Default); err != nil {
		logger.LogEvent("Config", "Region", err.Error(), logger.LOG_LEVEL_ERROR)
		os.Exit(1)
	}
	if orm.InitDatabase() {
		misc.UpdateAllData(region.Current())
	}
	if *reseed {
		logger.LogEvent("Reseed", "Forced", "Forcing reseed of the database...", logger.LOG_LEVEL_INFO)
		misc.UpdateAllData(region.Current())
	}
	server := connection.NewServer(loadedConfig.Belfast.BindAddress, loadedConfig.Belfast.Port, packets.Dispatch)
	server.SetMaintenance(loadedConfig.Belfast.Maintenance)
	if !*noAPI {
		cfg := api.LoadConfig(loadedConfig)
		go func() {
			if err := api.Start(cfg); err != nil {
				logger.LogEvent("API", "Start", err.Error(), logger.LOG_LEVEL_ERROR)
			}
		}()
	}

	// wait for SIGINT
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt)
	go func() {
		<-sigChannel
		fmt.Printf("\r")
		server.DisconnectAll(consts.DR_CONNECTION_TO_SERVER_LOST)
		os.Exit(0)
	}()
	// Prepare adb background task
	if *adb {
		tty, err := tty.Open()
		if err != nil {
			log.Println("failed to open tty:", err)
			log.Println("adb background routine will be disabled.")
		} else {
			go debug.ADBRoutine(tty, *flushLogcat, *restartGame)
		}
	}
	if err := server.Run(); err != nil {
		logger.LogEvent("Server", "Run", fmt.Sprintf("%v", err), logger.LOG_LEVEL_ERROR)
		os.Exit(1)
	}
}

func initRuntime() {
	runtimeOnce.Do(func() {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		currentRegion := region.Current()
		if _, ok := validRegions[currentRegion]; !ok {
			logger.LogEvent("Environment", "Invalid", fmt.Sprintf("AL_REGION is not a valid region ('%s' was supplied)", currentRegion), logger.LOG_LEVEL_ERROR)
			os.Exit(1)
		}
		registerPackets()
	})
}
