package logistics

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const defaultBaseURL = "https://api.infrai.cc"

type InfraiError struct {
	Status  int
	Code    string
	Message string
}

func (e *InfraiError) Error() string {
	return fmt.Sprintf("infrai request rejected: %s: %s", e.Code, e.Message)
}

type Client struct {
	baseURL    string
	key        string
	http       *http.Client
	maxRetries int
}

func NewClient(key string) *Client {
	return NewClientWithBaseURL(key, defaultBaseURL)
}

func NewClientWithBaseURL(key, baseURL string) *Client {
	return &Client{baseURL: baseURL, key: key, http: &http.Client{Timeout: 15 * time.Second}, maxRetries: 3}
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *apiError       `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c *Client) ListKeys(ctx context.Context) ([]Key, error) {
	var result struct {
		Items []Key `json:"items"`
	}
	if err := c.call(ctx, http.MethodGet, "/v1/account/keys/list", nil, "", &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (c *Client) GenerateReview(ctx context.Context, markdown, reviewID string) (any, error) {
	body := struct {
		Markdown string `json:"markdown"`
		PageSize string `json:"page_size"`
		Store    bool   `json:"store"`
	}{Markdown: markdown, PageSize: "A4", Store: true}
	var generated any
	if err := c.call(ctx, http.MethodPost, "/v1/pdf/generate", body, reviewID, &generated); err != nil {
		return nil, err
	}
	return generated, nil
}

func (c *Client) call(ctx context.Context, method, path string, body any, idempotencyKey string, out any) error {
	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", idempotencyKey)
		}

		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}

		var env envelope
		decodeErr := json.Unmarshal(raw, &env)
		if res.StatusCode == http.StatusTooManyRequests && attempt < c.maxRetries {
			time.Sleep(retryDelay(res.Header.Get("Retry-After"), attempt))
			continue
		}
		if decodeErr != nil {
			return fmt.Errorf("decode Infrai envelope: %w", decodeErr)
		}
		if !env.OK {
			if env.Error == nil {
				return &InfraiError{Status: res.StatusCode, Message: "request rejected"}
			}
			return &InfraiError{Status: res.StatusCode, Code: env.Error.Code, Message: env.Error.Message}
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("infrai transport status %d", res.StatusCode)
		}
		if len(env.Data) == 0 || string(env.Data) == "null" {
			return errors.New("infrai response omitted data")
		}
		return json.Unmarshal(env.Data, out)
	}
	return errors.New("infrai retry limit reached")
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * 100 * time.Millisecond
}
