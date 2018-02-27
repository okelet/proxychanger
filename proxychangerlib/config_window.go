package proxychangerlib

import (
	"fmt"
	"strings"

	"github.com/juju/loggo"

	"bytes"

	"github.com/gotk3/gotk3/gtk"
	"github.com/okelet/goutils"
	"github.com/pkg/errors"
)

type ConfigWindow struct {
	*goutils.BuilderBase
	Window *gtk.Window

	InfoBar             *gtk.InfoBar
	SwitchRunStartup    *gtk.Switch
	SwitchShowProxyName *gtk.Switch
	ComboBoxLogLevel    *gtk.ComboBox
	ListStoreLogLevel   *gtk.ListStore

	SwitchEnableAutoSwitch *gtk.Switch
	ComboBoxIpNoMatch      *gtk.ComboBox
	ListStoreNoMatch       *gtk.ListStore

	SwitchUpdateCheck *gtk.Switch

	TreeViewProxies   *gtk.TreeView
	ListStoreProxies  *gtk.ListStore
	ButtonProxyEdit   *gtk.Button
	ButtonProxyRemove *gtk.Button

	TreeViewApps  *gtk.TreeView
	ListStoreApps *gtk.ListStore

	TextViewOnProxyChangeScript       *gtk.TextView
	TextBufferOnProxyChangeScript     *gtk.TextBuffer
	TextViewOnProxyDeactivateScript   *gtk.TextView
	TextBufferOnProxyDeactivateScript *gtk.TextBuffer
	TextViewOnProxyActivateScript     *gtk.TextView
	TextBufferOnProxyActivateScript   *gtk.TextBuffer

	Indicator *Indicator
}

