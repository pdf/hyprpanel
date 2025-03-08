# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [hyprpanel/v1/hyprpanel.proto](#hyprpanel_v1_hyprpanel-proto)
    - [AppInfo](#hyprpanel-v1-AppInfo)
    - [AppInfo.Action](#hyprpanel-v1-AppInfo-Action)
    - [HostServiceAudioSinkMuteToggleRequest](#hyprpanel-v1-HostServiceAudioSinkMuteToggleRequest)
    - [HostServiceAudioSinkMuteToggleResponse](#hyprpanel-v1-HostServiceAudioSinkMuteToggleResponse)
    - [HostServiceAudioSinkVolumeAdjustRequest](#hyprpanel-v1-HostServiceAudioSinkVolumeAdjustRequest)
    - [HostServiceAudioSinkVolumeAdjustResponse](#hyprpanel-v1-HostServiceAudioSinkVolumeAdjustResponse)
    - [HostServiceAudioSourceMuteToggleRequest](#hyprpanel-v1-HostServiceAudioSourceMuteToggleRequest)
    - [HostServiceAudioSourceMuteToggleResponse](#hyprpanel-v1-HostServiceAudioSourceMuteToggleResponse)
    - [HostServiceAudioSourceVolumeAdjustRequest](#hyprpanel-v1-HostServiceAudioSourceVolumeAdjustRequest)
    - [HostServiceAudioSourceVolumeAdjustResponse](#hyprpanel-v1-HostServiceAudioSourceVolumeAdjustResponse)
    - [HostServiceBrightnessAdjustRequest](#hyprpanel-v1-HostServiceBrightnessAdjustRequest)
    - [HostServiceBrightnessAdjustResponse](#hyprpanel-v1-HostServiceBrightnessAdjustResponse)
    - [HostServiceCaptureFrameRequest](#hyprpanel-v1-HostServiceCaptureFrameRequest)
    - [HostServiceCaptureFrameResponse](#hyprpanel-v1-HostServiceCaptureFrameResponse)
    - [HostServiceExecRequest](#hyprpanel-v1-HostServiceExecRequest)
    - [HostServiceExecResponse](#hyprpanel-v1-HostServiceExecResponse)
    - [HostServiceFindApplicationRequest](#hyprpanel-v1-HostServiceFindApplicationRequest)
    - [HostServiceFindApplicationResponse](#hyprpanel-v1-HostServiceFindApplicationResponse)
    - [HostServiceNotificationActionRequest](#hyprpanel-v1-HostServiceNotificationActionRequest)
    - [HostServiceNotificationActionResponse](#hyprpanel-v1-HostServiceNotificationActionResponse)
    - [HostServiceNotificationClosedRequest](#hyprpanel-v1-HostServiceNotificationClosedRequest)
    - [HostServiceNotificationClosedResponse](#hyprpanel-v1-HostServiceNotificationClosedResponse)
    - [HostServiceSystrayActivateRequest](#hyprpanel-v1-HostServiceSystrayActivateRequest)
    - [HostServiceSystrayActivateResponse](#hyprpanel-v1-HostServiceSystrayActivateResponse)
    - [HostServiceSystrayMenuAboutToShowRequest](#hyprpanel-v1-HostServiceSystrayMenuAboutToShowRequest)
    - [HostServiceSystrayMenuAboutToShowResponse](#hyprpanel-v1-HostServiceSystrayMenuAboutToShowResponse)
    - [HostServiceSystrayMenuContextActivateRequest](#hyprpanel-v1-HostServiceSystrayMenuContextActivateRequest)
    - [HostServiceSystrayMenuContextActivateResponse](#hyprpanel-v1-HostServiceSystrayMenuContextActivateResponse)
    - [HostServiceSystrayMenuEventRequest](#hyprpanel-v1-HostServiceSystrayMenuEventRequest)
    - [HostServiceSystrayMenuEventResponse](#hyprpanel-v1-HostServiceSystrayMenuEventResponse)
    - [HostServiceSystrayScrollRequest](#hyprpanel-v1-HostServiceSystrayScrollRequest)
    - [HostServiceSystrayScrollResponse](#hyprpanel-v1-HostServiceSystrayScrollResponse)
    - [HostServiceSystraySecondaryActivateRequest](#hyprpanel-v1-HostServiceSystraySecondaryActivateRequest)
    - [HostServiceSystraySecondaryActivateResponse](#hyprpanel-v1-HostServiceSystraySecondaryActivateResponse)
    - [ImageNRGBA](#hyprpanel-v1-ImageNRGBA)
    - [PanelServiceCloseRequest](#hyprpanel-v1-PanelServiceCloseRequest)
    - [PanelServiceCloseResponse](#hyprpanel-v1-PanelServiceCloseResponse)
    - [PanelServiceInitRequest](#hyprpanel-v1-PanelServiceInitRequest)
    - [PanelServiceInitResponse](#hyprpanel-v1-PanelServiceInitResponse)
    - [PanelServiceNotificationCloseRequest](#hyprpanel-v1-PanelServiceNotificationCloseRequest)
    - [PanelServiceNotificationCloseResponse](#hyprpanel-v1-PanelServiceNotificationCloseResponse)
    - [PanelServiceNotifyRequest](#hyprpanel-v1-PanelServiceNotifyRequest)
    - [PanelServiceNotifyResponse](#hyprpanel-v1-PanelServiceNotifyResponse)
  
    - [NotificationClosedReason](#hyprpanel-v1-NotificationClosedReason)
    - [SystrayMenuEvent](#hyprpanel-v1-SystrayMenuEvent)
    - [SystrayScrollOrientation](#hyprpanel-v1-SystrayScrollOrientation)
  
    - [HostService](#hyprpanel-v1-HostService)
    - [PanelService](#hyprpanel-v1-PanelService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="hyprpanel_v1_hyprpanel-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## hyprpanel/v1/hyprpanel.proto



<a name="hyprpanel-v1-AppInfo"></a>

### AppInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| desktop_file | [string](#string) |  |  |
| name | [string](#string) |  |  |
| icon | [string](#string) |  |  |
| try_exec | [string](#string) |  |  |
| exec | [string](#string) | repeated |  |
| raw_exec | [string](#string) |  |  |
| path | [string](#string) |  |  |
| startup_wm_class | [string](#string) |  |  |
| terminal | [bool](#bool) |  |  |
| actions | [AppInfo.Action](#hyprpanel-v1-AppInfo-Action) | repeated |  |






<a name="hyprpanel-v1-AppInfo-Action"></a>

### AppInfo.Action



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| icon | [string](#string) |  |  |
| exec | [string](#string) | repeated |  |
| raw_exec | [string](#string) |  |  |






<a name="hyprpanel-v1-HostServiceAudioSinkMuteToggleRequest"></a>

### HostServiceAudioSinkMuteToggleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="hyprpanel-v1-HostServiceAudioSinkMuteToggleResponse"></a>

### HostServiceAudioSinkMuteToggleResponse







<a name="hyprpanel-v1-HostServiceAudioSinkVolumeAdjustRequest"></a>

### HostServiceAudioSinkVolumeAdjustRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| direction | [hyprpanel.event.v1.Direction](#hyprpanel-event-v1-Direction) |  |  |






<a name="hyprpanel-v1-HostServiceAudioSinkVolumeAdjustResponse"></a>

### HostServiceAudioSinkVolumeAdjustResponse







<a name="hyprpanel-v1-HostServiceAudioSourceMuteToggleRequest"></a>

### HostServiceAudioSourceMuteToggleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="hyprpanel-v1-HostServiceAudioSourceMuteToggleResponse"></a>

### HostServiceAudioSourceMuteToggleResponse







<a name="hyprpanel-v1-HostServiceAudioSourceVolumeAdjustRequest"></a>

### HostServiceAudioSourceVolumeAdjustRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| direction | [hyprpanel.event.v1.Direction](#hyprpanel-event-v1-Direction) |  |  |






<a name="hyprpanel-v1-HostServiceAudioSourceVolumeAdjustResponse"></a>

### HostServiceAudioSourceVolumeAdjustResponse







<a name="hyprpanel-v1-HostServiceBrightnessAdjustRequest"></a>

### HostServiceBrightnessAdjustRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dev_name | [string](#string) |  |  |
| direction | [hyprpanel.event.v1.Direction](#hyprpanel-event-v1-Direction) |  |  |






<a name="hyprpanel-v1-HostServiceBrightnessAdjustResponse"></a>

### HostServiceBrightnessAdjustResponse







<a name="hyprpanel-v1-HostServiceCaptureFrameRequest"></a>

### HostServiceCaptureFrameRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | [uint64](#uint64) |  |  |
| width | [int32](#int32) |  |  |
| height | [int32](#int32) |  |  |






<a name="hyprpanel-v1-HostServiceCaptureFrameResponse"></a>

### HostServiceCaptureFrameResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| image | [ImageNRGBA](#hyprpanel-v1-ImageNRGBA) |  |  |






<a name="hyprpanel-v1-HostServiceExecRequest"></a>

### HostServiceExecRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| action | [AppInfo.Action](#hyprpanel-v1-AppInfo-Action) |  |  |






<a name="hyprpanel-v1-HostServiceExecResponse"></a>

### HostServiceExecResponse







<a name="hyprpanel-v1-HostServiceFindApplicationRequest"></a>

### HostServiceFindApplicationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| query | [string](#string) |  |  |






<a name="hyprpanel-v1-HostServiceFindApplicationResponse"></a>

### HostServiceFindApplicationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| app_info | [AppInfo](#hyprpanel-v1-AppInfo) |  |  |






<a name="hyprpanel-v1-HostServiceNotificationActionRequest"></a>

### HostServiceNotificationActionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [uint32](#uint32) |  |  |
| action_key | [string](#string) |  |  |






<a name="hyprpanel-v1-HostServiceNotificationActionResponse"></a>

### HostServiceNotificationActionResponse







<a name="hyprpanel-v1-HostServiceNotificationClosedRequest"></a>

### HostServiceNotificationClosedRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [uint32](#uint32) |  |  |
| reason | [NotificationClosedReason](#hyprpanel-v1-NotificationClosedReason) |  |  |






<a name="hyprpanel-v1-HostServiceNotificationClosedResponse"></a>

### HostServiceNotificationClosedResponse







<a name="hyprpanel-v1-HostServiceSystrayActivateRequest"></a>

### HostServiceSystrayActivateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bus_name | [string](#string) |  |  |
| x | [int32](#int32) |  |  |
| y | [int32](#int32) |  |  |






<a name="hyprpanel-v1-HostServiceSystrayActivateResponse"></a>

### HostServiceSystrayActivateResponse







<a name="hyprpanel-v1-HostServiceSystrayMenuAboutToShowRequest"></a>

### HostServiceSystrayMenuAboutToShowRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bus_name | [string](#string) |  |  |
| menu_item_id | [string](#string) |  |  |






<a name="hyprpanel-v1-HostServiceSystrayMenuAboutToShowResponse"></a>

### HostServiceSystrayMenuAboutToShowResponse







<a name="hyprpanel-v1-HostServiceSystrayMenuContextActivateRequest"></a>

### HostServiceSystrayMenuContextActivateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bus_name | [string](#string) |  |  |
| x | [int32](#int32) |  |  |
| y | [int32](#int32) |  |  |






<a name="hyprpanel-v1-HostServiceSystrayMenuContextActivateResponse"></a>

### HostServiceSystrayMenuContextActivateResponse







<a name="hyprpanel-v1-HostServiceSystrayMenuEventRequest"></a>

### HostServiceSystrayMenuEventRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bus_name | [string](#string) |  |  |
| id | [int32](#int32) |  |  |
| event_id | [SystrayMenuEvent](#hyprpanel-v1-SystrayMenuEvent) |  |  |
| data | [google.protobuf.Any](#google-protobuf-Any) |  |  |
| timestamp | [uint32](#uint32) |  |  |






<a name="hyprpanel-v1-HostServiceSystrayMenuEventResponse"></a>

### HostServiceSystrayMenuEventResponse







<a name="hyprpanel-v1-HostServiceSystrayScrollRequest"></a>

### HostServiceSystrayScrollRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bus_name | [string](#string) |  |  |
| delta | [int32](#int32) |  |  |
| orientation | [SystrayScrollOrientation](#hyprpanel-v1-SystrayScrollOrientation) |  |  |






<a name="hyprpanel-v1-HostServiceSystrayScrollResponse"></a>

### HostServiceSystrayScrollResponse







<a name="hyprpanel-v1-HostServiceSystraySecondaryActivateRequest"></a>

### HostServiceSystraySecondaryActivateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bus_name | [string](#string) |  |  |
| x | [int32](#int32) |  |  |
| y | [int32](#int32) |  |  |






<a name="hyprpanel-v1-HostServiceSystraySecondaryActivateResponse"></a>

### HostServiceSystraySecondaryActivateResponse







<a name="hyprpanel-v1-ImageNRGBA"></a>

### ImageNRGBA



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pixels | [bytes](#bytes) |  |  |
| stride | [uint32](#uint32) |  |  |
| width | [uint32](#uint32) |  |  |
| height | [uint32](#uint32) |  |  |






<a name="hyprpanel-v1-PanelServiceCloseRequest"></a>

### PanelServiceCloseRequest







<a name="hyprpanel-v1-PanelServiceCloseResponse"></a>

### PanelServiceCloseResponse







<a name="hyprpanel-v1-PanelServiceInitRequest"></a>

### PanelServiceInitRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | [uint32](#uint32) |  |  |
| id | [string](#string) |  |  |
| log_level | [hyprpanel.config.v1.LogLevel](#hyprpanel-config-v1-LogLevel) |  |  |
| config | [hyprpanel.config.v1.Panel](#hyprpanel-config-v1-Panel) |  |  |
| stylesheet | [bytes](#bytes) |  |  |






<a name="hyprpanel-v1-PanelServiceInitResponse"></a>

### PanelServiceInitResponse







<a name="hyprpanel-v1-PanelServiceNotificationCloseRequest"></a>

### PanelServiceNotificationCloseRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [uint32](#uint32) |  |  |






<a name="hyprpanel-v1-PanelServiceNotificationCloseResponse"></a>

### PanelServiceNotificationCloseResponse







<a name="hyprpanel-v1-PanelServiceNotifyRequest"></a>

### PanelServiceNotifyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [hyprpanel.event.v1.Event](#hyprpanel-event-v1-Event) |  |  |






<a name="hyprpanel-v1-PanelServiceNotifyResponse"></a>

### PanelServiceNotifyResponse






 


<a name="hyprpanel-v1-NotificationClosedReason"></a>

### NotificationClosedReason


| Name | Number | Description |
| ---- | ------ | ----------- |
| NOTIFICATION_CLOSED_REASON_UNSPECIFIED | 0 |  |
| NOTIFICATION_CLOSED_REASON_EXPIRED | 1 |  |
| NOTIFICATION_CLOSED_REASON_DISMISSED | 2 |  |
| NOTIFICATION_CLOSED_REASON_SIGNAL | 3 |  |



<a name="hyprpanel-v1-SystrayMenuEvent"></a>

### SystrayMenuEvent


| Name | Number | Description |
| ---- | ------ | ----------- |
| SYSTRAY_MENU_EVENT_UNSPECIFIED | 0 |  |
| SYSTRAY_MENU_EVENT_CLICKED | 1 |  |
| SYSTRAY_MENU_EVENT_HOVERED | 2 |  |



<a name="hyprpanel-v1-SystrayScrollOrientation"></a>

### SystrayScrollOrientation


| Name | Number | Description |
| ---- | ------ | ----------- |
| SYSTRAY_SCROLL_ORIENTATION_UNSPECIFIED | 0 |  |
| SYSTRAY_SCROLL_ORIENTATION_VERTICAL | 1 |  |
| SYSTRAY_SCROLL_ORIENTATION_HORIZONTAL | 2 |  |


 

 


<a name="hyprpanel-v1-HostService"></a>

### HostService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Exec | [HostServiceExecRequest](#hyprpanel-v1-HostServiceExecRequest) | [HostServiceExecResponse](#hyprpanel-v1-HostServiceExecResponse) |  |
| FindApplication | [HostServiceFindApplicationRequest](#hyprpanel-v1-HostServiceFindApplicationRequest) | [HostServiceFindApplicationResponse](#hyprpanel-v1-HostServiceFindApplicationResponse) |  |
| SystrayActivate | [HostServiceSystrayActivateRequest](#hyprpanel-v1-HostServiceSystrayActivateRequest) | [HostServiceSystrayActivateResponse](#hyprpanel-v1-HostServiceSystrayActivateResponse) |  |
| SystraySecondaryActivate | [HostServiceSystraySecondaryActivateRequest](#hyprpanel-v1-HostServiceSystraySecondaryActivateRequest) | [HostServiceSystraySecondaryActivateResponse](#hyprpanel-v1-HostServiceSystraySecondaryActivateResponse) |  |
| SystrayScroll | [HostServiceSystrayScrollRequest](#hyprpanel-v1-HostServiceSystrayScrollRequest) | [HostServiceSystrayScrollResponse](#hyprpanel-v1-HostServiceSystrayScrollResponse) |  |
| SystrayMenuContextActivate | [HostServiceSystrayMenuContextActivateRequest](#hyprpanel-v1-HostServiceSystrayMenuContextActivateRequest) | [HostServiceSystrayMenuContextActivateResponse](#hyprpanel-v1-HostServiceSystrayMenuContextActivateResponse) |  |
| SystrayMenuAboutToShow | [HostServiceSystrayMenuAboutToShowRequest](#hyprpanel-v1-HostServiceSystrayMenuAboutToShowRequest) | [HostServiceSystrayMenuAboutToShowResponse](#hyprpanel-v1-HostServiceSystrayMenuAboutToShowResponse) |  |
| SystrayMenuEvent | [HostServiceSystrayMenuEventRequest](#hyprpanel-v1-HostServiceSystrayMenuEventRequest) | [HostServiceSystrayMenuEventResponse](#hyprpanel-v1-HostServiceSystrayMenuEventResponse) |  |
| NotificationClosed | [HostServiceNotificationClosedRequest](#hyprpanel-v1-HostServiceNotificationClosedRequest) | [HostServiceNotificationClosedResponse](#hyprpanel-v1-HostServiceNotificationClosedResponse) |  |
| NotificationAction | [HostServiceNotificationActionRequest](#hyprpanel-v1-HostServiceNotificationActionRequest) | [HostServiceNotificationActionResponse](#hyprpanel-v1-HostServiceNotificationActionResponse) |  |
| AudioSinkVolumeAdjust | [HostServiceAudioSinkVolumeAdjustRequest](#hyprpanel-v1-HostServiceAudioSinkVolumeAdjustRequest) | [HostServiceAudioSinkVolumeAdjustResponse](#hyprpanel-v1-HostServiceAudioSinkVolumeAdjustResponse) |  |
| AudioSinkMuteToggle | [HostServiceAudioSinkMuteToggleRequest](#hyprpanel-v1-HostServiceAudioSinkMuteToggleRequest) | [HostServiceAudioSinkMuteToggleResponse](#hyprpanel-v1-HostServiceAudioSinkMuteToggleResponse) |  |
| AudioSourceVolumeAdjust | [HostServiceAudioSourceVolumeAdjustRequest](#hyprpanel-v1-HostServiceAudioSourceVolumeAdjustRequest) | [HostServiceAudioSourceVolumeAdjustResponse](#hyprpanel-v1-HostServiceAudioSourceVolumeAdjustResponse) |  |
| AudioSourceMuteToggle | [HostServiceAudioSourceMuteToggleRequest](#hyprpanel-v1-HostServiceAudioSourceMuteToggleRequest) | [HostServiceAudioSourceMuteToggleResponse](#hyprpanel-v1-HostServiceAudioSourceMuteToggleResponse) |  |
| BrightnessAdjust | [HostServiceBrightnessAdjustRequest](#hyprpanel-v1-HostServiceBrightnessAdjustRequest) | [HostServiceBrightnessAdjustResponse](#hyprpanel-v1-HostServiceBrightnessAdjustResponse) |  |
| CaptureFrame | [HostServiceCaptureFrameRequest](#hyprpanel-v1-HostServiceCaptureFrameRequest) | [HostServiceCaptureFrameResponse](#hyprpanel-v1-HostServiceCaptureFrameResponse) |  |


<a name="hyprpanel-v1-PanelService"></a>

### PanelService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Init | [PanelServiceInitRequest](#hyprpanel-v1-PanelServiceInitRequest) | [PanelServiceInitResponse](#hyprpanel-v1-PanelServiceInitResponse) |  |
| Notify | [PanelServiceNotifyRequest](#hyprpanel-v1-PanelServiceNotifyRequest) | [PanelServiceNotifyResponse](#hyprpanel-v1-PanelServiceNotifyResponse) |  |
| Close | [PanelServiceCloseRequest](#hyprpanel-v1-PanelServiceCloseRequest) | [PanelServiceCloseResponse](#hyprpanel-v1-PanelServiceCloseResponse) |  |

 



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

