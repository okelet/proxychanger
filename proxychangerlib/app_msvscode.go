package proxychangerlib

import (
	"os"
	"path"

	"github.com/okelet/goutils"
)

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewMsVsCodeProxySetter())
}

type MsVsCodeProxySetter struct {
}

func NewMsVsCodeProxySetter() *MsVsCodeProxySetter {
	return &MsVsCodeProxySetter{}
}

func (a *MsVsCodeProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	var err error

	codePath, err := goutils.Which("code")
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error checking if command %v exists: %v", "code", err)}
	}

	if codePath == "" {
		return &AppProxyChangeResult{a, MyGettextv("Command %v not found", "code"), ""}
	}

	// Create dirs
	codeConfDirPath := path.Join(HOME_DIR, ".config", "Code", "User")
	exists, err := goutils.DirExists(codeConfDirPath)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error checking if directory %v exists: %v", codeConfDirPath, err)}
	}

	if !exists {
		err = os.MkdirAll(codeConfDirPath, os.ModeDir)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error creating directory %v: %v", codeConfDirPath, err)}
		}
	}

	codeConfFilePath := path.Join(codeConfDirPath, "settings.json")
	confData, err := goutils.LoadJsonFileAsMap(codeConfFilePath, false)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error reading file %v: %v", codeConfFilePath, err)}
	}

	if p != nil {
		confData["http.proxy"] = p.ToSimpleUrl()
	} else {
		delete(confData, "http.proxy")
	}

	err = goutils.SaveMapAsJsonFile(codeConfFilePath, confData)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error saving file %v: %v", codeConfFilePath, err)}
	}

	return &AppProxyChangeResult{a, "", ""}

}

func (a *MsVsCodeProxySetter) GetId() string {
	return "msvscode"
}

func (a *MsVsCodeProxySetter) GetSimpleName() string {
	return MyGettextv("MS VS Code")
}

func (a *MsVsCodeProxySetter) GetDescription() string {
	return MyGettextv("Microsoft Visual Studio Code")
}

func (a *MsVsCodeProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
