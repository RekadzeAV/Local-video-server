package camera

import "context"

// RTSPClient placeholder for connect/describe/play logic.
type RTSPClient struct{}

func NewRTSPClient() *RTSPClient {
	return &RTSPClient{}
}

func (c *RTSPClient) Probe(ctx context.Context, url string) error {
	_ = ctx
	_ = url
	// TODO: implement OPTIONS/DESCRIBE, reconnect strategy, auth.
	return nil
}

