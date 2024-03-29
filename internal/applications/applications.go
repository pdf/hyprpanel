// Package applications provides an API for querying Desktop entries.
package applications

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/hashicorp/go-hclog"
	hyprpanelv1 "github.com/pdf/hyprpanel/proto/hyprpanel/v1"
	"github.com/rkoesters/xdg/desktop"
)

const (
	defaultName = `Unknown`
	defaultIcon = `wayland`
)

var (
	// ErrNotFound is returned when an application is not found.
	ErrNotFound = errors.New(`application not found`)
)

// AppCache holds an auto-updated list of Desktop application data.
type AppCache struct {
	log        hclog.Logger
	mu         sync.RWMutex
	watcher    *fsnotify.Watcher
	quitCh     chan struct{}
	targets    []string
	cache      map[string]*hyprpanelv1.AppInfo
	cacheLower map[string]*hyprpanelv1.AppInfo
}

// Find an application by class.
func (a *AppCache) Find(class string) *hyprpanelv1.AppInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if app, ok := a.cache[class]; ok {
		return app
	} else if app, ok := a.cacheLower[strings.ToLower(class)]; ok {
		return app
	} else if idx := strings.Index(class, `-`); idx > 0 {
		class = class[:idx]
		if app, ok := a.cache[class]; ok {
			return app
		}
		if app, ok := a.cacheLower[strings.ToLower(class)]; ok {
			return app
		}
	}

	return &hyprpanelv1.AppInfo{
		Name: defaultName,
		Icon: defaultIcon,
	}
}

// Refresh the cache.
func (a *AppCache) Refresh() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, target := range a.targets {
		stat, err := os.Stat(target)
		if err != nil || !stat.IsDir() {
			continue
		}

		err = filepath.WalkDir(target, a.cacheWalk)

		if err != nil {
			return err
		}
	}

	return nil
}

// Close this instance.
func (a *AppCache) Close() error {
	close(a.quitCh)
	return a.watcher.Close()
}

func (a *AppCache) cacheWalk(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if d.IsDir() {
		return nil
	}

	if !strings.HasSuffix(path, `.desktop`) {
		return nil
	}

	p, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	app, err := newAppInfo(p)
	if err != nil {
		a.log.Error(`Failed parsing desktop file`, `file`, p, `err`, err)
		return nil
	}

	name := strings.TrimSuffix(d.Name(), `.desktop`)
	// Match app_id with missing prefix
	if idx := strings.LastIndex(name, `.`); idx > 0 {
		dotPrefixed := name[idx+1:]
		a.cache[dotPrefixed] = app
		a.cacheLower[strings.ToLower(dotPrefixed)] = app
	}
	// Match app_id for mismatched .desktop and WmClass
	if app.StartupWmClass != `` {
		a.cache[app.StartupWmClass] = app
		a.cacheLower[strings.ToLower(app.StartupWmClass)] = app
	}
	// Last ditch, by Name for apps with missing app_id
	if app.Name != `` {
		a.cache[app.Name] = app
		a.cacheLower[strings.ToLower(app.Name)] = app
	}
	// Standard match app_id by .desktop name
	a.cache[name] = app
	a.cacheLower[strings.ToLower(name)] = app

	return nil
}

func (a *AppCache) watch() {
	for _, target := range a.targets {
		if err := a.watcher.Add(target); err != nil {
			a.log.Warn(`Failed adding application watcher`, `target`, target, `err`, err)
		}
	}
	debounce := time.NewTimer(200 * time.Millisecond)
	if !debounce.Stop() {
		select {
		case <-debounce.C:
		default:
		}
	}
	defer debounce.Stop()

	for {
		select {
		case <-a.quitCh:
			return
		default:
			select {
			case <-a.quitCh:
				return
			case <-a.watcher.Events:
				if !debounce.Stop() {
					select {
					case <-debounce.C:
					default:
					}
				}
				debounce.Reset(200 * time.Millisecond)
			case <-debounce.C:
				if err := a.Refresh(); err != nil {
					a.log.Warn(`Failed refreshing AppCache`, `err`, err)
				}
			}
		}
	}
}

