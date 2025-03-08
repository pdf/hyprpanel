# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [hyprpanel/event/v1/event.proto](#hyprpanel_event_v1_event-proto)
    - [AudioSinkChangeValue](#hyprpanel-event-v1-AudioSinkChangeValue)
    - [AudioSinkMuteToggle](#hyprpanel-event-v1-AudioSinkMuteToggle)
    - [AudioSinkVolumeAdjust](#hyprpanel-event-v1-AudioSinkVolumeAdjust)
    - [AudioSourceChangeValue](#hyprpanel-event-v1-AudioSourceChangeValue)
    - [AudioSourceMuteToggle](#hyprpanel-event-v1-AudioSourceMuteToggle)
    - [AudioSourceVolumeAdjust](#hyprpanel-event-v1-AudioSourceVolumeAdjust)
    - [BrightnessAdjustValue](#hyprpanel-event-v1-BrightnessAdjustValue)
    - [BrightnessChangeValue](#hyprpanel-event-v1-BrightnessChangeValue)
    - [Event](#hyprpanel-event-v1-Event)
    - [HudNotificationValue](#hyprpanel-event-v1-HudNotificationValue)
    - [HyprActiveWindowValue](#hyprpanel-event-v1-HyprActiveWindowValue)
    - [HyprCreateWorkspaceV2Value](#hyprpanel-event-v1-HyprCreateWorkspaceV2Value)
    - [HyprDestroyWorkspaceV2Value](#hyprpanel-event-v1-HyprDestroyWorkspaceV2Value)
    - [HyprMoveWindowV2Value](#hyprpanel-event-v1-HyprMoveWindowV2Value)
    - [HyprMoveWindowValue](#hyprpanel-event-v1-HyprMoveWindowValue)
    - [HyprMoveWorkspaceV2Value](#hyprpanel-event-v1-HyprMoveWorkspaceV2Value)
    - [HyprMoveWorkspaceValue](#hyprpanel-event-v1-HyprMoveWorkspaceValue)
    - [HyprOpenWindowValue](#hyprpanel-event-v1-HyprOpenWindowValue)
    - [HyprRenameWorkspaceValue](#hyprpanel-event-v1-HyprRenameWorkspaceValue)
    - [HyprWorkspaceV2Value](#hyprpanel-event-v1-HyprWorkspaceV2Value)
    - [NotificationValue](#hyprpanel-event-v1-NotificationValue)
    - [NotificationValue.Action](#hyprpanel-event-v1-NotificationValue-Action)
    - [NotificationValue.Hint](#hyprpanel-event-v1-NotificationValue-Hint)
    - [NotificationValue.Pixmap](#hyprpanel-event-v1-NotificationValue-Pixmap)
    - [PowerChangeValue](#hyprpanel-event-v1-PowerChangeValue)
    - [StatusNotifierValue](#hyprpanel-event-v1-StatusNotifierValue)
    - [StatusNotifierValue.Icon](#hyprpanel-event-v1-StatusNotifierValue-Icon)
    - [StatusNotifierValue.Menu](#hyprpanel-event-v1-StatusNotifierValue-Menu)
    - [StatusNotifierValue.Menu.Properties](#hyprpanel-event-v1-StatusNotifierValue-Menu-Properties)
    - [StatusNotifierValue.Pixmap](#hyprpanel-event-v1-StatusNotifierValue-Pixmap)
    - [StatusNotifierValue.Tooltip](#hyprpanel-event-v1-StatusNotifierValue-Tooltip)
    - [UpdateIconValue](#hyprpanel-event-v1-UpdateIconValue)
    - [UpdateMenuValue](#hyprpanel-event-v1-UpdateMenuValue)
    - [UpdateStatusValue](#hyprpanel-event-v1-UpdateStatusValue)
    - [UpdateTitleValue](#hyprpanel-event-v1-UpdateTitleValue)
    - [UpdateTooltipValue](#hyprpanel-event-v1-UpdateTooltipValue)
  
    - [Direction](#hyprpanel-event-v1-Direction)
    - [EventKind](#hyprpanel-event-v1-EventKind)
    - [PowerState](#hyprpanel-event-v1-PowerState)
    - [PowerType](#hyprpanel-event-v1-PowerType)
  
- [Scalar Value Types](#scalar-value-types)



<a name="hyprpanel_event_v1_event-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## hyprpanel/event/v1/event.proto



<a name="hyprpanel-event-v1-AudioSinkChangeValue"></a>

### AudioSinkChangeValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| volume | [int32](#int32) |  |  |
| percent | [double](#double) |  |  |
| percent_max | [double](#double) |  |  |
| mute | [bool](#bool) |  |  |
| default | [bool](#bool) |  |  |






<a name="hyprpanel-event-v1-AudioSinkMuteToggle"></a>

### AudioSinkMuteToggle



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="hyprpanel-event-v1-AudioSinkVolumeAdjust"></a>

### AudioSinkVolumeAdjust



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| direction | [Direction](#hyprpanel-event-v1-Direction) |  |  |






<a name="hyprpanel-event-v1-AudioSourceChangeValue"></a>

### AudioSourceChangeValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| volume | [int32](#int32) |  |  |
| percent | [double](#double) |  |  |
| percent_max | [double](#double) |  |  |
| mute | [bool](#bool) |  |  |
| default | [bool](#bool) |  |  |






<a name="hyprpanel-event-v1-AudioSourceMuteToggle"></a>

### AudioSourceMuteToggle



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="hyprpanel-event-v1-AudioSourceVolumeAdjust"></a>

### AudioSourceVolumeAdjust



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| direction | [Direction](#hyprpanel-event-v1-Direction) |  |  |






<a name="hyprpanel-event-v1-BrightnessAdjustValue"></a>

### BrightnessAdjustValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dev_name | [string](#string) |  |  |
| direction | [Direction](#hyprpanel-event-v1-Direction) |  |  |






<a name="hyprpanel-event-v1-BrightnessChangeValue"></a>

### BrightnessChangeValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| brightness | [int32](#int32) |  |  |
| brightness_max | [int32](#int32) |  |  |






<a name="hyprpanel-event-v1-Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [EventKind](#hyprpanel-event-v1-EventKind) |  |  |
| data | [google.protobuf.Any](#google-protobuf-Any) |  |  |






<a name="hyprpanel-event-v1-HudNotificationValue"></a>

### HudNotificationValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| icon | [string](#string) |  |  |
| icon_symbolic | [bool](#bool) |  |  |
| title | [string](#string) |  |  |
| body | [string](#string) |  |  |
| percent | [double](#double) |  |  |
| percent_max | [double](#double) |  |  |






<a name="hyprpanel-event-v1-HyprActiveWindowValue"></a>

### HyprActiveWindowValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| class | [string](#string) |  |  |
| title | [string](#string) |  |  |






<a name="hyprpanel-event-v1-HyprCreateWorkspaceV2Value"></a>

### HyprCreateWorkspaceV2Value



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |
| name | [string](#string) |  |  |






<a name="hyprpanel-event-v1-HyprDestroyWorkspaceV2Value"></a>

### HyprDestroyWorkspaceV2Value



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |
| name | [string](#string) |  |  |






<a name="hyprpanel-event-v1-HyprMoveWindowV2Value"></a>

### HyprMoveWindowV2Value



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | [string](#string) |  |  |
| workspace_id | [int32](#int32) |  |  |
| workspace_name | [string](#string) |  |  |






<a name="hyprpanel-event-v1-HyprMoveWindowValue"></a>

### HyprMoveWindowValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | [string](#string) |  |  |
| workspace_name | [string](#string) |  |  |






<a name="hyprpanel-event-v1-HyprMoveWorkspaceV2Value"></a>

### HyprMoveWorkspaceV2Value



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |
| name | [string](#string) |  |  |
| monitor | [string](#string) |  |  |






<a name="hyprpanel-event-v1-HyprMoveWorkspaceValue"></a>

### HyprMoveWorkspaceValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| monitor | [string](#string) |  |  |






<a name="hyprpanel-event-v1-HyprOpenWindowValue"></a>

### HyprOpenWindowValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | [string](#string) |  |  |
| workspace_name | [string](#string) |  |  |
| class | [string](#string) |  |  |
| title | [string](#string) |  |  |






<a name="hyprpanel-event-v1-HyprRenameWorkspaceValue"></a>

### HyprRenameWorkspaceValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |
| name | [string](#string) |  |  |






<a name="hyprpanel-event-v1-HyprWorkspaceV2Value"></a>

### HyprWorkspaceV2Value



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |
| name | [string](#string) |  |  |






<a name="hyprpanel-event-v1-NotificationValue"></a>

### NotificationValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [uint32](#uint32) |  |  |
| app_name | [string](#string) |  |  |
| replaces_id | [uint32](#uint32) |  |  |
| app_icon | [string](#string) |  |  |
| summary | [string](#string) |  |  |
| body | [string](#string) |  |  |
| actions | [NotificationValue.Action](#hyprpanel-event-v1-NotificationValue-Action) | repeated |  |
| hints | [NotificationValue.Hint](#hyprpanel-event-v1-NotificationValue-Hint) | repeated |  |
| timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="hyprpanel-event-v1-NotificationValue-Action"></a>

### NotificationValue.Action



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="hyprpanel-event-v1-NotificationValue-Hint"></a>

### NotificationValue.Hint



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [google.protobuf.Any](#google-protobuf-Any) |  |  |






<a name="hyprpanel-event-v1-NotificationValue-Pixmap"></a>

### NotificationValue.Pixmap



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| width | [int32](#int32) |  | Width of image in pixels |
| height | [int32](#int32) |  | Height of image in pixels |
| row_stride | [int32](#int32) |  | Distance in bytes between row starts |
| has_alpha | [bool](#bool) |  | Whether the image has an alpha channel |
| bits_per_sample | [int32](#int32) |  | Must always be 8 |
| channels | [int32](#int32) |  | If has_alpha is TRUE, must be 4, otherwise 3 |
| data | [bytes](#bytes) |  | The image data, in RGB byte order |






<a name="hyprpanel-event-v1-PowerChangeValue"></a>

### PowerChangeValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| vendor | [string](#string) |  |  |
| model | [string](#string) |  |  |
| type | [PowerType](#hyprpanel-event-v1-PowerType) |  |  |
| power_supply | [bool](#bool) |  |  |
| online | [bool](#bool) |  |  |
| time_to_empty | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| time_to_full | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| percentage | [uint32](#uint32) |  |  |
| state | [PowerState](#hyprpanel-event-v1-PowerState) |  |  |
| icon | [string](#string) |  |  |
| energy | [double](#double) |  |  |
| energy_empty | [double](#double) |  |  |
| energy_full | [double](#double) |  |  |






<a name="hyprpanel-event-v1-StatusNotifierValue"></a>

### StatusNotifierValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bus_name | [string](#string) |  |  |
| object_path | [string](#string) |  |  |
| id | [string](#string) |  |  |
| title | [string](#string) |  |  |
| status | [hyprpanel.module.v1.Systray.Status](#hyprpanel-module-v1-Systray-Status) |  |  |
| tooltip | [StatusNotifierValue.Tooltip](#hyprpanel-event-v1-StatusNotifierValue-Tooltip) |  |  |
| icon | [StatusNotifierValue.Icon](#hyprpanel-event-v1-StatusNotifierValue-Icon) |  |  |
| menu | [StatusNotifierValue.Menu](#hyprpanel-event-v1-StatusNotifierValue-Menu) |  |  |
| menu_revision | [int32](#int32) |  |  |






<a name="hyprpanel-event-v1-StatusNotifierValue-Icon"></a>

### StatusNotifierValue.Icon



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| icon_name | [string](#string) |  |  |
| icon_theme_path | [string](#string) |  |  |
| icon_pixmap | [StatusNotifierValue.Pixmap](#hyprpanel-event-v1-StatusNotifierValue-Pixmap) |  |  |






<a name="hyprpanel-event-v1-StatusNotifierValue-Menu"></a>

### StatusNotifierValue.Menu



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |
| properties | [StatusNotifierValue.Menu.Properties](#hyprpanel-event-v1-StatusNotifierValue-Menu-Properties) |  |  |
| children | [StatusNotifierValue.Menu](#hyprpanel-event-v1-StatusNotifierValue-Menu) | repeated |  |






<a name="hyprpanel-event-v1-StatusNotifierValue-Menu-Properties"></a>

### StatusNotifierValue.Menu.Properties



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| label | [string](#string) |  |  |
| icon_name | [string](#string) |  |  |
| icon_data | [bytes](#bytes) |  |  |
| toggle_state | [int32](#int32) |  |  |
| is_separator | [bool](#bool) |  |  |
| is_parent | [bool](#bool) |  |  |
| is_hidden | [bool](#bool) |  |  |
| is_disabled | [bool](#bool) |  |  |
| is_radio | [bool](#bool) |  |  |
| is_checkbox | [bool](#bool) |  |  |






<a name="hyprpanel-event-v1-StatusNotifierValue-Pixmap"></a>

### StatusNotifierValue.Pixmap



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| width | [int32](#int32) |  |  |
| height | [int32](#int32) |  |  |
| data | [bytes](#bytes) |  |  |






<a name="hyprpanel-event-v1-StatusNotifierValue-Tooltip"></a>

### StatusNotifierValue.Tooltip



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| icon_name | [string](#string) |  |  |
| icon_pixmap | [StatusNotifierValue.Pixmap](#hyprpanel-event-v1-StatusNotifierValue-Pixmap) |  |  |
| title | [string](#string) |  |  |
| body | [string](#string) |  |  |






<a name="hyprpanel-event-v1-UpdateIconValue"></a>

### UpdateIconValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bus_name | [string](#string) |  |  |
| icon | [StatusNotifierValue.Icon](#hyprpanel-event-v1-StatusNotifierValue-Icon) |  |  |






<a name="hyprpanel-event-v1-UpdateMenuValue"></a>

### UpdateMenuValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bus_name | [string](#string) |  |  |
| menu | [StatusNotifierValue.Menu](#hyprpanel-event-v1-StatusNotifierValue-Menu) |  |  |






<a name="hyprpanel-event-v1-UpdateStatusValue"></a>

### UpdateStatusValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bus_name | [string](#string) |  |  |
| status | [hyprpanel.module.v1.Systray.Status](#hyprpanel-module-v1-Systray-Status) |  |  |






<a name="hyprpanel-event-v1-UpdateTitleValue"></a>

### UpdateTitleValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bus_name | [string](#string) |  |  |
| title | [string](#string) |  |  |






<a name="hyprpanel-event-v1-UpdateTooltipValue"></a>

### UpdateTooltipValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bus_name | [string](#string) |  |  |
| tooltip | [StatusNotifierValue.Tooltip](#hyprpanel-event-v1-StatusNotifierValue-Tooltip) |  |  |





 


<a name="hyprpanel-event-v1-Direction"></a>

### Direction


| Name | Number | Description |
| ---- | ------ | ----------- |
| DIRECTION_UNSPECIFIED | 0 |  |
| DIRECTION_UP | 1 |  |
| DIRECTION_DOWN | 2 |  |



<a name="hyprpanel-event-v1-EventKind"></a>

### EventKind


| Name | Number | Description |
| ---- | ------ | ----------- |
| EVENT_KIND_UNSPECIFIED | 0 |  |
| EVENT_KIND_HYPR_WORKSPACE | 1 |  |
| EVENT_KIND_HYPR_FOCUSEDMON | 2 |  |
| EVENT_KIND_HYPR_ACTIVEWINDOW | 4 |  |
| EVENT_KIND_HYPR_ACTIVEWINDOWV2 | 5 |  |
| EVENT_KIND_HYPR_FULLSCREEN | 6 |  |
| EVENT_KIND_HYPR_MONITORREMOVED | 7 |  |
| EVENT_KIND_HYPR_MONITORADDED | 8 |  |
| EVENT_KIND_HYPR_CREATEWORKSPACE | 9 |  |
| EVENT_KIND_HYPR_DESTROYWORKSPACE | 10 |  |
| EVENT_KIND_HYPR_MOVEWORKSPACE | 11 |  |
| EVENT_KIND_HYPR_RENAMEWORKSPACE | 12 |  |
| EVENT_KIND_HYPR_ACTIVESPECIAL | 13 |  |
| EVENT_KIND_HYPR_ACTIVELAYOUT | 14 |  |
| EVENT_KIND_HYPR_OPENWINDOW | 15 |  |
| EVENT_KIND_HYPR_CLOSEWINDOW | 16 |  |
| EVENT_KIND_HYPR_MOVEWINDOW | 17 |  |
| EVENT_KIND_HYPR_OPENLAYER | 18 |  |
| EVENT_KIND_HYPR_CLOSELAYER | 19 |  |
| EVENT_KIND_HYPR_SUBMAP | 20 |  |
| EVENT_KIND_HYPR_CHANGEFLOATINGMODE | 21 |  |
| EVENT_KIND_HYPR_URGENT | 22 |  |
| EVENT_KIND_HYPR_MINIMIZE | 23 |  |
| EVENT_KIND_HYPR_SCREENCAST | 24 |  |
| EVENT_KIND_HYPR_WINDOWTITLE | 25 |  |
| EVENT_KIND_HYPR_IGNOREGROUPLOCK | 26 |  |
| EVENT_KIND_HYPR_EVENTLOCKGROUPS | 27 |  |
| EVENT_KIND_DBUS_REGISTERSTATUSNOTIFIER | 28 |  |
| EVENT_KIND_DBUS_UNREGISTERSTATUSNOTIFIER | 29 |  |
| EVENT_KIND_DBUS_UPDATETITLE | 30 |  |
| EVENT_KIND_DBUS_UPDATETOOLTIP | 31 |  |
| EVENT_KIND_DBUS_UPDATEICON | 32 |  |
| EVENT_KIND_DBUS_UPDATEMENU | 33 |  |
| EVENT_KIND_DBUS_UPDATESTATUS | 34 |  |
| EVENT_KIND_DBUS_NOTIFICATION | 35 |  |
| EVENT_KIND_DBUS_CLOSENOTIFICATION | 36 |  |
| EVENT_KIND_DBUS_BRIGHTNESS_CHANGE | 37 |  |
| EVENT_KIND_DBUS_BRIGHTNESS_ADJUST | 38 |  |
| EVENT_KIND_AUDIO_SINK_NEW | 39 |  |
| EVENT_KIND_AUDIO_SINK_CHANGE | 40 |  |
| EVENT_KIND_AUDIO_SINK_REMOVE | 41 |  |
| EVENT_KIND_AUDIO_SOURCE_NEW | 42 |  |
| EVENT_KIND_AUDIO_SOURCE_CHANGE | 43 |  |
| EVENT_KIND_AUDIO_SOURCE_REMOVE | 44 |  |
| EVENT_KIND_AUDIO_CARD_NEW | 45 |  |
| EVENT_KIND_AUDIO_CARD_CHANGE | 46 |  |
| EVENT_KIND_AUDIO_CARD_REMOVE | 47 |  |
| EVENT_KIND_AUDIO_SINK_VOLUME_ADJUST | 48 |  |
| EVENT_KIND_AUDIO_SINK_MUTE_TOGGLE | 49 |  |
| EVENT_KIND_AUDIO_SOURCE_VOLUME_ADJUST | 50 |  |
| EVENT_KIND_AUDIO_SOURCE_MUTE_TOGGLE | 51 |  |
| EVENT_KIND_HUD_NOTIFY | 52 |  |
| EVENT_KIND_DBUS_POWER_CHANGE | 53 |  |
| EVENT_KIND_HYPR_MOVEWORKSPACEV2 | 54 |  |
| EVENT_KIND_HYPR_MOVEWINDOWV2 | 55 |  |
| EVENT_KIND_HYPR_CREATEWORKSPACEV2 | 56 |  |
| EVENT_KIND_HYPR_DESTROYWORKSPACEV2 | 57 |  |
| EVENT_KIND_HYPR_WORKSPACEV2 | 58 |  |
| EVENT_KIND_EXEC | 59 |  |



<a name="hyprpanel-event-v1-PowerState"></a>

### PowerState


| Name | Number | Description |
| ---- | ------ | ----------- |
| POWER_STATE_UNSPECIFIED | 0 |  |
| POWER_STATE_CHARGING | 1 |  |
| POWER_STATE_DISCHARGING | 2 |  |
| POWER_STATE_EMPTY | 3 |  |
| POWER_STATE_FULLY_CHARGED | 4 |  |
| POWER_STATE_PENDING_CHARGE | 5 |  |
| POWER_STATE_PENDING_DISCHARGE | 6 |  |



<a name="hyprpanel-event-v1-PowerType"></a>

### PowerType


| Name | Number | Description |
| ---- | ------ | ----------- |
| POWER_TYPE_UNSPECIFIED | 0 |  |
| POWER_TYPE_LINE_POWER | 1 |  |
| POWER_TYPE_BATTERY | 2 |  |
| POWER_TYPE_UPS | 3 |  |
| POWER_TYPE_MONITOR | 4 |  |
| POWER_TYPE_MOUSE | 5 |  |
| POWER_TYPE_KEYBOARD | 6 |  |
| POWER_TYPE_PDA | 7 |  |
| POWER_TYPE_PHONE | 8 |  |


 

 

 



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

