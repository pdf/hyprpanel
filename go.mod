module github.com/pdf/hyprpanel

go 1.24.0

toolchain go1.24.6

require (
	github.com/disintegration/imaging v1.6.2
	github.com/fsnotify/fsnotify v1.9.0
	github.com/godbus/dbus/v5 v5.1.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-plugin v1.7.0
	github.com/iancoleman/strcase v0.3.0
	github.com/jfreymuth/pulse v0.1.1
	github.com/jwijenbergh/purego v0.0.0-20250812133547-b5852df1402b
	github.com/jwijenbergh/puregotk v0.0.0-20250812133623-7203178b5172
	github.com/mattn/go-shellwords v1.0.12
	github.com/pdf/go-wayland v0.0.3
	github.com/peterbourgon/ff/v4 v4.0.0-beta.1
	github.com/rkoesters/xdg v0.0.1
	golang.org/x/sync v0.16.0
	golang.org/x/sys v0.35.0
	google.golang.org/grpc v1.75.0
	google.golang.org/protobuf v1.36.8
)

require (
	github.com/fatih/color v1.18.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/oklog/run v1.2.0 // indirect
	github.com/pelletier/go-toml/v2 v2.1.1 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/image v0.30.0 // indirect
	golang.org/x/mod v0.26.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/tools v0.35.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250826171959-ef028d996bc1 // indirect
	mvdan.cc/gofumpt v0.7.0 // indirect
)

tool github.com/pdf/go-wayland/cmd/go-wayland-scanner
