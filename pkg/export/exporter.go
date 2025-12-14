package export

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"time"

	"github.com/local-video-server/internal/models"
	"gopkg.in/yaml.v3"
)

// Exporter - интерфейс для экспорта результатов
type Exporter interface {
	Export(devices []*models.Device, filename string) error
}

// JSONExporter - экспорт в JSON
type JSONExporter struct{}

// Export экспортирует устройства в JSON файл
func (e *JSONExporter) Export(devices []*models.Device, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(devices); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// CSVExporter - экспорт в CSV
type CSVExporter struct{}

// Export экспортирует устройства в CSV файл
func (e *CSVExporter) Export(devices []*models.Device, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Заголовки
	headers := []string{
		"IP", "MAC", "Hostname", "Manufacturer", "Model",
		"Protocols", "RTSP Streams Count", "Discovered At", "Last Seen",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	// Данные
	for _, device := range devices {
		protocols := ""
		for i, p := range device.Protocols {
			if i > 0 {
				protocols += "; "
			}
			protocols += fmt.Sprintf("%s:%d", p.Type, p.Port)
		}

		record := []string{
			device.IP,
			device.MAC,
			device.Hostname,
			device.Manufacturer,
			device.Model,
			protocols,
			fmt.Sprintf("%d", len(device.RTSPStreams)),
			device.DiscoveredAt.Format(time.RFC3339),
			device.LastSeen.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// XMLExporter - экспорт в XML
type XMLExporter struct{}

// XMLDeviceList - структура для XML экспорта
type XMLDeviceList struct {
	XMLName xml.Name         `xml:"devices"`
	Devices []*models.Device `xml:"device"`
}

// Export экспортирует устройства в XML файл
func (e *XMLExporter) Export(devices []*models.Device, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	file.WriteString(xml.Header)

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")

	deviceList := XMLDeviceList{
		Devices: devices,
	}

	if err := encoder.Encode(deviceList); err != nil {
		return fmt.Errorf("failed to encode XML: %w", err)
	}

	return nil
}

// YAMLExporter - экспорт в YAML
type YAMLExporter struct{}

// Export экспортирует устройства в YAML файл
func (e *YAMLExporter) Export(devices []*models.Device, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	data, err := yaml.Marshal(devices)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportToFile экспортирует устройства в указанный формат
func ExportToFile(devices []*models.Device, format string, filename string) error {
	var exporter Exporter

	switch format {
	case "json":
		exporter = &JSONExporter{}
	case "csv":
		exporter = &CSVExporter{}
	case "xml":
		exporter = &XMLExporter{}
	case "yaml", "yml":
		exporter = &YAMLExporter{}
	default:
		return fmt.Errorf("unsupported format: %s (supported: json, csv, xml, yaml)", format)
	}

	return exporter.Export(devices, filename)
}

// ExportToMultipleFormats экспортирует устройства в несколько форматов одновременно
func ExportToMultipleFormats(devices []*models.Device, baseFilename string, formats []string) error {
	for _, format := range formats {
		var ext string
		switch format {
		case "json":
			ext = ".json"
		case "csv":
			ext = ".csv"
		case "xml":
			ext = ".xml"
		case "yaml", "yml":
			ext = ".yaml"
		default:
			continue
		}

		filename := baseFilename + ext
		if err := ExportToFile(devices, format, filename); err != nil {
			return fmt.Errorf("failed to export to %s: %w", format, err)
		}
	}

	return nil
}