func NewConfigWindow(indicator *Indicator) (*ConfigWindow, error) {

	var err error

	w := ConfigWindow{}
	w.Indicator = indicator

	w.BuilderBase, err = goutils.NewBuilderBase(indicator, "assets/config.glade")
	if err != nil {
		return nil, errors.Wrap(err, "Error loading asset assets/config.glade")
	}

	// ------------------------------------------------------------------------------------

	w.Window, err = w.GetWindow("preferences_window")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting preferences_window")
	}
	w.Window.Resize(700, 600)
	w.Window.SetPosition(gtk.WIN_POS_CENTER)
	w.Window.SetIconName(ICON_NAME)

	// ------------------------------------------------------------------------------------

	w.InfoBar, err = w.GetInfoBar("infobar_log_level")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting infobar_log_level")
	}

	w.SwitchRunStartup, err = w.GetSwitch("switch_run_startup")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting switch_run_startup")
	}

	w.SwitchShowProxyName, err = w.GetSwitch("switch_show_proxy_name")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting switch_show_proxy_name")
	}

	w.ListStoreLogLevel, err = w.GetListStore("liststore_log_level")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting liststore_log_level")
	}

	w.ComboBoxLogLevel, err = w.GetComboBox("combobox_log_level")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting combobox_log_level")
	}

	// ------------------------------------------------------------------------------------

	w.SwitchEnableAutoSwitch, err = w.GetSwitch("switch_enable_auto_change")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting switch_enable_auto_change")
	}

	w.ComboBoxIpNoMatch, err = w.GetComboBox("combobox_ip_no_match")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting combobox_ip_no_match")
	}

	w.ListStoreNoMatch, err = w.GetListStore("liststore_ip_no_match")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting liststore_ip_no_match")
	}

	// ------------------------------------------------------------------------------------

	w.SwitchUpdateCheck, err = w.GetSwitch("switch_check_updates")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting switch_check_updates")
	}

	// ------------------------------------------------------------------------------------

	w.TreeViewProxies, err = w.GetTreeView("treeview_proxies")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting treeview_proxies")
	}

	w.ListStoreProxies, err = w.GetListStore("liststore_proxies")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting liststore_proxies")
	}

	w.ButtonProxyEdit, err = w.GetButton("button_proxy_edit")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting button_proxy_edit")
	}

	w.ButtonProxyRemove, err = w.GetButton("button_proxy_remove")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting button_proxy_remove")
	}

	// ------------------------------------------------------------------------------------

	w.TreeViewApps, err = w.GetTreeView("treeview_apps")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting treeview_apps")
	}

	w.ListStoreApps, err = w.GetListStore("liststore_apps")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting liststore_apps")
	}

	// ------------------------------------------------------------------------------------

	w.TextViewOnProxyChangeScript, err = w.GetTextView("textview_proxy_change_script")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting textview_proxy_change_script")
	}

	w.TextBufferOnProxyChangeScript, err = w.GetTextBuffer("textbuffer_proxy_change_script")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting textbuffer_proxy_change_script")
	}

	w.TextViewOnProxyDeactivateScript, err = w.GetTextView("textview_proxy_deactivate_script")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting textview_proxy_deactivate_script")
	}

	w.TextBufferOnProxyDeactivateScript, err = w.GetTextBuffer("textbuffer_proxy_deactivate_script")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting textbuffer_proxy_deactivate_script")
	}

	w.TextViewOnProxyActivateScript, err = w.GetTextView("textview_proxy_activate_script")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting textview_proxy_activate_script")
	}

	w.TextBufferOnProxyActivateScript, err = w.GetTextBuffer("textbuffer_proxy_activate_script")
	if err != nil {
		return nil, errors.Wrap(err, "Error getting textbuffer_proxy_activate_script")
	}

	// ------------------------------------------------------------------------------------
	// Signals
	// ------------------------------------------------------------------------------------

	w.Builder.ConnectSignals(map[string]interface{}{
		"on_window_deleted":                             w.OnWindowDeleted,
		"on_switch_run_startup_state_set":               w.OnSwitchRunStartupChanged,
		"on_switch_show_proxy_name_state_set":           w.OnSwitchShowProxyNameChanged,
		"on_combobox_log_level_changed":                 w.OnComboBoxLogLevelChanged,
		"on_enable_auto_change":                         w.OnSwitchEnableAutoChangeByIp,
		"on_combobox_ip_no_match_changed":               w.OnComboBoxWhatToDoWhenNoIpMatchesChanged,
		"on_switch_check_updates_changed":               w.OnSwitchUpdateCheckChanged,
		"on_treeview_proxies_row_activated":             w.OnTreeviewProxiesRowActivated,
		"on_treeview_proxies_selection_changed":         w.OnTreeviewProxiesSelectionChanged,
		"on_button_proxy_add_clicked":                   w.OnButtonProxyAddClicked,
		"on_button_proxy_edit_clicked":                  w.OnButtonProxyEditClicked,
		"on_button_proxy_remove_clicked":                w.OnButtonProxyRemoveClicked,
		"on_application_enabled_toggled":                w.OnApplicationEnabledToggled,
		"on_button_export_clicked":                      w.OnExportButtonClicked,
		"on_button_import_clicked":                      w.OnImportButtonClicked,
		"on_button_close_clicked":                       w.OnCloseButtonClicked,
		"on_textbuffer_proxy_change_script_changed":     w.OnTextbufferProxyChangeScriptChanged,
		"on_textbuffer_proxy_deactivate_script_changed": w.OnTextbufferProxyDeactivateScriptChanged,
		"on_textbuffer_proxy_activate_script_changed":   w.OnTextbufferProxyActivateScriptChanged,
	})

	return &w, nil

}

