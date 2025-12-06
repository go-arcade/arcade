package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-arcade/arcade/internal/pkg/notify/auth"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-resty/resty/v2"
)

// WeComChannel implements WeCom (WeChat Work) notification channel
type WeComChannel struct {
	webhookURL   string
	authProvider auth.IAuthProvider
	client       *resty.Client
}

// NewWeComChannel creates a new WeCom notification channel
func NewWeComChannel(webhookURL string) *WeComChannel {
	return &WeComChannel{
		webhookURL: webhookURL,
		client:     resty.New(),
	}
}

// SetAuth sets authentication provider (WeCom webhook typically doesn't need extra auth but supports custom auth)
func (c *WeComChannel) SetAuth(provider auth.IAuthProvider) error {
	if provider == nil {
		return nil
	}

	c.authProvider = provider
	return provider.Validate()
}

// GetAuth gets the authentication provider
func (c *WeComChannel) GetAuth() auth.IAuthProvider {
	return c.authProvider
}

// Send sends message to WeCom
func (c *WeComChannel) Send(ctx context.Context, message string) error {
	if err := c.Validate(); err != nil {
		return err
	}

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}

	return c.sendRequest(ctx, payload)
}

// SendWithTemplate sends message using template
func (c *WeComChannel) SendWithTemplate(ctx context.Context, template string, data map[string]interface{}) error {
	if err := c.Validate(); err != nil {
		return err
	}

	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": template,
		},
	}

	return c.sendRequest(ctx, payload)
}

// sendRequest sends HTTP request
func (c *WeComChannel) sendRequest(ctx context.Context, payload map[string]interface{}) error {
	req := c.client.R().SetContext(ctx)

	// Add authentication header
	if c.authProvider != nil {
		key, value := c.authProvider.GetAuthHeader()
		if key != "" && value != "" {
			req.SetHeader(key, value)
		}
	}

	req.SetHeader("Content-Type", "application/json")
	req.SetBody(payload)

	resp, err := req.Post(c.webhookURL)
	if err != nil {
		log.Error("wecom send request failed: %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		log.Error("wecom request failed with status %d: %s", resp.StatusCode(), resp.String())
		return fmt.Errorf("wecom request failed with status %d", resp.StatusCode())
	}

	// Check for errors in response body
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err == nil {
		if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
			errmsg := result["errmsg"].(string)
			return fmt.Errorf("wecom API error: errcode=%v, errmsg=%s", errcode, errmsg)
		}
	}

	return nil
}

// Receive receives messages (webhook callback)
func (c *WeComChannel) Receive(ctx context.Context, message string) error {
	return nil
}

// Validate validates the configuration
func (c *WeComChannel) Validate() error {
	if c.webhookURL == "" {
		return fmt.Errorf("wecom webhook URL is required")
	}
	if c.authProvider != nil {
		return c.authProvider.Validate()
	}
	return nil
}

// Close closes the connection
func (c *WeComChannel) Close() error {
	return nil
}
