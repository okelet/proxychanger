package proxychangerlib

import (
	"strings"

	"github.com/gotk3/gotk3/gtk"
	"github.com/okelet/goutils"
	"github.com/pkg/errors"
)

type ProxyDialog struct {
	*goutils.BuilderBase
	ConfigWindow  *ConfigWindow
	Proxy         *Proxy
	AskToSetAfter bool

	Dialog                        *gtk.Dialog
	EntrySlug                     *gtk.Entry
	EntryName                     *gtk.Entry
	EntryAddress                  *gtk.Entry
	SpinButtonPort                *gtk.SpinButton
	EntryUsername                 *gtk.Entry
	EntryPassword                 *gtk.Entry
	TextViewExceptions            *gtk.TextView
	EntryMatchingIps              *gtk.Entry
	TextViewProxyActivateScript   *gtk.TextView
	TextBufferProxyActivateScript *gtk.TextBuffer
}

func NewProxyDialog(configWindow *ConfigWindow, p *Proxy, askToSetAfter bool) (*ProxyDialog, error) {

	var err error

	w := ProxyDialog{}
	w.ConfigWindow = configWindow
	w.Proxy = p
	w.AskToSetAfter = askToSetAfter

	w.BuilderBase, err = goutils.NewBuilderBase(w.ConfigWindow.Indicator, "assets/proxy.glade")
	if err != nil {
		return nil, errors.Wrap(err, MyGettextv("Error loading asset %v", "assets/proxy.glade"))
	}

	// ------------------------------------------------------------------------------------

	w.Dialog, err = w.GetDialog("dialog_proxy")
	if err != nil {
		return nil, errors.Wrap(err, MyGettextv("Error getting widget %v", "preferences_window"))
	}
	w.Dialog.SetTransientFor(configWindow.Window)
	w.Dialog.Resize(400, 500)
	w.Dialog.SetPosition(gtk.WIN_POS_CENTER)
	w.Dialog.SetIconName(ICON_NAME)

	if w.Proxy != nil {
		w.Dialog.SetTitle(MyGettextv("Edit proxy %v", w.Proxy.Name))
	} else {
		w.Dialog.SetTitle(MyGettextv("Add proxy"))
	}

	// ------------------------------------------------------------------------------------

	w.EntrySlug, err = w.GetEntry("entry_slug")
	if err != nil {
		return nil, errors.Wrap(err, MyGettextv("Error getting widget %v", "entry_slug"))
	}
	if w.Proxy != nil {
		w.EntrySlug.SetText(w.Proxy.Slug)
	}

	w.EntryName, err = w.GetEntry("entry_name")
	if err != nil {
		return nil, errors.Wrap(err, MyGettextv("Error getting widget %v", "widget entry_name"))
	}
	if w.Proxy != nil {
		w.EntryName.SetText(w.Proxy.Name)
	}

	w.EntryAddress, err = w.GetEntry("entry_address")
	if err != nil {
		return nil, errors.Wrap(err, MyGettextv("Error getting widget %v", "entry_address"))
	}
	if w.Proxy != nil {
		w.EntryAddress.SetText(w.Proxy.Address)
	}

	w.SpinButtonPort, err = w.GetSpinButton("spinbutton_port")
	if err != nil {
		return nil, errors.Wrap(err, MyGettextv("Error getting widget %v", "spinbutton_port"))
	}
	if w.Proxy != nil {
		w.SpinButtonPort.SetValue(float64(w.Proxy.Port))
	}

	w.EntryUsername, err = w.GetEntry("entry_username")
	if err != nil {
		return nil, errors.Wrap(err, MyGettextv("Error getting widget %v", "entry_username"))
	}
	if w.Proxy != nil {
		w.EntryUsername.SetText(w.Proxy.Username)
	}

	w.EntryPassword, err = w.GetEntry("entry_password")
	if err != nil {
		return nil, errors.Wrap(err, MyGettextv("Error getting widget %v", "entry_password"))
	}
	if w.Proxy != nil {
		password, err := w.Proxy.GetPassword()
		if err != nil {
			return nil, err
		}
		w.EntryPassword.SetText(password)
	}

	w.TextViewExceptions, err = w.GetTextView("textview_exceptions")
	if err != nil {
		return nil, errors.Wrap(err, MyGettextv("Error getting widget %v", "textview_exceptions"))
	}
	if w.Proxy != nil {
		w.SetTextViewText(w.TextViewExceptions, strings.Join(w.Proxy.Exceptions, "\n"))
	}

	w.EntryMatchingIps, err = w.GetEntry("entry_matching_ips")
	if err != nil {
		return nil, errors.Wrap(err, MyGettextv("Error getting widget %v", "entry_matching_ips"))
	}
	if w.Proxy != nil {
		w.EntryMatchingIps.SetText(strings.Join(w.Proxy.MatchingIps, ", "))
	}

	// ------------------------------------------------------------------------------------

	w.TextViewProxyActivateScript, err = w.GetTextView("textview_proxy_activate_script")
	if err != nil {
		return nil, errors.Wrap(err, MyGettextv("Error getting widget %v", "textview_proxy_activate"))
	}

	w.TextBufferProxyActivateScript, err = w.GetTextBuffer("textbuffer_proxy_activate_script")
	if err != nil {
		return nil, errors.Wrap(err, MyGettextv("Error getting widget %v", "textbuffer_proxy_activate_script"))
	}

	if w.Proxy != nil {
		w.SetTextViewText(w.TextViewProxyActivateScript, w.Proxy.ActivateScript)
	}

	// ------------------------------------------------------------------------------------
	// Signals
	// ------------------------------------------------------------------------------------

	w.Builder.ConnectSignals(map[string]interface{}{
		"on_togglebutton_show_password_toggled": w.OnToggleButtonShowPasswordChanged,
		"on_button_ok_clicked":                  w.OnButtonOkClicked,
		"on_button_cancel_clicked":              w.OnButtonCancelClicked,
	})

	w.Dialog.SetFocus(&w.EntrySlug.Widget)

	return &w, nil

}

