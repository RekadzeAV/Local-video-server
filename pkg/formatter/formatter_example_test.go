package formatter_test

import (
	"time"

	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/pkg/formatter"
)

// ExampleFormatter демонстрирует использование форматтера
func ExampleFormatter() {
	// Создаем форматтер с цветами и детальным выводом
	f := formatter.NewFormatter(true, false)

	// Создаем тестовые данные
	devices := []*models.Device{
		{
			IP:           "192.168.1.100",
			MAC:          "00:11:22:33:44:55",
			Hostname:     "camera-01",
			Manufacturer: "Hikvision",
			Model:        "DS-2CD2342WD-I",
			Protocols: []models.Protocol{
				{
					Type:      "RTSP",
					Port:      554,
					URL:       "rtsp://192.168.1.100:554/Streaming/Channels/101",
					Available: true,
					DetectedAt: time.Now(),
				},
			},
			DiscoveredAt: time.Now(),
			LastSeen:     time.Now(),
		},
	}

	// Вывод в табличном формате
	f.PrintDevices(devices)

	// Вывод сводки
	f.PrintSummary(devices)
}

// ExampleFormatterDetailed демонстрирует детальный вывод
func ExampleFormatterDetailed() {
	// Создаем форматтер с детальным выводом
	f := formatter.NewFormatter(true, true)

	devices := []*models.Device{
		{
			IP:           "192.168.1.100",
			MAC:          "00:11:22:33:44:55",
			Hostname:     "camera-01",
			Manufacturer: "Hikvision",
			Model:        "DS-2CD2342WD-I",
			Protocols: []models.Protocol{
				{
					Type:      "RTSP",
					Port:      554,
					URL:       "rtsp://192.168.1.100:554/Streaming/Channels/101",
					Available: true,
					DetectedAt: time.Now(),
				},
			},
			RTSPStreams: []models.RTSPStreamInfo{
				{
					URL:        "rtsp://192.168.1.100:554/Streaming/Channels/101",
					Codec:      "H.264",
					Resolution: "1920x1080",
					FPS:        25.0,
					Available:  true,
					CheckedAt:  time.Now(),
				},
			},
			DiscoveredAt: time.Now(),
			LastSeen:     time.Now(),
		},
	}

	// Детальный вывод
	f.PrintDevices(devices)
}
