# hyprpanel

[![Lint](https://github.com/pdf/hyprpanel/actions/workflows/lint.yml/badge.svg)](https://github.com/pdf/hyprpanel/actions/workflows/lint.yml)
[![Release](https://github.com/pdf/hyprpanel/actions/workflows/release.yml/badge.svg)](https://github.com/pdf/hyprpanel/actions/workflows/release.yml)
[![AUR](https://github.com/pdf/hyprpanel/actions/workflows/aur.yml/badge.svg)](https://github.com/pdf/hyprpanel/actions/workflows/aur.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/pdf/hyprpanel)](https://goreportcard.com/report/github.com/pdf/hyprpanel)
[![License](https://img.shields.io/badge/License-MIT-%23a31f34)](https://github.com/pdf/hyprpanel/blob/main/LICENSE)

An opinionated panel/shell for the Hyprland compositor.

> [!NOTE]
> This project was created as a holiday hackathon for my personal use only. There are almost certainly bugs, and possibly the occasional memory leak. Please don't read the code if you're easily offended 🙃.
> 
> Use at your own risk - I will accept contributions, but I don't expect to spend a lot of time maintaining this project.

https://github.com/pdf/hyprpanel/assets/146192/2da11a7e-b232-4c7d-a563-997d7e4e0be1

## Install

### Dependencies

This project depends on (required):
- gtk4
- gtk4-layer-shell
- Hyprland (version must be >= [5920c6a](https://github.com/hyprwm/Hyprland/commit/5920c6a6b8d059413377f0cb25f3dfb1dc8c4201))

Optional dependencies (required for default configuration):
- systemd
- pipewire-pulse/pulseaudio (for audio)
- upower (for battery state)

Please ensure that you have these packages installed.

### Install from Release

Download the [latest release](https://github.com/pdf/hyprpanel/releases/latest) for your operating system and architecture.

Unpack the archive and place the `hyprpanel` and `hyprpanel-client` binaries on your `$PATH`.

### Install from Source

Execute the following (requires the Go toolchain):

```shell
go install github.com/pdf/hyprpanel/cmd/hyprpanel@latest
go install github.com/pdf/hyprpanel/cmd/hyprpanel-client@latest
```

## Usage

Add the following to your `hyprland.conf` (assuming that `hyprpanel` is available on your PATH):

```
exec-once = hyprpanel
```

## Configuration

On first run, hyprpanel will create a default configuration file at:

`${XDG_CONFIG_HOME}/hyprpanel/config.json`

You may review the current default configuration at [config/default.json](config/default.json).

JSON is not my first choice for a human-writable config format, but due to internal protobuf usage this was by far the least painful format to implement.

For details on the configuration options, please see the [Configuration Reference](https://github.com/pdf/hyprpanel/wiki/Config) in the wiki.

## Panels

Multiple panels are supported, if that's your thing.

## Modules

Each panel is composed of modules.

### Pager

The pager module displays a stylized preview of your workspace contents.

#### Actions

- Left-click switches to workspace.
- Scroll-wheel switches between workspaces.

### Taskbar

The taskbar module displays an icon-only representation of running tasks, and optionally displays pinned launchers.

#### Actions

- Left-click launches a pinned application if it is not running. Focuses the application if it is running.
- Middle-click launches a new instance of the application (if supported).
- Right-click displays the application context-menu.
- Scroll-wheel cycles focus between application windows when grouped tasks is enabled.

### Systray

The systray module implements the StatusNotifierItem spec.

Some modules (where noted) support embedding in the systray.

> [!NOTE]
> The hyprpanel SNI implementation does not play well with others: hyprpanel does not support registering additional StatusNotifierHosts, and will fail to start if it can't own the bus. So if you use this module hyprpanel *must* be the only SNI implementation on your desktop.
>
> To disable systray support, set the config option `dbus.systray.enabled` to `false`, and remove the `systray` module from all panels.
>
> If you need xembed support, you can try `xembedsniproxy` from the KDE project, though expect some artifical delays as that project expects to communicate directly with the KDE SNI implementation.

### Audio

The audio module displays the current audio volume level as an icon.

This module supports embedding in systray.

#### Actions

- Left-click launches the specified mixer application.
- Middle-click mutes the default input device.
- Right-click mutes the default output device.
- Scroll-wheel adjusts default output volume. 

### Power

The power module displays the current battery level as an icon.

This modules supports embedding in systray.

#### Actions

- Scroll-wheel adjusts display brightness. 

### Clock

The clock module displays the current time/date. You may also specify a number of secondary regions to display via tooltip.

#### Actions

- Left-click to display a basic calendar.

### Session

The session module provides a basic session management screen.

#### Actions

- Left-click to display session management screen.

### Notifications

Displays system notifications.

> [!NOTE]
> The hyprpanel notifications implementation does not play well with others, and will fail to start if it can't own the bus. So if you use this module hyprpanel *must* be the only notifications implementation on your desktop.
>
> To disable notifications support, set the config option `dbus.notifications.enabled` to `false`, and remove the `notifications` module from all panels.

#### Actions

On notifications:

- Left-click on notifications that include a default action will execute that action and optionally focus the sending application if supported by the notification.
- Middle-click closes the notification.

### Hud

Displays heads-up notifications for hardware events (e.g. volume, display brightness changes, etc)

### Spacer

The spacer module simply adds empty space between modules.

## Global keybinds

Global keybinds are registered through the desktop portal, so they do not have prefixes in `hyprctl globalshortcuts`. The following keybinds are available:

```
:com.c0dedbad.hyprpanel.audioSinkVolumeUp -> Increase the volume of the default audio output device
:com.c0dedbad.hyprpanel.audioSinkVolumeDown -> Decrease the volume of the default audio output device
:com.c0dedbad.hyprpanel.audioSinkMuteToggle -> Toggle the mute status of the default audio output device
:com.c0dedbad.hyprpanel.audioSourceVolumeUp -> Increase the volume of the default audio input device
:com.c0dedbad.hyprpanel.audioSourceVolumeDown -> Decrease the volume of the default audio input device
:com.c0dedbad.hyprpanel.audioSourceMuteToggle -> Toggle the mute status of the default audio input device
:com.c0dedbad.hyprpanel.brightnessUp -> Increase display brightness
:com.c0dedbad.hyprpanel.brightnessDown -> Increase display brightness
```

## Styling

You may apply custom styling by providing a GTK4-compatible CSS file. By default hyprpanel will look for this file at:

`${XDG_CONFIG_HOME}/hyprpanel/style.css`

You may review the default stylesheet at [style/default.css](style/default.css).

If you simply wish to update the colour scheme, you may define only the colours you wish to change in your stylesheet. For example, the following sets the highlight to blue and creates floating modules by setting a transparent panel background, and opaque module background:

```css
@define-color Highlight rgba(0, 102, 255, 1.0);
@define-color PanelBackground rgba(0, 0, 0, 0.0);
@define-color ModuleBackground rgba(16, 16, 16, 1.0);
```

Alternatively, you can supply entirely custom styles, have fun.

## Roadmap

- [ ] Granular config reloads - reloads currently restart the whole panel plugin process
- [X] (Pulse)Audio module
- [X] Power/Battery/Brightness module
- [ ] Notification history
- [ ] GUI configuration (e.g. pinned launchers, pinned tray items, etc) (maybe)