func (w *ProxyDialog) OnToggleButtonShowPasswordChanged(button *gtk.ToggleButton) {
	w.EntryPassword.SetVisibility(button.GetActive())
}

func (w *ProxyDialog) OnButtonOkClicked() {

	slug, err := w.EntrySlug.GetText()
	if err != nil {
		Log.Errorf("Error getting data: %v", err)
		goutils.ShowMessage(&w.Dialog.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	name, err := w.EntryName.GetText()
	if err != nil {
		Log.Errorf("Error getting data: %v", err)
		goutils.ShowMessage(&w.Dialog.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	address, err := w.EntryAddress.GetText()
	if err != nil {
		Log.Errorf("Error getting data: %v", err)
		goutils.ShowMessage(&w.Dialog.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	port := w.SpinButtonPort.GetValueAsInt()

	username, err := w.EntryUsername.GetText()
	if err != nil {
		Log.Errorf("Error getting data: %v", err)
		goutils.ShowMessage(&w.Dialog.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	password, err := w.EntryPassword.GetText()
	if err != nil {
		Log.Errorf("Error getting data: %v", err)
		goutils.ShowMessage(&w.Dialog.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	exceptions, err := w.GetTextViewText(w.TextViewExceptions)
	if err != nil {
		Log.Errorf("Error getting data: %v", err)
		goutils.ShowMessage(&w.Dialog.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	cleanExceptions := []string{}
	for _, v := range strings.Split(exceptions, "\n") {
		cleanValue := strings.TrimSpace(v)
		if cleanValue != "" {
			cleanExceptions = append(cleanExceptions, cleanValue)
		}
	}

	matchingIps, err := w.EntryMatchingIps.GetText()
	if err != nil {
		Log.Errorf("Error getting data: %v", err)
		goutils.ShowMessage(&w.Dialog.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	cleanMatchingIps := []string{}
	for _, v := range strings.Split(matchingIps, ",") {
		cleanValue := strings.TrimSpace(v)
		if cleanValue != "" {
			cleanMatchingIps = append(cleanMatchingIps, cleanValue)
		}
	}

	activateScript, err := w.GetTextViewText(w.TextViewProxyActivateScript)
	if err != nil {
		Log.Errorf("Error getting data: %v", err)
		goutils.ShowMessage(&w.Dialog.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), MyGettextv("Please review the LOG."))
		return
	}

	p := w.Proxy
	if p == nil {
		p = NewEmptyProxy(w.ConfigWindow.Indicator.Config)
	}
	err, field := w.ConfigWindow.Indicator.Config.UpdateProxyFromData(
		true,
		p,
		true, slug,
		true, name,
		true, "http",
		true, address,
		true, port,
		true, username,
		true, password,
		true, cleanExceptions,
		true, cleanMatchingIps,
		true, activateScript,
	)
	if err != nil {
		goutils.ShowMessage(&w.Dialog.Window, gtk.MESSAGE_ERROR, MyGettextv("Error"), err.Error())
		if field == "slug" {
			w.Dialog.SetFocus(&w.EntrySlug.Widget)
		} else if field == "name" {
			w.Dialog.SetFocus(&w.EntryName.Widget)
		} else if field == "address" {
			w.Dialog.SetFocus(&w.EntryAddress.Widget)
		} else if field == "port" {
			w.Dialog.SetFocus(&w.SpinButtonPort.Widget)
		} else if field == "username" {
			w.Dialog.SetFocus(&w.EntryUsername.Widget)
		} else if field == "password" {
			w.Dialog.SetFocus(&w.EntryPassword.Widget)
		} else if field == "exceptions" {
			w.Dialog.SetFocus(&w.TextViewExceptions.Widget)
		} else if field == "matchingips" {
			w.Dialog.SetFocus(&w.EntryMatchingIps.Widget)
		}
		return
	}
	if w.AskToSetAfter {
		if goutils.ConfirmMessage(
			nil,
			MyGettextv("Set proxy"),
			MyGettextv("Do you want to activate this proxy?"),
		) {
			w.ConfigWindow.Indicator.Config.SetActiveProxy(p, MyGettextv("Proxy %v added", p.Name), true)
		}
	}
	w.Dialog.Response(gtk.RESPONSE_OK)
}

func (w *ProxyDialog) OnButtonCancelClicked() {
	w.Dialog.Response(gtk.RESPONSE_CANCEL)
}
