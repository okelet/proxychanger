package proxychangerlib

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/godbus/dbus"
	"github.com/juju/loggo"

	"github.com/go-ini/ini"
	"github.com/gosimple/slug"
	"github.com/okelet/goutils"
	"github.com/pkg/errors"
	"github.com/zalando/go-keyring"
)

const KEEP_CURRENT_PROXY = "keep"
const DEACTIVATE_PROXY = "deactivate"

type Configuration struct {
	Filename string

	IndicatorAlreadyRun                 bool
	ShowCurrentProxyNameNextToIndicator bool
	LogLevel                            loggo.Level

	ExcludedInterfacesRegexps       []string
	ExcludedInterfacesRegexpsParsed []*regexp.Regexp

	EnableUpdateCheck       bool
	TimeBetweenUpdateChecks int

	EnableAutoChangeByIp    bool
	WhatToDoWhenNoIpMatches string
	TimeBetweenIpChecks     int

	ProxyChangeScript     string
	ProxyDeactivateScript string

	DisabledApplicationsIds []string

	Proxies     []*Proxy
	ActiveProxy *Proxy

	Listeners []ConfigListener

	LastExecutionResults *GlobalProxyChangeResult
}

func NewConfig(sessionBus *dbus.Conn, configPath string, setActiveProxy bool) (*Configuration, []string, error) {

	config := &Configuration{}
	config.Listeners = []ConfigListener{}
	config.Proxies = []*Proxy{}

	if configPath == "" {
		configPath = DEFAULT_CONFIG_PATH
	}

	warnings, err := config.Load(configPath, setActiveProxy, false, false)
	if err != nil {
		return nil, []string{}, err
	}

	config.Filename = configPath

	err = sessionBus.Export(config, DBUS_PATH, DBUS_INTERFACE)
	if err != nil {
		return nil, []string{}, errors.Wrap(err, "Error exporting to dbus")
	}

	return config, warnings, nil

}

func (c *Configuration) Load(configPath string, setActiveProxy bool, loadPasswordsFromMap bool, saveAfterLoad bool) ([]string, error) {

	warnings := []string{}

	failIfNotFound := true
	if configPath == DEFAULT_CONFIG_PATH {
		// Use default file, but don't get an error if it doesn't exist
		failIfNotFound = false
	}

	helper, err := goutils.NewMapHelperFromJsonFile(configPath, failIfNotFound)
	if err != nil {
		return warnings, errors.Wrapf(err, "Error loading configuration")
	}

	// Deactivate current proxy, if any
	if c.ActiveProxy != nil {
		_, err := c.SetActiveProxy(nil, "Deactivating current proxy while loading configuration", false)
		if err != nil {
			return warnings, errors.Wrap(err, "Error deactivating proxy")
		}
	}

	// Remove old proxies, sending notification to listeners
	proxiesCopy := c.Proxies[:]
	for ind, _ := range proxiesCopy {
		err := c.DeleteProxy(proxiesCopy[ind], false)
		if err != nil {
			return []string{}, errors.Wrapf(err, "Error deleting proxy %v", proxiesCopy[ind].Name)
		}
	}

	version := helper.GetInt("version", 0)

	c.IndicatorAlreadyRun = helper.GetBoolean("indicator_already_run", false)

	if version == 0 {
		c.SetShowCurrentProxyNameNextToIndicator(helper.GetBoolean("show_current_proxy_next_icon", true))
	} else {
		c.SetShowCurrentProxyNameNextToIndicator(helper.GetBoolean("show_current_proxy_name_next_to_indicator", true))
	}
	logLevelStr := helper.GetString("log_level", "warning")
	level, ok := loggo.ParseLevel(logLevelStr)
	if !ok {
		Log.Errorf("Invalid LOG level %v.", logLevelStr)
	} else {
		c.LogLevel = level
	}

	c.EnableUpdateCheck = helper.GetBoolean("enable_update_check", true)
	c.TimeBetweenUpdateChecks = helper.GetInt("time_between_update_checks", DEFAULT_TIME_BETWEEN_UPDATE_CHECKS)

	c.EnableAutoChangeByIp = helper.GetBoolean("enable_auto_change_by_ip", false)
	c.WhatToDoWhenNoIpMatches = helper.GetString("what_to_do_when_no_ip_matches", DEACTIVATE_PROXY)
	c.TimeBetweenIpChecks = helper.GetInt("time_between_ips_checks", DEFAULT_TIME_BETWEEN_IP_CHECKS)

	c.ProxyChangeScript = helper.GetString("proxy_change_script", "")
	c.ProxyDeactivateScript = helper.GetString("proxy_deactivate_script", "")

	c.ExcludedInterfacesRegexps = helper.GetListOfStrings("excluded_interfaces_regexps", DEFAULT_EXCLUDED_INTERFACES_REGEXPS)
	c.ExcludedInterfacesRegexpsParsed = []*regexp.Regexp{}
	for _, s := range c.ExcludedInterfacesRegexps {
		regexp, err := regexp.Compile(s)
		if err != nil {
			return nil, errors.Wrapf(err, "Error compiling regexp %v", s)
		}
		c.ExcludedInterfacesRegexpsParsed = append(c.ExcludedInterfacesRegexpsParsed, regexp)
	}

	c.DisabledApplicationsIds = helper.GetListOfStrings("disabled_applications", []string{})

	for _, v := range helper.GetListOfHelpers("proxies") {
		p, err := NewProxyFromMap(c, v, loadPasswordsFromMap)
		if err != nil {
			return []string{}, errors.Wrapf(err, "Error loading proxy")
		}
		Log.Debugf("Loaded proxy %v", p.Name)
		c.AddProxy(false, p)
	}

	activeProxyUuid := helper.GetString("active_proxy", "")
	if activeProxyUuid != "" {
		c.ActiveProxy = c.GetProxyWithUuid(activeProxyUuid)
		if c.ActiveProxy != nil {
			if setActiveProxy {
				c.SetActiveProxy(c.ActiveProxy, MyGettextv("Configuration loaded"), false)
			}
		} else {
			// TODO WARN/LOG
			Log.Warningf("Default proxy with UUID %v not found", activeProxyUuid)
		}
	}

	for _, l := range c.Listeners {
		l.OnConfigLoaded()
	}

	if saveAfterLoad {
		return warnings, c.Save("Saving after load")
	} else {
		return warnings, nil
	}

}

