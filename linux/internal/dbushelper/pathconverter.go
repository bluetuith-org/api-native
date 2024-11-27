//go:build linux

package dbushelper

import (
	bluetooth "github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/godbus/dbus/v5"
	"github.com/puzpuzpuz/xsync/v3"
)

// DbusPathType represents the type of DBus path in the Bluez DBus service.
// For example, adapter paths will have a path type of DbusPathAdapter and will
// be mapped to an adapter address (/org/bluez/hci0 => DBusPathAdapter).
// For other DBus path types like DbusPathObexSession and DbusPathObexTransfer,
// their paths will be mapped to device addresses.
type DbusPathType int

// The different Bluez DBus path types.
const (
	DbusPathDevice DbusPathType = iota
	DbusPathAdapter
	DbusPathObexSession
	DbusPathObexTransfer
)

// DbusPath holds the Bluez DBus path and its type.
type DbusPath struct {
	pathType DbusPathType
	path     dbus.ObjectPath
}

// DbusPathConverter holds a list of Bluez DBus paths and maps them
// to their respective Bluetooth addresses.
type DbusPathConverter struct {
	paths *xsync.MapOf[DbusPath, bluetooth.MacAddress]
}

// PathConverter is used to obtain respective Bluetooth addresses that are mapped to
// Bluez DBus paths. This is mainly used to identify adapters and devices.
var PathConverter = DbusPathConverter{paths: xsync.NewMapOf[DbusPath, bluetooth.MacAddress]()}

// AddDbusPath adds a mapping of a Bluez DBus path and a Bluetooth address to the path converter.
func (d *DbusPathConverter) AddDbusPath(pathType DbusPathType, path dbus.ObjectPath, address bluetooth.MacAddress) {
	d.paths.Store(DbusPath{pathType: pathType, path: path}, address)
}

// RemoveDbusPath removes a mapping of a Bluez DBus path and a Bluetooth address from the path converter.
func (d *DbusPathConverter) RemoveDbusPath(pathType DbusPathType, path dbus.ObjectPath) {
	d.paths.Delete(DbusPath{pathType: pathType, path: path})
}

// Address returns a Bluetooth address that is mapped to the provided Bluez DBus path.
func (d *DbusPathConverter) Address(pathType DbusPathType, path dbus.ObjectPath) (bluetooth.MacAddress, bool) {
	return d.paths.Load(DbusPath{pathType: pathType, path: path})
}

// DbusPath returns a Bluez DBus path that is mapped to the provided Bluetooth address.
func (d *DbusPathConverter) DbusPath(pathType DbusPathType, address bluetooth.MacAddress) (dbus.ObjectPath, bool) {
	var dpath dbus.ObjectPath

	d.paths.Range(func(p DbusPath, addr bluetooth.MacAddress) bool {
		if address == addr && p.pathType == pathType {
			dpath = p.path

			return false
		}

		return true
	})

	return dpath, dpath != ""
}