func newAppInfo(file string) (*hyprpanelv1.AppInfo, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	a := &hyprpanelv1.AppInfo{
		DesktopFile: file,
	}

	entry, err := desktop.New(r)
	if err != nil {
		return nil, err
	}

	a.Name = entry.Name
	a.Icon = entry.Icon
	a.TryExec = entry.TryExec
	a.Exec = parseExec(entry.Exec, entry, file)
	a.RawExec = entry.Exec
	a.Path = entry.Path
	a.StartupWmClass = entry.StartupWMClass
	a.Terminal = entry.Terminal
	if len(entry.Actions) > 0 {
		a.Actions = make([]*hyprpanelv1.AppInfo_Action, len(entry.Actions))
		for i, action := range entry.Actions {
			a.Actions[i] = &hyprpanelv1.AppInfo_Action{
				Name:    action.Name,
				Icon:    action.Icon,
				Exec:    parseExec(action.Exec, entry, file),
				RawExec: action.Exec,
			}
		}
	}

	if a.Name == `` {
		a.Name = defaultName
	}
	if a.Icon == `` {
		a.Icon = defaultIcon
	}

	return a, nil
}

// New instantiate a new AppCache.
func New(log hclog.Logger) (*AppCache, error) {
	a := &AppCache{
		log:        log,
		cache:      make(map[string]*hyprpanelv1.AppInfo),
		cacheLower: make(map[string]*hyprpanelv1.AppInfo),
		quitCh:     make(chan struct{}),
	}

	var err error
	a.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	dedup := make(map[string]struct{})
	a.targets = make([]string, 0)
	xdgDataDirs := os.Getenv(`XDG_DATA_DIRS`)
	xdgDataHome := os.Getenv(`XDG_DATA_HOME`)
	home := os.Getenv(`HOME`)
	if xdgDataDirs != `` {
		for _, dir := range strings.Split(xdgDataDirs, `:`) {
			target := filepath.Join(dir, `applications`)
			dedup[target] = struct{}{}
			a.targets = append(a.targets, target)
		}
	} else {
		usr := `/usr/share/applications`
		usrLocal := `/usr/local/share/applications`
		dedup[usr] = struct{}{}
		a.targets = append(a.targets, usr)
		dedup[usrLocal] = struct{}{}
		a.targets = append(a.targets, usrLocal)
	}
	varFlatpak := `/var/lib/flatpak/exports/share/applications`
	if _, found := dedup[varFlatpak]; !found {
		dedup[varFlatpak] = struct{}{}
		a.targets = append(a.targets, varFlatpak)
	}
	var homeApps, homeFlatpak string
	if xdgDataHome != `` {
		homeApps = filepath.Join(xdgDataHome, `applications`)
		homeFlatpak = filepath.Join(xdgDataHome, `flatpak`, `exports`, `share`, `applications`)
	} else if home != `` {
		homeApps = filepath.Join(home, `.local`, `share`, `applications`)
		homeFlatpak = filepath.Join(home, `.local`, `share`, `flatpak`, `exports`, `share`, `applications`)
	}
	if _, found := dedup[homeApps]; !found {
		dedup[homeApps] = struct{}{}
		a.targets = append(a.targets, homeApps)
	}
	if _, found := dedup[homeFlatpak]; !found {
		dedup[homeFlatpak] = struct{}{}
		a.targets = append(a.targets, homeFlatpak)
	}

	go func() {
		if err := a.Refresh(); err != nil {
			log.Error(`Failed walking application directories`, `err`, err)
		}
	}()
	go a.watch()

	return a, nil
}

func parseExec(exec string, i *desktop.Entry, desktopFile string) string {
	if exec == `` {
		return exec
	}

	exec = strings.ReplaceAll(exec, `%f`, ``)
	exec = strings.ReplaceAll(exec, `%F`, ``)
	exec = strings.ReplaceAll(exec, `%u`, ``)
	exec = strings.ReplaceAll(exec, `%U`, ``)
	exec = strings.ReplaceAll(exec, `%d`, ``)
	exec = strings.ReplaceAll(exec, `%D`, ``)
	exec = strings.ReplaceAll(exec, `%n`, ``)
	exec = strings.ReplaceAll(exec, `%N`, ``)
	exec = strings.ReplaceAll(exec, `%v`, ``)
	exec = strings.ReplaceAll(exec, `%m`, ``)
	if i.Icon != `` {
		exec = strings.ReplaceAll(exec, `%i`, fmt.Sprintf("--icon %s", i.Icon))
	}
	if i.Name != `` {
		exec = strings.ReplaceAll(exec, `%c`, fmt.Sprintf("'%s'", i.Name))
	}
	exec = strings.ReplaceAll(exec, `%k`, fmt.Sprintf("'%s'", desktopFile))

	return strings.TrimSpace(exec)
}