func (c *Configuration) AddListener(l ConfigListener) {
	c.Listeners = append(c.Listeners, l)
}

func (c *Configuration) ToMap(includePasswords bool) (*goutils.MapHelper, error) {

	h := goutils.NewEmptyMapHelper()
	h.SetInt("version", 1)
	if c.IndicatorAlreadyRun {
		h.SetBoolean("indicator_already_run", c.IndicatorAlreadyRun)
	}
	if c.ShowCurrentProxyNameNextToIndicator {
		h.SetBoolean("show_current_proxy_name_next_to_indicator", c.ShowCurrentProxyNameNextToIndicator)
	}
	if c.LogLevel != loggo.WARNING {
		h.SetString("log_level", c.LogLevel.String())
	}

	if c.EnableAutoChangeByIp {
		h.SetBoolean("enable_auto_change_by_ip", c.EnableAutoChangeByIp)
	}
	if c.WhatToDoWhenNoIpMatches != "" && c.WhatToDoWhenNoIpMatches != DEACTIVATE_PROXY {
		h.SetString("what_to_do_when_no_ip_matches", c.WhatToDoWhenNoIpMatches)
	}
	if c.TimeBetweenIpChecks != DEFAULT_TIME_BETWEEN_IP_CHECKS {
		h.SetInt("time_between_ips_checks", c.TimeBetweenIpChecks)
	}

	if c.ProxyChangeScript != "" {
		h.SetString("proxy_change_script", c.ProxyChangeScript)
	}
	if c.ProxyDeactivateScript != "" {
		h.SetString("proxy_deactivate_script", c.ProxyDeactivateScript)
	}

	if !goutils.StringListsAreEqual(c.ExcludedInterfacesRegexps, DEFAULT_EXCLUDED_INTERFACES_REGEXPS) {
		h.SetListOfStrings("excluded_interfaces_regexps", c.ExcludedInterfacesRegexps)
	}

	if !c.EnableUpdateCheck {
		h.SetBoolean("enable_update_check", c.EnableUpdateCheck)
	}
	if c.TimeBetweenUpdateChecks != DEFAULT_TIME_BETWEEN_UPDATE_CHECKS {
		h.SetInt("time_between_update_checks", c.TimeBetweenUpdateChecks)
	}

	if len(c.DisabledApplicationsIds) > 0 {
		h.SetListOfStrings("disabled_applications", c.DisabledApplicationsIds)
	}

	if len(c.Proxies) > 0 {
		l := []*goutils.MapHelper{}
		for _, v := range c.Proxies {
			data, err := v.ToMap(c, includePasswords)
			if err != nil {
				return nil, errors.Wrapf(err, "Error exporting proxy %v", v.Name)
			}
			l = append(l, data)
		}
		h.SetListOfHelpers("proxies", l)
	}
	if c.ActiveProxy != nil {
		h.SetString("active_proxy", c.ActiveProxy.UUID)
	}

	return h, nil

}

