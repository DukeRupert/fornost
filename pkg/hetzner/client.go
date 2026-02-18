package hetzner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const baseURL = "https://api.hetzner.cloud/v1"

// Client wraps the Hetzner Cloud API v1.
type Client struct {
	Token      string
	httpClient *http.Client
}

// NewClient returns a Client authenticated with the given API token.
func NewClient(token string) *Client {
	return &Client{
		Token:      token,
		httpClient: &http.Client{},
	}
}

// --- HTTP helpers ---

func (c *Client) doRequest(method, endpoint string, body io.Reader, out any) error {
	req, err := http.NewRequest(method, baseURL+endpoint, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr errorResponse
		if json.Unmarshal(raw, &apiErr) == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("hetzner api error: %s (code: %s)", apiErr.Error.Message, apiErr.Error.Code)
		}
		return fmt.Errorf("hetzner api error: status %d", resp.StatusCode)
	}

	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) get(endpoint string, out any) error {
	return c.doRequest(http.MethodGet, endpoint, nil, out)
}

func (c *Client) post(endpoint string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}
	return c.doRequest(http.MethodPost, endpoint, bytes.NewReader(body), out)
}

func (c *Client) delete(endpoint string) error {
	return c.doRequest(http.MethodDelete, endpoint, nil, nil)
}

// getAll fetches all pages from a paginated endpoint.
// key is the JSON key holding the array (e.g. "servers", "ssh_keys").
func getAll[T any](c *Client, endpoint string, key string) ([]T, error) {
	var all []T
	page := 1

	for {
		url := fmt.Sprintf("%s?page=%d&per_page=50", endpoint, page)

		var rawBody json.RawMessage
		if err := c.doRequest(http.MethodGet, url, nil, &rawBody); err != nil {
			return nil, err
		}

		var envelope map[string]json.RawMessage
		if err := json.Unmarshal(rawBody, &envelope); err != nil {
			return nil, fmt.Errorf("decode envelope: %w", err)
		}

		itemsRaw, ok := envelope[key]
		if !ok {
			return nil, fmt.Errorf("response missing key %q", key)
		}

		var items []T
		if err := json.Unmarshal(itemsRaw, &items); err != nil {
			return nil, fmt.Errorf("decode %s: %w", key, err)
		}
		all = append(all, items...)

		// Check pagination
		metaRaw, ok := envelope["meta"]
		if !ok {
			break
		}
		var meta Meta
		if err := json.Unmarshal(metaRaw, &meta); err != nil {
			break
		}
		if meta.Pagination.NextPage == nil {
			break
		}
		page = *meta.Pagination.NextPage
	}

	return all, nil
}

// --- API models ---

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type Meta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	TotalPages int  `json:"last_page"`
	NextPage   *int `json:"next_page"`
}

type Server struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	PublicNet  PublicNet  `json:"public_net"`
	ServerType ServerType `json:"server_type"`
	Datacenter Datacenter `json:"datacenter"`
	Created    string     `json:"created"`
}

type PublicNet struct {
	IPv4 IPv4 `json:"ipv4"`
}

type IPv4 struct {
	IP string `json:"ip"`
}

type ServerType struct {
	Name string `json:"name"`
}

type Datacenter struct {
	Name     string   `json:"name"`
	Location Location `json:"location"`
}

type Location struct {
	Name string `json:"name"`
	City string `json:"city"`
}

type SSHKey struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key"`
}

type Firewall struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	Rules     []FirewallRule `json:"rules"`
	AppliedTo []AppliedTo    `json:"applied_to"`
}

type FirewallRule struct {
	Direction   string   `json:"direction"`
	Protocol    string   `json:"protocol"`
	Port        string   `json:"port"`
	SourceIPs   []string `json:"source_ips"`
	DestIPs     []string `json:"destination_ips"`
	Description string   `json:"description"`
}

type AppliedTo struct {
	Type   string         `json:"type"`
	Server *AppliedServer `json:"server,omitempty"`
}

type AppliedServer struct {
	ID int `json:"id"`
}

type Action struct {
	ID      int    `json:"id"`
	Status  string `json:"status"`
	Command string `json:"command"`
}

// --- API methods ---

// Ping validates the API token by hitting GET /actions with a low per_page.
func (c *Client) Ping() error {
	var resp struct {
		Actions []Action `json:"actions"`
	}
	return c.get("/actions?per_page=1", &resp)
}

// ListServers returns all servers in the project, handling pagination.
func (c *Client) ListServers() ([]Server, error) {
	return getAll[Server](c, "/servers", "servers")
}

// GetServer retrieves a single server by name or numeric ID.
func (c *Client) GetServer(nameOrID string) (*Server, error) {
	// Try numeric ID first
	if id, err := strconv.Atoi(nameOrID); err == nil {
		var resp struct {
			Server Server `json:"server"`
		}
		if err := c.get(fmt.Sprintf("/servers/%d", id), &resp); err != nil {
			return nil, err
		}
		return &resp.Server, nil
	}

	// Resolve by name: list all and filter
	servers, err := c.ListServers()
	if err != nil {
		return nil, err
	}
	for _, s := range servers {
		if s.Name == nameOrID {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("server not found: %s", nameOrID)
}

// ListSSHKeys returns all SSH keys in the project.
func (c *Client) ListSSHKeys() ([]SSHKey, error) {
	return getAll[SSHKey](c, "/ssh_keys", "ssh_keys")
}

// AddSSHKey uploads a new SSH public key to the project.
func (c *Client) AddSSHKey(name, publicKey string) (*SSHKey, error) {
	payload := struct {
		Name      string `json:"name"`
		PublicKey string `json:"public_key"`
	}{Name: name, PublicKey: publicKey}

	var resp struct {
		SSHKey SSHKey `json:"ssh_key"`
	}
	if err := c.post("/ssh_keys", payload, &resp); err != nil {
		return nil, err
	}
	return &resp.SSHKey, nil
}

// DeleteSSHKey removes an SSH key by name or numeric ID.
func (c *Client) DeleteSSHKey(nameOrID string) error {
	id, err := strconv.Atoi(nameOrID)
	if err != nil {
		// Resolve by name
		keys, err := c.ListSSHKeys()
		if err != nil {
			return err
		}
		found := false
		for _, k := range keys {
			if k.Name == nameOrID {
				id = k.ID
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("ssh key not found: %s", nameOrID)
		}
	}
	return c.delete(fmt.Sprintf("/ssh_keys/%d", id))
}

// ListFirewalls returns all firewalls in the project.
func (c *Client) ListFirewalls() ([]Firewall, error) {
	return getAll[Firewall](c, "/firewalls", "firewalls")
}

// GetFirewall retrieves a single firewall by name or numeric ID.
func (c *Client) GetFirewall(nameOrID string) (*Firewall, error) {
	if id, err := strconv.Atoi(nameOrID); err == nil {
		var resp struct {
			Firewall Firewall `json:"firewall"`
		}
		if err := c.get(fmt.Sprintf("/firewalls/%d", id), &resp); err != nil {
			return nil, err
		}
		return &resp.Firewall, nil
	}

	firewalls, err := c.ListFirewalls()
	if err != nil {
		return nil, err
	}
	for _, f := range firewalls {
		if f.Name == nameOrID {
			return &f, nil
		}
	}
	return nil, fmt.Errorf("firewall not found: %s", nameOrID)
}
