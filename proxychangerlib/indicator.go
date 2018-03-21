package proxychangerlib

import (
	"fmt"
	"strings"

	"github.com/juju/loggo"

	"github.com/esiqveland/notify"
	"github.com/godbus/dbus"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	appindicator "github.com/jamesadney/appindicator"
	appindicatorgtk3 "github.com/jamesadney/appindicator/gtk-extensions/gotk3"
	"github.com/okelet/goutils"
	"github.com/okelet/goutils/updatechecker"
	"github.com/pkg/errors"
)

type Indicator struct {
	Config             *Configuration
	CurrentVersion     string
	TestMode           bool
	NewVersionDetected string
	CmdLogLevelSet     bool

	CheckUpdatesThread *updatechecker.CheckUpdatesThread

	CheckIpsThread *CheckIpsThread

	SessionBus   *dbus.Conn
	AppIndicator *appindicatorgtk3.AppIndicatorGotk3
	ConfigWindow *ConfigWindow

	NoProxyRadioItem       *gtk.RadioMenuItem
	NoProxyRadioItemHandle glib.SignalHandle

	NotificationId uint32
}

func NewIndicator(sessionBus *dbus.Conn, config *Configuration, currentVersion string, cmdLogLevelSet bool, testMode bool) (*Indicator, error) {

	var err error

	i := Indicator{}

	i.Config = config
	i.CurrentVersion = currentVersion
	i.TestMode = testMode
	i.CmdLogLevelSet = cmdLogLevelSet

	i.SessionBus = sessionBus
	i.AppIndicator = appindicatorgtk3.NewAppIndicator(APP_ID, ICON_NAME, appindicator.CategoryApplicationStatus)
	i.AppIndicator.SetStatus(appindicator.StatusActive)

	i.ConfigWindow, err = NewConfigWindow(&i)
	if err != nil {
		return nil, errors.Wrap(err, MyGettextv("Error creating configuration window"))
	}

	i.CheckIpsThread = NewCheckIpsThread(i.Config.TimeBetweenIpChecks, i.Config)

	if i.TestMode {
		i.CheckUpdatesThread = updatechecker.NewCheckUpdatesThread(30, "okelet", "proxychanger", "master", true)
	} else {
		i.CheckUpdatesThread = updatechecker.NewCheckUpdatesThread(i.Config.TimeBetweenUpdateChecks, "okelet", "proxychanger", i.CurrentVersion, true)
	}

	return &i, nil

}

