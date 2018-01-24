package proxychangerlib

import (
	"github.com/godbus/dbus"
)

type DbusListProxiesResponse struct {
	Proxies []ProxyStruct
}

type DbusSetActiveProxyBySlugResponse struct {
	Error string
}

type ProxyStruct struct {
	UUID        string
	Name        string
	Slug        string
	Description string
	Protocol    string
	Address     string
	Port        int
	Username    string
	Password    string
	Exceptions  []string
	MatchingIps []string
	Active      bool
}

type ConfigInterface interface {
	DbusListProxies() (string, *dbus.Error)
	SetActiveProxyBySlug(slug string) (string, *dbus.Error)
}
