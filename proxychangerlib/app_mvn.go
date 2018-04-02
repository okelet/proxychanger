package proxychangerlib

import (
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/okelet/goutils"
)

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewMavenProxySetter())
}

type MavenProxySetter struct {
}

func NewMavenProxySetter() *MavenProxySetter {
	return &MavenProxySetter{}
}

func (a *MavenProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	var err error
	var password string

	if p != nil {
		password, err = p.GetPassword()
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error getting the proxy password: %v", err)}
		}
	}

	_, err = exec.LookPath("mvn")
	if err != nil {
		return &AppProxyChangeResult{a, MyGettextv("Command %v not found", "mvn"), "", ""}
	}

	// Create dirs
	mvnConfDirPath := path.Join(HOME_DIR, ".m2")
	dirExists, err := goutils.DirExists(mvnConfDirPath)
	if err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error checking if directory %v exists: %v", mvnConfDirPath, err)}
	}

	if !dirExists {
		err = os.MkdirAll(mvnConfDirPath, 0755)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error creating directory %v: %v", mvnConfDirPath, err)}
		}
	}

	mvnConfFilePath := path.Join(mvnConfDirPath, "settings.xml")
	fileExists, err := goutils.FileExists(mvnConfFilePath)
	if err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error checking if file %v exists: %v", mvnConfDirPath, err)}
	}

	if fileExists {
		mvnBackup := mvnConfFilePath + ".proxychanger_backup"
		mvnBackupExists, err := goutils.FileExists(mvnBackup)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error checking if file %v exists: %v", mvnBackup, err)}
		}
		if !mvnBackupExists {
			err := goutils.CopyFile(mvnConfFilePath, mvnBackup)
			if err != nil {
				return &AppProxyChangeResult{a, "", "", MyGettextv("Error backing up file %v to %v: %v", mvnConfFilePath, mvnBackup, err)}
			}
		}
	}

	doc := etree.NewDocument()
	if fileExists {
		err = doc.ReadFromFile(mvnConfFilePath)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error reading file %v: %v", mvnConfFilePath, err)}
		}
	} else {
		doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	}

	settings := doc.SelectElement("settings")
	if settings == nil {
		settings = doc.CreateElement("settings")
		settings.CreateAttr("xmlns", "http://maven.apache.org/SETTINGS/1.0.0")
		settings.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
		settings.CreateAttr("xsi:schemaLocation", "http://maven.apache.org/SETTINGS/1.0.0 http://maven.apache.org/xsd/settings-1.0.0.xsd")
	}
	proxies := settings.SelectElement("proxies")
	if proxies != nil {
		settings.RemoveChild(proxies)
	}
	if p != nil {
		proxies = settings.CreateElement("proxies")
		for _, proxyType := range []string{"http", "https"} {
			proxy := proxies.CreateElement("proxy")
			proxy.CreateElement("id").SetText(proxyType)
			proxy.CreateElement("protocol").SetText(proxyType)
			proxy.CreateElement("active").SetText("true")
			proxy.CreateElement("host").SetText(p.Address)
			proxy.CreateElement("port").SetText(strconv.Itoa(p.Port))
			if p.Username != "" {
				proxy.CreateElement("username").SetText(p.Username)
				if password != "" {
					proxy.CreateElement("password").SetText(password)
				}
			}
			if len(p.Exceptions) > 0 {
				proxy.CreateElement("nonProxyHosts").SetText(strings.Join(p.Exceptions, "|"))
			}
		}
	}

	doc.Indent(2)

	err = doc.WriteToFile(mvnConfFilePath)
	if err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error writing the file %v: %v", mvnConfFilePath, err)}
	}

	return &AppProxyChangeResult{a, "", "", ""}

}

func (a *MavenProxySetter) GetId() string {
	return "mvn"
}

func (a *MavenProxySetter) GetSimpleName() string {
	return MyGettextv("Maven")
}

func (a *MavenProxySetter) GetDescription() string {
	return MyGettextv("Maven")
}

func (a *MavenProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
