package admin

import (
	"github.com/docker/libchan"
	"github.com/docker/libchan/rpc"
)

func emit(s libchan.Sender, stream string, msg string) error {
	return s.Send(&rpc.Event{
		Stream: stream,
		Msg:    msg,
	})
}
