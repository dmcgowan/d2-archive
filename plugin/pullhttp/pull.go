package pullhttp

import (
	"fmt"
	"io"
	"net/http"

	"github.com/docker/d2/plugin"
)

func init() {
	plugin.Add("http_pull", &HttpPullPlugin{})
}

type HttpPullPlugin struct {
	registry string
}

func (h *HttpPullPlugin) Load(context plugin.LoadContext) error {
	h.registry = context.Args()[0]
	return context.Register("pull_layer", h)
}

func (h *HttpPullPlugin) PullLayer(id string) (io.ReadCloser, error) {
	meta, err := http.Get(fmt.Sprintf("%s/images/%s/blob", h.registry, id))
	if err != nil {
		return nil, err
	}
	return meta.Body, nil
}
