package proxychangerlib

import (
	"strings"

	"github.com/okelet/goutils"
)

var NPM_PATH string
var NPM_INIT_ERROR string

// Register this application in the list of applications
func init() {
	var err error
	NPM_PATH, err = goutils.Which("npm")
	if err != nil {
		NPM_INIT_ERROR = err.Error()
	}
	RegisterProxifiedApplication(NewNpmProxySetter())
}

type NpmProxySetter struct {
}

func NewNpmProxySetter() *NpmProxySetter {
	return &NpmProxySetter{}
}

func (a *NpmProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	if NPM_INIT_ERROR != "" {
		return &AppProxyChangeResult{a, "", MyGettextv("Error initializing function: %v", NPM_INIT_ERROR)}
	}

	if NPM_PATH == "" {
		return &AppProxyChangeResult{a, MyGettextv("Command %v not found", "npm"), ""}
	}

	var err error
	var url string

	if p != nil {
		url, err = p.ToUrl(true)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error generating proxy URL: %v", err)}
		}
	}

	var params [][]string
	if p != nil {
		params = [][]string{
			{"config", "set", "proxy", url},
			{"config", "set", "https-proxy", url},
		}
	} else {
		params = [][]string{
			{"config", "delete", "proxy"},
			{"config", "delete", "https-proxy"},
		}
	}
	for _, commandParams := range params {
		err, _, exitCode, outBuff, errBuff := goutils.RunCommandAndWait("", nil, NPM_PATH, commandParams, map[string]string{})
		if err != nil {
			fullCommand := strings.Join(append([]string{NPM_PATH}, commandParams...), " ")
			return &AppProxyChangeResult{a, "", MyGettextv("Error running command %v (%v): %v/%v", fullCommand, exitCode, outBuff, errBuff)}
		}
	}
	return &AppProxyChangeResult{a, "", ""}
}

func (a *NpmProxySetter) GetId() string {
	return "npm"
}

func (a *NpmProxySetter) GetSimpleName() string {
	return MyGettextv("Npm (Node.js Package Manager)")
}

func (a *NpmProxySetter) GetDescription() string {
	return MyGettextv("Node.js Package Manager")
}

func (a *NpmProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
