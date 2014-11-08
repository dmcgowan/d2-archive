package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/docker/d2/admin"
	"github.com/docker/d2/daemon"
)

var (
	logger = logrus.New()
)

func main() {
	a := admin.New(daemon.New(logger), logger)
	if err := a.Listen("admin.chan"); err != nil {
		logger.Fatal(err)
	}
}
