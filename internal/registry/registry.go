package registry

import (
	"sync"
	"time"

	"github.com/local-video-server/internal/models"
)

// DeviceRegistry - реестр обнаруженных устройств
type DeviceRegistry struct {
	devices map[string]*models.Device
	mu      sync.RWMutex
	cache   *Cache
}

// Cache - кэш для результатов сканирования
type Cache struct {
	devices     map[string]*CachedDevice
	mu          sync.RWMutex
	defaultTTL  time.Duration
	lastScan    time.Time
	scanResults []*models.Device
}

// CachedDevice - кэшированное устройство
type CachedDevice struct {
	Device    *models.Device
	ExpiresAt time.Time
}

// NewDeviceRegistry создает новый реестр устройств
func NewDeviceRegistry(cacheTTL time.Duration) *DeviceRegistry {
	return &DeviceRegistry{
		devices: make(map[string]*models.Device),
		cache: &Cache{
			devices:    make(map[string]*CachedDevice),
			defaultTTL: cacheTTL,
		},
	}
}

// AddDevice добавляет или обновляет устройство в реестре
func (r *DeviceRegistry) AddDevice(device *models.Device) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Обновляем время последнего обнаружения
	now := time.Now()
	if existing, exists := r.devices[device.IP]; exists {
		// Обновляем существующее устройство
		existing.MAC = device.MAC
		existing.Hostname = device.Hostname
		existing.Manufacturer = device.Manufacturer
		existing.Model = device.Model
		existing.Protocols = device.Protocols
		existing.RTSPStreams = device.RTSPStreams
		existing.LastSeen = now
	} else {
		// Добавляем новое устройство
		device.DiscoveredAt = now
		device.LastSeen = now
		r.devices[device.IP] = device
	}
}

// GetDevice возвращает устройство по IP адресу
func (r *DeviceRegistry) GetDevice(ip string) (*models.Device, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	device, exists := r.devices[ip]
	if !exists {
		return nil, false
	}
	return device, true
}

// GetAllDevices возвращает все устройства
func (r *DeviceRegistry) GetAllDevices() []*models.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()

	devices := make([]*models.Device, 0, len(r.devices))
	for _, device := range r.devices {
		devices = append(devices, device)
	}
	return devices
}

// RemoveDevice удаляет устройство из реестра
func (r *DeviceRegistry) RemoveDevice(ip string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.devices, ip)
}

// UpdateDeviceState обновляет состояние устройства
func (r *DeviceRegistry) UpdateDeviceState(ip string, updateFunc func(*models.Device)) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	device, exists := r.devices[ip]
	if !exists {
		return false
	}

	updateFunc(device)
	device.LastSeen = time.Now()
	return true
}

// GetDeviceCount возвращает количество устройств в реестре
func (r *DeviceRegistry) GetDeviceCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.devices)
}

// Clear очищает реестр
func (r *DeviceRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.devices = make(map[string]*models.Device)
}

// Cache methods

// SaveToCache сохраняет результаты сканирования в кэш
func (r *DeviceRegistry) SaveToCache(devices []*models.Device) {
	r.cache.mu.Lock()
	defer r.cache.mu.Unlock()

	now := time.Now()
	expiresAt := now.Add(r.cache.defaultTTL)

	// Очищаем старый кэш
	r.cache.devices = make(map[string]*CachedDevice)

	// Сохраняем новые результаты
	for _, device := range devices {
		r.cache.devices[device.IP] = &CachedDevice{
			Device:    device,
			ExpiresAt: expiresAt,
		}
	}

	r.cache.lastScan = now
	r.cache.scanResults = devices
}

// GetFromCache возвращает результаты из кэша, если они еще не истекли
func (r *DeviceRegistry) GetFromCache() ([]*models.Device, bool) {
	r.cache.mu.RLock()
	defer r.cache.mu.RUnlock()

	now := time.Now()

	// Проверяем, не истек ли кэш
	if r.cache.lastScan.IsZero() || now.After(r.cache.lastScan.Add(r.cache.defaultTTL)) {
		return nil, false
	}

	// Проверяем, не истекли ли отдельные устройства
	validDevices := make([]*models.Device, 0)
	for _, cached := range r.cache.devices {
		if now.Before(cached.ExpiresAt) {
			validDevices = append(validDevices, cached.Device)
		}
	}

	if len(validDevices) == 0 {
		return nil, false
	}

	return validDevices, true
}

// ClearCache очищает кэш
func (r *DeviceRegistry) ClearCache() {
	r.cache.mu.Lock()
	defer r.cache.mu.Unlock()

	r.cache.devices = make(map[string]*CachedDevice)
	r.cache.lastScan = time.Time{}
	r.cache.scanResults = nil
}

// GetLastScanTime возвращает время последнего сканирования
func (r *DeviceRegistry) GetLastScanTime() time.Time {
	r.cache.mu.RLock()
	defer r.cache.mu.RUnlock()

	return r.cache.lastScan
}

// FilterDevices фильтрует устройства по критериям
func (r *DeviceRegistry) FilterDevices(filterFunc func(*models.Device) bool) []*models.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Device
	for _, device := range r.devices {
		if filterFunc(device) {
			result = append(result, device)
		}
	}
	return result
}

// GetDevicesByProtocol возвращает устройства с указанным протоколом
func (r *DeviceRegistry) GetDevicesByProtocol(protocolType string) []*models.Device {
	return r.FilterDevices(func(device *models.Device) bool {
		for _, protocol := range device.Protocols {
			if protocol.Type == protocolType {
				return true
			}
		}
		return false
	})
}

// GetDevicesWithRTSP возвращает устройства с RTSP потоками
func (r *DeviceRegistry) GetDevicesWithRTSP() []*models.Device {
	return r.FilterDevices(func(device *models.Device) bool {
		return len(device.RTSPStreams) > 0
	})
}