func (i *Indicator) BuildMenu() error {

	var err error

	menu, err := gtk.MenuNew()
	if err != nil {
		Log.Errorf("Error building GUI: %v.", err)
		goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return nil
	}

	noProxyRadioItem, err := gtk.RadioMenuItemNewWithLabel(nil, MyGettextv("No proxy"))
	if err != nil {
		Log.Errorf("Error building GUI: %v.", err)
		goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return nil
	} else {
		if i.Config.ActiveProxy == nil {
			noProxyRadioItem.SetActive(true)
		}
		handle, err := noProxyRadioItem.Connect("activate", i.OnProxyItemDeActivated)
		if err != nil {
			Log.Errorf("Error building GUI: %v.", err)
			goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		} else {
			menu.Append(noProxyRadioItem)
			i.NoProxyRadioItem = noProxyRadioItem
			i.NoProxyRadioItemHandle = handle
		}
		if i.Config.EnableAutoChangeByIp && i.Config.WhatToDoWhenNoIpMatches == DEACTIVATE_PROXY {
			i.NoProxyRadioItem.SetSensitive(false)
		} else {
			i.NoProxyRadioItem.SetSensitive(true)
		}
	}

	group, err := noProxyRadioItem.GetGroup()
	if err != nil {
		Log.Errorf("Error building GUI: %v.", err)
		goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return nil
	}

	if len(i.Config.Proxies) > 0 {
		for _, p := range i.Config.Proxies {
			radioItem, err := gtk.RadioMenuItemNewWithLabel(group, p.Name)
			if err != nil {
				Log.Errorf("Error building GUI: %v.", err)
				goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
				return nil
			} else {
				if i.Config.ActiveProxy == p {
					radioItem.SetActive(true)
				}
				handle, err := radioItem.Connect("activate", i.OnProxyItemActivated, p)
				if err != nil {
					Log.Errorf("Error building GUI: %v.", err)
					goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
				} else {
					menu.Append(radioItem)
					p.RadioMenuItem = radioItem
					p.RadioMenuItemHandle = handle
				}
				if i.Config.EnableAutoChangeByIp && i.Config.WhatToDoWhenNoIpMatches == DEACTIVATE_PROXY {
					radioItem.SetSensitive(false)
				} else {
					radioItem.SetSensitive(true)
				}
			}
		}
	} else {
		radioItem, err := gtk.RadioMenuItemNewWithLabel(group, MyGettextv("No proxies defined"))
		if err != nil {
			Log.Errorf("Error building GUI: %v.", err)
			goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
			return nil
		} else {
			radioItem.SetSensitive(false)
			menu.Append(radioItem)
		}
	}

	sepItem, err := gtk.SeparatorMenuItemNew()
	if err != nil {
		Log.Errorf("Error building GUI: %v.", err)
		goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
	} else {
		menu.Append(sepItem)
	}

	item, err := gtk.MenuItemNewWithLabel(MyGettextv("Configuration"))
	if err != nil {
		Log.Errorf("Error building GUI: %v.", err)
		goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return nil
	} else {
		_, err := item.Connect("activate", i.ShowConfigurationWindow)
		if err != nil {
			Log.Errorf("Error building GUI: %v.", err)
			goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		}
		menu.Append(item)
	}

	itemLabel := MyGettextv("Status")
	if i.Config.LastExecutionResults != nil {
		if i.Config.LastExecutionResults.GetNumberOfErrors() > 0 {
			itemLabel = MyGettextv("%v Status", "\u2716")
		} else {
			itemLabel = MyGettextv("%v Status", "\u2714")
		}
	}
	item, err = gtk.MenuItemNewWithLabel(itemLabel)
	if err != nil {
		Log.Errorf("Error building GUI: %v.", err)
		goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return nil
	} else {
		if i.Config.LastExecutionResults != nil {
			_, err := item.Connect("activate", i.ShowLastExecutionResults)
			if err != nil {
				Log.Errorf("Error building GUI: %v.", err)
				goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
			}
		} else {
			item.SetSensitive(false)
		}
		menu.Append(item)
	}

	if i.NewVersionDetected != "" {
		item, err = gtk.MenuItemNewWithLabel(MyGettextv("New version %v released, click to update", i.NewVersionDetected))
		if err != nil {
			Log.Errorf("Error building GUI: %v.", err)
			goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
			return nil
		} else {
			_, err := item.Connect("activate", goutils.XdgOpenFromMenuItem, fmt.Sprintf("https://github.com/okelet/proxychanger/releases/tag/%v", i.NewVersionDetected))
			if err != nil {
				Log.Errorf("Error building GUI: %v.", err)
				goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
			}
			menu.Append(item)
		}
	}

	item, err = gtk.MenuItemNewWithLabel(MyGettextv("Show LOG"))
	if err != nil {
		Log.Errorf("Error building GUI: %v.", err)
		goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return nil
	} else {
		_, err := item.Connect("activate", goutils.XdgOpenFromMenuItem, LOG_PATH)
		if err != nil {
			Log.Errorf("Error building GUI: %v.", err)
			goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		}
		menu.Append(item)
	}

	item, err = gtk.MenuItemNewWithLabel(MyGettextv("Help"))
	if err != nil {
		Log.Errorf("Error building GUI: %v.", err)
		goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return nil
	} else {
		_, err := item.Connect("activate", goutils.XdgOpenFromMenuItem, fmt.Sprintf("https://github.com/okelet/proxychanger?currentVersion=%v", i.CurrentVersion))
		if err != nil {
			Log.Errorf("Error building GUI: %v.", err)
			goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		}
		menu.Append(item)
	}

	item, err = gtk.MenuItemNewWithLabel(MyGettextv("Quit"))
	if err != nil {
		Log.Errorf("Error building GUI: %v.", err)
		goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return nil
	} else {
		_, err := item.Connect("activate", i.Quit)
		if err != nil {
			Log.Errorf("Error building GUI: %v.", err)
			goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		}
		menu.Append(item)
	}

	menu.ShowAll()
	i.AppIndicator.SetMenu(menu)

	return nil

}

