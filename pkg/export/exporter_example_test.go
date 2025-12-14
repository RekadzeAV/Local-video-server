package export_test

import (
	"os"
	"time"

	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/pkg/export"
)

// ExampleExportToFile демонстрирует экспорт устройств в различные форматы
func ExampleExportToFile() {
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

	// Экспорт в JSON
	_ = export.ExportToFile(devices, "json", "devices.json")
	defer os.Remove("devices.json")

	// Экспорт в CSV
	_ = export.ExportToFile(devices, "csv", "devices.csv")
	defer os.Remove("devices.csv")

	// Экспорт в XML
	_ = export.ExportToFile(devices, "xml", "devices.xml")
	defer os.Remove("devices.xml")

	// Экспорт в YAML
	_ = export.ExportToFile(devices, "yaml", "devices.yaml")
	defer os.Remove("devices.yaml")
}

// ExampleExportToMultipleFormats демонстрирует экспорт в несколько форматов одновременно
func ExampleExportToMultipleFormats() {
	devices := []*models.Device{
		{
			IP:           "192.168.1.100",
			MAC:          "00:11:22:33:44:55",
			Hostname:     "camera-01",
			Manufacturer: "Hikvision",
			Model:        "DS-2CD2342WD-I",
			DiscoveredAt: time.Now(),
		},
	}

	formats := []string{"json", "csv", "xml", "yaml"}
	_ = export.ExportToMultipleFormats(devices, "devices", formats)
	
	// Очистка
	for _, format := range formats {
		var ext string
		switch format {
		case "json":
			ext = ".json"
		case "csv":
			ext = ".csv"
		case "xml":
			ext = ".xml"
		case "yaml":
			ext = ".yaml"
		}
		os.Remove("devices" + ext)
	}
}
