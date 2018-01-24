package proxychangerlib

import (
	"strings"

	"github.com/okelet/goutils"
)

var APM_PATH string
var APM_INIT_ERROR string

// Register this application in the list of applications
func init() {
	var err error
	APM_PATH, err = goutils.Which("apm")
	if err != nil {
		APM_INIT_ERROR = err.Error()
	}
	RegisterProxifiedApplication(NewApmProxySetter())
}

type ApmProxySetter struct {
}

func NewApmProxySetter() *ApmProxySetter {
	return &ApmProxySetter{}
}

func (a *ApmProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	if APM_INIT_ERROR != "" {
		return &AppProxyChangeResult{a, "", MyGettextv("Error initializing function: %v", APM_INIT_ERROR)}
	}

	if APM_PATH == "" {
		return &AppProxyChangeResult{a, MyGettextv("Command %v not found", "apm"), ""}
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
		err, _, exitCode, outBuff, errBuff := goutils.RunCommandAndWait("", nil, APM_PATH, commandParams, map[string]string{})
		if err != nil {
			fullCommand := strings.Join(append([]string{APM_PATH}, commandParams...), " ")
			return &AppProxyChangeResult{a, "", MyGettextv("Error running command %v (%v): %v/%v", fullCommand, exitCode, outBuff, errBuff)}
		}
	}
	return &AppProxyChangeResult{a, "", ""}
}

func (a *ApmProxySetter) GetId() string {
	return "apm"
}

func (a *ApmProxySetter) GetSimpleName() string {
	return MyGettextv("APM")
}

func (a *ApmProxySetter) GetDescription() string {
	return MyGettextv("Atom Package Manager")
}

func (a *ApmProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
