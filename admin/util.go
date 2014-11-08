package admin

import "github.com/docker/libchan"

func emit(s libchan.Sender, stream string, msg string) error {
	return s.Send(event{
		Stream: stream,
		Msg:    msg,
	})
}
