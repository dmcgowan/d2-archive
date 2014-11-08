package main

import (
	"github.com/codegangsta/cli"
	"github.com/docker/d2/admin"
	"github.com/docker/d2/daemon"
)

var daemonCommand = cli.Command{
	Name:   "daemon",
	Usage:  "run the docker daemon",
	Action: daemonAction,
}

func daemonAction(context *cli.Context) {
	a := admin.New(daemon.New(logger), logger)
	if err := a.Listen("admin.chan"); err != nil {
		logger.Fatal(err)
	}
}