func (c *Configuration) Save(reason string) error {
	if reason == "" {
		reason = "No reason"
	}
	Log.Infof("Saving configuration (reason: %v)...", reason)
	data, err := c.ToMap(false)
	if err != nil {
		return errors.Wrap(err, "Error exporting configuration")
	}
	return data.SaveToJsonFile(c.Filename)
}

func (c *Configuration) Export(filename string) error {
	Log.Debugf("Exporting configuration to file %v...", filename)
	data, err := c.ToMap(true)
	if err != nil {
		return errors.Wrap(err, "Error exporting configuration")
	}
	return data.SaveToJsonFile(filename)
}

func (c *Configuration) NoOp() {

}

func (c *Configuration) GetProxyWithUuid(uuid string) *Proxy {
	for ind, _ := range c.Proxies {
		if c.Proxies[ind].UUID == uuid {
			return c.Proxies[ind]
		}
	}
	return nil
}

func (c *Configuration) GetProxyWithSlug(slug string) *Proxy {
	for ind, _ := range c.Proxies {
		if c.Proxies[ind].Slug == slug {
			return c.Proxies[ind]
		}
	}
	return nil
}

// 3rd return: error when proxy not found
// 2nd return: error when saving configuration
// 1st return: results from applications applying proxy
func (c *Configuration) SetActiveProxyFromUuid(uuid string, reason string, save bool) (*GlobalProxyChangeResult, error, error) {
	p := c.GetProxyWithUuid(uuid)
	if p == nil {
		return nil, nil, errors.Errorf("Proxy with UUID %v not found", uuid)
	}
	results, saveError := c.SetActiveProxy(p, reason, save)
	return results, saveError, nil
}

