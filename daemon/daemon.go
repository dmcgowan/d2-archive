package daemon

import "github.com/Sirupsen/logrus"

type Daemon interface {
	ListPlugins() ([]Plugin, error)
	LoadPlugin(name string, args []string) error
	GetPlugin(name string) (Plugin, error)
}

func New(logger *logrus.Logger) Daemon {
	return nil
}
