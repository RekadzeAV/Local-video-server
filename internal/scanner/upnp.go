package scanner

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/pkg/utils"
	"github.com/sirupsen/logrus"
)

// UPnPScanner выполняет обнаружение устройств через UPnP/SSDP
type UPnPScanner struct {
	config *models.ScanConfig
	logger *logrus.Logger
}

// NewUPnPScanner создает новый экземпляр UPnPScanner
func NewUPnPScanner(config *models.ScanConfig) *UPnPScanner {
	return &UPnPScanner{
		config: config,
		logger: utils.GetLogger(),
	}
}

// SSDP константы
const (
	SSDPMulticastIPv4 = "239.255.255.250"
	SSDPPort          = 1900
	SSDPMaxAge        = 1800
)

// SSDPResponse представляет SSDP ответ от устройства
type SSDPResponse struct {
	CacheControl string
	Location     string
	Server       string
	ST           string // Search Target
	USN          string // Unique Service Name
	EXT          string
	Date         string
}

// Discover выполняет UPnP/SSDP Discovery для обнаружения устройств
func (us *UPnPScanner) Discover(ctx context.Context) ([]*models.Device, error) {
	us.logger.Infof("Starting UPnP/SSDP Discovery")

	// Создаем UDP соединение для отправки M-SEARCH запросов
	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP socket: %w", err)
	}
	defer conn.Close()

	// Устанавливаем таймаут для чтения
	timeout := us.config.DiscoveryTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	conn.SetReadDeadline(time.Now().Add(timeout))

	// Отправляем M-SEARCH запросы для различных типов устройств
	searchTargets := []string{
		"urn:schemas-upnp-org:device:MediaServer:1",
		"urn:schemas-upnp-org:device:MediaRenderer:1",
		"urn:schemas-upnp-org:device:InternetGatewayDevice:1",
		"upnp:rootdevice",
		"ssdp:all",
	}

	// Отправляем M-SEARCH для каждого типа
	multicastAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", SSDPMulticastIPv4, SSDPPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve multicast address: %w", err)
	}

	for _, st := range searchTargets {
		msearch := us.buildMSearchRequest(st)
		_, err = conn.WriteToUDP([]byte(msearch), multicastAddr)
		if err != nil {
			us.logger.Warnf("Failed to send M-SEARCH for %s: %v", st, err)
			continue
		}
		us.logger.Debugf("Sent M-SEARCH request for %s", st)
	}

	// Слушаем ответы
	devices := make(map[string]*models.Device)
	buffer := make([]byte, 4096)

	// Читаем ответы в течение таймаута
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			us.logger.Debugf("UPnP discovery cancelled")
			return us.devicesToSlice(devices), nil
		default:
			// Устанавливаем таймаут для каждого чтения
			remaining := time.Until(deadline)
			if remaining > 1*time.Second {
				remaining = 1 * time.Second
			}
			conn.SetReadDeadline(time.Now().Add(remaining))

			n, addr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Таймаут - продолжаем, если общее время не истекло
					continue
				}
				us.logger.Debugf("Error reading UDP response: %v", err)
				continue
			}

			// Парсим SSDP ответ
			device, err := us.parseSSDPResponse(buffer[:n], addr.IP.String())
			if err != nil {
				us.logger.Debugf("Failed to parse SSDP response: %v", err)
				continue
			}

			if device != nil {
				// Объединяем информацию, если устройство уже найдено
				if existing, exists := devices[device.IP]; exists {
					us.mergeDeviceInfo(existing, device)
				} else {
					devices[device.IP] = device
				}
			}
		}
	}

	us.logger.Infof("UPnP/SSDP Discovery completed. Found %d devices", len(devices))
	return us.devicesToSlice(devices), nil
}

// buildMSearchRequest создает M-SEARCH SSDP запрос
func (us *UPnPScanner) buildMSearchRequest(searchTarget string) string {
	return fmt.Sprintf(`M-SEARCH * HTTP/1.1
HOST: %s:%d
MAN: "ssdp:discover"
ST: %s
MX: 3
USER-AGENT: Local-video-server/1.0

`, SSDPMulticastIPv4, SSDPPort, searchTarget)
}

// parseSSDPResponse парсит SSDP ответ и извлекает информацию об устройстве
func (us *UPnPScanner) parseSSDPResponse(data []byte, sourceIP string) (*models.Device, error) {
	response := string(data)
	
	// Проверяем, что это HTTP ответ
	if !strings.HasPrefix(response, "HTTP/1.1") && !strings.HasPrefix(response, "HTTP/1.0") {
		return nil, fmt.Errorf("not an HTTP response")
	}

	// Парсим заголовки
	ssdpResp := us.parseSSDPHeaders(response)
	
	// Извлекаем IP из Location или используем source IP
	deviceIP := sourceIP
	if ssdpResp.Location != "" {
		if ip, err := us.extractIPFromURL(ssdpResp.Location); err == nil {
			deviceIP = ip
		}
	}

	device := &models.Device{
		IP:           deviceIP,
		Protocols:    []models.Protocol{},
		DiscoveredAt: time.Now(),
	}

	// Добавляем UPnP протокол
	upnpProtocol := models.Protocol{
		Type:       "UPnP",
		Port:       80, // UPnP обычно на порту 80
		URL:        ssdpResp.Location,
		Available:  true,
		DetectedAt: time.Now(),
	}

	// Определяем порт из Location
	if ssdpResp.Location != "" {
		if port := us.extractPortFromURL(ssdpResp.Location); port > 0 {
			upnpProtocol.Port = port
		}
	}

	device.Protocols = append(device.Protocols, upnpProtocol)

	// Парсим Server для получения информации о производителе
	if ssdpResp.Server != "" {
		us.parseServerHeader(device, ssdpResp.Server)
	}

	// Парсим USN для получения дополнительной информации
	if ssdpResp.USN != "" {
		us.parseUSN(device, ssdpResp.USN)
	}

	// Парсим ST для определения типа устройства
	if ssdpResp.ST != "" {
		us.parseSearchTarget(device, ssdpResp.ST)
	}

	us.logger.Debugf("Found UPnP device: %s at %s", deviceIP, ssdpResp.Location)
	return device, nil
}

