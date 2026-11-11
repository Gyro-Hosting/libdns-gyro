// Package gyro implements a libdns provider for Gyro DNS, limited to what an
// ACME key may do: append and delete _acme-challenge TXT records (DNS-01).
// It is meant for ACME clients such as Caddy/CertMagic.
package gyro

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// Provider is a libdns provider for Gyro DNS ACME challenges.
type Provider struct {
	// APIToken is the Gyro DNS ACME key (Gyro DNS service page > ACME / Let's Encrypt).
	APIToken string `json:"api_token,omitempty"`
	// APIBase overrides the API URL (default https://www.gyrohosting.com/api/v1/client/).
	APIBase string `json:"api_base,omitempty"`
	// SkipWait disables waiting until both Gyro DNS name servers publish the record.
	SkipWait bool `json:"skip_wait,omitempty"`
}

func (p *Provider) client() *Client { return &Client{Token: p.APIToken, APIBase: p.APIBase} }

func fqdnOf(rr libdns.RR, zone string) (string, error) {
	if rr.Type != "TXT" {
		return "", fmt.Errorf("gyro dns: only TXT records are supported (got %s)", rr.Type)
	}
	n := strings.TrimSuffix(libdns.AbsoluteName(rr.Name, zone), ".")
	if !strings.HasPrefix(n, "_acme-challenge.") {
		return "", fmt.Errorf("gyro dns: ACME keys can only write _acme-challenge records (got %s)", n)
	}
	return n, nil
}

// AppendRecords adds the TXT challenge records.
func (p *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	c := p.client()
	var done []libdns.Record
	for _, r := range recs {
		rr := r.RR()
		n, err := fqdnOf(rr, zone)
		if err != nil {
			return done, err
		}
		if err := c.Present(ctx, n, rr.Data); err != nil {
			return done, err
		}
		if !p.SkipWait {
			if err := c.WaitReady(ctx, n, rr.Data, 2*time.Minute, 5*time.Second); err != nil {
				return done, err
			}
		}
		done = append(done, libdns.TXT{Name: rr.Name, TTL: 60 * time.Second, Text: rr.Data})
	}
	return done, nil
}

// DeleteRecords removes the TXT challenge records.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	c := p.client()
	var done []libdns.Record
	for _, r := range recs {
		rr := r.RR()
		n, err := fqdnOf(rr, zone)
		if err != nil {
			return done, err
		}
		if err := c.CleanUp(ctx, n, rr.Data); err != nil {
			return done, err
		}
		done = append(done, r)
	}
	return done, nil
}

var (
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
