package daemon

import (
	"errors"

	"github.com/Sirupsen/logrus"
)

var (
	ErrPluginNotLoaded = errors.New("plugin not loaded")
	ErrNilPlugin       = errors.New("cannot load nil plugin")
)

type Daemon interface {
	ListPlugins() ([]Plugin, error)
	LoadPlugin(name string, p Plugin) error
	GetPlugin(name string) (Plugin, error)
}

func New(logger *logrus.Logger) Daemon {
	return &testDaemon{
		logger:  logger,
		plugins: make(map[string]Plugin),
	}
}

type testDaemon struct {
	logger  *logrus.Logger
	plugins map[string]Plugin
}

func (t *testDaemon) ListPlugins() ([]Plugin, error) {
	out := []Plugin{}
	for _, p := range t.plugins {
		out = append(out, p)
	}
	return out, nil
}

func (t *testDaemon) LoadPlugin(name string, p Plugin) error {
	if p == nil {
		return ErrNilPlugin
	}
	t.plugins["name"] = p
	return nil
}

func (t *testDaemon) GetPlugin(name string) (Plugin, error) {
	p, exists := t.plugins[name]
	if !exists {
		return nil, ErrPluginNotLoaded
	}
	return p, nil
}
