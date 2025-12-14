package rtsp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/local-video-server/internal/models"
)

// StreamInfo содержит информацию о RTSP потоке
type StreamInfo struct {
	URL        string
	Codec      string
	Resolution string
	FPS        float64
	Bitrate    int
	AudioCodec string
	Channels   int
	Available  bool
	VideoTracks []VideoTrack
	AudioTracks []AudioTrack
}

// VideoTrack содержит информацию о видео дорожке
type VideoTrack struct {
	Codec      string
	Resolution string
	FPS        float64
	Bitrate    int
	Profile    string
	Level      string
}

// AudioTrack содержит информацию об аудио дорожке
type AudioTrack struct {
	Codec    string
	Channels int
	SampleRate int
	Bitrate  int
}

// ParseSDP парсит SDP (Session Description Protocol) и извлекает информацию о потоке
func ParseSDP(sdp string) (*StreamInfo, error) {
	info := &StreamInfo{
		VideoTracks: []VideoTrack{},
		AudioTracks: []AudioTrack{},
	}

	lines := strings.Split(sdp, "\n")
	var currentMedia *MediaDescription

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// SDP формат: тип=значение
		if len(line) < 2 || line[1] != '=' {
			continue
		}

		mediaType := line[0]
		value := line[2:]

		switch mediaType {
		case 'v': // version
			// Игнорируем
		case 'o': // origin
			// Игнорируем
		case 's': // session name
			// Игнорируем
		case 't': // timing
			// Игнорируем
		case 'm': // media description
			// Парсим медиа описание (m=video 0 RTP/AVP 96)
			parts := strings.Fields(value)
			if len(parts) >= 2 {
				mediaType := parts[0]
				if mediaType == "video" || mediaType == "audio" {
					currentMedia = &MediaDescription{
						Type: mediaType,
					}
					if mediaType == "video" {
						info.VideoTracks = append(info.VideoTracks, VideoTrack{})
					} else {
						info.AudioTracks = append(info.AudioTracks, AudioTrack{})
					}
				}
			}
		case 'a': // attribute
			if currentMedia != nil {
				parseAttribute(line, currentMedia, info)
			}
		}
	}

	// Извлекаем основную информацию из первой видео дорожки
	if len(info.VideoTracks) > 0 {
		videoTrack := info.VideoTracks[0]
		info.Codec = videoTrack.Codec
		info.Resolution = videoTrack.Resolution
		info.FPS = videoTrack.FPS
		info.Bitrate = videoTrack.Bitrate
	}

	// Извлекаем основную информацию из первой аудио дорожки
	if len(info.AudioTracks) > 0 {
		audioTrack := info.AudioTracks[0]
		info.AudioCodec = audioTrack.Codec
		info.Channels = audioTrack.Channels
	}

	return info, nil
}

// MediaDescription представляет описание медиа в SDP
type MediaDescription struct {
	Type      string
	Codec     string
	Format    string
	Width     int
	Height    int
	FPS       float64
	Bitrate   int
	Profile   string
	Level     string
	Channels  int
	SampleRate int
}

// parseAttribute парсит атрибут SDP
func parseAttribute(line string, media *MediaDescription, info *StreamInfo) {
	// Убираем "a="
	attr := line[2:]

	// rtpmap: формат кодека
	if strings.HasPrefix(attr, "rtpmap:") {
		// Пример: rtpmap:96 H264/90000
		parts := strings.Fields(attr[7:])
		if len(parts) >= 2 {
			codecInfo := parts[1]
			codecParts := strings.Split(codecInfo, "/")
			codec := codecParts[0]

			if media.Type == "video" {
				// Обновляем последнюю видео дорожку
				if len(info.VideoTracks) > 0 {
					idx := len(info.VideoTracks) - 1
					info.VideoTracks[idx].Codec = normalizeCodec(codec)
					info.Codec = normalizeCodec(codec)
				}
			} else if media.Type == "audio" {
				// Обновляем последнюю аудио дорожку
				if len(info.AudioTracks) > 0 {
					idx := len(info.AudioTracks) - 1
					info.AudioTracks[idx].Codec = normalizeAudioCodec(codec)
					info.AudioCodec = normalizeAudioCodec(codec)
				}
			}
		}
	}

	// fmtp: параметры формата
	if strings.HasPrefix(attr, "fmtp:") {
		// Пример: fmtp:96 profile-level-id=420029; packetization-mode=1; sprop-parameter-sets=...
		parts := strings.SplitN(attr[5:], " ", 2)
		if len(parts) == 2 {
			params := parts[1]
			parseFmtpParams(params, media, info)
		}
	}

	// framerate: FPS
	if strings.HasPrefix(attr, "framerate:") {
		fpsStr := strings.TrimSpace(attr[10:])
		if fps, err := strconv.ParseFloat(fpsStr, 64); err == nil {
			if media.Type == "video" && len(info.VideoTracks) > 0 {
				idx := len(info.VideoTracks) - 1
				info.VideoTracks[idx].FPS = fps
				info.FPS = fps
			}
		}
	}

	// x-dimensions: разрешение (нестандартный атрибут)
	if strings.HasPrefix(attr, "x-dimensions:") {
		resolution := strings.TrimSpace(attr[13:])
		// Пример: 1920x1080
		if parts := strings.Split(resolution, "x"); len(parts) == 2 {
			if width, err := strconv.Atoi(parts[0]); err == nil {
				if height, err := strconv.Atoi(parts[1]); err == nil {
					resolutionStr := fmt.Sprintf("%dx%d", width, height)
					if media.Type == "video" && len(info.VideoTracks) > 0 {
						idx := len(info.VideoTracks) - 1
						info.VideoTracks[idx].Resolution = resolutionStr
						info.Resolution = resolutionStr
					}
				}
			}
		}
	}

	// Извлекаем разрешение из sprop-parameter-sets (H.264)
	if strings.Contains(attr, "sprop-parameter-sets=") {
		// Пытаемся извлечь разрешение из SPS
		if resolution := extractResolutionFromSPS(attr); resolution != "" {
			if media.Type == "video" && len(info.VideoTracks) > 0 {
				idx := len(info.VideoTracks) - 1
				info.VideoTracks[idx].Resolution = resolution
				info.Resolution = resolution
			}
		}
	}
}

