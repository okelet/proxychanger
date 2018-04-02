package proxychangerlib

import (
	"os"
	"strings"
)

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewEnvProxySetter())
}

type EnvProxySetter struct {
}

func NewEnvProxySetter() *EnvProxySetter {
	return &EnvProxySetter{}
}

func (a *EnvProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	var err error
	var url string

	if p != nil {
		url, err = p.ToUrl(true)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error generating proxy URL: %v", err)}
		}
	}

	if p != nil {
		for _, v := range []string{"http_proxy", "https_proxy", "ftp_proxy"} {
			if err = os.Setenv(v, url); err != nil {
				return &AppProxyChangeResult{a, "", "", MyGettextv("Error setting environment variable %v: %v", v, err)}
			}
			if err = os.Setenv(strings.ToUpper(v), url); err != nil {
				return &AppProxyChangeResult{a, "", "", MyGettextv("Error setting environment variable %v: %v", strings.ToUpper(v), err)}
			}
		}
		if len(p.Exceptions) > 0 {
			if err = os.Setenv("no_proxy", strings.Join(p.Exceptions, ",")); err != nil {
				return &AppProxyChangeResult{a, "", "", MyGettextv("Error setting environment variable %v: %v", "no_proxy", err)}
			}
			if err = os.Setenv("NO_PROXY", strings.Join(p.Exceptions, ",")); err != nil {
				return &AppProxyChangeResult{a, "", "", MyGettextv("Error setting environment variable %v: %v", "NO_PROXY", err)}
			}
		} else {
			if err = os.Unsetenv("no_proxy"); err != nil {
				return &AppProxyChangeResult{a, "", "", MyGettextv("Error unsetting environment variable %v: %v", "no_proxy", err)}
			}
			if err = os.Unsetenv("NO_PROXY"); err != nil {
				return &AppProxyChangeResult{a, "", "", MyGettextv("Error unsetting environment variable %v: %v", "NO_PROXY", err)}
			}
		}

	} else {
		for _, v := range []string{"http_proxy", "https_proxy", "ftp_proxy", "no_proxy"} {
			if err = os.Unsetenv(v); err != nil {
				return &AppProxyChangeResult{a, "", "", MyGettextv("Error unsetting environment variable %v: %v", v, err)}
			}
			if err = os.Unsetenv(strings.ToUpper(v)); err != nil {
				return &AppProxyChangeResult{a, "", "", MyGettextv("Error unsetting environment variable %v: %v", strings.ToUpper(v), err)}
			}
		}
	}
	return &AppProxyChangeResult{a, "", "", ""}
}

func (a *EnvProxySetter) GetId() string {
	return "env"
}

func (a *EnvProxySetter) GetSimpleName() string {
	return MyGettextv("Environment")
}

func (a *EnvProxySetter) GetDescription() string {
	return MyGettextv("Sets the proxy in the current environment")
}

func (a *EnvProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
