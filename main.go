package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	log "github.com/sirupsen/logrus"

	fkt "github.com/clingclangclick/fkt/fkt"
	utils "github.com/clingclangclick/fkt/utils"
)

var CLI struct {
	ConfigFile    string `type:"existingfile" short:"f" help:"YAML configuration file" env:"CONFIG_FILE"`
	BaseDirectory string `type:"existingdirectory" short:"b" help:"Sources and overlays base directory" env:"BASE_DIRECTORY" default:"${base_directory}"`
	DryRun        bool   `short:"d" help:"Validate and return error if changes are needed" env:"DRY_RUN" default:"false"`
	Logging       struct {
		Level  string `enum:"default,none,trace,debug,info,warn,error" short:"l" help:"Log level" env:"LOG_LEVEL" default:"${logging_level}"`
		File   string `type:"path" short:"o" help:"Log file" env:"LOG_FILE"`
		Format string `enum:"default,console,json" short:"t" help:"Log format" env:"LOG_FORMAT" default:"${logging_format}"`
	} `embed:"" prefix:"logging."`
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	ctx := kong.Parse(&CLI,
		kong.Name("fkt"),
		kong.Description("FluxCD Kind of Templater."),
		kong.Vars{
			"base_directory": utils.RelWD(cwd),
			"logging_level":  "default",
			"logging_format": "default",
		},
	)

	_, err = os.Stat(CLI.BaseDirectory)
	if os.IsNotExist(err) {
		fmt.Println("Base Directory does not exist.")
		ctx.Exit(1)
	}

	if CLI.ConfigFile == "" {
		fmt.Print("Configuration file not supplied.")
		_ = ctx.PrintUsage(true)
		ctx.Exit(1)
	}
	_, err = os.Stat(CLI.ConfigFile)
	if os.IsNotExist(err) {
		fmt.Println("Configuration file does not exist.")
		ctx.Exit(1)
	}

	config, err := fkt.LoadConfig(CLI.ConfigFile)
	if err != nil {
		log.Error("Error loading config file: ", CLI.ConfigFile, " (", err, ")")
		ctx.Exit(1)
	}

	err = config.Settings.Defaults(CLI.BaseDirectory, CLI.DryRun, fkt.LogConfig{
		Level:  fkt.LogLevel(CLI.Logging.Level),
		Format: fkt.LogFormat(CLI.Logging.Format),
		File:   CLI.Logging.File,
	})
	if err != nil {
		log.Error("Error setting configuration; ", err)
		ctx.Exit(1)
	}

	log.Debug("Configuration file: ", CLI.ConfigFile)

	settings := config.Settings

	err = settings.Validate()
	if err != nil {
		log.Error("Error validating settings: ", err)
		ctx.Exit(1)
	}

	err = config.Validate(settings)
	if err != nil {
		log.Error("Error validating configuration: ", err)
		ctx.Exit(1)
	}

	err = config.Process(settings)
	if err != nil {
		log.Error("Error processing configuration: ", err)
		ctx.Exit(1)
	}

	ctx.Exit(0)
}
