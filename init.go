package main

import (
	"encoding/json"
	"os"
	"runtime"

	"github.com/codegangsta/cli"
	"github.com/docker/libcontainer"
	"github.com/docker/libcontainer/namespaces"
)

var initCommand = cli.Command{
	Name:   "init",
	Usage:  "INTERNAL USE ONLY",
	Action: initAction,
}

func initAction(context *cli.Context) {
	runtime.LockOSThread()
	config, err := loadConfig()
	if err != nil {
		logger.Fatal(err)
	}
	rootfs, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}
	pipe := os.NewFile(3, "pipe")
	if err := namespaces.Init(config, rootfs, "", pipe, []string(context.Args())); err != nil {
		logger.Fatal(err)
	}
}

func loadConfig() (*libcontainer.Config, error) {
	f, err := os.Open("container.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config *libcontainer.Config
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}