func (w *ConfigWindow) FillData() {

	if w.Indicator.CmdLogLevelSet {
		w.InfoBar.Show()
	} else {
		w.InfoBar.Hide()
	}

	autoStart, err := w.Indicator.Config.GetIndicatorAutostart()
	if err != nil {
		Log.Errorf("Error checking auto start: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Error filling information; please, check the LOG."))
		w.SwitchRunStartup.SetSensitive(false)
		w.SwitchRunStartup.SetTooltipText(err.Error())
	} else {
		w.SwitchRunStartup.SetSensitive(true)
		w.SwitchRunStartup.SetActive(autoStart)
	}
	w.SwitchShowProxyName.SetActive(w.Indicator.Config.ShowCurrentProxyNameNextToIndicator)
	w.ComboBoxLogLevel.SetActiveID(w.Indicator.Config.LogLevel.String())
	w.SwitchEnableAutoSwitch.SetActive(w.Indicator.Config.EnableAutoChangeByIp)
	w.ComboBoxIpNoMatch.SetActiveID(w.Indicator.Config.WhatToDoWhenNoIpMatches)
	w.SwitchUpdateCheck.SetActive(w.Indicator.Config.EnableUpdateCheck)
	if w.Indicator.Config.EnableAutoChangeByIp {
		w.ComboBoxIpNoMatch.SetSensitive(true)
	} else {
		w.ComboBoxIpNoMatch.SetSensitive(false)
	}
	w.FillProxiesTreeView()
	w.FillApplicationsTreeView()

	w.SetTextViewText(w.TextViewOnProxyChangeScript, w.Indicator.Config.ProxyChangeScript)
	w.SetTextViewText(w.TextViewOnProxyDeactivateScript, w.Indicator.Config.ProxyDeactivateScript)
	w.SetTextViewText(w.TextViewOnProxyActivateScript, w.Indicator.Config.ProxyActivateScript)

}

func (w *ConfigWindow) FillProxiesTreeView() {
	var err error
	w.ListStoreProxies.Clear()
	for _, p := range w.Indicator.Config.Proxies {
		iter := w.ListStoreProxies.Append()
		err = w.ListStoreProxies.SetValue(iter, 0, p.UUID)
		if err != nil {
			Log.Errorf("Can't set value in liststoreproxies: %v", err)
			goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		} else {
			url, err := p.ToUrl(false)
			if err != nil {
				Log.Errorf("Error generating URL for proxy %v: %v.", p.Name, err)
				err = w.ListStoreProxies.SetValue(iter, 1, p.Name+" ("+p.ToSimpleUrl()+")")
				if err != nil {
					Log.Errorf("Can't set value in liststoreproxies: %v", err)
					goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
				}
			} else {
				err = w.ListStoreProxies.SetValue(iter, 1, p.Name+" ("+url+")")
				if err != nil {
					Log.Errorf("Can't set value in liststoreproxies: %v", err)
					goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
				}
			}
		}
	}
}

func (w *ConfigWindow) FillApplicationsTreeView() {
	var err error
	w.ListStoreApps.Clear()
	for _, a := range ProxifiedApplications {
		iter := w.ListStoreApps.Append()
		err = w.ListStoreApps.SetValue(iter, 0, a.GetId())
		if err != nil {
			Log.Errorf("Can't set value in liststoreproxies: %v", err)
			goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		}
		err = w.ListStoreApps.SetValue(iter, 1, a.GetSimpleName())
		if err != nil {
			Log.Errorf("Can't set value in liststoreproxies: %v", err)
			goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		}
		err = w.ListStoreApps.SetValue(iter, 2, w.Indicator.Config.IsApplicationEnabled(a.GetId()))
		if err != nil {
			Log.Errorf("Can't set value in liststoreproxies: %v", err)
			goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		}
		err = w.ListStoreApps.SetValue(iter, 3, a.GetDescription())
		if err != nil {
			Log.Errorf("Can't set value in liststoreproxies: %v", err)
			goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		}
	}
}

func (w *ConfigWindow) OnSwitchRunStartupChanged() {
	err := w.Indicator.Config.SetIndicatorAutostart(w.SwitchRunStartup.GetActive())
	if err != nil {
		Log.Errorf("Error updating auto start: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
	} else {
		Log.Debugf("Run startup is now %v", w.SwitchRunStartup.GetActive())
	}
}

func (w *ConfigWindow) OnSwitchShowProxyNameChanged() {
	w.Indicator.Config.SetShowCurrentProxyNameNextToIndicator(w.SwitchShowProxyName.GetActive())
	err := w.Indicator.Config.Save(fmt.Sprintf("Show proxy name is now %v", w.Indicator.Config.ShowCurrentProxyNameNextToIndicator))
	if err != nil {
		Log.Errorf("Error saving configuration: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Error saving configuration: %v.", err))
	}
}

