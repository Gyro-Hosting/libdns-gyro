package gyro

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultAPIBase is the public Gyro API.
const DefaultAPIBase = "https://www.gyrohosting.com/api/v1/client/"

// Version of this package (sent in the User-Agent).
const Version = "0.1.0"

// Client talks to the Gyro DNS ACME challenge endpoints with an ACME key
// (a key that can only add and remove _acme-challenge TXT records).
type Client struct {
	Token   string
	APIBase string
	HTTP    *http.Client
}

// Error is an API error answer.
type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("gyro dns: HTTP %d %s: %s", e.Status, e.Code, e.Message)
}

func (c *Client) base() string {
	b := c.APIBase
	if b == "" {
		b = DefaultAPIBase
	}
	if !strings.HasSuffix(b, "/") {
		b += "/"
	}
	return b
}

func (c *Client) do(ctx context.Context, method, fqdn, value string) (map[string]any, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("gyro dns: API token is empty")
	}
	name := strings.TrimSuffix(fqdn, ".")
	u := c.base() + "cloud/dns/acme-challenge"
	var body io.Reader
	if method == http.MethodPost {
		b, _ := json.Marshal(map[string]string{"name": name, "value": value})
		body = bytes.NewReader(b)
	} else {
		q := url.Values{"name": {name}}
		if value != "" {
			q.Set("value", value)
		}
		u += "?" + q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "libdns-gyro/"+Version)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	hc := c.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gyro dns: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var out struct {
		Data  map[string]any `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(raw, &out)
	if resp.StatusCode >= 300 {
		msg := out.Error.Message
		if msg == "" {
			msg = strings.TrimSpace(string(raw))
			if len(msg) > 200 {
				msg = msg[:200]
			}
		}
		return nil, &Error{Status: resp.StatusCode, Code: out.Error.Code, Message: msg}
	}
	return out.Data, nil
}

// Present adds value to the TXT record set at fqdn (other values stay).
func (c *Client) Present(ctx context.Context, fqdn, value string) error {
	_, err := c.do(ctx, http.MethodPost, fqdn, value)
	return err
}

// CleanUp removes value from the TXT record set at fqdn.
func (c *Client) CleanUp(ctx context.Context, fqdn, value string) error {
	_, err := c.do(ctx, http.MethodDelete, fqdn, value)
	return err
}

// Ready reports whether both authoritative name servers publish value at fqdn.
func (c *Client) Ready(ctx context.Context, fqdn, value string) (bool, error) {
	d, err := c.do(ctx, http.MethodGet, fqdn, value)
	if err != nil {
		return false, err
	}
	r, _ := d["ready"].(bool)
	return r, nil
}

// WaitReady polls Ready every interval until it is true, ctx ends or timeout passes.
// It returns nil on timeout as well (the CA retries); only API errors are returned.
func (c *Client) WaitReady(ctx context.Context, fqdn, value string, timeout, interval time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		ok, err := c.Ready(ctx, fqdn, value)
		if err != nil {
			return err
		}
		if ok || time.Now().After(deadline) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}
