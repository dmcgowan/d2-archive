package daemon

import (
	"io"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libchan"
	"github.com/docker/libchan/rpc"
)

type Plugin interface {
	Name() string
	GetPlug() rpc.Plug
}

func NewExecPlugin() Plugin {
	return &ExecPlugin{}
}

type ExecPlugin struct {
}

func (e *ExecPlugin) Name() string {
	return "exec"
}

func (e *ExecPlugin) GetPlug() rpc.Plug {
	return func(receiver libchan.Receiver, sender libchan.Sender) error {
		cmd := &rpc.Cmd{
			Op:   "register",
			Args: []string{"start"},
		}
		if err := sender.Send(cmd); err != nil {
			return err
		}

		for {
			var cmd rpc.Cmd
			if err := receiver.Receive(&cmd); err != nil {
				return err
			}
			switch cmd.Op {
			case "start":
				if err := e.Process(cmd.Args, cmd.Out); err != nil {
					return err
				}
			default:
				log.Debugf("unhandled op: %s", cmd.Op)
			}
			if err := cmd.Out.Close(); err != nil {
				log.Errorf("Error closing send out channel")
			}
		}
	}
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
