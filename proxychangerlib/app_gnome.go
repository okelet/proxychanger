package proxychangerlib

import (
	"github.com/okelet/goutils"
)

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewGnomeProxySetter())
}

type GnomeProxySetter struct {
}

func NewGnomeProxySetter() *GnomeProxySetter {
	return &GnomeProxySetter{}
}

func (a *GnomeProxySetter) Apply(p *Proxy) *AppProxyChangeResult {
	var err error
	if p != nil {
		err = goutils.SetGnomeProxy(p.Proxy)
	} else {
		err = goutils.SetGnomeProxy(nil)
	}
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error setting the Gnome settings: %v", err)}
	}
	return &AppProxyChangeResult{a, "", ""}
}

func (a *GnomeProxySetter) GetId() string {
	return "gnome"
}

func (a *GnomeProxySetter) GetSimpleName() string {
	return MyGettextv("Gnome")
}

func (a *GnomeProxySetter) GetDescription() string {
	return MyGettextv("Sets the proxy in the Gnome Desktop environment")
}

func (a *GnomeProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
