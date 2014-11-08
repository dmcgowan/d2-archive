package admin

import (
	"net"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/d2/daemon"
	"github.com/docker/libchan"
	"github.com/docker/libchan/spdy"
)

var avaliablePlugins = map[string]daemon.Plugin{
	"exec": daemon.NewExecPlugin(),
}

func New(d daemon.Daemon, logger *logrus.Logger) *Admin {
	return &Admin{
		daemon: d,
		logger: logger,
	}
}

type Admin struct {
	running  bool
	daemon   daemon.Daemon
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
		if err := a.handleConn(conn); err != nil {
			a.logger.WithField("error", err).Error("handle connection")
			continue
		}
	}
	return nil
}

func (a *Admin) Close() error {
	a.running = false
	return a.listener.Close()
}

func (a *Admin) handleConn(conn net.Conn) error {
	transport, err := spdy.NewServerTransport(conn)
	if err != nil {
		conn.Close()
		return err
	}
	defer transport.Close()
	receiver, err := transport.WaitReceiveChannel()
	if err != nil {
		return err
	}

	go a.receiveLoop(receiver)
	return nil
}

func (a *Admin) receiveLoop(receiver libchan.Receiver) {
	for {
		var c *command
		if err := receiver.Receive(&c); err != nil {
			a.logger.WithField("error", err).Error("receive command")
			break
		}
		a.logger.WithFields(logrus.Fields{
			"op": c.Op,
		}).Debug("receive command")

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
			plugin, err := a.daemon.GetPlugin(c.Op)
			if err != nil {
				emit(c.Out, "error", err.Error())
				c.Out.Close()
				continue
			}
			if err := plugin.Process(c.Args, c.Out); err != nil {
				a.logger.WithField("error", err).Errorf("%s process command", c.Op)
				emit(c.Out, "error", err.Error())
			}
			c.Out.Close()
		}
	}
}
