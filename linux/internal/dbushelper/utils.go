//go:build linux

package dbushelper

import "github.com/godbus/dbus/v5"

// ListBusNames returns a list of bus names from the provided DBus connection.
func ListBusNames(conn *dbus.Conn) ([]string, error) {
	var names []string

	if err := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").
		Call("org.freedesktop.DBus.ListNames", 0).
		Store(&names); err != nil {
		return nil, err
	}

	return names, nil
}
