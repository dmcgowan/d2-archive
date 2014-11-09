package plugin

import "github.com/docker/d2/core"

type CreatePlugin interface {
	Create(*core.Image, *core.UserConfig) (*core.Container, error)
}

type StartPlugin interface {
	Start(*core.Container) (int, error)
}
