package proxychangerlib

import "github.com/juju/loggo"

type GlobalProxyChangeResult struct {
	Proxy                        *Proxy
	Reason                       string
	Results                      []*AppProxyChangeResult
	ChangeScriptResult           *ScritpResult
	ProxyActivateScriptResult    *ScritpResult
	GlobalDeactivateScriptResult *ScritpResult
	GlobalActivateScriptResult   *ScritpResult
}

func (n *GlobalProxyChangeResult) GetNumberOfErrors() int {
	counter := 0
	for _, r := range n.Results {
		if !r.Skipped() && !r.Success() {
			counter += 1
		}
	}
	if n.ChangeScriptResult != nil && (n.ChangeScriptResult.Error != nil || n.ChangeScriptResult.Code != 0) {
		counter += 1
	}
	if n.ProxyActivateScriptResult != nil && (n.ProxyActivateScriptResult.Error != nil || n.ProxyActivateScriptResult.Code != 0) {
		counter += 1
	}
	if n.GlobalDeactivateScriptResult != nil && (n.GlobalDeactivateScriptResult.Error != nil || n.GlobalDeactivateScriptResult.Code != 0) {
		counter += 1
	}
	if n.GlobalActivateScriptResult != nil && (n.GlobalActivateScriptResult.Error != nil || n.GlobalActivateScriptResult.Code != 0) {
		counter += 1
	}
	return counter
}

type ConfigListener interface {
	OnConfigLoaded()
	OnProxyActivated(notification *GlobalProxyChangeResult)
	OnProxyAdded(p *Proxy)
	OnProxyUpdated(p *Proxy)
	OnProxyRemoved(p *Proxy)
	OnShowProxyNameNextToIndicatorChanged(newValue bool)
	OnEnableAutoChangeByIpChanged(newValue bool)
	OnEnableUpdateCheckChanged(newValue bool)
	OnWhatToDoWhenNoIpMatchesChanged(newValue string)
	OnLogLevelChanged(newValue loggo.Level)
}
