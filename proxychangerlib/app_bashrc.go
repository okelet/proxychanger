package proxychangerlib

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/okelet/goutils"
)

var BASHRC_INITIALIZED bool
var BASHRC_PATH string
var BASHRC_SET_PROXY_REGEXPS []*regexp.Regexp
var BASHRC_INIT_ERROR string

// Register this application in the list of applications
func init() {
	RegisterProxifiedApplication(NewBashrcProxySetter())
}

type BashrcProxySetter struct {
}

func NewBashrcProxySetter() *BashrcProxySetter {
	return &BashrcProxySetter{}
}

func (a *BashrcProxySetter) Apply(p *Proxy) *AppProxyChangeResult {

	if !BASHRC_INITIALIZED {
		BASHRC_PATH = path.Join(HOME_DIR, ".bashrc")
		BASHRC_SET_PROXY_REGEXPS = []*regexp.Regexp{}
		for _, v := range []string{"(?i)^export (http|ftp|https|no)_proxy$", "(?i)^(export )?(http|ftp|https|no)_proxy=.*", "(?i)^(set )?(http|ftp|https|no)_proxy=.*"} {
			r, err := regexp.Compile(v)
			if err != nil {
				BASHRC_INIT_ERROR = err.Error()
				break
			} else {
				BASHRC_SET_PROXY_REGEXPS = append(BASHRC_SET_PROXY_REGEXPS, r)
			}
		}
		BASHRC_INITIALIZED = true
	}

	if BASHRC_INIT_ERROR != "" {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error initializing function: %v", BASHRC_INIT_ERROR)}
	}

	var err error
	var url string

	if p != nil {
		url, err = p.ToUrl(true)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error generating proxy URL: %v", err)}
		}
	}

	bashrcExists, err := goutils.FileExists(BASHRC_PATH)
	if err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error checking if file %v exists: %v", BASHRC_PATH, err)}
	} else if bashrcExists {
		bashrcBackup := BASHRC_PATH + ".proxychanger_backup"
		bashrcBackupExists, err := goutils.FileExists(bashrcBackup)
		if err != nil {
			return &AppProxyChangeResult{a, "", "", MyGettextv("Error checking if file %v exists: %v", BASHRC_PATH, err)}
		}
		if !bashrcBackupExists {
			err := goutils.CopyFile(BASHRC_PATH, bashrcBackup)
			if err != nil {
				return &AppProxyChangeResult{a, "", "", MyGettextv("Error backing up file %v to %v: %v", BASHRC_PATH, bashrcBackup, err)}
			}
		}
	}

	file, err := os.Open(BASHRC_PATH)
	if err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error opening file %v: %v", BASHRC_PATH, err)}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := bytes.NewBufferString("")
	for scanner.Scan() {
		line := scanner.Text()
		matched := false
		for _, r := range BASHRC_SET_PROXY_REGEXPS {
			if r.MatchString(line) {
				matched = true
				break
			}
		}
		if !matched {
			lines.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error reading file %v: %v", BASHRC_PATH, err)}
	}

	if p != nil {
		lines.WriteString(fmt.Sprintf("export http_proxy=%v\n", url))
		lines.WriteString(fmt.Sprintf("export https_proxy=%v\n", url))
		lines.WriteString(fmt.Sprintf("export ftp_proxy=%v\n", url))
		lines.WriteString(fmt.Sprintf("export HTTP_PROXY=%v\n", url))
		lines.WriteString(fmt.Sprintf("export HTTPS_PROXY=%v\n", url))
		lines.WriteString(fmt.Sprintf("export FTP_PROXY=%v\n", url))
		if len(p.Exceptions) > 0 {
			lines.WriteString(fmt.Sprintf("export no_proxy=%v\n", strings.Join(p.Exceptions, ",")))
			lines.WriteString(fmt.Sprintf("export NO_PROXY=%v\n", strings.Join(p.Exceptions, ",")))
		}
	}

	err = ioutil.WriteFile(BASHRC_PATH, lines.Bytes(), 0666)
	if err != nil {
		return &AppProxyChangeResult{a, "", "", MyGettextv("Error writing the file %v: %v", BASHRC_PATH, err)}
	}

	return &AppProxyChangeResult{a, "", "", ""}

}

func (a *BashrcProxySetter) GetId() string {
	return "bashrc"
}

func (a *BashrcProxySetter) GetSimpleName() string {
	return MyGettextv("bashrc")
}

func (a *BashrcProxySetter) GetDescription() string {
	return MyGettextv("Sets the proxy in the file %v", BASHRC_PATH)
}

func (a *BashrcProxySetter) GetHomepage() string {
	return "https://github.com/okelet/proxychanger"
}