// 2nd return: error when saving configuration
// 1st return: results from applications applying proxy
func (c *Configuration) SetActiveProxy(p *Proxy, reason string, save bool) (*GlobalProxyChangeResult, error) {

	var err error

	if p != nil {
		Log.Infof("Activating proxy %v", p.Name)
	} else {
		Log.Infof("Deactivating proxy")
	}

	var proxyPassword string
	if p != nil {
		proxyPassword, err = p.GetPassword()
		if err != nil {
			return nil, errors.Wrap(err, MyGettextv("Failed to get proxy password"))
		}
	}

	var proxyUrl string
	if p != nil {
		proxyUrl, err = p.ToUrl(true)
		if err != nil {
			return nil, errors.Wrap(err, MyGettextv("Failed to get proxy URL"))
		}
	}

	var changeScriptResult *ScritpResult
	if c.ProxyChangeScript != "" {
		env := map[string]string{
			"PC_ACTION":                "change",
			"PC_HTTP_PROXY_SIMPLE_URL": "",
			"PC_HTTP_PROXY_FULL_URL":   "",
			"PC_HTTP_PROXY_HOST":       "",
			"PC_HTTP_PROXY_PORT":       "",
			"PC_HTTP_PROXY_USERNAME":   "",
			"PC_HTTP_PROXY_PASSWORD":   "",
		}
		err, pid, exitCode, stdOut, stdErr := goutils.RunCommandAndWait("", strings.NewReader(c.ProxyChangeScript), "bash", []string{}, env)
		changeScriptResult = &ScritpResult{
			Error:  err,
			Pid:    pid,
			Code:   exitCode,
			Stdout: stdOut,
			Stderr: stdErr,
		}
	}

	results := []*AppProxyChangeResult{}
	for _, a := range c.GetEnabledApplications() {
		Log.Debugf("Applying proxy to %v", a.GetSimpleName())
		var result *AppProxyChangeResult
		if p != nil {
			result = a.Apply(p)
		} else {
			result = a.Apply(nil)
		}
		if !result.Skipped() && !result.Success() {
			Log.Errorf("Error applying proxy in application %v: %v.\n", a.GetSimpleName(), result.ErrorMessage)
		}
		results = append(results, result)
	}
	c.ActiveProxy = p

	var activateScriptResult *ScritpResult
	if p != nil {
		if p.ActivateScript != "" {
			env := map[string]string{
				"PC_ACTION":                "activate",
				"PC_HTTP_PROXY_SIMPLE_URL": p.ToSimpleUrl(),
				"PC_HTTP_PROXY_FULL_URL":   proxyUrl,
				"PC_HTTP_PROXY_HOST":       p.Address,
				"PC_HTTP_PROXY_PORT":       strconv.Itoa(p.Port),
				"PC_HTTP_PROXY_USERNAME":   p.Username,
				"PC_HTTP_PROXY_PASSWORD":   proxyPassword,
				"PC_HTTP_PROXY_EXCEPTIONS": strings.Join(p.Exceptions, ","),
			}
			err, pid, exitCode, stdOut, stdErr := goutils.RunCommandAndWait("", strings.NewReader(p.ActivateScript), "bash", []string{}, env)
			activateScriptResult = &ScritpResult{
				Error:  err,
				Pid:    pid,
				Code:   exitCode,
				Stdout: stdOut,
				Stderr: stdErr,
			}
		}
	} else {
		if c.ProxyDeactivateScript != "" {
			env := map[string]string{
				"PC_ACTION":                "deactivate",
				"PC_HTTP_PROXY_SIMPLE_URL": "",
				"PC_HTTP_PROXY_FULL_URL":   "",
				"PC_HTTP_PROXY_HOST":       "",
				"PC_HTTP_PROXY_PORT":       "",
				"PC_HTTP_PROXY_USERNAME":   "",
				"PC_HTTP_PROXY_PASSWORD":   "",
				"PC_HTTP_PROXY_EXCEPTIONS": "",
			}
			err, pid, exitCode, stdOut, stdErr := goutils.RunCommandAndWait("", strings.NewReader(c.ProxyDeactivateScript), "bash", []string{}, env)
			activateScriptResult = &ScritpResult{
				Error:  err,
				Pid:    pid,
				Code:   exitCode,
				Stdout: stdOut,
				Stderr: stdErr,
			}
		}
	}

	n := &GlobalProxyChangeResult{
		Proxy:                p,
		Reason:               reason,
		Results:              results,
		ChangeScriptResult:   changeScriptResult,
		ActivateScriptResult: activateScriptResult,
	}
	c.LastExecutionResults = n
	for _, l := range c.Listeners {
		l.OnProxyActivated(n)
	}

	var saveError error
	if save {
		if p != nil {
			saveError = c.Save(MyGettextv("Proxy %v activated", p.Name))
		} else {
			saveError = c.Save(MyGettextv("Proxy deactivated"))
		}
	}
	return n, saveError

}

func (c *Configuration) IsNameAlreadyInUse(name string, proxyToExclude *Proxy) bool {
	for _, p := range c.Proxies {
		if p != proxyToExclude {
			if p.Name == name {
				return true
			}
		}
	}
	return false
}

func (c *Configuration) IsSlugAlreadyInUse(slug string, proxyToExclude *Proxy) bool {
	for _, p := range c.Proxies {
		if p != proxyToExclude {
			if p.Slug == slug {
				return true
			}
		}
	}
	return false
}

func (c *Configuration) AddProxyFromData(save bool, newSlug, newName, newProtocol, newAddress, newUsername, newPassword string, newPort int, newExceptions, newMatchingIps []string, activateScript string) (error, string) {
	p, err := NewProxyFromMap(c, goutils.NewEmptyMapHelper(), false)
	if err != nil {
		return errors.Wrapf(err, "Error creating new proxy"), "_error_creating"
	}
	return c.UpdateProxyFromData(save, p, newSlug, newName, newProtocol, newAddress, newPort, newUsername, newPassword, newExceptions, newMatchingIps, activateScript)
}

