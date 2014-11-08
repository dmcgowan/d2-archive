package daemon

import (
	"io"
	"os/exec"

	"github.com/docker/libchan"
)

type Plugin interface {
	Name() string
	Process(args []string, sender libchan.Sender) error
}

func NewExecPlugin() Plugin {
	return &ExecPlugin{}
}

type ExecPlugin struct {
}

func (e *ExecPlugin) Name() string {
	return "exec"
}

func (e *ExecPlugin) Process(args []string, sender libchan.Sender) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = newChunkedWriter("stdout", sender)
	cmd.Stderr = newChunkedWriter("stderr", sender)
	return cmd.Run()
}

func newChunkedWriter(stream string, sender libchan.Sender) io.Writer {
	return &chunkedWriter{
		stream: stream,
		sender: sender,
	}
}

type chunkedWriter struct {
	stream string
	sender libchan.Sender
}

func (c *chunkedWriter) Write(p []byte) (int, error) {
	return len(p), c.sender.Send(struct {
		Stream string
		Msg    string
	}{
		Stream: c.stream,
		Msg:    string(p),
	})
}
