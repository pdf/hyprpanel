# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [hyprpanel/config/v1/config.proto](#hyprpanel_config_v1_config-proto)
    - [Config](#hyprpanel-config-v1-Config)
    - [Config.Audio](#hyprpanel-config-v1-Config-Audio)
    - [Config.DBUS](#hyprpanel-config-v1-Config-DBUS)
    - [Config.DBUS.Brightness](#hyprpanel-config-v1-Config-DBUS-Brightness)
    - [Config.DBUS.Notifications](#hyprpanel-config-v1-Config-DBUS-Notifications)
    - [Config.DBUS.Power](#hyprpanel-config-v1-Config-DBUS-Power)
    - [Config.DBUS.Shortcuts](#hyprpanel-config-v1-Config-DBUS-Shortcuts)
    - [Config.DBUS.Systray](#hyprpanel-config-v1-Config-DBUS-Systray)
    - [IconOverride](#hyprpanel-config-v1-IconOverride)
    - [Panel](#hyprpanel-config-v1-Panel)
  
    - [Edge](#hyprpanel-config-v1-Edge)
    - [LogLevel](#hyprpanel-config-v1-LogLevel)
  
- [Scalar Value Types](#scalar-value-types)



<a name="hyprpanel_config_v1_config-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## hyprpanel/config/v1/config.proto



<a name="hyprpanel-config-v1-Config"></a>

### Config



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| log_level | [LogLevel](#hyprpanel-config-v1-LogLevel) |  | specifies the maximum log level for output. |
| log_subprocesses_to_journal | [bool](#bool) |  | send processes spawned by e.g. taskbar launchers to the systemd journal via sytstemd-cat. |
| dbus | [Config.DBUS](#hyprpanel-config-v1-Config-DBUS) |  | dbus configuration section. |
| audio | [Config.Audio](#hyprpanel-config-v1-Config-Audio) |  | audio configuration section. |
| panels | [Panel](#hyprpanel-config-v1-Panel) | repeated | list of panels to display. |
| icon_overrides | [IconOverride](#hyprpanel-config-v1-IconOverride) | repeated | list of icon overrides. |






<a name="hyprpanel-config-v1-Config-Audio"></a>

### Config.Audio



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | [bool](#bool) |  | if false, no Audio functionality is available. |
| volume_step_percent | [uint32](#uint32) |  | percentage that volume should change on each adjustment. |
| volume_exceed_maximum | [bool](#bool) |  | allow increasing volume above 100%. |
| hud_notifications | [bool](#bool) |  | display HUD notifications on volume change (requires at least one HUD module). |






<a name="hyprpanel-config-v1-Config-DBUS"></a>

### Config.DBUS



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | [bool](#bool) |  | if false, no DBUS functionality is available. |
| connect_timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  | specifies the maximum time we will attempt to connect to the bus before failing (format: &#34;20s&#34;). |
| connect_interval | [google.protobuf.Duration](#google-protobuf-Duration) |  | specifies the interval that we will attempt to connect to the session bus on startup (format: &#34;0.200s&#34;). |
| notifications | [Config.DBUS.Notifications](#hyprpanel-config-v1-Config-DBUS-Notifications) |  | notifications configuration. |
| systray | [Config.DBUS.Systray](#hyprpanel-config-v1-Config-DBUS-Systray) |  | systray configuration. |
| shortcuts | [Config.DBUS.Shortcuts](#hyprpanel-config-v1-Config-DBUS-Shortcuts) |  | shortcuts configuration. |
| brightness | [Config.DBUS.Brightness](#hyprpanel-config-v1-Config-DBUS-Brightness) |  | brightness configuration. |
| power | [Config.DBUS.Power](#hyprpanel-config-v1-Config-DBUS-Power) |  | power configuration. |






<a name="hyprpanel-config-v1-Config-DBUS-Brightness"></a>

### Config.DBUS.Brightness



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | [bool](#bool) |  | enables brightness control functionality. |
| adjust_step_percent | [uint32](#uint32) |  | percentage that brightness should change on each adjustment. |
| min_brightness | [uint32](#uint32) |  | minimum brightness value. |
| enable_logind | [bool](#bool) |  | set brightness via systemd-logind DBUS interface instead of direct sysfs. Requires logind session, and DBUS.enabled = true. |
| hud_notifications | [bool](#bool) |  | display HUD notifications on change (requires at least one HUD module). |






<a name="hyprpanel-config-v1-Config-DBUS-Notifications"></a>

### Config.DBUS.Notifications



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | [bool](#bool) |  | toggles the notification host functionality, required for &#34;notifications&#34; module. |






<a name="hyprpanel-config-v1-Config-DBUS-Power"></a>

### Config.DBUS.Power



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | [bool](#bool) |  | enables power functionality. |
| low_percent | [uint32](#uint32) |  | percentage below which we should consider low power. |
| critical_percent | [uint32](#uint32) |  | percentage below which we should consider critical power. |
| low_command | [string](#string) |  | command to execute on low power. |
| critical_command | [string](#string) |  | command to execute on critical power. |
| hud_notifications | [bool](#bool) |  | display HUD notifications on power state change or low power. |






<a name="hyprpanel-config-v1-Config-DBUS-Shortcuts"></a>

### Config.DBUS.Shortcuts



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | [bool](#bool) |  | enables GlobalShortcuts support. |






<a name="hyprpanel-config-v1-Config-DBUS-Systray"></a>

### Config.DBUS.Systray



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | [bool](#bool) |  | toggles the StatusNotifierItem host, required for &#34;systray&#34; module. Must be the only SNI implementation running in the session. |






<a name="hyprpanel-config-v1-IconOverride"></a>

### IconOverride



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| window_class | [string](#string) |  | window class of the application to match. |
| icon | [string](#string) |  | icon name to use for this application. |






<a name="hyprpanel-config-v1-Panel"></a>

### Panel



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | unique identifier for this panel. |
| edge | [Edge](#hyprpanel-config-v1-Edge) |  | screen edge to place this panel. |
| size | [uint32](#uint32) |  | either width or height in pixels, depending on orientation for screen edge. |
| monitor | [string](#string) |  | monitor to display this panel on. |
| modules | [hyprpanel.module.v1.Module](#hyprpanel-module-v1-Module) | repeated | list of modules for this panel. |





 


<a name="hyprpanel-config-v1-Edge"></a>

### Edge


| Name | Number | Description |
| ---- | ------ | ----------- |
| EDGE_UNSPECIFIED | 0 |  |
| EDGE_TOP | 1 |  |
| EDGE_RIGHT | 2 |  |
| EDGE_BOTTOM | 3 |  |
| EDGE_LEFT | 4 |  |



<a name="hyprpanel-config-v1-LogLevel"></a>

### LogLevel


| Name | Number | Description |
| ---- | ------ | ----------- |
| LOG_LEVEL_UNSPECIFIED | 0 |  |
| LOG_LEVEL_TRACE | 1 |  |
| LOG_LEVEL_DEBUG | 2 |  |
| LOG_LEVEL_INFO | 3 |  |
| LOG_LEVEL_WARN | 4 |  |
| LOG_LEVEL_ERROR | 5 |  |
| LOG_LEVEL_OFF | 6 |  |


 

 

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

