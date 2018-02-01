package proxychangerlib

import (
	"github.com/godbus/dbus"
)

type ListProxiesResponse struct {
	Error   string
	Proxies []ProxyStruct
}

type GetActiveProxySlugResponse struct {
	Error string
	Slug  string
}

type SetActiveProxyBySlugResponse struct {
	Error string
}

type ApplyActiveProxyResponse struct {
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

type ConfigService interface {
	ListProxies(includePasswords bool) (string, *dbus.Error)
	ApplyActiveProxy() (string, *dbus.Error)
	GetActiveProxySlug() (string, *dbus.Error)
	SetActiveProxyBySlug(slug string) (string, *dbus.Error)
}
