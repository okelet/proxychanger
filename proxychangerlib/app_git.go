package proxychangerlib

import (
	"os/exec"
	"strings"

	"github.com/okelet/goutils"
)

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewGitProxySetter())
}

type GitProxySetter struct {
}

func NewGitProxySetter() *GitProxySetter {
	return &GitProxySetter{}
}

func (a *GitProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	var err error
	var url string

	gitPath, err := exec.LookPath("git")
	if err != nil {
		return &AppProxyChangeResult{a, MyGettextv("Command %v not found", "git"), "", ""}
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
		err, _, exitCode, outBuff, errBuff := goutils.RunCommandAndWait("", nil, gitPath, commandParams, map[string]string{})
		if err != nil {
			fullCommand := strings.Join(append([]string{gitPath}, commandParams...), " ")
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error running command %v (%v): %v/%v", fullCommand, exitCode, outBuff, errBuff)}
		}
	}
	return &AppProxyChangeResult{a, "", "", ""}
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