// parseFmtpParams парсит параметры fmtp
func parseFmtpParams(params string, media *MediaDescription, info *StreamInfo) {
	// Разделяем параметры по точке с запятой
	paramPairs := strings.Split(params, ";")
	
	for _, pair := range paramPairs {
		pair = strings.TrimSpace(pair)
		if idx := strings.Index(pair, "="); idx != -1 {
			key := strings.TrimSpace(pair[:idx])
			value := strings.TrimSpace(pair[idx+1:])

			if media.Type == "video" && len(info.VideoTracks) > 0 {
				idx := len(info.VideoTracks) - 1
				track := &info.VideoTracks[idx]

				switch key {
				case "profile-level-id":
					// Пример: 420029 (H.264)
					if len(value) >= 6 {
						profile := value[0:2]
						level := value[4:6]
						track.Profile = profile
						track.Level = level
					}
				case "sprop-parameter-sets":
					// Пытаемся извлечь разрешение из SPS
					if resolution := extractResolutionFromSPS(value); resolution != "" {
						track.Resolution = resolution
						info.Resolution = resolution
					}
				}
			}
		}
	}
}

// extractResolutionFromSPS пытается извлечь разрешение из SPS (Sequence Parameter Set)
// Это сложная задача, так как SPS закодирован в base64 и требует декодирования
// Здесь упрощенная версия, которая пытается найти известные паттерны
func extractResolutionFromSPS(sps string) string {
	// SPS обычно в формате: sprop-parameter-sets=Z0IAHpWoKA9puAgICBA=,aM48gA==
	// Это base64 закодированные данные
	// Для полного парсинга нужна библиотека для декодирования H.264 SPS
	
	// Упрощенный подход: пытаемся найти известные разрешения в других атрибутах
	// Или используем значения по умолчанию для популярных камер
	return ""
}

// normalizeCodec нормализует название кодека
func normalizeCodec(codec string) string {
	codec = strings.ToUpper(codec)
	
	// H.264 варианты
	if strings.Contains(codec, "H264") || strings.Contains(codec, "H.264") || codec == "AVC" {
		return "H.264"
	}
	
	// H.265 варианты
	if strings.Contains(codec, "H265") || strings.Contains(codec, "H.265") || codec == "HEVC" {
		return "H.265"
	}
	
	// MJPEG
	if strings.Contains(codec, "MJPEG") || strings.Contains(codec, "JPEG") {
		return "MJPEG"
	}
	
	// MPEG4
	if strings.Contains(codec, "MPEG4") || strings.Contains(codec, "MPEG-4") {
		return "MPEG-4"
	}
	
	return codec
}

// normalizeAudioCodec нормализует название аудио кодека
func normalizeAudioCodec(codec string) string {
	codec = strings.ToUpper(codec)
	
	// AAC
	if strings.Contains(codec, "AAC") {
		return "AAC"
	}
	
	// PCM
	if strings.Contains(codec, "PCM") {
		return "PCM"
	}
	
	// G.711
	if strings.Contains(codec, "PCMU") || strings.Contains(codec, "PCMA") {
		return "G.711"
	}
	
	// G.722
	if strings.Contains(codec, "G722") {
		return "G.722"
	}
	
	return codec
}

// ExtractResolutionFromFmtp пытается извлечь разрешение из fmtp параметров
func ExtractResolutionFromFmtp(fmtp string) (width, height int, err error) {
	// Ищем паттерны типа width=1920, height=1080
	widthRe := regexp.MustCompile(`width[=:](\d+)`)
	heightRe := regexp.MustCompile(`height[=:](\d+)`)

	widthMatch := widthRe.FindStringSubmatch(fmtp)
	heightMatch := heightRe.FindStringSubmatch(fmtp)

	if len(widthMatch) == 2 && len(heightMatch) == 2 {
		width, err1 := strconv.Atoi(widthMatch[1])
		height, err2 := strconv.Atoi(heightMatch[1])
		if err1 == nil && err2 == nil {
			return width, height, nil
		}
	}

	return 0, 0, fmt.Errorf("resolution not found in fmtp")
}

// ToRTSPStreamInfo конвертирует StreamInfo в models.RTSPStreamInfo
func (s *StreamInfo) ToRTSPStreamInfo() models.RTSPStreamInfo {
	return models.RTSPStreamInfo{
		URL:        s.URL,
		Codec:      s.Codec,
		Resolution: s.Resolution,
		FPS:        s.FPS,
		Bitrate:    s.Bitrate,
		AudioCodec: s.AudioCodec,
		Channels:   s.Channels,
		Available:  s.Available,
	}
}
