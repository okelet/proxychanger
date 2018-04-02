package proxychangerlib

import (
	"bytes"
	"strings"
)

type ScritpResult struct {
	Error  error
	Pid    int
	Code   int
	Stdout string
	Stderr string
}

func (s *ScritpResult) GetCombinedOutput() string {
	buff := bytes.NewBufferString("")
	if s.Stdout != "" && s.Stderr != "" {
		buff.WriteString(strings.Trim(s.Stdout, "\n"))
		buff.WriteString("\n")
		buff.WriteString(strings.Trim(s.Stderr, "\n"))
	} else if s.Stdout != "" {
		buff.WriteString(strings.Trim(s.Stdout, "\n"))
	} else if s.Stderr != "" {
		buff.WriteString(strings.Trim(s.Stderr, "\n"))
	}
	return buff.String()
}

type AppProxyChangeResult struct {
	Application    ProxifiedApplication
	SkippedMessage string
	WarningMessage string
	ErrorMessage   string
}
