// Package main provides the hyprpanel host binary
package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pdf/hyprpanel/config"
	"github.com/pdf/hyprpanel/style"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"golang.org/x/sys/unix"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	name         = `hyprpanel`
	crashTimeout = 2 * time.Second
	crashRetry   = 200 * time.Millisecond
	crashCount   = 3
)

func reapChildren(log hclog.Logger) {
	// Wait for the child process to exit, this will also reap any zombie processes
	for {
		pid, err := unix.Wait4(-1, nil, unix.WNOHANG, nil)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			if errors.Is(err, unix.ECHILD) {
				log.Debug(`No more child processes`)
				break
			}
			log.Error(`Failed waiting for child process`, `err`, err)
			break
		}
		if pid > 0 {
			log.Debug(`Reaped child process`, `pid`, pid)
		} else {
			log.Debug(`No child processes to reap`)
			break
		}
	}
}

func sigHandler(log hclog.Logger, h *host) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGINT, syscall.SIGCHLD)

	for s := range sigChan {
		switch s {
		case syscall.SIGTERM, syscall.SIGINT:
			log.Warn(`Quitting`, `sig`, s.String())
			go h.Close()
		case syscall.SIGCHLD:
			log.Debug(`Child process exited`, `sig`, s.String())
			go reapChildren(log)
		default:
			log.Warn(`Unhandled signal`, `sig`, s.String())
		}
	}
}

func main() {
	fs := ff.NewFlagSet(name)
	var configPath string
	xdgConfigPath := os.Getenv(`XDG_CONFIG_HOME`)
	if xdgConfigPath == `` {
		configPath = filepath.Join(os.Getenv(`HOME`), `.config`, `hyprpanel`)
	} else {
		configPath = filepath.Join(xdgConfigPath, `hyprpanel`)
	}
	configFileDefault := filepath.Join(configPath, `config.json`)
	configFile := fs.String('c', `config`, configFileDefault, `Path to configuration file`)
	styleFileDefault := filepath.Join(configPath, `style.css`)
	styleFile := fs.String('s', `style`, styleFileDefault, `Path to stylesheet`)
	version := fs.BoolLong(`version`, `Display the application version`)

	log := hclog.New(&hclog.LoggerOptions{
		Name:   `host`,
		Output: os.Stdout,
	})

	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix(`HYPRPANEL`)); err != nil {
		fmt.Printf("%s\n", ffhelp.Flags(fs))
		if errors.Is(err, ff.ErrHelp) {
			os.Exit(0)
		}

		fmt.Printf("err=%v\n", err)
		os.Exit(1)
	}

	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
		log.Error(`Failed setting child subreaper`, `err`, err)
		os.Exit(1)
	}

	if *version {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Printf("%s unknown version", name)
			os.Exit(1)
		}
		fmt.Printf("%s version %s built with %s\n", name, info.Main.Version, info.GoVersion)
		os.Exit(0)
	}

	cfg, err := config.Load(*configFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Error(`Failed loading configuration file`, `file`, *configFile, `err`, err)
			os.Exit(1)
		}
		log.Warn(`Failed loading configuration file, creating with defaults`, `file`, *configFile)
		cfg, err = config.Default()
		if err != nil {
			log.Error(`Failed loading default configuration file`, `err`, err)
			os.Exit(1)
		}
		if err := os.MkdirAll(configPath, 0o755); err != nil && err != os.ErrExist {
			log.Error(`Failed creating configuration directory`, `path`, configPath, `err`, err)
			os.Exit(1)
		}

		marshal := protojson.MarshalOptions{
			Multiline:       true,
			Indent:          "\t",
			EmitUnpopulated: true,
			UseProtoNames:   true,
		}
		b, err := marshal.Marshal(cfg)
		if err != nil {
			log.Error(`Failed encoding default configuration file`, `err`, err)
			os.Exit(1)
		}

		if err := os.WriteFile(*configFile, b, 0o644); err != nil {
			log.Error(`Failed writing default configuration file`, `file`, *configFile, `err`, err)
			os.Exit(1)
		}
	}

	log.SetLevel(hclog.Level(cfg.LogLevel))

	stylesheet, err := style.Load(*styleFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Error(`Failed loading stylesheet`, `file`, *styleFile, `err`, err)
			os.Exit(1)
		}
		log.Warn(`Failed loading stylesheet, continuing with defaults`, `file`, *styleFile)
	}

	h, err := newHost(cfg, stylesheet, log)
	if err != nil {
		log.Error(`Failed initializing hyprpanel`, `err`, err)
		os.Exit(1)
	}
	go sigHandler(log, h)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(`Failed initializaing filesystem watcher`, `err`, err)
		os.Exit(1)
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.Error(`Failed closing filesystem watcher`, `err`, err)
		}
	}()

	go func() {
		for {
			select {
			case evt := <-watcher.Events:
				if !evt.Has(fsnotify.Write) && !evt.Has(fsnotify.Create) && !evt.Has(fsnotify.Remove) {
					continue
				}
				switch evt.Name {
				case *configFile:
					cfg, err := config.Load(*configFile)
					if os.IsNotExist(err) {
						cfg, err = config.Default()
					}
					if err != nil {
						log.Error(`Failed reloading config`, `err`, err)
						continue
					}
					log.SetLevel(hclog.Level(cfg.LogLevel))
					h.updateConfig(cfg)
				case *styleFile:
					stylesheet, err := style.Load(*styleFile)
					if os.IsNotExist(err) {
						stylesheet = style.Default
					} else if err != nil {
						log.Error(`Failed reloading stylesheet`, `err`, err)
						continue
					}
					h.updateStyle(stylesheet)
				}
			case err := <-watcher.Errors:
				if err != nil {
					log.Error(`Filesystem watcher failed`, `err`, err)
					os.Exit(1)
				}
			}
		}
	}()

	configDir := filepath.Dir(*configFile)
	styleDir := filepath.Dir(*styleFile)
	if err := watcher.Add(configDir); err != nil {
		log.Error(`Failed adding filesystem watch path`, `path`, configDir, `err`, err)
		os.Exit(1)
	}
	if configDir != styleDir {
		if err := watcher.Add(styleDir); err != nil {
			log.Error(`Failed adding filesystem watch path`, `path`, styleDir, `err`, err)
			os.Exit(1)
		}
	}
	defer plugin.CleanupClients()

	count := 0
	timer := time.NewTimer(crashTimeout)
	for {
		err := h.run()
		if err != nil && err != errReload {
			count += 1
			log.Error(`Clients failed`, `err`, err)
			select {
			case <-timer.C:
				if count >= crashCount {
					log.Error(`Restarting too quickly, terminating`)
					os.Exit(1)
				}
				count = 0
				timer.Reset(crashTimeout)
			default:
			}
			time.Sleep(crashRetry)
			continue
		} else if err != nil && err == errReload {
			continue
		}

		os.Exit(0)
	}
}