func (w *ConfigWindow) OnComboBoxLogLevelChanged() {
	logLevelStr := w.ComboBoxLogLevel.GetActiveID()
	level, ok := loggo.ParseLevel(logLevelStr)
	if !ok {
		Log.Errorf("Invalid log level %v", logLevelStr)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
	} else {
		w.Indicator.Config.SetLogLevel(level)
		err := w.Indicator.Config.Save(fmt.Sprintf("Configuration log level is now %v; runtime level is %v", w.Indicator.Config.LogLevel.String(), Log.LogLevel().String()))
		if err != nil {
			Log.Errorf("Error saving configuration: %v", err)
			goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Error saving configuration: %v.", err))
		}
	}
}

func (w *ConfigWindow) OnSwitchEnableAutoChangeByIp() {
	w.Indicator.Config.SetEnableAutoChangeByIp(w.SwitchEnableAutoSwitch.GetActive())
	if w.Indicator.Config.EnableAutoChangeByIp {
		w.ComboBoxIpNoMatch.SetSensitive(true)
	} else {
		w.ComboBoxIpNoMatch.SetSensitive(false)
	}
	err := w.Indicator.Config.Save(fmt.Sprintf("Enable auto change by ip is now %v", w.Indicator.Config.EnableAutoChangeByIp))
	if err != nil {
		Log.Errorf("Error saving configuration: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Error saving configuration: %v.", err))
	}
}

func (w *ConfigWindow) OnComboBoxWhatToDoWhenNoIpMatchesChanged() {
	w.Indicator.Config.SetWhatToDoWhenNoIpMatches(w.ComboBoxIpNoMatch.GetActiveID())
	err := w.Indicator.Config.Save(fmt.Sprintf("What to do when no ip matches is now %v", w.Indicator.Config.WhatToDoWhenNoIpMatches))
	if err != nil {
		Log.Errorf("Error saving configuration: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Error saving configuration: %v.", err))
	}
}

func (w *ConfigWindow) OnSwitchUpdateCheckChanged() {
	w.Indicator.Config.SetEnableUpdateCheck(w.SwitchUpdateCheck.GetActive())
	err := w.Indicator.Config.Save(fmt.Sprintf("Update check is now %v", w.Indicator.Config.EnableUpdateCheck))
	if err != nil {
		Log.Errorf("Error saving configuration: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Error saving configuration: %v.", err))
	}
}

