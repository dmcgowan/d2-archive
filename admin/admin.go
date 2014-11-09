package admin

import (
	"net"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/d2/daemon"
	"github.com/docker/libchan/rpc"
	"github.com/docker/libchan/spdy"
)

var avaliablePlugins = map[string]daemon.Plugin{
	"exec": daemon.NewExecPlugin(),
	"net":  daemon.NewNetPlugin(),
}

func New(d daemon.Daemon, logger *logrus.Logger) *Admin {
	return &Admin{
		daemon: d,
		logger: logger,
		sb:     rpc.NewSwitchBoard(),
	}
}

type Admin struct {
	running  bool
	daemon   daemon.Daemon
	sb       *rpc.SwitchBoard
	listener net.Listener
	logger   *logrus.Logger
}

func (a *Admin) Listen(socketPath string) error {
	a.logger.WithField("socket", socketPath).Debug("creating chan")
	a.running = true
	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	a.listener = l
	defer func() {
		l.Close()
		os.Remove(socketPath)
	}()

	for a.running {
		conn, err := l.Accept()
		if err != nil {
			if !a.running {
				return nil
			}
			a.logger.WithField("error", err).Error("accept connection")
			continue
		}
		go a.handleConn(conn)
	}
	return nil
}

func (a *Admin) Close() error {
	a.running = false
	return a.listener.Close()
}

func (a *Admin) handleConn(conn net.Conn) {
	transport, err := spdy.NewServerTransport(conn)
	if err != nil {
		conn.Close()
		a.logger.WithField("error", err).Error("new spdy transport")
		return
	}
	defer transport.Close()
	receiver, err := transport.WaitReceiveChannel()
	if err != nil {
		a.logger.WithField("error", err).Error("new receive channel")
		return
	}

	for {
		var c command
		if err := receiver.Receive(&c); err != nil {
			a.logger.WithField("error", err).Error("receive command")
			break
		}
		a.logger.WithFields(logrus.Fields{
			"op": c.Op,
		}).Debugf("receive command %#v", c)

		switch c.Op {
		case "addplugin":
			// <name> <args...>
			name := c.Args[0]
			p := avaliablePlugins[name]
			if err := a.daemon.LoadPlugin(name, p); err != nil {
				emit(c.Out, "error", err.Error())
				c.Out.Close()
				continue
			}
			a.sb.StartRouting(p.GetPlug())
			//if err := <-; err != nil {
			//	emit(c.Out, "error", err.Error())
			//	c.Out.Close()
			//	continue
			//}
			emit(c.Out, "status", "OK")
			c.Out.Close()
		case "listplugins":
			plugins, err := a.daemon.ListPlugins()
			if err != nil {
				emit(c.Out, "error", err.Error())
				c.Out.Close()
				continue
			}
			e := event{Stream: "data", KV: make(map[string]string, len(plugins))}
			for _, p := range plugins {
				e.KV[p.Name()] = ""
			}
			c.Out.Send(e)
			c.Out.Close()
		default:
			a.logger.Debugf("getting plugin for %q", c.Op)
			var command rpc.Cmd
			command.Op = c.Op
			command.Args = c.Args
			command.KV = c.KV
			command.Out = c.Out
			if err := a.sb.Call(&command); err != nil {
				emit(c.Out, "error", err.Error())
				c.Out.Close()
				continue
			}
		}
	}
}