// parseSSDPHeaders парсит заголовки SSDP ответа
func (us *UPnPScanner) parseSSDPHeaders(response string) SSDPResponse {
	ssdpResp := SSDPResponse{}
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Парсим заголовки в формате "Header: Value"
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		header := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch header {
		case "cache-control":
			ssdpResp.CacheControl = value
		case "location":
			ssdpResp.Location = value
		case "server":
			ssdpResp.Server = value
		case "st", "search-target":
			ssdpResp.ST = value
		case "usn":
			ssdpResp.USN = value
		case "ext":
			ssdpResp.EXT = value
		case "date":
			ssdpResp.Date = value
		}
	}

	return ssdpResp
}

// extractIPFromURL извлекает IP адрес из URL
func (us *UPnPScanner) extractIPFromURL(url string) (string, error) {
	// Убираем протокол
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	
	// Убираем путь
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid URL")
	}

	// Извлекаем host:port
	hostPort := parts[0]
	host := strings.Split(hostPort, ":")[0]

	// Проверяем, что это валидный IP
	ip := net.ParseIP(host)
	if ip == nil {
		return "", fmt.Errorf("not a valid IP address")
	}

	return ip.String(), nil
}

// extractPortFromURL извлекает порт из URL
func (us *UPnPScanner) extractPortFromURL(url string) int {
	// Убираем протокол
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	
	// Убираем путь
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return 0
	}

	// Извлекаем host:port
	hostPort := parts[0]
	portParts := strings.Split(hostPort, ":")
	if len(portParts) == 2 {
		var port int
		fmt.Sscanf(portParts[1], "%d", &port)
		return port
	}

	// Порт по умолчанию для HTTP
	if strings.HasPrefix(url, "https://") {
		return 443
	}
	return 80
}

// parseServerHeader парсит Server заголовок для получения информации о производителе
func (us *UPnPScanner) parseServerHeader(device *models.Device, server string) {
	// Формат Server заголовка: "OS/version UPnP/version product/version"
	// Например: "Linux/3.14 UPnP/1.0 DLNADOC/1.50"
	parts := strings.Fields(server)
	
	for _, part := range parts {
		// Ищем информацию о производителе/модели
		if strings.Contains(part, "/") {
			productParts := strings.Split(part, "/")
			if len(productParts) >= 1 {
				productName := productParts[0]
				// Попытка определить производителя по известным маркам
				knownManufacturers := []string{"Samsung", "LG", "Sony", "Panasonic", "TP-Link", "D-Link", "Netgear"}
				for _, mfg := range knownManufacturers {
					if strings.Contains(strings.ToLower(productName), strings.ToLower(mfg)) {
						device.Manufacturer = mfg
						break
					}
				}
			}
		}
	}
}

// parseUSN парсит USN (Unique Service Name) для получения дополнительной информации
func (us *UPnPScanner) parseUSN(device *models.Device, usn string) {
	// USN формат: "uuid:device-UUID::urn:schemas-upnp-org:device:..."
	// Можем извлечь UUID и тип устройства
	parts := strings.Split(usn, "::")
	if len(parts) > 0 {
		uuidPart := parts[0]
		if strings.HasPrefix(uuidPart, "uuid:") {
			// UUID устройства
		}
	}
	
	// Парсим тип устройства из USN
	if strings.Contains(usn, "urn:schemas-upnp-org:device:") {
		deviceParts := strings.Split(usn, "urn:schemas-upnp-org:device:")
		if len(deviceParts) > 1 {
			deviceType := strings.Split(deviceParts[1], ":")[0]
			// deviceType может быть MediaServer, MediaRenderer и т.д.
		}
	}
}

// parseSearchTarget парсит ST (Search Target) для определения типа устройства
func (us *UPnPScanner) parseSearchTarget(device *models.Device, st string) {
	// ST может содержать информацию о типе устройства
	// Например: "urn:schemas-upnp-org:device:MediaServer:1"
	if strings.Contains(st, "MediaServer") {
		// Это медиа-сервер
	} else if strings.Contains(st, "MediaRenderer") {
		// Это медиа-рендерер
	} else if strings.Contains(st, "InternetGatewayDevice") {
		// Это интернет-шлюз
	}
}

// mergeDeviceInfo объединяет информацию об устройстве
func (us *UPnPScanner) mergeDeviceInfo(existing, new *models.Device) {
	// Объединяем протоколы
	protocolMap := make(map[string]models.Protocol)
	for _, p := range existing.Protocols {
		protocolMap[p.Type] = p
	}
	for _, p := range new.Protocols {
		if _, exists := protocolMap[p.Type]; !exists {
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
}

// devicesToSlice преобразует map устройств в slice
func (us *UPnPScanner) devicesToSlice(devices map[string]*models.Device) []*models.Device {
	result := make([]*models.Device, 0, len(devices))
	for _, device := range devices {
		result = append(result, device)
	}
	return result
}
