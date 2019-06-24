package main

import (
	systemd "github.com/coreos/go-systemd/dbus"
	"github.com/godbus/dbus"
)

func newProperty(name string, value interface{}) systemd.Property {
	return systemd.Property{
		Name: name,
		Value: dbus.MakeVariant(value),
	}
}