func (c *Configuration) AddProxy(save bool, p *Proxy) (error, string) {
	return c.UpdateProxy(save, p)
}

func (c *Configuration) UpdateProxyFromUuid(save bool, uuid, newSlug, newName, newProtocol, newAddress string, newPort int, newUsername, newPassword string, newExceptions, newMatchingIps []string, activateScript string) (error, string) {
	p := c.GetProxyWithUuid(uuid)
	if p == nil {
		return errors.Errorf("Proxy with UUID %v not found", uuid), "_uuid_not_found"
	}
	return c.UpdateProxyFromData(save, p, newSlug, newName, newProtocol, newAddress, newPort, newUsername, newPassword, newExceptions, newMatchingIps, activateScript)
}

func (c *Configuration) CreateUniqueSlug(name string, proxyToExclude *Proxy) string {
	nameToSlug := name
	for i := 1; ; i++ {
		slug := slug.Make(nameToSlug)
		if !c.IsSlugAlreadyInUse(slug, proxyToExclude) {
			return slug
		} else {
			nameToSlug = name + "-" + strconv.Itoa(i)
		}
	}
}

func (c *Configuration) CreateUniqueName(baseName string, proxyToExclude *Proxy) string {
	if baseName == "" {
		baseName = MyGettextv("Random name")
	}
	generatedName := baseName
	for i := 1; ; i++ {
		if !c.IsNameAlreadyInUse(generatedName, proxyToExclude) {
			return generatedName
		} else {
			if baseName == "" {
				generatedName = MyGettextv("Random name #%v", strconv.Itoa(i))
			} else {
				generatedName = MyGettextv("%v #%v", baseName, strconv.Itoa(i))
			}
		}
	}
}

func (c *Configuration) UpdateProxyFromData(save bool, p *Proxy, newSlug, newName, newProtocol, newAddress string, newPort int, newUsername, newPassword string, newExceptions, newMatchingIps []string, activateScript string) (error, string) {

	var err error

	if newName == "" {
		return errors.New(MyGettextv("Name can not be empty")), "name"
	}

	if c.IsNameAlreadyInUse(newName, p) {
		return errors.New(MyGettextv("Proxy with name %v already exists", newName)), "name"
	}

	if newSlug == "" {
		newSlug = c.CreateUniqueSlug(newName, p)
	} else if c.IsSlugAlreadyInUse(newSlug, p) {
		return errors.New(MyGettextv("Proxy with slug %v already exists", newSlug)), "slug"
	}

	if newAddress == "" {
		return errors.New(MyGettextv("Address can not be empty")), "slug"
	}

	if newPort <= 0 {
		return errors.New(MyGettextv("Port must be greater than 0")), "port"
	}

	if newPort >= 65535 {
		return errors.New(MyGettextv("Port must be lower than 65535")), "port"
	}

	/*
		for i, v := range newMatchingIps {

		}
	*/

	p.Slug = newSlug
	p.Name = newName
	p.Protocol = newProtocol
	p.Address = newAddress
	p.Port = newPort
	p.Username = newUsername
	p.Exceptions = newExceptions
	p.MatchingIps = newMatchingIps
	p.ActivateScript = activateScript
	if newPassword != "" {
		err = c.SetPassword(p.UUID, newPassword)
		if err != nil {
			return errors.Wrap(err, MyGettextv("Error saving password in keyring")), "password"
		}
	} else {
		err = c.DeletePassword(p.UUID)
		if err != nil {
			return errors.Wrap(err, MyGettextv("Error saving password in keyring")), "password"
		}
	}

	return c.UpdateProxy(save, p)

}

func (c *Configuration) UpdateProxy(save bool, p *Proxy) (error, string) {

	var foundProxy *Proxy
	for ind, _ := range c.Proxies {
		if p == c.Proxies[ind] {
			foundProxy = c.Proxies[ind]
			break
		}
	}

	if foundProxy == nil {
		c.Proxies = append(c.Proxies, p)
		for _, l := range c.Listeners {
			l.OnProxyAdded(p)
		}
		if save {
			return c.Save(MyGettextv("Proxy %v added", p.Name)), "_saving"
		} else {
			return nil, ""
		}
	} else {
		for _, l := range c.Listeners {
			l.OnProxyUpdated(p)
		}
		if save {
			return c.Save(MyGettextv("Proxy %v updated", p.Name)), "_saving"
		} else {
			return nil, ""
		}
	}

}

