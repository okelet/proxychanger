package proxychangerlib

import (
	"os"
	"os/exec"
	"path"

	"github.com/okelet/goutils"
)

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewS3tpcProxySetter())
}

type S3tpcProxySetter struct {
}

func NewS3tpcProxySetter() *S3tpcProxySetter {
	return &S3tpcProxySetter{}
}

func (a *S3tpcProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	var err error

	_, err = exec.LookPath("subl")
	if err != nil {
		return &AppProxyChangeResult{a, MyGettextv("Command %v not found", "subl"), ""}
	}

	var password string

	if p != nil {
		password, err = p.GetPassword()
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error getting proxy password: %v", err)}
		}
	}

	// Create dirs
	sublConfDirPath := path.Join(HOME_DIR, ".config", "sublime-text-3", "Packages", "User")
	exists, err := goutils.DirExists(sublConfDirPath)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error checking if directory %v exists: %v", sublConfDirPath, err)}
	}

	if !exists {
		err = os.MkdirAll(sublConfDirPath, 0777)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error creating directory %v: %v", sublConfDirPath, err)}
		}
	}

	sublConfFilePath := path.Join(sublConfDirPath, "settings.json")
	confData, err := goutils.LoadJsonFileAsMap(sublConfFilePath, false)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error reading file %v: %v", sublConfFilePath, err)}
	}

	if p != nil {
		confData["http_proxy"] = p.ToSimpleUrl()
		confData["https_proxy"] = p.ToSimpleUrl()
		if p.Username != "" {
			confData["proxy_username"] = p.Username
			if password != "" {
				confData["proxy_password"] = password
			}
		}
	} else {
		delete(confData, "http_proxy")
		delete(confData, "https_proxy")
		delete(confData, "proxy_username")
		delete(confData, "proxy_password")
	}

	err = goutils.SaveMapAsJsonFile(sublConfFilePath, confData)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error saving file %v: %v", sublConfFilePath, err)}
	}

	return &AppProxyChangeResult{a, "", ""}

}

func (a *S3tpcProxySetter) GetId() string {
	return "st3pc"
}

func (a *S3tpcProxySetter) GetSimpleName() string {
	return MyGettextv("Sublime Text 3 Package Control")
}

func (a *S3tpcProxySetter) GetDescription() string {
	return MyGettextv("Sublime Text 3 Package Control")
}

func (a *S3tpcProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
