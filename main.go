package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var (
	logger = logrus.New()

	globalFlags = []cli.Flag{
		cli.BoolFlag{Name: "debug", Usage: "enable debug output for the logs"},
	}
)

func preload(context *cli.Context) error {
	if context.GlobalBool("debug") {
		logger.Level = logrus.DebugLevel
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "docker"
	app.Usage = "docker is an app to manage distriubuted systems"
	app.Version = "2.0"
	app.Email = "core@docker.com"
	app.Before = preload
	app.Flags = globalFlags
	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}
