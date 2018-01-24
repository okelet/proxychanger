package proxychangerlib

import (
	"net"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/okelet/goutils"
	"github.com/pkg/errors"
)

type Proxy struct {
	*goutils.Proxy
	Slug                string
	Name                string
	MatchingIps         []string
	RadioMenuItem       *gtk.RadioMenuItem
	RadioMenuItemHandle glib.SignalHandle
	ActivateScript      string
}

func NewEmptyProxy(passwordManager goutils.ProxyPasswordManager) *Proxy {
	p := Proxy{
		Proxy:       goutils.NewEmptyProxy(passwordManager),
		MatchingIps: []string{},
	}
	return &p
}

func NewProxyFromMap(c *Configuration, h *goutils.MapHelper, loadPasswordFromMap bool) (*Proxy, error) {
	p := Proxy{
		Proxy:          goutils.NewProxyFromMap(h, c, loadPasswordFromMap),
		Slug:           h.GetString("slug", ""),
		Name:           h.GetString("name", ""),
		MatchingIps:    h.GetListOfStrings("matching_ips", []string{}),
		ActivateScript: h.GetString("activate_script", ""),
	}
	if p.Name == "" || c.IsNameAlreadyInUse(p.Name, &p) {
		p.Name = c.CreateUniqueName(p.Name, &p)
	}
	if p.Slug == "" || c.IsSlugAlreadyInUse(p.Slug, &p) {
		p.Slug = c.CreateUniqueSlug(p.Name, nil)
	}
	return &p, nil
}

func NewImportedProxy(c *Configuration, importedProxy *goutils.Proxy, name, slug string) *Proxy {
	p := Proxy{
		Proxy: importedProxy,
		Slug:  slug,
		Name:  name,
	}
	if p.Slug == "" && p.Name != "" {
		p.Slug = c.CreateUniqueSlug(p.Name, nil)
	}
	return &p
}

func (p *Proxy) ToMap(c *Configuration, includePassword bool) (*goutils.MapHelper, error) {
	h, err := p.Proxy.ToMap(false)
	if err != nil {
		return nil, errors.Wrap(err, "Error generating map")
	}
	h.SetString("uuid", p.UUID)
	h.SetString("slug", p.Slug)
	h.SetString("name", p.Name)
	if len(p.MatchingIps) > 0 {
		h.SetListOfStrings("matching_ips", p.MatchingIps)
	}
	if includePassword {
		password, err := c.GetPassword(p.UUID)
		if err != nil {
			return nil, err
		}
		h.SetString("password", password)
	}
	if p.ActivateScript != "" {
		h.SetString("activate_script", p.ActivateScript)
	}
	return h, nil
}

func (p *Proxy) MatchesIps(ips []string) bool {
	parsedIps := []net.IP{}
	for _, i := range ips {
		ip := net.ParseIP(i)
		if ip != nil {
			parsedIps = append(parsedIps, ip)
		} else {
			Log.Errorf("IP %v is not valid.", i)
		}
	}
	found := false
	for _, r := range p.MatchingIps {
		_, subnet, err := net.ParseCIDR(r)
		if err != nil {
			Log.Errorf("Error parsing CIDR %v: %v.", r, err)
		} else {
			for _, i := range parsedIps {
				if subnet.Contains(i) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}
	return found
}
