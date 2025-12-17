package media

// WebRTCRelay handles relaying media to browsers with low latency.
type WebRTCRelay struct{}

func NewWebRTCRelay() *WebRTCRelay {
	return &WebRTCRelay{}
}

func (r *WebRTCRelay) Start() error { return nil }
func (r *WebRTCRelay) Stop() error  { return nil }

