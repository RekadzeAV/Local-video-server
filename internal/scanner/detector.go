package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/pkg/utils"
	"github.com/sirupsen/logrus"
)

// Detector объединяет все методы обнаружения устройств
type Detector struct {
	config        *models.ScanConfig
	logger        *logrus.Logger
	networkScanner *NetworkScanner
	onvifScanner  *ONVIFScanner
	upnpScanner   *UPnPScanner
}

// NewDetector создает новый экземпляр Detector
func NewDetector(config *models.ScanConfig) *Detector {
	return &Detector{
		config:         config,
		logger:         utils.GetLogger(),
		networkScanner: NewNetworkScanner(config),
		onvifScanner:   NewONVIFScanner(config),
		upnpScanner:    NewUPnPScanner(config),
	}
}

// Scan выполняет полное сканирование сети всеми доступными методами
func (d *Detector) Scan(ctx context.Context, subnet string) ([]*models.Device, error) {
	d.logger.Infof("Starting comprehensive network scan for subnet: %s", subnet)

	// Map для объединения результатов (по IP адресу)
	devicesMap := make(map[string]*models.Device)
	var mu sync.Mutex

	// Функция для объединения устройств
	mergeDevice := func(device *models.Device) {
		if device == nil {
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if existing, exists := devicesMap[device.IP]; exists {
			// Объединяем информацию об устройстве
			d.mergeDevices(existing, device)
		} else {
			devicesMap[device.IP] = device
		}
	}

	var wg sync.WaitGroup

	// 1. Базовое сканирование сети (ARP + порты)
	wg.Add(1)
	go func() {
		defer wg.Done()
		devices, err := d.networkScanner.ScanNetwork(ctx, subnet)
		if err != nil {
			d.logger.Warnf("Network scan failed: %v", err)
			return
		}
		for _, device := range devices {
			mergeDevice(device)
		}
		d.logger.Infof("Network scan completed: found %d devices", len(devices))
	}()

	// 2. ONVIF Discovery (если включен)
	if d.config.EnableONVIF {
		wg.Add(1)
		go func() {
			defer wg.Done()
			devices, err := d.onvifScanner.Discover(ctx)
			if err != nil {
				d.logger.Warnf("ONVIF discovery failed: %v", err)
				return
			}
			for _, device := range devices {
				mergeDevice(device)
			}
			d.logger.Infof("ONVIF discovery completed: found %d devices", len(devices))
		}()
	}

	// 3. UPnP/SSDP Discovery (если включен)
	if d.config.EnableUPnP {
		wg.Add(1)
		go func() {
			defer wg.Done()
			devices, err := d.upnpScanner.Discover(ctx)
			if err != nil {
				d.logger.Warnf("UPnP discovery failed: %v", err)
				return
			}
			for _, device := range devices {
				mergeDevice(device)
			}
			d.logger.Infof("UPnP discovery completed: found %d devices", len(devices))
		}()
	}

	// Ждем завершения всех методов сканирования
	wg.Wait()

	// Преобразуем map в slice
	devices := make([]*models.Device, 0, len(devicesMap))
	for _, device := range devicesMap {
		devices = append(devices, device)
	}

	d.logger.Infof("Comprehensive scan completed: found %d unique devices", len(devices))
	return devices, nil
}

// mergeDevices объединяет информацию о двух устройствах с одинаковым IP
func (d *Detector) mergeDevices(existing, new *models.Device) {
	// Объединяем протоколы
	protocolMap := make(map[string]models.Protocol)
	for _, p := range existing.Protocols {
		key := fmt.Sprintf("%s:%d", p.Type, p.Port)
		protocolMap[key] = p
	}

	for _, p := range new.Protocols {
		key := fmt.Sprintf("%s:%d", p.Type, p.Port)
		if _, exists := protocolMap[key]; !exists {
			existing.Protocols = append(existing.Protocols, p)
		}
	}

	// Обновляем информацию, если она отсутствует
	if existing.Manufacturer == "" && new.Manufacturer != "" {
		existing.Manufacturer = new.Manufacturer
	}
	if existing.Model == "" && new.Model != "" {
		existing.Model = new.Model
	}
	if existing.Hostname == "" && new.Hostname != "" {
		existing.Hostname = new.Hostname
	}
	if existing.MAC == "" && new.MAC != "" {
		existing.MAC = new.MAC
	}

	// Обновляем LastSeen
	if new.DiscoveredAt.After(existing.DiscoveredAt) {
		existing.LastSeen = time.Now()
	}
}

// ScanWithTimeout выполняет сканирование с таймаутом
func (d *Detector) ScanWithTimeout(subnet string, timeout time.Duration) ([]*models.Device, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return d.Scan(ctx, subnet)
}
