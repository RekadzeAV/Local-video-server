package scanner

import (
	"context"
	"encoding/xml"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/pkg/utils"
	"github.com/sirupsen/logrus"
)

// ONVIFScanner выполняет обнаружение устройств через ONVIF WS-Discovery
type ONVIFScanner struct {
	config *models.ScanConfig
	logger *logrus.Logger
}

// NewONVIFScanner создает новый экземпляр ONVIFScanner
func NewONVIFScanner(config *models.ScanConfig) *ONVIFScanner {
	return &ONVIFScanner{
		config: config,
		logger: utils.GetLogger(),
	}
}

// WS-Discovery константы
const (
	WSDiscoveryMulticastIPv4 = "239.255.255.250"
	WSDiscoveryPort          = 3702
	WSDiscoveryNamespace     = "http://schemas.xmlsoap.org/ws/2005/04/discovery"
	ONVIFNamespace           = "http://www.onvif.org/ver10/network/wsdl"
)

// ProbeMessage представляет WS-Discovery Probe сообщение
type ProbeMessage struct {
	XMLName xml.Name `xml:"Envelope"`
	Xmlns   string   `xml:"xmlns:a,attr"`
	Xmlnsd  string   `xml:"xmlns:d,attr"`
	Header  struct {
		MessageID string `xml:"a:MessageID"`
		To        string `xml:"a:To"`
		Action    string `xml:"a:Action"`
	} `xml:"Header"`
	Body struct {
		Probe struct {
			Types string `xml:"d:Types"`
		} `xml:"Probe"`
	} `xml:"Body"`
}

// ProbeMatchMessage представляет WS-Discovery ProbeMatch ответ
type ProbeMatchMessage struct {
	XMLName xml.Name `xml:"Envelope"`
	Header  struct {
		RelatesTo string `xml:"RelatesTo"`
		To        string `xml:"To"`
		Action    string `xml:"Action"`
	} `xml:"Header"`
	Body struct {
		ProbeMatches struct {
			ProbeMatch []struct {
				EndpointReference struct {
					Address string `xml:"Address"`
				} `xml:"EndpointReference"`
				Types            string `xml:"Types"`
				Scopes           string `xml:"Scopes"`
				XAddrs           string `xml:"XAddrs"`
				MetadataVersion  int    `xml:"MetadataVersion"`
			} `xml:"ProbeMatch"`
		} `xml:"ProbeMatches"`
	} `xml:"Body"`
}

// Discover выполняет ONVIF WS-Discovery для обнаружения устройств
func (os *ONVIFScanner) Discover(ctx context.Context) ([]*models.Device, error) {
	os.logger.Infof("Starting ONVIF WS-Discovery")

	// Создаем UDP соединение для отправки Probe сообщений
	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP socket: %w", err)
	}
	defer conn.Close()

	// Устанавливаем таймаут для чтения
	timeout := os.config.DiscoveryTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	conn.SetReadDeadline(time.Now().Add(timeout))

	// Отправляем Probe сообщение
	probeMsg := os.buildProbeMessage()
	multicastAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", WSDiscoveryMulticastIPv4, WSDiscoveryPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve multicast address: %w", err)
	}

	_, err = conn.WriteToUDP([]byte(probeMsg), multicastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to send Probe message: %w", err)
	}

	os.logger.Debugf("Sent WS-Discovery Probe message")

	// Слушаем ответы
	devices := make(map[string]*models.Device)
	buffer := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			os.logger.Debugf("ONVIF discovery cancelled")
			return os.devicesToSlice(devices), nil
		default:
			// Читаем ответы с таймаутом
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, addr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				// Проверяем, это таймаут или реальная ошибка
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Таймаут - проверяем, не истекло ли общее время
					if time.Now().After(time.Now().Add(timeout)) {
						return os.devicesToSlice(devices), nil
					}
					continue
				}
				os.logger.Debugf("Error reading UDP response: %v", err)
				continue
			}

			// Парсим ProbeMatch ответ
			device, err := os.parseProbeMatch(buffer[:n], addr.IP.String())
			if err != nil {
				os.logger.Debugf("Failed to parse ProbeMatch: %v", err)
				continue
			}

			if device != nil {
				// Объединяем информацию, если устройство уже найдено
				if existing, exists := devices[device.IP]; exists {
					os.mergeDeviceInfo(existing, device)
				} else {
					devices[device.IP] = device
				}
			}
		}
	}
}

