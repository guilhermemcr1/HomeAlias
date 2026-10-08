package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/homealias/homealias/backend/internal/dns"
)

type Client struct {
	HTTP *http.Client
	Base string
}

func New() *Client {
	return &Client{
		HTTP: &http.Client{Timeout: 15 * time.Second},
		Base: "https://api.cloudflare.com/client/v4",
	}
}

type apiResp struct {
	Success bool            `json:"success"`
	Errors  []apiError      `json:"errors"`
	Result  json.RawMessage `json:"result"`
}

type apiError struct {
	Message string `json:"message"`
}

func (c *Client) do(ctx context.Context, method, path, token string, body any, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.Base+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var parsed apiResp
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return fmt.Errorf("cloudflare decode: %w", err)
	}
	if !parsed.Success {
		msg := "cloudflare error"
		if len(parsed.Errors) > 0 {
			msg = parsed.Errors[0].Message
		}
		return fmt.Errorf("%s", msg)
	}
	if out != nil && len(parsed.Result) > 0 {
		return json.Unmarshal(parsed.Result, out)
	}
	return nil
}

func (c *Client) ValidateCredentials(ctx context.Context, token string) ([]dns.Zone, error) {
	var zones []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := c.do(ctx, http.MethodGet, "/zones?per_page=50", token, nil, &zones); err != nil {
		return nil, err
	}
	out := make([]dns.Zone, 0, len(zones))
	for _, z := range zones {
		out = append(out, dns.Zone{ID: z.ID, Name: z.Name})
	}
	return out, nil
}

func (c *Client) GetRecord(ctx context.Context, in dns.GetInput) (*dns.Record, error) {
	path := fmt.Sprintf("/zones/%s/dns_records?type=%s&name=%s", in.ZoneID, in.Type, in.Name)
	var recs []struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Name    string `json:"name"`
		Content string `json:"content"`
		TTL     int    `json:"ttl"`
		Proxied bool   `json:"proxied"`
	}
	token, ok := tokenFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("missing token in context")
	}
	if err := c.do(ctx, http.MethodGet, path, token, nil, &recs); err != nil {
		return nil, err
	}
	if len(recs) == 0 {
		return nil, nil
	}
	r := recs[0]
	return &dns.Record{ID: r.ID, Type: r.Type, Name: r.Name, Content: r.Content, TTL: r.TTL, Proxied: r.Proxied}, nil
}

func (c *Client) UpsertRecord(ctx context.Context, in dns.UpsertInput) error {
	token, ok := tokenFrom(ctx)
	if !ok {
		return fmt.Errorf("missing token in context")
	}
	existing, err := c.GetRecord(ctx, dns.GetInput{ZoneID: in.ZoneID, Type: in.Type, Name: in.Name})
	if err != nil {
		return err
	}
	payload := map[string]any{
		"type":    in.Type,
		"name":    in.Name,
		"content": in.Content,
		"ttl":     in.TTL,
		"proxied": in.Proxied,
	}
	if existing == nil {
		path := fmt.Sprintf("/zones/%s/dns_records", in.ZoneID)
		return c.do(ctx, http.MethodPost, path, token, payload, nil)
	}
	path := fmt.Sprintf("/zones/%s/dns_records/%s", in.ZoneID, existing.ID)
	return c.do(ctx, http.MethodPut, path, token, payload, nil)
}

func (c *Client) DeleteRecord(ctx context.Context, in dns.DeleteInput) error {
	token, ok := tokenFrom(ctx)
	if !ok {
		return fmt.Errorf("missing token in context")
	}
	existing, err := c.GetRecord(ctx, dns.GetInput{ZoneID: in.ZoneID, Type: in.Type, Name: in.Name})
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}
	path := fmt.Sprintf("/zones/%s/dns_records/%s", in.ZoneID, existing.ID)
	return c.do(ctx, http.MethodDelete, path, token, nil, nil)
}

type tokenKey struct{}

func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey{}, token)
}

func tokenFrom(ctx context.Context) (string, bool) {
	t, ok := ctx.Value(tokenKey{}).(string)
	return t, ok && strings.TrimSpace(t) != ""
}

// Ensure Client implements dns.Provider
var _ dns.Provider = (*Client)(nil)