func (i *Indicator) OnProxyItemDeActivated(item *gtk.RadioMenuItem) {
	if item.GetActive() {
		i.Config.SetActiveProxy(nil, "", true)
	}
}

func (i *Indicator) OnProxyItemActivated(item *gtk.RadioMenuItem, p *Proxy) {
	if item.GetActive() {
		i.Config.SetActiveProxy(p, "", true)
	}
}

func (i *Indicator) Run(setProxyNow bool) error {

	var err error

	// FIXME: try to import both, and detect if they are different
	var p *Proxy
	firstRun := !i.Config.IndicatorAlreadyRun
	if firstRun {
		gnomeProxy, err := goutils.GetGnomeProxy(i.Config)
		if err != nil {
			Log.Errorf("Error loading gnome proxy: %v", err)
		} else if gnomeProxy != nil {
			Log.Infof("Importing gnome proxy %v", gnomeProxy.ToSimpleUrl())
			p = NewImportedProxy(i.Config, gnomeProxy, MyGettextv("Gnome imported"), "")
			i.Config.AddProxy(false, p)
		} else {
			envProxy, err := goutils.GetEnvironmentProxy(i.Config)
			if err != nil {
				Log.Errorf("Error loading environment proxy: %v", err)
			} else if envProxy != nil {
				Log.Infof("Importing environment proxy %v", envProxy.ToSimpleUrl())
				p = NewImportedProxy(i.Config, envProxy, MyGettextv("Environment imported"), "")
				i.Config.AddProxy(false, p)
			}
		}
		i.Config.IndicatorAlreadyRun = true
		i.Config.Save(MyGettextv("Initial indicator configuration and proxy import"))
	}

	err = i.BuildMenu()
	if err != nil {
		return errors.Wrap(err, "Error building indicator menu")
	}

	i.UpdateLabel()

	i.Config.AddListener(i)

	if setProxyNow {
		// If just imported
		if p != nil {
			i.Config.SetActiveProxy(p, MyGettextv("Startup"), true)
		} else {
			i.Config.SetActiveProxy(i.Config.ActiveProxy, MyGettextv("Startup"), false)
		}
	}

	// Don not wait to cron to check the ips, check them now
	i.CheckIpsThread.AddListener(i)
	if i.Config.EnableAutoChangeByIp {
		i.CheckIpsThread.Check()
		err = i.CheckIpsThread.Start()
		if err != nil {
			return errors.Wrap(err, MyGettextv("Error starting IP check thread"))
		}
	}

	i.CheckUpdatesThread.AddListener(i)
	if i.Config.EnableUpdateCheck {
		if i.TestMode {
			i.CheckUpdatesThread.Check()
		}
		err = i.CheckUpdatesThread.Start()
		if err != nil {
			return errors.Wrap(err, MyGettextv("Error starting update check thread"))
		}
	}

	if firstRun && len(i.Config.Proxies) == 0 {
		if goutils.ConfirmMessage(
			nil,
			MyGettextv("Initial configuration"),
			MyGettextv("Do you want to add a proxy?"),
		) {
			i.ConfigWindow.FillData()
			i.ConfigWindow.Window.Show()
			i.ConfigWindow.Window.Present()
			i.ConfigWindow.ShowAddProxyDialog(true)
		}

	}

	return nil

}

func (i *Indicator) UpdateLabel() {
	if i.Config.ShowCurrentProxyNameNextToIndicator {
		n := i.Config.LastExecutionResults
		if n != nil {
			if n.Proxy != nil {
				if n.GetNumberOfErrors() > 0 {
					i.AppIndicator.SetLabel(MyGettextv("%v %v", "\u2757", n.Proxy.Name), "")
				} else {
					i.AppIndicator.SetLabel(n.Proxy.Name, "")
				}
			} else {
				if n.GetNumberOfErrors() > 0 {
					i.AppIndicator.SetLabel(MyGettextv("%v No proxy", "\u2757"), "")
				} else {
					i.AppIndicator.SetLabel(MyGettextv("No proxy"), "")
				}
			}
		} else {
			if i.Config.ActiveProxy != nil {
				i.AppIndicator.SetLabel(i.Config.ActiveProxy.Name, "")
			} else {
				i.AppIndicator.SetLabel(MyGettextv("No proxy"), "")
			}
		}
	} else {
		i.AppIndicator.SetLabel("", "")
	}
}