func (c *Configuration) DeleteProxyFromUuid(uuid string, save bool) error {
	p := c.GetProxyWithUuid(uuid)
	if p == nil {
		return errors.New(MyGettextv("Proxy with UUID %v not found", uuid))
	}
	return c.DeleteProxy(p, save)
}

func (c *Configuration) DeleteProxy(p *Proxy, save bool) error {

	if c.ActiveProxy == p {
		return errors.New(MyGettextv("The proxy %v is currently in use", p.Name))
	}
	err := c.DeletePassword(p.UUID)
	if err != nil {
		return errors.Wrap(err, MyGettextv("Error deleting password in keyring"))
	}
	newProxies := []*Proxy{}
	for ind, _ := range c.Proxies {
		if p != c.Proxies[ind] {
			newProxies = append(newProxies, c.Proxies[ind])
		}
	}
	c.Proxies = newProxies
	for _, l := range c.Listeners {
		l.OnProxyRemoved(p)
	}

	if save {
		return c.Save(MyGettextv("Proxy %v deleted", p.Name))
	} else {
		return nil
	}

}

func (c *Configuration) ListProxies() ([]*Proxy, error) {
	return c.Proxies, nil
}

func (c *Configuration) GetEnabledApplications() []ProxifiedApplication {
	apps := []ProxifiedApplication{}
	for _, app := range ProxifiedApplications {
		if c.IsApplicationEnabled(app.GetId()) {
			apps = append(apps, app)
		}
	}
	return apps
}

func (c *Configuration) IsApplicationEnabled(appName string) bool {
	return !goutils.ListContainsString(c.DisabledApplicationsIds, appName)
}

func (c *Configuration) EnableApplication(appName string) {
	c.DisabledApplicationsIds = goutils.RemoveStringFromList(c.DisabledApplicationsIds, appName)
}

func (c *Configuration) DisableApplication(appName string) {
	c.DisabledApplicationsIds = goutils.AddStringToList(c.DisabledApplicationsIds, appName)
}

func (c *Configuration) GetPassword(uuid string) (string, error) {
	password, err := keyring.Get(APP_ID, uuid)
	if err != nil && err != keyring.ErrNotFound {
		return "", err
	} else {
		return password, nil
	}
}

func (c *Configuration) SetPassword(uuid string, password string) error {
	return keyring.Set(APP_ID, uuid, password)
}

func (c *Configuration) DeletePassword(uuid string) error {
	err := keyring.Delete(APP_ID, uuid)
	if err != nil && err != keyring.ErrNotFound {
		return err
	} else {
		return nil
	}
}

func (c *Configuration) SetShowCurrentProxyNameNextToIndicator(value bool) {
	c.ShowCurrentProxyNameNextToIndicator = value
	for _, l := range c.Listeners {
		l.OnShowProxyNameNextToIndicatorChanged(value)
	}
}

func (c *Configuration) GetIndicatorAutostart() (bool, error) {

	exists, err := goutils.FileExists(AUTOSTART_FILE)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, errors.New(MyGettextv("File %v doesn't exist", AUTOSTART_FILE))
	}

	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, AUTOSTART_FILE)
	if err != nil {
		return false, err
	}

	section := cfg.Section("Desktop Entry")
	key, err := section.GetKey("X-GNOME-Autostart-enabled")
	if err != nil {
		// If section doesn't exists, the default value is true
		return true, nil
	}

	val, err := key.Bool()
	if err != nil {
		return false, err
	}

	return val, nil

}

func (c *Configuration) SetIndicatorAutostart(value bool) error {

	exists, err := goutils.FileExists(AUTOSTART_FILE)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New(MyGettextv("File %v doesn't exist", AUTOSTART_FILE))
	}

	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, AUTOSTART_FILE)
	if err != nil {
		return err
	}

	cfg.Section("Desktop Entry").Key("X-GNOME-Autostart-enabled").SetValue(map[bool]string{true: "true", false: "false"}[value])
	err = cfg.SaveTo(AUTOSTART_FILE)
	if err != nil {
		return err
	}

	return nil

}

