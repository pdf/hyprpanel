// Package main provides the hyprpanel panel plugin binary
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pdf/hyprpanel/internal/panelplugin"
)

var log = hclog.New(&hclog.LoggerOptions{
	Level:      hclog.Trace,
	Output:     os.Stderr,
	JSONFormat: true,
})

func sigHandler(p *panel) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGINT)

	for s := range sigChan {
		switch s {
		case syscall.SIGTERM, syscall.SIGINT:
			log.Warn(`Quitting`, `sig`, s.String())
			p.app.Quit()
		case syscall.SIGUSR1:
			log.Warn(`Quitting`, `sig`, s.String())
			// TODO: Implement reload
			p.app.Quit()
		default:
			log.Warn(`Unhandled signal`, `sig`, s.String())
		}
	}
}

func main() {
	p, err := newPanel()
	if err != nil {
		log.Error(`hyprpanel initialization failed`, `err`, err)
		os.Exit(1)
	}

	go sigHandler(p)

	go func() {
		plugin.Serve(&plugin.ServeConfig{
			HandshakeConfig: panelplugin.Handshake,
			Plugins: map[string]plugin.Plugin{
				panelplugin.PanelPluginName: &panelplugin.PanelPlugin{Impl: p},
			},
			GRPCServer: plugin.DefaultGRPCServer,
			Logger:     log,
		})
	}()

	os.Exit(p.run())
}
