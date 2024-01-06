package dbus

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
)

const (
	fdoName                   = `org.freedesktop.DBus`
	fdoPath                   = dbus.ObjectPath(`/org/freedesktop/DBus`)
	fdoSignalNameOwnerChanged = fdoName + `.NameOwnerChanged`
	fdoIntrospectableName     = fdoName + `.Introspectable`

	fdoPropertiesName                    = fdoName + `.Properties`
	fdoPropertiesMethodGetAll            = fdoPropertiesName + `.GetAll`
	fdoPropertiesMemberPropertiesChanged = `PropertiesChanged`
	fdoPropertiesSignalPropertiesChanged = fdoPropertiesName + `.` + fdoPropertiesMemberPropertiesChanged

	fdoLogindName                       = `org.freedesktop.login1`
	fdoLogindSessionName                = fdoLogindName + `.Session`
	fdoLogindSessionPath                = `/org/freedesktop/login1/session/auto`
	fdoLogindSessionMethodSetBrightness = fdoLogindSessionName + `.SetBrightness`

	fdoSystemdName       = `org.freedesktop.systemd1`
	fdoSystemdUnitPath   = `/org/freedesktop/systemd1/unit`
	fdoSystemdDeviceName = `org.freedesktop.systemd1.Device`

	fdoUPowerName                   = `org.freedesktop.UPower`
	fdoUPowerPath                   = `/org/freedesktop/UPower`
	fdoUPowerMethodGetDisplayDevice = fdoUPowerName + `.GetDisplayDevice`

	fdoUPowerDeviceName                = fdoUPowerName + `.Device`
	fdoUPowerDevicePropertyVendor      = `Vendor`
	fdoUPowerDevicePropertyModel       = `Model`
	fdoUPowerDevicePropertyType        = `Type`
	fdoUPowerDevicePropertyPowerSupply = `PowerSupply`
	fdoUPowerDevicePropertyOnline      = `Online`
	fdoUPowerDevicePropertyTimeToEmpty = `TimeToEmpty`
	fdoUPowerDevicePropertyTimeToFull  = `TimeToFull`
	fdoUPowerDevicePropertyPercentage  = `Percentage`
	fdoUPowerDevicePropertyIsPresent   = `IsPresent`
	fdoUPowerDevicePropertyState       = `State`
	fdoUPowerDevicePropertyIconName    = `IconName`
	fdoUPowerDevicePropertyEnergy      = `Energy`
	fdoUPowerDevicePropertyEnergyEmpty = `EnergyEmpty`
	fdoUPowerDevicePropertyEnergyFull  = `EnergyFull`
)

func systemdUnitToObjectPath(unitName string) (dbus.ObjectPath, error) {
	unitName = strings.ReplaceAll(unitName, `-`, `\x2d`)

	var result strings.Builder
	for _, r := range unitName {
		if !isValidObjectPathChar(r) {
			if _, err := result.WriteString(fmt.Sprintf("%d", r)); err != nil {
				return ``, err
			}
			continue
		}

		if _, err := result.WriteRune(r); err != nil {
			return ``, err
		}
	}

	return dbus.ObjectPath(result.String()), nil
}