func (c *Configuration) GetProxyPassword(p *goutils.Proxy) (string, error) {
	password, err := c.GetPassword(p.UUID)
	if err != nil {
		return "", err
	}
	return password, nil
}
func (c *Configuration) SetProxyPassword(p *goutils.Proxy, password string) error {
	return c.SetPassword(p.UUID, password)
}

func (c *Configuration) DbusListProxies() (string, *dbus.Error) {
	Log.Debugf("Received dbus request to DbusListProxies...")
	response := DbusListProxiesResponse{}
	response.Proxies = []ProxyStruct{}
	for _, v := range c.Proxies {
		p := ProxyStruct{
			UUID:        v.UUID,
			Name:        v.Name,
			Slug:        v.Slug,
			Protocol:    v.Protocol,
			Address:     v.Address,
			Port:        v.Port,
			Username:    v.Username,
			Exceptions:  v.Exceptions,
			MatchingIps: v.MatchingIps,
			Active:      c.ActiveProxy == v,
		}
		response.Proxies = append(response.Proxies, p)
	}
	b, err := json.Marshal(response)
	if err != nil {
		return "", dbus.NewError("Error marshaling", nil)
	}
	return string(b), nil
}

func (c *Configuration) SetActiveProxyBySlug(slug string) (string, *dbus.Error) {
	Log.Debugf("Received dbus request to SetActiveProxyBySlug...")
	response := DbusSetActiveProxyBySlugResponse{}
	proxy := c.GetProxyWithSlug(slug)
	if slug == "none" {
		_, err := c.SetActiveProxy(nil, MyGettextv("Proxy deactivated from dbus"), true)
		if err != nil {
			response.Error = err.Error()
		}
	} else {
		if proxy == nil {
			response.Error = MyGettextv("Proxy with slug %v not found", slug)
		} else {
			_, err := c.SetActiveProxy(proxy, MyGettextv("Proxy activated from dbus"), true)
			if err != nil {
				response.Error = err.Error()
			}
		}
	}
	b, err := json.Marshal(response)
	if err != nil {
		return "", dbus.NewError("Error marshaling", nil)
	}
	return string(b), nil

}

func (c *Configuration) SetProxyForIps(ips []string) {

	Log.Debugf("Received new ips %v.", ips)
	if !c.EnableAutoChangeByIp {
		return
	}

	var foundProxy *Proxy
	for _, p := range c.Proxies {
		matches := p.MatchesIps(ips)
		if matches {
			foundProxy = p
			break
		}
	}

	if foundProxy != nil {
		Log.Debugf("Detected proxy %v for ips %v.", foundProxy.Name, ips)
		if foundProxy != c.ActiveProxy {
			Log.Tracef("Changing proxy to %v", foundProxy.Name)
			c.SetActiveProxy(foundProxy, MyGettextv("Proxy activated according to current IPs"), true)
		} else {
			Log.Tracef("Not changing proxy because is the same than the current active")
		}
	} else if c.WhatToDoWhenNoIpMatches == DEACTIVATE_PROXY {
		Log.Tracef("No proxy found for IPs %v", ips)
		if foundProxy != c.ActiveProxy {
			Log.Tracef("Deactivating proxy")
			c.SetActiveProxy(nil, MyGettextv("No proxy found for current IPs"), true)
		} else {
			Log.Tracef("Not deactivating proxy because already deactivated")
		}
	}

}

func (c *Configuration) SetEnableAutoChangeByIp(value bool) {
	c.EnableAutoChangeByIp = value
	for _, l := range c.Listeners {
		l.OnEnableAutoChangeByIpChanged(value)
	}
}

func (c *Configuration) SetEnableUpdateCheck(value bool) {
	c.EnableUpdateCheck = value
	for _, l := range c.Listeners {
		l.OnEnableUpdateCheckChanged(value)
	}
}

func (c *Configuration) SetWhatToDoWhenNoIpMatches(value string) {
	c.WhatToDoWhenNoIpMatches = value
	for _, l := range c.Listeners {
		l.OnWhatToDoWhenNoIpMatchesChanged(value)
	}
}

func (c *Configuration) SetLogLevel(value loggo.Level) {
	c.LogLevel = value
	for _, l := range c.Listeners {
		l.OnLogLevelChanged(value)
	}
}
