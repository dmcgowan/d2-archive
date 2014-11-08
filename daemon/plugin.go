package daemon

import "github.com/docker/libchan"

type Plugin interface {
	Name() string
	Process(args []string, sender libchan.Sender) error
}
