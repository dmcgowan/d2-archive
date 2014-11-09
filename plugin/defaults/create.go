package defaults

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/docker/d2/core"
	"github.com/docker/d2/plugin"
	"github.com/docker/docker/daemon/execdriver/native/template"
	"github.com/docker/libcontainer/namespaces"
)

func init() {
	plugin.Add("defaults", &DefaultPlugin{})
}

type DefaultPlugin struct {
	currentID int
}

func (d *DefaultPlugin) Load(context plugin.LoadContext) error {
	if err := context.Register("create", d); err != nil {
		return err
	}
	return context.Register("start", d)
}

func (d *DefaultPlugin) Create(image *core.Image, config *core.UserConfig) (*core.Container, error) {
	d.currentID++
	container := &core.Container{
		ID:    fmt.Sprint(d.currentID),
		Image: image,
		Args:  config.Args,
	}
	return container, nil
}

func (d *DefaultPlugin) Start(container *core.Container, stdout, stderr io.Writer) (int, error) {
	config := template.New()
	config.RootFs = filepath.Join("/tmp/d2/images", container.Image.ID)
	config.Hostname = "testing"
	return namespaces.Exec(config, nil, stdout, stderr, "", "", container.Args, namespaces.DefaultCreateCommand, nil)
}