func (i *Indicator) ShowLastExecutionResults() {

	lines := []string{}
	if i.Config.LastExecutionResults != nil {

		changeScriptResult := i.Config.LastExecutionResults.ChangeScriptResult
		if changeScriptResult != nil {
			if changeScriptResult.Error != nil {
				lines = append(lines, MyGettextv("%v: ERROR (%v)", "Before proxy change script", changeScriptResult.Error))
			} else if changeScriptResult.Code != 0 {
				lines = append(lines, MyGettextv("%v: WARNING (%v, %v)", "Before proxy change script", changeScriptResult.Code, changeScriptResult.GetCombinedOutput()))
			} else {
				combinedOutput := changeScriptResult.GetCombinedOutput()
				if combinedOutput != "" {
					lines = append(lines, MyGettextv("%v: OK (%v)", "Before proxy change script", combinedOutput))
				} else {
					lines = append(lines, MyGettextv("%v: OK", "Before proxy change script"))
				}
			}
		} else {
			lines = append(lines, MyGettextv("%v: Skipped (%v)", "Before proxy change script", MyGettextv("not configured")))
		}

		globalActivateScriptResult := i.Config.LastExecutionResults.GlobalActivateScriptResult
		if globalActivateScriptResult != nil {
			if globalActivateScriptResult.Error != nil {
				lines = append(lines, MyGettextv("%v: ERROR (%v)", "Global activate proxy script", globalActivateScriptResult.Error))
			} else if globalActivateScriptResult.Code != 0 {
				lines = append(lines, MyGettextv("%v: WARNING (%v, %v)", "Global activate proxy script", globalActivateScriptResult.Code, globalActivateScriptResult.GetCombinedOutput()))
			} else {
				combinedOutput := globalActivateScriptResult.GetCombinedOutput()
				if combinedOutput != "" {
					lines = append(lines, MyGettextv("%v: OK (%v)", "Global activate proxy script", combinedOutput))
				} else {
					lines = append(lines, MyGettextv("%v: OK", "Global activate proxy script"))
				}
			}
		} else {
			lines = append(lines, MyGettextv("%v: Skipped (%v)", "Global activate proxy script", MyGettextv("not configured")))
		}

		globalDeactivateScriptResult := i.Config.LastExecutionResults.GlobalDeactivateScriptResult
		if globalDeactivateScriptResult != nil {
			if globalDeactivateScriptResult.Error != nil {
				lines = append(lines, MyGettextv("%v: ERROR (%v)", "Global deactivate proxy script", globalDeactivateScriptResult.Error))
			} else if globalDeactivateScriptResult.Code != 0 {
				lines = append(lines, MyGettextv("%v: WARNING (%v, %v)", "Global deactivate proxy script", globalDeactivateScriptResult.Code, globalDeactivateScriptResult.GetCombinedOutput()))
			} else {
				combinedOutput := globalDeactivateScriptResult.GetCombinedOutput()
				if combinedOutput != "" {
					lines = append(lines, MyGettextv("%v: OK (%v)", "Global deactivate proxy script", combinedOutput))
				} else {
					lines = append(lines, MyGettextv("%v: OK", "Global deactivate proxy script"))
				}
			}
		} else {
			lines = append(lines, MyGettextv("%v: Skipped (%v)", "Global deactivate proxy script", MyGettextv("not configured")))
		}

		proxyActivateScriptResult := i.Config.LastExecutionResults.ProxyActivateScriptResult
		if proxyActivateScriptResult != nil {
			if proxyActivateScriptResult.Error != nil {
				lines = append(lines, MyGettextv("%v: ERROR (%v)", "Activate proxy script", proxyActivateScriptResult.Error))
			} else if proxyActivateScriptResult.Code != 0 {
				lines = append(lines, MyGettextv("%v: WARNING (%v, %v)", "Activate proxy script", proxyActivateScriptResult.Code, proxyActivateScriptResult.GetCombinedOutput()))
			} else {
				combinedOutput := proxyActivateScriptResult.GetCombinedOutput()
				if combinedOutput != "" {
					lines = append(lines, MyGettextv("%v: OK (%v)", "Activate proxy script", combinedOutput))
				} else {
					lines = append(lines, MyGettextv("%v: OK", "Activate proxy script"))
				}
			}
		} else {
			lines = append(lines, MyGettextv("%v: Skipped (%v)", "Activate proxy script", MyGettextv("not configured")))
		}

		for _, r := range i.Config.LastExecutionResults.Results {
			if r.Skipped() {
				lines = append(lines, MyGettextv("%v: Skipped (%v)", r.Application.GetSimpleName(), r.SkippedMessage))
			} else if r.Success() {
				lines = append(lines, MyGettextv("%v: OK", r.Application.GetSimpleName()))
			} else {
				lines = append(lines, MyGettextv("%v: ERROR (%v)", r.Application.GetSimpleName(), r.ErrorMessage))
			}
		}

	} else {
		lines = append(lines, MyGettextv("No proxy set yet."))
	}

	goutils.ShowMessage(nil, gtk.MESSAGE_INFO, "Execution results", strings.Join(lines, "\n"))

}

