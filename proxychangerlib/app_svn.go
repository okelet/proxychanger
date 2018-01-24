package proxychangerlib

import (
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/go-ini/ini"
	"github.com/okelet/goutils"
)

var SVN_INITIALIZED bool
var SVN_PATH string
var SVN_INIT_ERROR string
var SVN_CONFIG_DIR string
var SVN_CONFIG_FILE string

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
	if !SVN_INITIALIZED {
		SVN_PATH, err = goutils.Which("svn")
		if err != nil {
			SVN_INIT_ERROR = err.Error()
		}
		SVN_CONFIG_DIR = path.Join(HOME_DIR, ".subversion")
		SVN_CONFIG_FILE = path.Join(SVN_CONFIG_DIR, "servers")
		SVN_INITIALIZED = true
	}

	if SVN_INIT_ERROR != "" {
		return &AppProxyChangeResult{a, "", MyGettextv("Error initializing function: %v", SVN_INIT_ERROR)}
	}

	if SVN_PATH == "" {
		return &AppProxyChangeResult{a, MyGettextv("Command %v not found", "svn"), ""}
	}

	var password string

	if p != nil {
		password, err = p.GetPassword()
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error getting the proxy password: %v", err)}
		}
	}

	exists, err := goutils.DirExists(SVN_CONFIG_DIR)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error checking if directory %v exists: %v", SVN_CONFIG_DIR, err)}
	}

	if !exists {
		err = os.Mkdir(SVN_CONFIG_DIR, os.ModeDir)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error creating directory %v: %v", SVN_CONFIG_DIR, err)}
		}
	}

	serversExists, err := goutils.FileExists(SVN_CONFIG_FILE)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error checking if file %v exists: %v", SVN_CONFIG_FILE, err)}
	} else if serversExists {
		serversBackup := SVN_CONFIG_FILE + ".proxychanger_backup"
		serversBackupExists, err := goutils.FileExists(serversBackup)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error checking if file %v exists: %v", serversBackup, err)}
		}
		if !serversBackupExists {
			err := goutils.CopyFile(SVN_CONFIG_FILE, serversBackup)
			if err != nil {
				return &AppProxyChangeResult{a, "", MyGettextv("Error backing up file %v to %v: %v", SVN_CONFIG_FILE, serversBackup, err)}
			}
		}
	}

	cfg, err := ini.LoadSources(ini.LoadOptions{Loose: true}, SVN_CONFIG_FILE)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error reading file %v: %v", SVN_CONFIG_FILE, err)}
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

	err = cfg.SaveTo(SVN_CONFIG_FILE)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error writing the file %v: %v", SVN_CONFIG_FILE, err)}
	}

	return &AppProxyChangeResult{a, "", ""}

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
