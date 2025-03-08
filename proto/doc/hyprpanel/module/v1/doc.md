# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [hyprpanel/module/v1/module.proto](#hyprpanel_module_v1_module-proto)
    - [Audio](#hyprpanel-module-v1-Audio)
    - [Clock](#hyprpanel-module-v1-Clock)
    - [Hud](#hyprpanel-module-v1-Hud)
    - [Module](#hyprpanel-module-v1-Module)
    - [Notifications](#hyprpanel-module-v1-Notifications)
    - [Pager](#hyprpanel-module-v1-Pager)
    - [Power](#hyprpanel-module-v1-Power)
    - [Session](#hyprpanel-module-v1-Session)
    - [Spacer](#hyprpanel-module-v1-Spacer)
    - [Systray](#hyprpanel-module-v1-Systray)
    - [SystrayModule](#hyprpanel-module-v1-SystrayModule)
    - [Taskbar](#hyprpanel-module-v1-Taskbar)
  
    - [Position](#hyprpanel-module-v1-Position)
    - [Systray.Status](#hyprpanel-module-v1-Systray-Status)
  
- [Scalar Value Types](#scalar-value-types)



<a name="hyprpanel_module_v1_module-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## hyprpanel/module/v1/module.proto



<a name="hyprpanel-module-v1-Audio"></a>

### Audio



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| icon_size | [uint32](#uint32) |  | size in pixels for panel icon. |
| icon_symbolic | [bool](#bool) |  | display symbolic or coloured icon in panel. |
| command_mixer | [string](#string) |  | command to execute on mixer button. |
| enable_source | [bool](#bool) |  | display source (mic) icon in panel. |






<a name="hyprpanel-module-v1-Clock"></a>

### Clock



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| time_format | [string](#string) |  | Go time layout string for panel time display formatting, see https://pkg.go.dev/time#pkg-constants for details. |
| date_format | [string](#string) |  | Go time layout string for panel date display formatting, see https://pkg.go.dev/time#pkg-constants for details. |
| tooltip_time_format | [string](#string) |  | Go time layout string for tooltip time display formatting, see https://pkg.go.dev/time#pkg-constants for details. |
| tooltip_date_format | [string](#string) |  | Go time layout string for tooltip time display formatting, see https://pkg.go.dev/time#pkg-constants for details. |
| additional_regions | [string](#string) | repeated | list of addtional regions to display in the tooltip. |






<a name="hyprpanel-module-v1-Hud"></a>

### Hud



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| notification_icon_size | [uint32](#uint32) |  | size in pixels for icons in notifications. |
| timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  | delay before notifications are hidden (format: &#34;7s&#34;). |
| position | [Position](#hyprpanel-module-v1-Position) |  | screen position to display notifications. |
| margin | [uint32](#uint32) |  | space in pixels between notifications. |






<a name="hyprpanel-module-v1-Module"></a>

### Module



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pager | [Pager](#hyprpanel-module-v1-Pager) |  |  |
| taskbar | [Taskbar](#hyprpanel-module-v1-Taskbar) |  |  |
| systray | [Systray](#hyprpanel-module-v1-Systray) |  |  |
| notifications | [Notifications](#hyprpanel-module-v1-Notifications) |  |  |
| hud | [Hud](#hyprpanel-module-v1-Hud) |  |  |
| audio | [Audio](#hyprpanel-module-v1-Audio) |  |  |
| power | [Power](#hyprpanel-module-v1-Power) |  |  |
| clock | [Clock](#hyprpanel-module-v1-Clock) |  |  |
| session | [Session](#hyprpanel-module-v1-Session) |  |  |
| spacer | [Spacer](#hyprpanel-module-v1-Spacer) |  |  |






<a name="hyprpanel-module-v1-Notifications"></a>

### Notifications



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| icon_size | [uint32](#uint32) |  | size in pixels for the panel notification icon. Currently unused as notification history is unimplemented. |
| notification_icon_size | [uint32](#uint32) |  | size in pixels for icons in notifications. |
| default_timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  | delay before notifications are hidden, if the notification does not specify a timemout (format: &#34;7s&#34;). |
| position | [Position](#hyprpanel-module-v1-Position) |  | screen position to display notifications. |
| margin | [uint32](#uint32) |  | space in pixels between notifications. |
| persistent | [string](#string) | repeated | list of application names to retain notification history for. Currently unused as notification history is unimplemented. |






<a name="hyprpanel-module-v1-Pager"></a>

### Pager



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| icon_size | [uint32](#uint32) |  | size in pixels for pager window preview application icons. |
| active_monitor_only | [bool](#bool) |  | show only workspaces from the monitor the panel is running on. |
| scroll_wrap_workspaces | [bool](#bool) |  | when switching workspaces via mouse scroll, wrap to start/end on over-scroll. |
| scroll_include_inactive | [bool](#bool) |  | when scrolling workspaces, include inactive workspaces |
| enable_workspace_names | [bool](#bool) |  | display workspace name labels. |
| pinned | [int32](#int32) | repeated | list of workspace IDs that will always be included in the pager, regardless of activation state. |
| ignore_windows | [string](#string) | repeated | list of window classes that will be excluded from preview on the pager. |
| preview_width | [uint32](#uint32) |  | width in pixels for task preview windows. |
| follow_window_on_move | [bool](#bool) |  | when moving a window, switch to the workspace the window is being moved to. |






<a name="hyprpanel-module-v1-Power"></a>

### Power



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| icon_size | [uint32](#uint32) |  | size in pixels for panel icon. |
| icon_symbolic | [bool](#bool) |  | display symbolic or coloured icon in panel. |






<a name="hyprpanel-module-v1-Session"></a>

### Session



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| icon_size | [uint32](#uint32) |  | size in pixels for panel icon. |
| icon_symbolic | [bool](#bool) |  | display symbolic or coloured icon in panel. |
| overlay_icon_size | [uint32](#uint32) |  | size in pixels for overlay popup icons. |
| overlay_icon_symbolic | [bool](#bool) |  | display symbolic or coloured icons in overlay popup. |
| command_logout | [string](#string) |  | command that will be executed for logout action, empty disabled the button. |
| command_reboot | [string](#string) |  | command that will be executed for reboot action, empty disabled the button. |
| command_suspend | [string](#string) |  | command that will be executed for suspend action, empty disabled the button. |
| command_shutdown | [string](#string) |  | command that will be executed for shutdown action, empty disabled the button. |






<a name="hyprpanel-module-v1-Spacer"></a>

### Spacer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [uint32](#uint32) |  | size in pixels for this spacer. |
| expand | [bool](#bool) |  | expand to fill available space. |






<a name="hyprpanel-module-v1-Systray"></a>

### Systray



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| icon_size | [uint32](#uint32) |  | size in pixels for icons in the systray. |
| menu_icon_size | [uint32](#uint32) |  | size in pixels for menu icons. Currently unused because GNOME developers hate user/developer choice. |
| auto_hide_statuses | [Systray.Status](#hyprpanel-module-v1-Systray-Status) | repeated | list of statuses that should be auto-hidden. |
| auto_hide_delay | [google.protobuf.Duration](#google-protobuf-Duration) |  | delay before new (or status-changed) icons are auto-hidden (format &#34;4s&#34;, zero to disable). |
| pinned | [string](#string) | repeated | list of SNI IDs that should never be hidden. There&#39;s no convention for ID values - if you want to collect IDs, start hyprpanel with LOG_LEVEL_DEBUG and look for SNI registration events. |
| modules | [SystrayModule](#hyprpanel-module-v1-SystrayModule) | repeated | list of modules to dislpay in systray. Currently supported modules: [&#34;audio&#34;, &#34;power&#34;] |






<a name="hyprpanel-module-v1-SystrayModule"></a>

### SystrayModule



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| audio | [Audio](#hyprpanel-module-v1-Audio) |  |  |
| power | [Power](#hyprpanel-module-v1-Power) |  |  |






<a name="hyprpanel-module-v1-Taskbar"></a>

### Taskbar



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| icon_size | [uint32](#uint32) |  | size in pixels for task icons. |
| active_workspace_only | [bool](#bool) |  | show only tasks from the current workspace. |
| active_monitor_only | [bool](#bool) |  | show only tasks from the monitor the panel is running on. |
| group_tasks | [bool](#bool) |  | group tasks for the same application into a single icon. Scroll wheel cycles tasks. |
| hide_indicators | [bool](#bool) |  | if you&#39;re not using pinned tasks, you may wish to hide the running task indicators. |
| expand | [bool](#bool) |  | expand this module to fill available space in the panel. |
| max_size | [uint32](#uint32) |  | maximum size in pixels for this module. Zero means no limit. |
| pinned | [string](#string) | repeated | list of window classes that should always be displayed on the taskbar. Allows the taskbar to act as a launcher. |
| preview_width | [uint32](#uint32) |  | width in pixels for task preview windows. |





 


<a name="hyprpanel-module-v1-Position"></a>

### Position


| Name | Number | Description |
| ---- | ------ | ----------- |
| POSITION_UNSPECIFIED | 0 |  |
| POSITION_TOP_LEFT | 1 |  |
| POSITION_TOP | 2 |  |
| POSITION_TOP_RIGHT | 3 |  |
| POSITION_RIGHT | 4 |  |
| POSITION_BOTTOM_RIGHT | 5 |  |
| POSITION_BOTTOM | 6 |  |
| POSITION_BOTTOM_LEFT | 7 |  |
| POSITION_LEFT | 8 |  |
| POSITION_CENTER | 9 |  |



<a name="hyprpanel-module-v1-Systray-Status"></a>

### Systray.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_UNSPECIFIED | 0 |  |
| STATUS_PASSIVE | 1 |  |
| STATUS_ACTIVE | 2 |  |
| STATUS_NEEDS_ATTENTION | 3 |  |


 

 

 



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