func (i *Indicator) Quit() {

	i.CheckIpsThread.Stop()
	i.CheckUpdatesThread.Stop()

	gtk.MainQuit()

}

func (i *Indicator) OnProxyActivated(n *GlobalProxyChangeResult) {
	glib.IdleAdd(i.OnProxyActivatedInternal, n)
}

func (i *Indicator) OnProxyActivatedInternal(n *GlobalProxyChangeResult) {

	i.UpdateLabel()

	// Set the selected proxy in the menu
	var item *gtk.RadioMenuItem
	var handle glib.SignalHandle
	if n.Proxy == nil {
		item = i.NoProxyRadioItem
		handle = i.NoProxyRadioItemHandle
		if handle == 0 {
			Log.Errorf("Handle not found for deactivate menu item")
			goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		}
	} else {
		item = n.Proxy.RadioMenuItem
		handle = n.Proxy.RadioMenuItemHandle
		if handle == 0 {
			Log.Errorf("Handle not found for menu item for proxy %v", n.Proxy.Name)
			goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		}
	}

	if handle != 0 {
		item.HandlerBlock(handle)
	}
	item.SetActive(true)
	if handle != 0 {
		item.HandlerUnblock(handle)
	}

	// Show notification
	if n.Proxy != nil {
		if n.Reason != "" && n.GetNumberOfErrors() > 0 {
			i.ShowNotification(MyGettextv("Proxy activated"), MyGettextv("Proxy %v has been activated (%v).\n\n\n%v errors occurred.", n.Proxy.Name, goutils.EnsureFirstSentenceLetterLowercase(n.Reason), n.GetNumberOfErrors()))
		} else if n.Reason != "" {
			i.ShowNotification(MyGettextv("Proxy activated"), MyGettextv("Proxy %v has been activated (%v).", n.Proxy.Name, goutils.EnsureFirstSentenceLetterLowercase(n.Reason)))
		} else if n.GetNumberOfErrors() > 0 {
			i.ShowNotification(MyGettextv("Proxy activated"), MyGettextv("Proxy %v has been activated.\n\n\n%v errors occurred.", n.Proxy.Name, n.GetNumberOfErrors()))
		} else {
			i.ShowNotification(MyGettextv("Proxy activated"), MyGettextv("Proxy %v has been activated.", n.Proxy.Name))
		}
	} else {
		if n.Reason != "" && n.GetNumberOfErrors() > 0 {
			i.ShowNotification(MyGettextv("Proxy deactivated"), MyGettextv("Proxy has been deactivated (%v).\n\n\n%v errors occurred.", goutils.EnsureFirstSentenceLetterLowercase(n.Reason), n.GetNumberOfErrors()))
		} else if n.Reason != "" {
			i.ShowNotification(MyGettextv("Proxy deactivated"), MyGettextv("Proxy has been deactivated (%v).", goutils.EnsureFirstSentenceLetterLowercase(n.Reason)))
		} else if n.GetNumberOfErrors() > 0 {
			i.ShowNotification(MyGettextv("Proxy deactivated"), MyGettextv("Proxy has been deactivated.\n\n\n%v errors occurred.", n.GetNumberOfErrors()))
		} else {
			i.ShowNotification(MyGettextv("Proxy deactivated"), MyGettextv("Proxy has been deactivated."))
		}
	}

	// Rebuild menu
	i.BuildMenu()

}