// buildProbeMessage создает WS-Discovery Probe сообщение
func (os *ONVIFScanner) buildProbeMessage() string {
	messageID := fmt.Sprintf("uuid:%d", time.Now().UnixNano())
	
	probe := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" 
            xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing"
            xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery">
    <s:Header>
        <a:Action s:mustUnderstand="1">http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</a:Action>
        <a:MessageID>urn:uuid:%s</a:MessageID>
        <a:To s:mustUnderstand="1">urn:schemas-xmlsoap-org:ws:2005:04:discovery</a:To>
    </s:Header>
    <s:Body>
        <d:Probe>
            <d:Types>dn:NetworkVideoTransmitter</d:Types>
        </d:Probe>
    </s:Body>
</s:Envelope>`, messageID)

	return probe
}

// parseProbeMatch парсит ProbeMatch ответ и извлекает информацию об устройстве
func (os *ONVIFScanner) parseProbeMatch(data []byte, sourceIP string) (*models.Device, error) {
	var envelope ProbeMatchMessage
	
	// Пробуем распарсить XML
	err := xml.Unmarshal(data, &envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	// Проверяем, что это ProbeMatch ответ
	if !strings.Contains(envelope.Header.Action, "ProbeMatches") {
		return nil, fmt.Errorf("not a ProbeMatch message")
	}

	if len(envelope.Body.ProbeMatches.ProbeMatch) == 0 {
		return nil, fmt.Errorf("no ProbeMatch entries")
	}

	probeMatch := envelope.Body.ProbeMatches.ProbeMatch[0]
	
	// Извлекаем XAddrs (адреса устройств)
	xaddrs := strings.TrimSpace(probeMatch.XAddrs)
	if xaddrs == "" {
		return nil, fmt.Errorf("no XAddrs in ProbeMatch")
	}

	// Парсим первый адрес из XAddrs
	addresses := strings.Split(xaddrs, " ")
	if len(addresses) == 0 {
		return nil, fmt.Errorf("empty XAddrs")
	}

	// Извлекаем IP из URL
	deviceURL := addresses[0]
	deviceIP, err := os.extractIPFromURL(deviceURL)
	if err != nil {
		// Используем source IP как fallback
		deviceIP = sourceIP
	}

	device := &models.Device{
		IP:           deviceIP,
		Protocols:    []models.Protocol{},
		DiscoveredAt: time.Now(),
	}

	// Добавляем ONVIF протокол
	onvifProtocol := models.Protocol{
		Type:       "ONVIF",
		Port:       80, // ONVIF обычно на порту 80 или 8080
		URL:        deviceURL,
		Available:  true,
		DetectedAt: time.Now(),
	}

	// Определяем порт из URL
	if port := os.extractPortFromURL(deviceURL); port > 0 {
		onvifProtocol.Port = port
	}

	device.Protocols = append(device.Protocols, onvifProtocol)

	// Парсим Scopes для получения дополнительной информации
	scopes := strings.TrimSpace(probeMatch.Scopes)
	if scopes != "" {
		os.parseScopes(device, scopes)
	}

	// Парсим Types
	types := strings.TrimSpace(probeMatch.Types)
	if types != "" {
		os.parseTypes(device, types)
	}

	os.logger.Debugf("Found ONVIF device: %s at %s", deviceIP, deviceURL)
	return device, nil
}

// extractIPFromURL извлекает IP адрес из URL
func (os *ONVIFScanner) extractIPFromURL(url string) (string, error) {
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
func (os *ONVIFScanner) extractPortFromURL(url string) int {
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

// parseScopes парсит Scopes и извлекает информацию о производителе и модели
func (os *ONVIFScanner) parseScopes(device *models.Device, scopes string) {
	scopeList := strings.Split(scopes, " ")
	
	for _, scope := range scopeList {
		scope = strings.TrimSpace(scope)
		
		// Парсим формат onvif://www.onvif.org/name/...
		if strings.HasPrefix(scope, "onvif://www.onvif.org/name/") {
			name := strings.TrimPrefix(scope, "onvif://www.onvif.org/name/")
			parts := strings.Split(name, "/")
			if len(parts) >= 1 {
				device.Model = parts[0]
			}
		}
		
		// Парсим формат onvif://www.onvif.org/hardware/...
		if strings.HasPrefix(scope, "onvif://www.onvif.org/hardware/") {
			hardware := strings.TrimPrefix(scope, "onvif://www.onvif.org/hardware/")
			parts := strings.Split(hardware, "/")
			if len(parts) >= 1 {
				device.Manufacturer = parts[0]
			}
		}
	}
}

// parseTypes парсит Types для получения дополнительной информации
func (os *ONVIFScanner) parseTypes(device *models.Device, types string) {
	// Types обычно содержат информацию о типах устройств
	// Например: "dn:NetworkVideoTransmitter"
	if strings.Contains(types, "NetworkVideoTransmitter") {
		// Это сетевая камера
	}
}

// mergeDeviceInfo объединяет информацию об устройстве
func (os *ONVIFScanner) mergeDeviceInfo(existing, new *models.Device) {
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
func (os *ONVIFScanner) devicesToSlice(devices map[string]*models.Device) []*models.Device {
	result := make([]*models.Device, 0, len(devices))
	for _, device := range devices {
		result = append(result, device)
	}
	return result
}
