package storage

// LocalRecorder implements cyclic recording to disk.
type LocalRecorder struct{}

func NewLocalRecorder() *LocalRecorder { return &LocalRecorder{} }

func (r *LocalRecorder) Record(streamID string) error {
	_ = streamID
	// TODO: ring buffer, retention policy.
	return nil
}

