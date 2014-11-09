package admin

import "github.com/docker/libchan"

type command struct {
	Target string // FIXME: dont ever use this!!!
	Op     string
	Args   []string
	KV     map[string]interface{}
	Out    libchan.Sender
}

type event struct {
	Stream string
	Msg    string
	KV     map[string]string
}
