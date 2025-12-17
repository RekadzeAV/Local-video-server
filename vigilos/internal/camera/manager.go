package camera

// Manager coordinates camera lifecycle, discovery, and configuration.
type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Start() error {
	// TODO: wire RTSP client, ONVIF discovery, status tracking.
	return nil
}

func (m *Manager) Stop() error {
	return nil
}

