package media

import (
	"context"
	"net"
	"testing"
)

type resolver struct{ ips []net.IPAddr }

func (r resolver) LookupIPAddr(context.Context, string) ([]net.IPAddr, error) { return r.ips, nil }
func TestURLPolicy(t *testing.T) {
	public := resolver{[]net.IPAddr{{IP: net.ParseIP("93.184.216.34")}}}
	p := URLPolicy{Resolver: public}
	provider, normalized, err := p.Validate(context.Background(), "https://www.youtube.com/watch?v=abc#fragment")
	if err != nil || provider != YouTube || normalized != "https://www.youtube.com/watch?v=abc" {
		t.Fatalf("provider=%s url=%s err=%v", provider, normalized, err)
	}
	for _, raw := range []string{"http://youtube.com/watch?v=x", "https://evil.example/video", "https://youtube.com/watch?list=playlist", "https://youtube.com/watch?v=x"} {
		policy := p
		if raw == "https://youtube.com/watch?v=x" {
			policy.Resolver = resolver{[]net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}}
		}
		if _, _, err := policy.Validate(context.Background(), raw); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
}
