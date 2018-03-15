package proxychangerlib

import (
	"os"
	"path"
	"strings"

	"github.com/okelet/goutils"
)

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewDockerCliProxySetter())
}

type DockerCliProxySetter struct {
}

func NewDockerCliProxySetter() *DockerCliProxySetter {
	return &DockerCliProxySetter{}
}

func (a *DockerCliProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	var err error

	dockerPath, err := goutils.Which("docker")
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error checking if command %v exists: %v", "docker", err)}
	}

	if dockerPath == "" {
		return &AppProxyChangeResult{a, MyGettextv("Command %v not found", "docker"), ""}
	}

	// Create dirs
	dockerConfDirPath := path.Join(HOME_DIR, ".docker")
	exists, err := goutils.DirExists(dockerConfDirPath)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error checking if directory %v exists: %v", dockerConfDirPath, err)}
	}

	if !exists {
		err = os.MkdirAll(dockerConfDirPath, 0777)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error creating directory %v: %v", dockerConfDirPath, err)}
		}
	}

	dockerConfFilePath := path.Join(dockerConfDirPath, "config.json")
	helper, err := goutils.NewMapHelperFromJsonFile(dockerConfFilePath, false)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error loading file %v: %v", dockerConfFilePath, err)}
	}

	if p != nil {
		var url string
		url, err = p.ToUrl(true)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error generating proxy URL: %v", err)}
		}
		helper.GetHelper("proxies").GetHelper("default").SetString("httpProxy", url)
		helper.GetHelper("proxies").GetHelper("default").SetString("httpsProxy", url)
		helper.GetHelper("proxies").GetHelper("default").SetString("ftpProxy", url)
		if len(p.Exceptions) > 0 {
			helper.GetHelper("proxies").GetHelper("default").SetString("noProxy", strings.Join(p.Exceptions, ","))
		} else {
			helper.GetHelper("proxies").GetHelper("default").Delete("noProxy")
		}
	} else {
		if helper.Exists("proxies") {
			helper.GetHelper("proxies").Delete("default")
			if len(helper.GetHelper("proxies").Keys()) == 0 {
				helper.Delete("proxies")
			}
		}
	}

	err = helper.SaveToJson(true)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error saving file %v: %v", dockerConfFilePath, err)}
	}

	return &AppProxyChangeResult{a, "", ""}

}

func (a *DockerCliProxySetter) GetId() string {
	return "docker-cli"
}

func (a *DockerCliProxySetter) GetSimpleName() string {
	return MyGettextv("Docker CLI")
}

func (a *DockerCliProxySetter) GetDescription() string {
	return MyGettextv("Docker CLI")
}

func (a *DockerCliProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
