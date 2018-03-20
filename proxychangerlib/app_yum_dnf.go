package proxychangerlib

import (
	"github.com/go-ini/ini"
	"github.com/okelet/goutils"
)

const YUM_CONF_PATH = "/etc/yum.conf"
const DNF_CONF_PATH = "/etc/dnf/dnf.conf"

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewYumDnfProxySetter())
}

type YumDnfProxySetter struct {
}

func NewYumDnfProxySetter() *YumDnfProxySetter {
	return &YumDnfProxySetter{}
}

func (a *YumDnfProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	var err error
	var password string

	if p != nil {
		password, err = p.GetPassword()
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error getting the proxy password: %v", err)}
		}
	}

	confFile := YUM_CONF_PATH
	yumExists, err := goutils.FileExists(YUM_CONF_PATH)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error checking if file %v exists: %v", YUM_CONF_PATH, err)}
	}
	if !yumExists {
		dnfExists, err := goutils.FileExists(DNF_CONF_PATH)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error checking if file %v exists: %v", DNF_CONF_PATH, err)}
		}
		if !dnfExists {
			return &AppProxyChangeResult{a, MyGettextv("Current system doesn't seem to have Yum/Dnf (none of %v and %v files exist)", YUM_CONF_PATH, DNF_CONF_PATH), ""}
		} else {
			confFile = DNF_CONF_PATH
		}
	}

	cfg, err := ini.LoadSources(ini.LoadOptions{Loose: true}, confFile)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error reading file %v: %v", confFile, err)}
	}

	section := cfg.Section("main")
	if p != nil {
		section.Key("proxy").SetValue(p.ToSimpleUrl())
		if p.Username != "" {
			section.Key("proxy_username").SetValue(p.Username)
			if password != "" {
				section.Key("proxy_password").SetValue(password)
			}
		} else {
			section.DeleteKey("proxy_username")
			section.DeleteKey("proxy_password")
		}
	} else {
		section.DeleteKey("proxy")
		section.DeleteKey("proxy_username")
		section.DeleteKey("proxy_password")
	}

	err = cfg.SaveTo(confFile)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error writing the file %v: %v; <a href=\"%v\">click here</a> for possible solutions", confFile, err, "https://github.com/okelet/proxychanger/wiki/Yum")}
	}

	return &AppProxyChangeResult{a, "", ""}

}

func (a *YumDnfProxySetter) GetId() string {
	return "yum-dnf"
}

func (a *YumDnfProxySetter) GetSimpleName() string {
	return MyGettextv("Yum/Dnf")
}

func (a *YumDnfProxySetter) GetDescription() string {
	return MyGettextv("Sets the proxy in the files %v or %v for Red Hat based distributions", YUM_CONF_PATH, DNF_CONF_PATH)
}

func (a *YumDnfProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
