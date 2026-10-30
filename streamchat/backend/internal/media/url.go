package media

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"
)

type Resolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}
type Provider string

const (
	YouTube   Provider = "youtube"
	TikTok    Provider = "tiktok"
	Facebook  Provider = "facebook"
	Instagram Provider = "instagram"
)

var allowedHosts = map[string]Provider{"youtube.com": YouTube, "www.youtube.com": YouTube, "m.youtube.com": YouTube, "youtu.be": YouTube, "youtube-nocookie.com": YouTube, "www.youtube-nocookie.com": YouTube, "tiktok.com": TikTok, "www.tiktok.com": TikTok, "vm.tiktok.com": TikTok, "vt.tiktok.com": TikTok, "facebook.com": Facebook, "www.facebook.com": Facebook, "m.facebook.com": Facebook, "fb.watch": Facebook, "instagram.com": Instagram, "www.instagram.com": Instagram}

type URLPolicy struct{ Resolver Resolver }

func (p URLPolicy) Validate(ctx context.Context, raw string) (Provider, string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u == nil || u.Scheme != "https" || u.User != nil || u.Port() != "" {
		return "", "", errors.New("URL must be a public HTTPS video link")
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	provider, ok := allowedHosts[host]
	if !ok {
		return "", "", errors.New("video host is not supported")
	}
	if net.ParseIP(host) != nil {
		return "", "", errors.New("IP literals are not allowed")
	}
	if p.Resolver == nil {
		return "", "", errors.New("DNS validation is unavailable")
	}
	ips, err := p.Resolver.LookupIPAddr(ctx, host)
	if err != nil || len(ips) == 0 {
		return "", "", errors.New("video host could not be resolved")
	}
	for _, addr := range ips {
		if !isPublic(addr.IP) {
			return "", "", errors.New("video host resolved to a non-public address")
		}
	}
	if strings.Contains(strings.ToLower(u.Path), "playlist") || u.Query().Has("list") {
		return "", "", errors.New("playlists are not supported")
	}
	u.Fragment = ""
	u.RawFragment = ""
	return provider, u.String(), nil
}
func isPublic(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
		return false
	}
	if v4 := ip.To4(); v4 != nil {
		if v4[0] == 100 && (v4[1]&0xc0) == 64 {
			return false
		}
		if v4[0] == 192 && v4[1] == 0 && v4[2] == 0 {
			return false
		}
		if v4[0] == 198 && (v4[1] == 18 || v4[1] == 19) {
			return false
		}
		if v4[0] >= 240 {
			return false
		}
	}
	return true
}
