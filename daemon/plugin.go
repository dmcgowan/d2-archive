package daemon

import (
	"fmt"
	"io"
	"net"
	"os/exec"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/networkdriver/portallocator"
	"github.com/docker/docker/daemon/networkdriver/portmapper"
	"github.com/docker/libchan"
	"github.com/docker/libchan/rpc"
)

type Plugin interface {
	Name() string
	GetPlug() rpc.Plug
}

func NewNetPlugin() Plugin {
	return &NetPlugin{}
}

type NetPlugin struct {
	//currentInterfaces
}

func (p *NetPlugin) Name() string {
	return "net"
}

func emit(s libchan.Sender, stream string, msg string) error {
	return s.Send(&rpc.Event{
		Stream: stream,
		Msg:    msg,
	})
}

func (p *NetPlugin) GetPlug() rpc.Plug {
	return func(receiver libchan.Receiver, sender libchan.Sender) error {
		c := &rpc.Cmd{
			Op:   "register",
			Args: []string{"allocip", "releaseip", "portmap", "linkcontainers"},
		}
		if err := sender.Send(c); err != nil {
			return err
		}

		for {
			var cmd rpc.Cmd
			if err := receiver.Receive(&cmd); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			switch cmd.Op {
			case "allocip":
			case "relaseip":
			case "portmap":
				if err := CmdAllocatePort(&cmd); err != nil {
					if err := emit(cmd.Out, "error", err.Error()); err != nil {
						log.WithField("error", err).Errorf("Error sending error event")
					}
					continue
				}
			case "linkcontainers":
			}

		}
		return nil
	}

}

func extractString(iv interface{}) string {
	if iv == nil {
		return ""
	}
	switch v := iv.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	return ""
}

func extractInt(iv interface{}) int {
	if iv == nil {
		return 0
	}
	switch v := iv.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case string:
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return 0
		}
		return int(i)
	}
	return 0
}

func CmdAllocatePort(cmd *rpc.Cmd) error {
	log.Debugf("KeyValues\n%#v", cmd.KV)
	var (
		err error

		ip = net.ParseIP("0.0.0.0")
		//containerIP   = extractString(cmd.Args[0])
		//hostIP        = extractString(cmd.Args[1])
		//hostPort      = extractInt(cmd.Args[2])
		//containerPort = extractInt(cmd.Args[3])
		//proto         = extractString(cmd.Args[4])
		containerIP   = extractString(cmd.KV["ContainerIP"])
		hostIP        = extractString(cmd.KV["HostIP"])
		hostPort      = extractInt(cmd.KV["HostPort"])
		containerPort = extractInt(cmd.KV["ContainerPort"])
		proto         = extractString(cmd.KV["Proto"])
	)

	if hostIP != "" {
		ip = net.ParseIP(hostIP)
		if ip == nil {
			return fmt.Errorf("bad parameter: invalid host ip %s", hostIP)
		}
	}
	var cIP net.IP
	if containerIP != "" {
		cIP = net.ParseIP(containerIP)
		if cIP == nil {
			return fmt.Errorf("bad parameter: invalid host ip %s", containerIP)
		}
	}

	// host ip, proto, and host port
	var container net.Addr
	switch proto {
	case "tcp":
		container = &net.TCPAddr{IP: cIP, Port: containerPort}
	case "udp":
		container = &net.UDPAddr{IP: cIP, Port: containerPort}
	default:
		return fmt.Errorf("unsupported address type %s", proto)
	}

	//
	// Try up to 10 times to get a port that's not already allocated.
	//
	// In the event of failure to bind, return the error that portmapper.Map
	// yields.
	//

	var host net.Addr
	for i := 0; i < 3; i++ {
		if host, err = portmapper.Map(container, ip, hostPort); err == nil {
			break
		}

		if _, ok := err.(portallocator.ErrPortAlreadyAllocated); ok {
			// There is no point in immediately retrying to map an explicitly
			// chosen port.
			if hostPort != 0 {
				//job.Logf("Failed to bind %s for container address %s: %s", allocerr.IPPort(), container.String(), allocerr.Error())
				break
			}

			// Automatically chosen 'free' port failed to bind: move on the next.
			//job.Logf("Failed to bind %s for container address %s. Trying another port.", allocerr.IPPort(), container.String())
		} else {
			// some other error during mapping
			//job.Logf("Received an unexpected error during port allocation: %s", err.Error())
			break
		}
	}

	if err != nil {
		return err
	}

	//network.PortMappings = append(network.PortMappings, host)

	e := &rpc.Event{
		KV: map[string]interface{}{},
	}
	switch netAddr := host.(type) {
	case *net.TCPAddr:
		e.KV["HostIP"] = netAddr.IP.String()
		e.KV["HostPort"] = netAddr.Port
	case *net.UDPAddr:
		e.KV["HostIP"] = netAddr.IP.String()
		e.KV["HostPort"] = netAddr.Port
	}

	cmd.Out.Close()

	return nil

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
