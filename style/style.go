package style

import (
	_ "embed"
	"os"
)

const (
	PanelID                = `panel`
	PagerID                = `pager`
	TaskbarID              = `taskbar`
	SystrayID              = `systray`
	AudioID                = `audio`
	PowerID                = `power`
	ClockID                = `clock`
	ClockTimeID            = `clockTime`
	ClockDateID            = `clockDate`
	ClockCalendarID        = `clockCalendar`
	SessionID              = `session`
	SessionOverlayID       = `sessionOverlay`
	SpacerID               = `spacer`
	NotificationsID        = `notifications`
	NotificationsOverlayID = `notificationsOverlay`
	HudID                  = `hud`
	HudOverlayID           = `hudOverlay`

	ModuleClass                  = `module`
	WorkspaceClass               = `workspace`
	WorkspaceLabelClass          = `workspaceLabel`
	ClientClass                  = `client`
	LiveClass                    = `live`
	ActiveClass                  = `active`
	HoverClass                   = `hover`
	IndicatorClass               = `indicator`
	NotificationItemClass        = `notification`
	NotificationItemSummaryClass = `notificationSummary`
	NotificationItemBodyClass    = `notificationBody`
	NotificationItemActionsClass = `notificationActions`
	NotificationItemIconClass    = `notificationIcon`
	HudNotificationClass         = `hudNotification`
	HudIconClass                 = `hudIcon`
	HudTitleClass                = `hudTitle`
	HudBodyClass                 = `hudBody`
	HudPercentClass              = `hudPercent`
	HudGaugeClass                = `hudGauge`

	TopClass    = `top`
	RightClass  = `right`
	BottomClass = `bottom`
	LeftClass   = `left`

	HorizontalClass = `horizontal`
	VerticalClass   = `vertical`
)

//go:embed default.css
var Default []byte

func Load(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return b, nil
}
