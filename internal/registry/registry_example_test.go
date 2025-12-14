package registry_test

import (
	"time"

	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/internal/registry"
)

// ExampleDeviceRegistry демонстрирует использование реестра устройств
func ExampleDeviceRegistry() {
	// Создаем реестр с TTL кэша 1 час
	reg := registry.NewDeviceRegistry(1 * time.Hour)

	// Добавляем устройство
	device := &models.Device{
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
	}

	reg.AddDevice(device)

	// Получаем устройство
	dev, exists := reg.GetDevice("192.168.1.100")
	if exists {
		_ = dev
	}

	// Получаем все устройства
	allDevices := reg.GetAllDevices()
	_ = allDevices

	// Сохраняем в кэш
	reg.SaveToCache(allDevices)

	// Получаем из кэша
	cached, found := reg.GetFromCache()
	if found {
		_ = cached
	}
}

// ExampleDeviceRegistryFiltering демонстрирует фильтрацию устройств
func ExampleDeviceRegistryFiltering() {
	reg := registry.NewDeviceRegistry(1 * time.Hour)

	// Добавляем несколько устройств
	reg.AddDevice(&models.Device{
		IP: "192.168.1.100",
		Protocols: []models.Protocol{
			{Type: "RTSP", Port: 554, Available: true},
		},
	})

	reg.AddDevice(&models.Device{
		IP: "192.168.1.101",
		Protocols: []models.Protocol{
			{Type: "RTMP", Port: 1935, Available: true},
		},
	})

	// Фильтруем устройства с RTSP
	rtspDevices := reg.GetDevicesByProtocol("RTSP")
	_ = rtspDevices

	// Фильтруем устройства с RTSP потоками
	devicesWithRTSP := reg.GetDevicesWithRTSP()
	_ = devicesWithRTSP
}
