package camera

import "context"

// ONVIFDiscovery placeholder for WS-Discovery.
type ONVIFDiscovery struct{}

func NewONVIFDiscovery() *ONVIFDiscovery {
	return &ONVIFDiscovery{}
}

func (d *ONVIFDiscovery) Discover(ctx context.Context, iface string) ([]string, error) {
	_ = ctx
	_ = iface
	// TODO: send Probe, parse ProbeMatch, return device endpoints.
	return nil, nil
}

