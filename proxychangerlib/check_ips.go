package proxychangerlib

import (
	"fmt"
	"net"

	"github.com/pkg/errors"
	"github.com/robfig/cron"
)

type IpsChangedListener interface {
	OnIpsChanged(newIps []string)
}

type CheckIpsThread struct {
	Cron              *cron.Cron
	Listeners         []IpsChangedListener
	IntervalInSeconds int
	Config            *Configuration
}

func NewCheckIpsThread(intervalInSeconds int, config *Configuration) *CheckIpsThread {
	t := CheckIpsThread{
		Cron:              nil,
		Listeners:         []IpsChangedListener{},
		IntervalInSeconds: intervalInSeconds,
		Config:            config,
	}
	return &t
}

func (t *CheckIpsThread) AddListener(listener IpsChangedListener) {
	t.Listeners = append(t.Listeners, listener)
}

func (t *CheckIpsThread) Check() {
	list, err := net.Interfaces()
	if err != nil {
		Log.Errorf("Error getting interfaces: %v", err)
		return
	}
	ips := []string{}
	for _, iface := range list {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		interfaceExcluded := false
		for _, regex := range t.Config.ExcludedInterfacesRegexpsParsed {
			if regex.MatchString(iface.Name) {
				interfaceExcluded = true
				break
			}
		}
		if interfaceExcluded {
			Log.Tracef("Ignoring interface %v", iface.Name)
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			Log.Errorf("Error getting the addresses of the interface %v: %v", iface.Name, err)
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			ips = append(ips, ip.String())
		}
	}
	for _, l := range t.Listeners {
		l.OnIpsChanged(ips)
	}
}

func (t *CheckIpsThread) SetInterval(seconds int) error {
	t.IntervalInSeconds = seconds
	if t.Cron != nil {
		t.Stop()
		return t.Start()
	}
	return nil
}

func (t *CheckIpsThread) Start() error {
	var err error
	if t.Cron == nil {
		t.Cron = cron.New()
		err = t.Cron.AddFunc(fmt.Sprintf("@every %vs", t.IntervalInSeconds), t.Check)
		if err != nil {
			return errors.Wrap(err, MyGettextv("Error starting cron"))
		}
		t.Cron.Start()
	}
	return nil
}

func (t *CheckIpsThread) Stop() {
	if t.Cron != nil {
		t.Cron.Stop()
		t.Cron = nil
	}
}
