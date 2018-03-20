package proxychangerlib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

const DEBIAN_ETC_VERSION = "/etc/debian_version"
const APT_PROXY_FILE = "/etc/apt/apt.conf.d/90proxy"

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewAptProxySetter())
}

type AptProxySetter struct {
}

func NewAptProxySetter() *AptProxySetter {
	return &AptProxySetter{}
}

func (a *AptProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	var err error
	var url string

	if p != nil {
		url, err = p.ToUrl(true)
		if err != nil {
			return &AppProxyChangeResult{a, "", MyGettextv("Error generating proxy URL: %v", err)}
		}
	}

	stat, err := os.Stat(DEBIAN_ETC_VERSION)
	if err != nil {
		if !os.IsNotExist(err) {
			return &AppProxyChangeResult{a, "", MyGettextv("Error detecting if the current system is Debian Based using file %v: %v", DEBIAN_ETC_VERSION, err)}
		} else {
			return &AppProxyChangeResult{a, MyGettextv("Current system doesn't seem to be Debian Based (file %v doesn't exist)", DEBIAN_ETC_VERSION), ""}
		}
	}
	if stat.IsDir() {
		return &AppProxyChangeResult{a, "", MyGettextv("Path %v is a directory, and it should be a file", DEBIAN_ETC_VERSION)}
	}

	buff := bytes.NewBufferString("")
	if p != nil {
		for _, proxyType := range []string{"http", "https", "ftp"} {
			buff.WriteString(fmt.Sprintf("Acquire::%v::proxy \"%v\";\n", proxyType, url))
		}
	}
	err = ioutil.WriteFile(APT_PROXY_FILE, buff.Bytes(), 0666)
	if err != nil {
		return &AppProxyChangeResult{a, "", MyGettextv("Error writing the file %v: %v; <a href=\"%v\">click here</a> for possible solutions", APT_PROXY_FILE, err, "https://github.com/okelet/proxychanger/wiki/APT")}
	}
	return &AppProxyChangeResult{a, "", ""}
}

func (a *AptProxySetter) GetId() string {
	return "apt"
}

func (a *AptProxySetter) GetSimpleName() string {
	return MyGettextv("APT")
}

func (a *AptProxySetter) GetDescription() string {
	return MyGettextv("Sets the proxy in the file %v for Debian based distributions", APT_PROXY_FILE)
}

func (a *AptProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
