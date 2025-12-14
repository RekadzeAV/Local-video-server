package models

import "time"

// Device - обнаруженное устройство
type Device struct {
	IP           string            `json:"ip" yaml:"ip" xml:"ip"`
	MAC          string            `json:"mac,omitempty" yaml:"mac,omitempty" xml:"mac,omitempty"`
	Hostname     string            `json:"hostname,omitempty" yaml:"hostname,omitempty" xml:"hostname,omitempty"`
	Manufacturer string            `json:"manufacturer,omitempty" yaml:"manufacturer,omitempty" xml:"manufacturer,omitempty"`
	Model        string            `json:"model,omitempty" yaml:"model,omitempty" xml:"model,omitempty"`
	Protocols    []Protocol        `json:"protocols" yaml:"protocols" xml:"protocols>protocol"`
	RTSPStreams  []RTSPStreamInfo  `json:"rtsp_streams,omitempty" yaml:"rtsp_streams,omitempty" xml:"rtsp_streams>stream,omitempty"`
	DiscoveredAt time.Time         `json:"discovered_at" yaml:"discovered_at" xml:"discovered_at"`
	LastSeen     time.Time         `json:"last_seen,omitempty" yaml:"last_seen,omitempty" xml:"last_seen,omitempty"`
}

// Protocol - поддерживаемый протокол
type Protocol struct {
	Type      string    `json:"type" yaml:"type" xml:"type"` // RTSP, RTMP, HLS, etc.
	Port      int       `json:"port" yaml:"port" xml:"port"`
	URL       string    `json:"url,omitempty" yaml:"url,omitempty" xml:"url,omitempty"`
	Available bool      `json:"available" yaml:"available" xml:"available"`
	DetectedAt time.Time `json:"detected_at,omitempty" yaml:"detected_at,omitempty" xml:"detected_at,omitempty"`
}

// RTSPStreamInfo - информация о RTSP потоке
type RTSPStreamInfo struct {
	URL        string    `json:"url" yaml:"url" xml:"url"`
	Codec      string    `json:"codec" yaml:"codec" xml:"codec"`           // H.264, H.265, MJPEG
	Resolution string    `json:"resolution" yaml:"resolution" xml:"resolution"` // 1920x1080
	FPS        float64   `json:"fps" yaml:"fps" xml:"fps"`
	Bitrate    int       `json:"bitrate,omitempty" yaml:"bitrate,omitempty" xml:"bitrate,omitempty"`
	AudioCodec string    `json:"audio_codec,omitempty" yaml:"audio_codec,omitempty" xml:"audio_codec,omitempty"`
	Channels   int       `json:"channels,omitempty" yaml:"channels,omitempty" xml:"channels,omitempty"`
	Available  bool      `json:"available" yaml:"available" xml:"available"`
	CheckedAt  time.Time `json:"checked_at,omitempty" yaml:"checked_at,omitempty" xml:"checked_at,omitempty"`
}