func (i *Indicator) OnConfigLoaded() {
	i.BuildMenu()
	i.OnShowProxyNameNextToIndicatorChanged(i.Config.ShowCurrentProxyNameNextToIndicator)
}

func (i *Indicator) OnProxyAdded(p *Proxy) {
	glib.IdleAdd(i.OnProxyAddedInternal, p)
}

func (i *Indicator) OnProxyAddedInternal(p *Proxy) {
	i.BuildMenu()
}

func (i *Indicator) OnProxyUpdated(p *Proxy) {
	glib.IdleAdd(i.OnProxyUpdatedInternal, p)
}

func (i *Indicator) OnProxyUpdatedInternal(p *Proxy) {
	i.BuildMenu()
	i.UpdateLabel()
}

func (i *Indicator) OnProxyRemoved(p *Proxy) {
	glib.IdleAdd(i.OnProxyRemovedInternal, p)
}

func (i *Indicator) OnProxyRemovedInternal(p *Proxy) {
	i.BuildMenu()
}

func (i *Indicator) OnShowProxyNameNextToIndicatorChanged(newValue bool) {
	i.UpdateLabel()
}

func (i *Indicator) ShowNotification(title string, text string) {
	var err error
	i.NotificationId, err = notify.SendNotification(i.SessionBus, notify.Notification{
		AppIcon:       ICON_NAME,
		Summary:       title,
		Body:          text,
		ExpireTimeout: int32(5000),
		ReplacesID:    i.NotificationId,
	})
	if err != nil {
		Log.Errorf("Error showing notification: %v.", err)
		goutils.ShowMessage(nil, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
	}
}

func (i *Indicator) ShowConfigurationWindow() {
	i.ConfigWindow.FillData()
	i.ConfigWindow.Window.Show()
	i.ConfigWindow.Window.Present()
}

func (i *Indicator) GetAsset(assetName string) ([]byte, error) {
	return Asset(assetName)
}

func (i *Indicator) OnIpsChanged(ips []string) {
	Log.Tracef("New IPs notification received: %v", ips)
	i.Config.SetProxyForIps(ips)
}

func (i *Indicator) OnNewVersionDetecetd(newVersion string) {
	i.ShowNotification(
		MyGettextv("New version detected"),
		MyGettextv("Version %v has been released.", newVersion),
	)
	if newVersion != i.NewVersionDetected {
		i.NewVersionDetected = newVersion
		glib.IdleAdd(i.BuildMenu)
	}
}

func (i *Indicator) OnEnableAutoChangeByIpChanged(newValue bool) {
	if newValue {
		Log.Debugf("Starting IP check thread...")
		err := i.CheckIpsThread.Start()
		if err != nil {
			i.ShowNotification(
				MyGettextv("Error"),
				MyGettextv("Error starting IP check thread: %v", err),
			)
		}
	} else {
		Log.Debugf("Stopping IP check thread...")
		i.CheckIpsThread.Stop()
	}
	glib.IdleAdd(i.BuildMenu)
}

func (i *Indicator) OnEnableUpdateCheckChanged(newValue bool) {
	if newValue {
		Log.Debugf("Starting update check thread...")
		err := i.CheckUpdatesThread.Start()
		if err != nil {
			i.ShowNotification(
				MyGettextv("Error"),
				MyGettextv("Error staring update check thread: %v", err),
			)
		}
	} else {
		Log.Debugf("Stopping update check thread...")
		i.CheckUpdatesThread.Stop()
	}
}

func (i *Indicator) OnWhatToDoWhenNoIpMatchesChanged(newValue string) {
	glib.IdleAdd(i.BuildMenu)
}

func (i *Indicator) OnLogLevelChanged(newValue loggo.Level) {
	if !i.CmdLogLevelSet {
		Log.SetLogLevel(newValue)
	}
}
