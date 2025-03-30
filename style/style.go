// Package style handles application styles.
package style

import (
	_ "embed"
	"os"
)

const (
	// PanelID element identifier.
	PanelID = `panel`
	// PagerID element identifier.
	PagerID = `pager`
	// TaskbarID element identifier.
	TaskbarID = `taskbar`
	// SystrayID element identifier.
	SystrayID = `systray`
	// AudioID element identifier.
	AudioID = `audio`
	// PowerID element identifier.
	PowerID = `power`
	// ClockID element identifier.
	ClockID = `clock`
	// ClockTimeID element identifier.
	ClockTimeID = `clockTime`
	// ClockDateID element identifier.
	ClockDateID = `clockDate`
	// ClockCalendarID element identifier.
	ClockCalendarID = `clockCalendar`
	// SessionID element identifier.
	SessionID = `session`
	// SessionOverlayID element identifier.
	SessionOverlayID = `sessionOverlay`
	// SpacerID element identifier.
	SpacerID = `spacer`
	// NotificationsID element identifier.
	NotificationsID = `notifications`
	// NotificationsOverlayID element identifier.
	NotificationsOverlayID = `notificationsOverlay`
	// HudID element identifier.
	HudID = `hud`
	// HudOverlayID element identifier.
	HudOverlayID = `hudOverlay`

	// ModuleClass class name.
	ModuleClass = `module`
	// WorkspaceClass class name.
	WorkspaceClass = `workspace`
	// WorkspaceLabelClass class name.
	WorkspaceLabelClass = `workspaceLabel`
	// ClientClass class name.
	ClientClass = `client`
	// LiveClass class name.
	LiveClass = `live`
	// ActiveClass class name.
	ActiveClass = `active`
	// HoverClass class name.
	HoverClass = `hover`
	// OverlayClass class name.
	OverlayClass = `overlay`
	// OverlayClass class name.
	DisabledClass = `disabled`
	// IndicatorClass class name.
	IndicatorClass = `indicator`
	// NotificationItemClass class name.
	NotificationItemClass = `notification`
	// NotificationItemSummaryClass class name.
	NotificationItemSummaryClass = `notificationSummary`
	// NotificationItemBodyClass class name.
	NotificationItemBodyClass = `notificationBody`
	// NotificationItemActionsClass class name.
	NotificationItemActionsClass = `notificationActions`
	// NotificationItemIconClass class name.
	NotificationItemIconClass = `notificationIcon`
	// HudNotificationClass class name.
	HudNotificationClass = `hudNotification`
	// HudIconClass class name.
	HudIconClass = `hudIcon`
	// HudTitleClass class name.
	HudTitleClass = `hudTitle`
	// HudBodyClass class name.
	HudBodyClass = `hudBody`
	// HudPercentClass class name.
	HudPercentClass = `hudPercent`
	// HudGaugeClass class name.
	HudGaugeClass = `hudGauge`

	// TooltipImageClass class name.
	TooltipImageClass = `tooltipImage`
	// TooltipSubtitleClass class name.
	TooltipSubtitleClass = `tooltipSubtitle`

	// TopClass class name.
	TopClass = `top`
	// RightClass class name.
	RightClass = `right`
	// BottomClass class name.
	BottomClass = `bottom`
	// LeftClass class name.
	LeftClass = `left`

	// HorizontalClass class name.
	HorizontalClass = `horizontal`
	// VerticalClass class name.
	VerticalClass = `vertical`
)

// Default stylesheet as bytes.
//
//go:embed default.css
var Default []byte

// Load a stylesheet from disk.
func Load(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return b, nil
}
