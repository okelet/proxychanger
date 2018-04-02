package proxychangerlib

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/okelet/goutils"
	"github.com/okelet/ssh_config"
)

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewSshConfigProxySetter())
}

type SshConfigProxySetter struct {
}

func NewSshConfigProxySetter() *SshConfigProxySetter {
	return &SshConfigProxySetter{}
}

func (a *SshConfigProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	var err error

	var username string
	var password string
	var corkscrewPath string
	corkscrewFound := false

	sshConfigDir := filepath.Join(os.Getenv("HOME"), ".ssh")
	sshConfigFile := filepath.Join(sshConfigDir, "config")
	sshProxyAuthFile := filepath.Join(sshConfigDir, "proxyauth")

	if p != nil {
		username = p.Username
		password, err = p.GetPassword()
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error getting proxy password: %v", err)}
		}
		corkscrewPath, err = exec.LookPath("corkscrew")
		if err == nil {
			corkscrewFound = true
		} else {
			corkscrewPath = "corkscrew"
		}
	}
	Log.Debugf("corkscrew found: %v", corkscrewFound)
	Log.Debugf("corkscrew path: %v", corkscrewPath)

	// Create dirs
	exists, err := goutils.DirExists(sshConfigDir)
	if err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error checking if directory %v exists: %v", sshConfigDir, err)}
	}

	if !exists {
		err = os.MkdirAll(sshConfigDir, 0777)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error creating directory %v: %v", sshConfigDir, err)}
		}
	}

	if username != "" && password != "" {
		err = ioutil.WriteFile(sshProxyAuthFile, []byte(fmt.Sprintf("%v:%v\n", username, password)), 0600)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error writing the file %v: %v", sshProxyAuthFile, err)}
		}
	} else {
		exists, err = goutils.FileExists(sshProxyAuthFile)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error checking if file %v exists: %v", sshProxyAuthFile, err)}
		} else if exists {
			err = os.Remove(sshProxyAuthFile)
			if err != nil {
				return &AppProxyChangeResult{a, "", "", MyGettextv("Error deleting file %v: %v", sshProxyAuthFile, err)}
			}
		}
	}

	exists, err = goutils.FileExists(sshConfigFile)
	if err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error checking if file %v exists: %v", sshConfigFile, err)}
	} else if exists {
		sshConfigFileBackup := sshConfigFile + ".proxychanger_backup"
		sshConfigFileBackupExists, err := goutils.FileExists(sshConfigFileBackup)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error checking if file %v exists: %v", sshConfigFile, err)}
		}
		if !sshConfigFileBackupExists {
			err := goutils.CopyFile(sshConfigFile, sshConfigFileBackup)
			if err != nil {
				return &AppProxyChangeResult{a, "", "", MyGettextv("Error backing up file %v to %v: %v", sshConfigFile, sshConfigFileBackup, err)}
			}
		}
	}

	// Load configuration
	var sshConfig *ssh_config.Config
	if exists {
		openFile, err := os.Open(sshConfigFile)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error opening file %v: %v", sshConfigFile, err)}
		}
		sshConfig, err = ssh_config.Decode(openFile)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error parsing file %v: %v", sshConfigFile, err)}
		}
	} else {
		sshConfig = ssh_config.NewEmptyConfig()
	}

	h := sshConfig.GetHostForPattern("*")
	if h == nil {
		h, err = sshConfig.AddNewHost("*")
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error creating new wildcard SSH host: %v", err)}
		}
	}
	if p != nil {
		if username != "" && password != "" {
			h.Set("ProxyCommand", fmt.Sprintf("%v %v %v %%h %%p %v", corkscrewPath, p.Address, p.Port, sshProxyAuthFile))
		} else {
			h.Set("ProxyCommand", fmt.Sprintf("%v %v %v %%h %%p", corkscrewPath, p.Address, p.Port))
		}
	} else {
		h.Delete("ProxyCommand")
	}
	sshConfig.SetLeadingSpace(4)

	err = ioutil.WriteFile(sshConfigFile, []byte(sshConfig.String()), 0666)
	if err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error writing the file %v: %v", sshConfigFile, err)}
	}

	if p != nil && !corkscrewFound {
		return &AppProxyChangeResult{a, "", MyGettextv("SSH configuration applied, but corkscrew command not found"), ""}
	} else {
		return &AppProxyChangeResult{a, "", "", ""}
	}

}

func (a *SshConfigProxySetter) GetId() string {
	return "ssh-config"
}

func (a *SshConfigProxySetter) GetSimpleName() string {
	return MyGettextv("SSH configuration (~/.ssh/config)")
}

func (a *SshConfigProxySetter) GetDescription() string {
	return MyGettextv("Configures proxy for Host * in ~/.ssh/config")
}

func (a *SshConfigProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
