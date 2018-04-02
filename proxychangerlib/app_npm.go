package proxychangerlib

import (
	"os/exec"
	"strings"

	"github.com/okelet/goutils"
)

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewNpmProxySetter())
}

type NpmProxySetter struct {
}

func NewNpmProxySetter() *NpmProxySetter {
	return &NpmProxySetter{}
}

func (a *NpmProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	var err error
	var url string

	npmPath, err := exec.LookPath("npm")
	if err != nil {
		return &AppProxyChangeResult{a, MyGettextv("Command %v not found", "npm"), "", ""}
	}

	if p != nil {
		url, err = p.ToUrl(true)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error generating proxy URL: %v", err)}
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
		err, _, exitCode, outBuff, errBuff := goutils.RunCommandAndWait("", nil, npmPath, commandParams, map[string]string{})
		if err != nil {
			fullCommand := strings.Join(append([]string{npmPath}, commandParams...), " ")
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error running command %v (%v): %v/%v", fullCommand, exitCode, outBuff, errBuff)}
		}
	}
	return &AppProxyChangeResult{a, "", "", ""}
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
