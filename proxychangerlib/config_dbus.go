package proxychangerlib

import (
	"fmt"

	"github.com/godbus/dbus"
)

type ConfigDbus struct {
	DBusConnection *dbus.Conn
}

func NewConfigDbus(dbusConnection *dbus.Conn) (*ConfigDbus, error) {
	c := ConfigDbus{}
	c.DBusConnection = dbusConnection
	return &c, nil
}

func (c *ConfigDbus) ListProxies(includePasswords bool) (string, *dbus.Error) {

	obj := c.DBusConnection.Object(DBUS_INTERFACE, DBUS_PATH)

	call := obj.Call(fmt.Sprintf("%v.%v", DBUS_INTERFACE, "ListProxies"), 0, includePasswords)
	if call.Err != nil {
		return "", dbus.NewError(call.Err.Error(), nil)
	}

	var ret string
	err := call.Store(&ret)
	if err != nil {
		return "", dbus.NewError(call.Err.Error(), nil)
	}

	return ret, nil

}

func (c *ConfigDbus) ApplyActiveProxy() (string, *dbus.Error) {

	obj := c.DBusConnection.Object(DBUS_INTERFACE, DBUS_PATH)

	call := obj.Call(fmt.Sprintf("%v.%v", DBUS_INTERFACE, "ApplyActiveProxy"), 0)
	if call.Err != nil {
		return "", dbus.NewError(call.Err.Error(), nil)
	}

	var ret string
	err := call.Store(&ret)
	if err != nil {
		return "", dbus.NewError(call.Err.Error(), nil)
	}

	return ret, nil

}

func (c *ConfigDbus) GetActiveProxySlug() (string, *dbus.Error) {

	obj := c.DBusConnection.Object(DBUS_INTERFACE, DBUS_PATH)

	call := obj.Call(fmt.Sprintf("%v.%v", DBUS_INTERFACE, "GetActiveProxySlug"), 0)
	if call.Err != nil {
		return "", dbus.NewError(call.Err.Error(), nil)
	}

	var ret string
	err := call.Store(&ret)
	if err != nil {
		return "", dbus.NewError(call.Err.Error(), nil)
	}

	return ret, nil

}

func (c *ConfigDbus) SetActiveProxyBySlug(slug string) (string, *dbus.Error) {

	obj := c.DBusConnection.Object(DBUS_INTERFACE, DBUS_PATH)

	call := obj.Call(fmt.Sprintf("%v.%v", DBUS_INTERFACE, "SetActiveProxyBySlug"), 0, slug)
	if call.Err != nil {
		return "", dbus.NewError(call.Err.Error(), nil)
	}

	var ret string
	err := call.Store(&ret)
	if err != nil {
		return "", dbus.NewError(call.Err.Error(), nil)
	}

	return ret, nil

}
