package storage

// NFSClient placeholder for NFSv4 integration.
type NFSClient struct{}

func NewNFSClient() *NFSClient { return &NFSClient{} }

func (c *NFSClient) Mount(endpoint string) error {
	_ = endpoint
	return nil
}

