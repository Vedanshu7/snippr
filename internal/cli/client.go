package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	base  string
	token string
	http  *http.Client
}

func NewClient(cfg *Config) *Client {
	return &Client{
		base:  cfg.ServerURL,
		token: cfg.Token,
		http:  &http.Client{Timeout: 15 * time.Second},
	}
}

type Snippet struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Language  string    `json:"language"`
	IsPublic  bool      `json:"is_public"`
	ShareSlug string    `json:"share_slug"`
	Tags      []string  `json:"tags"`
	UpdatedAt time.Time `json:"updated_at"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResp struct {
	Token string `json:"token"`
}

type createReq struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Language string   `json:"language"`
	IsPublic bool     `json:"is_public"`
	Tags     []string `json:"tags"`
}

func (c *Client) do(method, path string, body any) (*http.Response, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.base+path, r)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return c.http.Do(req)
}

func decode[T any](resp *http.Response) (T, error) {
	defer resp.Body.Close()
	var v T
	if resp.StatusCode >= 400 {
		var errBody struct{ Error string }
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		msg := errBody.Error
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return v, fmt.Errorf("server: %s", msg)
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode response: %w", err)
	}
	return v, nil
}

func (c *Client) Login(email, password string) (string, error) {
	resp, err := c.do("POST", "/api/login", loginReq{Email: email, Password: password})
	if err != nil {
		return "", err
	}
	lr, err := decode[loginResp](resp)
	return lr.Token, err
}

func (c *Client) ListSnippets(q, tag string) ([]Snippet, error) {
	params := url.Values{}
	if q != "" {
		params.Set("q", q)
	}
	if tag != "" {
		params.Set("tag", tag)
	}
	path := "/api/snippets"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Snippet](resp)
}

func (c *Client) GetSnippet(id int64) (*Snippet, error) {
	resp, err := c.do("GET", "/api/snippets/"+strconv.FormatInt(id, 10), nil)
	if err != nil {
		return nil, err
	}
	s, err := decode[Snippet](resp)
	return &s, err
}

func (c *Client) CreateSnippet(title, content, language string, public bool, tags []string) (*Snippet, error) {
	resp, err := c.do("POST", "/api/snippets", createReq{
		Title:    title,
		Content:  content,
		Language: language,
		IsPublic: public,
		Tags:     tags,
	})
	if err != nil {
		return nil, err
	}
	s, err := decode[Snippet](resp)
	return &s, err
}

func (c *Client) DeleteSnippet(id int64) error {
	resp, err := c.do("DELETE", "/api/snippets/"+strconv.FormatInt(id, 10), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("server: %s", http.StatusText(resp.StatusCode))
	}
	return nil
}

// BoardEvent is the top-level SSE payload envelope.
type BoardEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// BoardSnippet carries snippet fields in snippet_added / snippet_updated events.
type BoardSnippet struct {
	ID       int64    `json:"id"`
	Title    string   `json:"title"`
	Language string   `json:"language"`
	Tags     []string `json:"tags"`
}

// BoardDeletedPayload carries the ID in a snippet_deleted event.
type BoardDeletedPayload struct {
	ID int64 `json:"id"`
}

// BoardEditingPayload carries presence info for editing_start / editing_stop events.
type BoardEditingPayload struct {
	SnippetID int64  `json:"snippet_id"`
	Email     string `json:"email"`
}

// WatchBoard connects to the SSE board for workspaceID and calls handler for each event.
// It blocks until the context is cancelled (caller sends os.Interrupt) or the connection drops.
func (c *Client) WatchBoard(workspaceID int64, handler func(BoardEvent)) error {
	req, err := http.NewRequest("GET", c.base+"/api/workspaces/"+strconv.FormatInt(workspaceID, 10)+"/board", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	// No timeout — SSE connection is long-lived.
	sseClient := &http.Client{}
	resp, err := sseClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server: %s", strings.TrimSpace(string(body)))
	}

	scanner := bufio.NewScanner(resp.Body)
	var eventType, dataLine string
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event: "):
			eventType = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			dataLine = strings.TrimPrefix(line, "data: ")
		case line == "" && eventType != "" && dataLine != "":
			var ev BoardEvent
			if err := json.Unmarshal([]byte(dataLine), &ev); err == nil {
				handler(ev)
			}
			eventType, dataLine = "", ""
		}
	}
	return scanner.Err()
}
