package proxychangerlib

import (
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/go-ini/ini"
	"github.com/okelet/goutils"
)

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewSvnProxySetter())
}

type SvnProxySetter struct {
}

func NewSvnProxySetter() *SvnProxySetter {
	return &SvnProxySetter{}
}

func (a *SvnProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	var err error

	_, err = exec.LookPath("svn")
	if err != nil {
		return &AppProxyChangeResult{a, MyGettextv("Command %v not found", "svn"), "", ""}
	}

	var password string

	if p != nil {
		password, err = p.GetPassword()
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error getting the proxy password: %v", err)}
		}
	}

	svnConfigDir := path.Join(HOME_DIR, ".subversion")
	exists, err := goutils.DirExists(svnConfigDir)
	if err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error checking if directory %v exists: %v", svnConfigDir, err)}
	}

	if !exists {
		err = os.Mkdir(svnConfigDir, 0777)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error creating directory %v: %v", svnConfigDir, err)}
		}
	}

	svnConfigFile := path.Join(svnConfigDir, "servers")
	serversExists, err := goutils.FileExists(svnConfigFile)
	if err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error checking if file %v exists: %v", svnConfigFile, err)}
	} else if serversExists {
		serversBackup := svnConfigFile + ".proxychanger_backup"
		serversBackupExists, err := goutils.FileExists(serversBackup)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error checking if file %v exists: %v", serversBackup, err)}
		}
		if !serversBackupExists {
			err := goutils.CopyFile(svnConfigFile, serversBackup)
			if err != nil {
				return &AppProxyChangeResult{a, "", "", MyGettextv("Error backing up file %v to %v: %v", svnConfigFile, serversBackup, err)}
			}
		}
	}

	cfg, err := ini.LoadSources(ini.LoadOptions{Loose: true}, svnConfigFile)
	if err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error reading file %v: %v", svnConfigFile, err)}
	}

	section := cfg.Section("global")
	if p != nil {
		section.Key("http-proxy-host").SetValue(p.Address)
		section.Key("http-proxy-port").SetValue(strconv.Itoa(p.Port))
		if len(p.Exceptions) > 0 {
			section.Key("http-proxy-exceptions").SetValue(strings.Join(p.Exceptions, ", "))
		} else {
			section.DeleteKey("http-proxy-exceptions")
		}
		if p.Username != "" {
			section.Key("http-proxy-username").SetValue(p.Username)
			if password != "" {
				section.Key("http-proxy-password").SetValue(password)
			}
		} else {
			section.DeleteKey("http-proxy-username")
			section.DeleteKey("http-proxy-password")
		}
	} else {
		section.DeleteKey("http-proxy-host")
		section.DeleteKey("http-proxy-port")
		section.DeleteKey("http-proxy-exceptions")
		section.DeleteKey("http-proxy-username")
		section.DeleteKey("http-proxy-password")
	}

	err = cfg.SaveTo(svnConfigFile)
	if err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error writing the file %v: %v", svnConfigFile, err)}
	}

	return &AppProxyChangeResult{a, "", "", ""}

}

func (a *SvnProxySetter) GetId() string {
	return "svn"
}

func (a *SvnProxySetter) GetSimpleName() string {
	return MyGettextv("Subversion")
}

func (a *SvnProxySetter) GetDescription() string {
	return MyGettextv("Subversion")
}

func (a *SvnProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
