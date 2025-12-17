package events

// WebhookSender dispatches events to external systems.
type WebhookSender struct{}

func NewWebhookSender() *WebhookSender { return &WebhookSender{} }

func (s *WebhookSender) Send(target string, payload []byte) error {
	_ = target
	_ = payload
	return nil
}

