package proxychangerlib

import (
	"fmt"
)

var ProxifiedApplications []ProxifiedApplication

func init() {
	ProxifiedApplications = []ProxifiedApplication{}
}

type ProxifiedApplication interface {
	Apply(p *Proxy) *AppProxyChangeResult
	GetId() string
	GetSimpleName() string
	GetDescription() string
	GetHomepage() string
}

func RegisterProxifiedApplication(p ProxifiedApplication) {
	for _, a := range ProxifiedApplications {
		if a.GetId() == p.GetId() {
			panic(fmt.Sprintf("Trying to register an application with and id already registered: %v", a.GetId()))
		}
	}
	ProxifiedApplications = append(ProxifiedApplications, p)
}
