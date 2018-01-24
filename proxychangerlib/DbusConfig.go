package proxychangerlib

import (
	"fmt"

	"github.com/godbus/dbus"
)

type DbusConfig struct {
	DBusConnection *dbus.Conn
}

func NewDbusConfig(dbusConnection *dbus.Conn) (*DbusConfig, error) {
	c := DbusConfig{}
	c.DBusConnection = dbusConnection
	return &c, nil
}

/*
dbus-send --session --print-reply --dest=com.github.okelet.gocm /com/github/okelet/gocm com.github.okelet.gocm.DbusListGroups
*/
func (c *DbusConfig) DbusListProxies() (string, *dbus.Error) {
	obj := c.DBusConnection.Object(DBUS_INTERFACE, DBUS_PATH)
	call := obj.Call(DBUS_INTERFACE+".DbusListProxies", 0)
	if call.Err != nil {
		fmt.Println("Panic 1")
		panic(call.Err)
	}
	if call.Err != nil {
		fmt.Println("Panic 2")
		panic(call.Err)
	}
	var ret string
	err := call.Store(&ret)
	if err != nil {
		fmt.Println("Panic 3")
		panic(err)
	}
	return ret, nil
}

func (c *DbusConfig) SetActiveProxyBySlug(slug string) (string, *dbus.Error) {
	obj := c.DBusConnection.Object(DBUS_INTERFACE, DBUS_PATH)
	call := obj.Call(DBUS_INTERFACE+".SetActiveProxyBySlug", 0, slug)
	if call.Err != nil {
		fmt.Println("Panic 1")
		panic(call.Err)
	}
	if call.Err != nil {
		fmt.Println("Panic 2")
		panic(call.Err)
	}
	var ret string
	err := call.Store(&ret)
	if err != nil {
		fmt.Println("Panic 3")
		panic(err)
	}
	return ret, nil
}
