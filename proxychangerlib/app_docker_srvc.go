package proxychangerlib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/okelet/goutils"
)

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewDockerSrvcProxySetter())
}

type DockerSrvcProxySetter struct {
}

func NewDockerSrvcProxySetter() *DockerSrvcProxySetter {
	return &DockerSrvcProxySetter{}
}

func (a *DockerSrvcProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	var err error
	var url string
	helpUrl := "https://github.com/okelet/proxychanger/wiki/Docker-Service-Daemon"

	if p != nil {
		url, err = p.ToUrl(true)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error generating proxy URL: %v", err)}
		}
	}

	_, err = exec.LookPath("docker")
	if err != nil {
		return &AppProxyChangeResult{a, MyGettextv("Command %v not found", "docker"), ""}
	}

	systemctlPath, err := exec.LookPath("systemctl")
	if err != nil {
		return &AppProxyChangeResult{a, MyGettextv("Command %v not found", "systemctl"), ""}
	}

	// Create dirs
	dockerConfDirPath := path.Join("/etc/systemd/system/docker.service.d")
	exists, err := goutils.DirExists(dockerConfDirPath)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error checking if directory %v exists: %v; <a href=\"%v\">click here</a> for possible solutions", dockerConfDirPath, err, helpUrl)}
	}

	if !exists {
		err = os.MkdirAll(dockerConfDirPath, 0777)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error creating the directory %v: %v; <a href=\"%v\">click here</a> for possible solutions", dockerConfDirPath, err, helpUrl)}
		}
	}

	dockerConfFilePath := path.Join(dockerConfDirPath, "http-proxy.conf")
	var oldContent []byte

	fileExists, err := goutils.FileExists(dockerConfFilePath)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error checking if file %v exists: %v; <a href=\"%v\">click here</a> for possible solutions", dockerConfFilePath, err, helpUrl)}
	}

	if fileExists {
		oldContent, err = ioutil.ReadFile(dockerConfFilePath)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error reading file %v: %v; <a href=\"%v\">click here</a> for possible solutions", dockerConfFilePath, err, helpUrl)}
		}
	}

	buff := bytes.NewBufferString("")
	if p != nil {
		buff.WriteString("[Service]\n")
		buff.WriteString(fmt.Sprintf("Environment=HTTP_PROXY=%v\n", url))
	}
	newContent := buff.Bytes()

	// Only update the file if it has changed (to avoid not needed daemon restarts)
	if bytes.Compare(oldContent, newContent) != 0 {

		Log.Debugf("%v content must be updated", dockerConfFilePath)

		err = ioutil.WriteFile(dockerConfFilePath, buff.Bytes(), 0666)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error writing the file %v: %v; <a href=\"%v\">click here</a> for possible solutions", dockerConfFilePath, err, helpUrl)}
		}

		err, _, exitCode, outBuff, errBuff := goutils.RunCommandAndWait("", nil, "sudo", []string{"-n", systemctlPath, "daemon-reload"}, map[string]string{})
		fullCommand := strings.Join([]string{"sudo", "-n", systemctlPath, "daemon-reload"}, " ")
		Log.Debugf("Running command %v...", fullCommand)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error running command %v (%v): %v; <a href=\"%v\">click here</a> for possible solutions", fullCommand, exitCode, goutils.CombineStdErrOutput(outBuff, errBuff), helpUrl)}
		}

		err, _, exitCode, outBuff, errBuff = goutils.RunCommandAndWait("", nil, "sudo", []string{"-n", systemctlPath, "restart", "docker.service"}, map[string]string{})
		fullCommand = strings.Join([]string{"sudo", "-n", systemctlPath, "restart", "docker.service"}, " ")
		Log.Debugf("Running command %v...", fullCommand)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error running command %v (%v): %v; <a href=\"%v\">click here</a> for possible solutions", fullCommand, exitCode, goutils.CombineStdErrOutput(outBuff, errBuff), helpUrl)}
		}

	}

	return &AppProxyChangeResult{a, "", ""}

}

func (a *DockerSrvcProxySetter) GetId() string {
	return "docker-srvc"
}

func (a *DockerSrvcProxySetter) GetSimpleName() string {
	return MyGettextv("Docker Service")
}

func (a *DockerSrvcProxySetter) GetDescription() string {
	return MyGettextv("Docker Service")
}

func (a *DockerSrvcProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
