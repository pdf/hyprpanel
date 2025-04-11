module github.com/pdf/hyprpanel

go 1.24

toolchain go1.24.0

require (
	github.com/disintegration/imaging v1.6.2
	github.com/fsnotify/fsnotify v1.9.0
	github.com/godbus/dbus/v5 v5.1.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-plugin v1.6.3
	github.com/iancoleman/strcase v0.3.0
	github.com/jfreymuth/pulse v0.1.1
	github.com/jwijenbergh/purego v0.0.0-20241210143217-aeaa0bfe09e0
	github.com/jwijenbergh/puregotk v0.0.0-20250407124134-bc1a52f44fd4
	github.com/mattn/go-shellwords v1.0.12
	github.com/pdf/go-wayland v0.0.2
	github.com/peterbourgon/ff/v4 v4.0.0-alpha.4
	github.com/rkoesters/xdg v0.0.1
	golang.org/x/sync v0.13.0
	golang.org/x/sys v0.32.0
	google.golang.org/grpc v1.71.1
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/fatih/color v1.18.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/pelletier/go-toml/v2 v2.1.1 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/image v0.26.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250409194420-de1ac958c67a // indirect
	mvdan.cc/gofumpt v0.7.0 // indirect
)

tool github.com/pdf/go-wayland/cmd/go-wayland-scanner
