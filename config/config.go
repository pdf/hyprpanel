// Package config handles config loading.
package config

import (
	_ "embed"
	"io"
	"os"

	configv1 "github.com/pdf/hyprpanel/proto/hyprpanel/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

//go:embed default.json
var defaultConfig []byte

// Default returns the default configuration values.
func Default() (*configv1.Config, error) {
	c := &configv1.Config{}
	if err := protojson.Unmarshal(defaultConfig, c); err != nil {
		return nil, err
	}

	return c, nil
}

// Load and parse a configuration file from disk.
func Load(filePath string) (*configv1.Config, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	c := &configv1.Config{}

	if err := protojson.Unmarshal(b, c); err != nil {
		return nil, err
	}

	return c, nil

}
