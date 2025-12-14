package models

// Stream - общая модель потока (используется для различных протоколов)
type Stream struct {
	// URL потока
	URL string `json:"url" yaml:"url"`

	// Тип протокола (RTSP, RTMP, HLS, MJPEG, etc.)
	Protocol string `json:"protocol" yaml:"protocol"`

	// Доступность потока
	Available bool `json:"available" yaml:"available"`

	// Метаданные потока (зависят от типа протокола)
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}
