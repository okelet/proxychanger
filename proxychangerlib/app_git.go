package proxychangerlib

import (
	"strings"

	"github.com/okelet/goutils"
)

var GIT_PATH string
var GIT_INIT_ERROR string

// Register this application in the list of applications
func init() {
	var err error
	GIT_PATH, err = goutils.Which("git")
	if err != nil {
		GIT_INIT_ERROR = err.Error()
	}
	RegisterProxifiedApplication(NewGitProxySetter())
}

type GitProxySetter struct {
}

func NewGitProxySetter() *GitProxySetter {
	return &GitProxySetter{}
}

func (a *GitProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	if GIT_INIT_ERROR != "" {
		return &AppProxyChangeResult{a, "", MyGettextv("Error initializing function: %v", GIT_INIT_ERROR)}
	}

	if GIT_PATH == "" {
		return &AppProxyChangeResult{a, MyGettextv("Command %v not found", "git"), ""}
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
			{"config", "--global", "http.proxy", url},
			{"config", "--global", "https.proxy", url},
		}
	} else {
		params = [][]string{
			{"config", "--global", "--unset", "http.proxy"},
			{"config", "--global", "--unset", "https.proxy"},
		}
	}
	for _, commandParams := range params {
		err, _, exitCode, outBuff, errBuff := goutils.RunCommandAndWait("", nil, GIT_PATH, commandParams, map[string]string{})
		if err != nil {
			fullCommand := strings.Join(append([]string{GIT_PATH}, commandParams...), " ")
			return &AppProxyChangeResult{a, "", MyGettextv("Error running command %v (%v): %v/%v", fullCommand, exitCode, outBuff, errBuff)}
		}
	}
	return &AppProxyChangeResult{a, "", ""}
}

func (a *GitProxySetter) GetId() string {
	return "git"
}

func (a *GitProxySetter) GetSimpleName() string {
	return MyGettextv("Git")
}

func (a *GitProxySetter) GetDescription() string {
	return MyGettextv("Git")
}

func (a *GitProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
