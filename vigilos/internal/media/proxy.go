package media

// Proxy handles ingest and transcoding pipelines.
type Proxy struct{}

func NewProxy() *Proxy {
	return &Proxy{}
}

func (p *Proxy) Start() error { return nil }
func (p *Proxy) Stop() error  { return nil }

