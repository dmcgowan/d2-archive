package main

import (
	"os"
	"os/signal"
	"syscall"

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
	handleSignals(a)
	if err := a.Listen("admin.chan"); err != nil {
		logger.Fatal(err)
	}
}

func handleSignals(a *admin.Admin) {
	s := make(chan os.Signal, 64)
	signal.Notify(s, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for _ = range s {
			a.Close()
		}
	}()
}
