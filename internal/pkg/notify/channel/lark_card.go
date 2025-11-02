package channel

import (
	"context"

	"github.com/go-arcade/arcade/internal/pkg/notify/auth"
)

// LarkCardChannel implements Lark card notification channel
type LarkCardChannel struct {
	*FeishuCardChannel
}

// NewLarkCardChannel creates a new Lark card notification channel
func NewLarkCardChannel(appID, appSecret string) *LarkCardChannel {
	return &LarkCardChannel{
		FeishuCardChannel: NewFeishuCardChannel(appID, appSecret),
	}
}

// SetAuth sets authentication provider
func (c *LarkCardChannel) SetAuth(provider auth.IAuthProvider) error {
	return c.FeishuCardChannel.SetAuth(provider)
}

// GetAuth gets the authentication provider
func (c *LarkCardChannel) GetAuth() auth.IAuthProvider {
	return c.FeishuCardChannel.GetAuth()
}

// Send sends card message
func (c *LarkCardChannel) Send(ctx context.Context, message string) error {
	return c.FeishuCardChannel.Send(ctx, message)
}

// SendWithTemplate sends card message using template
func (c *LarkCardChannel) SendWithTemplate(ctx context.Context, template string, data map[string]interface{}) error {
	return c.FeishuCardChannel.SendWithTemplate(ctx, template, data)
}

// Receive receives messages
func (c *LarkCardChannel) Receive(ctx context.Context, message string) error {
	return c.FeishuCardChannel.Receive(ctx, message)
}

// Validate validates the configuration
func (c *LarkCardChannel) Validate() error {
	return c.FeishuCardChannel.Validate()
}

// Close closes the connection
func (c *LarkCardChannel) Close() error {
	return c.FeishuCardChannel.Close()
}