func (w *ConfigWindow) OnTreeviewProxiesRowActivated(treeView *gtk.TreeView, path *gtk.TreePath) {

	selection, err := w.TreeViewProxies.GetSelection()
	if err != nil {
		Log.Errorf("Can't get selection: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	_, iter, ok := selection.GetSelected()
	if !ok {
		Log.Errorf("Selected not ok")
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	gval, err := w.ListStoreProxies.GetValue(iter, 0)
	if err != nil {
		Log.Errorf("Can't get value for iter: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	val, err := gval.GoValue()
	if err != nil {
		Log.Errorf("Can't get value for gvalue: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	proxyUuid, ok := val.(string)
	if !ok {
		Log.Errorf("Can't convert value to string for iter: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	p := w.Indicator.Config.GetProxyWithUuid(proxyUuid)
	if p == nil {
		Log.Errorf("Proxy with UUID %v not found", proxyUuid)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	d, err := NewProxyDialog(w, p)
	if err != nil {
		Log.Errorf("Error creating dialog: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	response := d.Dialog.Run()
	d.Dialog.Destroy()
	if response == int(gtk.RESPONSE_OK) {
		w.FillProxiesTreeView()
	}

}

func (w *ConfigWindow) OnTreeviewProxiesSelectionChanged(selection *gtk.TreeSelection) {

	_, _, ok := selection.GetSelected()
	if !ok {
		w.ButtonProxyEdit.SetSensitive(false)
		w.ButtonProxyRemove.SetSensitive(false)
	} else {
		w.ButtonProxyEdit.SetSensitive(true)
		w.ButtonProxyRemove.SetSensitive(true)
	}

}

func (w *ConfigWindow) OnButtonProxyAddClicked() {

	d, err := NewProxyDialog(w, nil)
	if err != nil {
		Log.Errorf("Error creating dialog: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	response := d.Dialog.Run()
	d.Dialog.Destroy()
	if response == int(gtk.RESPONSE_OK) {
		w.FillProxiesTreeView()
	}

}

func (w *ConfigWindow) OnButtonProxyEditClicked() {

	selection, err := w.TreeViewProxies.GetSelection()
	if err != nil {
		Log.Errorf("Can't get selection: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	_, iter, ok := selection.GetSelected()
	if !ok {
		Log.Errorf("Selected not ok")
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	gval, err := w.ListStoreProxies.GetValue(iter, 0)
	if err != nil {
		Log.Errorf("Can't get value for iter: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	val, err := gval.GoValue()
	if err != nil {
		Log.Errorf("Can't get value for gvalue: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	proxyUuid, ok := val.(string)
	if !ok {
		Log.Errorf("Can't convert value to string for iter: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	p := w.Indicator.Config.GetProxyWithUuid(proxyUuid)
	if p == nil {
		Log.Errorf("Proxy with UUID %v not found", proxyUuid)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	d, err := NewProxyDialog(w, p)
	if err != nil {
		Log.Errorf("Error creating dialog: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	response := d.Dialog.Run()
	d.Dialog.Destroy()
	if response == int(gtk.RESPONSE_OK) {
		w.FillProxiesTreeView()
	}

}

func (w *ConfigWindow) OnButtonProxyRemoveClicked() {

	selection, err := w.TreeViewProxies.GetSelection()
	if err != nil {
		Log.Errorf("Can't get selection: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	_, iter, ok := selection.GetSelected()
	if !ok {
		Log.Errorf("Selected not ok")
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	gval, err := w.ListStoreProxies.GetValue(iter, 0)
	if err != nil {
		Log.Errorf("Can't get value for iter: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	val, err := gval.GoValue()
	if err != nil {
		Log.Errorf("Can't get value for gvalue: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	proxyUuid, ok := val.(string)
	if !ok {
		Log.Errorf("Can't convert value to string for iter: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	p := w.Indicator.Config.GetProxyWithUuid(proxyUuid)
	if p == nil {
		Log.Errorf("Proxy with UUID %v not found", proxyUuid)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	if w.Indicator.Config.ActiveProxy == p {
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, "Active proxy", "The selected proxy is currently active and can't be removed.")
		return
	}
	if goutils.ConfirmMessage(w.Window, "Remove proxy", "Are you sure you want to remove this proxy?") {
		w.Indicator.Config.DeleteProxy(p, true)
		w.FillProxiesTreeView()
	}

}

func (w *ConfigWindow) OnApplicationEnabledToggled(cellRendererToggle *gtk.CellRendererToggle, path string) {

	iter, err := w.ListStoreApps.GetIterFromString(path)
	if err != nil {
		Log.Errorf("Can't get iter for path: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	// Now extract the application id in column 0
	gval, err := w.ListStoreApps.GetValue(iter, 0)
	if err != nil {
		Log.Errorf("Can't get value for iter: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	val, err := gval.GoValue()
	if err != nil {
		Log.Errorf("Can't get value for gvalue: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	appId, ok := val.(string)
	if !ok {
		Log.Errorf("Can't convert value to string for iter: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	// Switch the value in the model, in column 2
	newValue := !cellRendererToggle.GetActive()
	w.ListStoreApps.SetValue(iter, 2, newValue)

	// And set the value in the configuration
	if newValue {
		w.Indicator.Config.EnableApplication(appId)
		err = w.Indicator.Config.Save(fmt.Sprintf("App %v is now enabled.", appId))
	} else {
		w.Indicator.Config.DisableApplication(appId)
		err = w.Indicator.Config.Save(fmt.Sprintf("App %v is now disabled.", appId))
	}
	if err != nil {
		Log.Errorf("Error saving configuration: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Error saving configuration: %v.", err))
	}

}

func (w *ConfigWindow) OnWindowDeleted() bool {
	w.Window.Hide()
	return true
}

func (w *ConfigWindow) OnExportButtonClicked() {

	// Confirm export with passwords
	includePasswords := goutils.ConfirmMessage(
		w.Window,
		MyGettextv("Export configuration"),
		MyGettextv("Do you want to include passwords in the exported file?"),
	)

	chooser, err := gtk.FileChooserDialogNewWith2Buttons("Select file", w.Window, gtk.FILE_CHOOSER_ACTION_SAVE, "OK", gtk.RESPONSE_OK, "Cancel", gtk.RESPONSE_CANCEL)
	if err != nil {
		Log.Errorf("Error creating file chooser dialog: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}
	chooser.SetDoOverwriteConfirmation(true)

	filter, err := gtk.FileFilterNew()
	if err != nil {
		Log.Errorf("Error creating file filter: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}
	filter.SetName("JSON files (*.json)")
	filter.AddPattern("*.json")
	chooser.AddFilter(filter)
	chooser.SetDoOverwriteConfirmation(true)

	response := chooser.Run()
	defer chooser.Destroy()
	if response == int(gtk.RESPONSE_OK) {
		filename := chooser.GetFilename()
		if filename != "" {
			if !strings.HasSuffix(strings.ToLower(filename), ".json") {
				filename += ".json"
			}
			err := w.Indicator.Config.Export(filename, includePasswords)
			if err != nil {
				Log.Errorf("Error exporting configuration: %v", err)
				goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
			}
		} else {
			Log.Errorf("Empty filename")
			goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		}
	}

}

func (w *ConfigWindow) OnImportButtonClicked() {

	chooser, err := gtk.FileChooserDialogNewWith2Buttons("Select file", w.Window, gtk.FILE_CHOOSER_ACTION_OPEN, "OK", gtk.RESPONSE_OK, "Cancel", gtk.RESPONSE_CANCEL)
	if err != nil {
		Log.Errorf("Error creating file chooser dialog: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	filter, err := gtk.FileFilterNew()
	if err != nil {
		Log.Errorf("Error creating file filter: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}
	filter.SetName("JSON files (*.json)")
	filter.AddPattern("*.json")
	chooser.AddFilter(filter)

	response := chooser.Run()
	defer chooser.Destroy()
	if response == int(gtk.RESPONSE_OK) {
		filename := chooser.GetFilename()
		if filename != "" {

			warnings, err := w.Indicator.Config.Load(filename, true, true, true)
			if err != nil {
				Log.Errorf("Error importing configuration: %v", err)
				goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, "Error", MyGettextv("Error importing configuration: %v.", err))
			}

			err = w.Indicator.Config.Save(MyGettextv("Configuration imported"))
			if err != nil {
				Log.Errorf("Error saving configuration: %v", err)
				goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Error saving configuration: %v.", err))
			}

			w.FillData()
			if len(warnings) > 0 {
				warningText := bytes.NewBufferString("")
				for _, w := range warnings {
					Log.Warningf("Warning during import: %v", w)
					warningText.WriteString(fmt.Sprintf("* %v\n", w))
				}
				goutils.ShowMessage(w.Window, gtk.MESSAGE_WARNING, MyGettextv("Warnings during import"), warningText.String())

			}

		} else {
			Log.Errorf("Empty filename")
			goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		}
	}

}

func (w *ConfigWindow) OnCloseButtonClicked() {
	w.Window.Hide()
}

func (w *ConfigWindow) OnTextbufferProxyChangeScriptChanged() {
	text, err := w.GetTextViewText(w.TextViewOnProxyChangeScript)
	if err != nil {
		Log.Errorf("Error getting text: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
	} else {
		w.Indicator.Config.ProxyChangeScript = text
		w.Indicator.Config.Save("ProxyChangeScript changed")
	}
}

func (w *ConfigWindow) OnTextbufferProxyDeactivateScriptChanged() {
	text, err := w.GetTextViewText(w.TextViewOnProxyDeactivateScript)
	if err != nil {
		Log.Errorf("Error getting text: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
	} else {
		w.Indicator.Config.ProxyDeactivateScript = text
		w.Indicator.Config.Save("ProxyDeactivateScript changed")
	}
}

func (w *ConfigWindow) OnTextbufferProxyActivateScriptChanged() {
	text, err := w.GetTextViewText(w.TextViewOnProxyActivateScript)
	if err != nil {
		Log.Errorf("Error getting text: %v", err)
		goutils.ShowMessage(w.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
	} else {
		w.Indicator.Config.ProxyActivateScript = text
		w.Indicator.Config.Save("ProxyActivateScript changed")
	}
}
