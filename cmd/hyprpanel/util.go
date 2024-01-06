package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func findClient() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return ``, err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return ``, err
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return ``, err
	}

	curDir := filepath.Dir(exe)
	clientPath := filepath.Join(curDir, clientName)
	if _, err := os.Stat(clientPath); err != nil {
		clientPath, err = exec.LookPath(clientName)
		if err != nil {
			return ``, err
		}
	}

	return clientPath, nil
}

func findLayerShell() (string, error) {
	var searchPaths []string
	switch runtime.GOARCH {
	case `amd64`:
		searchPaths = []string{`/usr/lib/x86_64-linux-gnu/`, `/usr/lib64/`}
	case `arm64`:
		searchPaths = []string{"/usr/lib/aarch64-linux-gnu/", "/usr/lib64/"}
	default:
		return ``, fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	if path, err := findLayerShellBySearchPath(searchPaths); err != nil {
		return findLayerShellByPkgconfig()
	} else {
		return path, nil
	}
}

func findLayerShellBySearchPath(paths []string) (string, error) {
	for _, dir := range paths {
		path := filepath.Join(dir, layerShellLib)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return ``, errors.New(`could not find gtk4-layer-shell library`)
}

func findLayerShellByPkgconfig() (string, error) {
	cmd := exec.Command(`pkg-config`, `--libs-only-L`, layerShellPkg)
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdOut, &stdErr
	if err := cmd.Run(); err != nil {
		return ``, fmt.Errorf("pkg-config failed finding '%s' with error %s: %s", layerShellPkg, err, stdOut.String())
	}

	outs := strings.Split(stdOut.String(), "-L")
	for _, v := range outs {
		c := strings.TrimSpace(v)
		if c == "" {
			continue
		}
		g, err := findLayerShellBySearchPath([]string{c})
		if err != nil {
			return ``, err
		}
		if g != "" {
			return g, nil
		}
	}

	return ``, errors.New(`pkg-config search failed`)
}
