package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/pkg/utils"
	"github.com/sirupsen/logrus"
)

// NetworkScanner выполняет сканирование сети для обнаружения устройств
type NetworkScanner struct {
	config     *models.ScanConfig
	logger     *logrus.Logger
	activeHosts map[string]bool
	mu         sync.RWMutex
}

// NewNetworkScanner создает новый экземпляр NetworkScanner
func NewNetworkScanner(config *models.ScanConfig) *NetworkScanner {
	return &NetworkScanner{
		config:      config,
		logger:      utils.GetLogger(),
		activeHosts: make(map[string]bool),
	}
}

// ScanNetwork выполняет полное сканирование сети
func (ns *NetworkScanner) ScanNetwork(ctx context.Context, subnet string) ([]*models.Device, error) {
	ns.logger.Infof("Starting network scan for subnet: %s", subnet)

	// 1. Получение списка активных хостов через ARP
	hosts, err := ns.getActiveHosts(ctx, subnet)
	if err != nil {
		ns.logger.Warnf("Failed to get active hosts via ARP: %v, falling back to port scan", err)
		// Fallback: получаем все хосты из подсети
		hosts, err = utils.GetSubnetHosts(subnet)
		if err != nil {
			return nil, fmt.Errorf("failed to get subnet hosts: %w", err)
		}
	}

	ns.logger.Infof("Found %d potential hosts to scan", len(hosts))

	// 2. Параллельное сканирование портов
	devices := ns.scanPortsParallel(ctx, hosts)

	ns.logger.Infof("Scan completed. Found %d devices", len(devices))
	return devices, nil
}

// getActiveHosts получает список активных хостов через ARP таблицу
func (ns *NetworkScanner) getActiveHosts(ctx context.Context, subnet string) ([]string, error) {
	// Получаем сетевые интерфейсы
	interfaces, err := utils.GetNetworkInterfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var allHosts []string
	hostsMap := make(map[string]bool)

	// Пробуем получить хосты из ARP таблицы для каждого интерфейса
	for _, iface := range interfaces {
		// Проверяем, что интерфейс в нужной подсети
		ipNet, err := utils.ParseSubnet(subnet)
		if err != nil {
			continue
		}
		if !ipNet.Contains(iface.IP) {
			continue
		}

		// Получаем хосты из ARP таблицы через pcap
		hosts, err := ns.getHostsFromARP(iface.Name, subnet)
		if err != nil {
			ns.logger.Debugf("Failed to get hosts from ARP for interface %s: %v", iface.Name, err)
			continue
		}

		for _, host := range hosts {
			if !hostsMap[host] {
				hostsMap[host] = true
				allHosts = append(allHosts, host)
			}
		}
	}

	// Если не получилось через ARP, возвращаем все хосты из подсети
	if len(allHosts) == 0 {
		allHosts, err = utils.GetSubnetHosts(subnet)
		if err != nil {
			return nil, err
		}
	}

	return allHosts, nil
}

// getHostsFromARP получает активные хосты из ARP таблицы через pcap
func (ns *NetworkScanner) getHostsFromARP(interfaceName, subnet string) ([]string, error) {
	// Парсим подсеть для проверки IP
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}

	// Открываем интерфейс для чтения пакетов
	handle, err := pcap.OpenLive(interfaceName, 1600, true, pcap.BlockForever)
	if err != nil {
		// На Windows может не работать без прав администратора
		return nil, fmt.Errorf("failed to open interface: %w", err)
	}
	defer handle.Close()

	// Читаем ARP пакеты в течение короткого времени
	hostsMap := make(map[string]bool)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	
	timeout := time.After(2 * time.Second)
	
	for {
		select {
		case <-timeout:
			// Преобразуем map в slice
			var hosts []string
			for host := range hostsMap {
				if ipNet.Contains(net.ParseIP(host)) {
					hosts = append(hosts, host)
				}
			}
			return hosts, nil
		case packet := <-packetSource.Packets():
			if packet == nil {
				continue
			}
			
			// Проверяем ARP слой
			arpLayer := packet.Layer(layers.LayerTypeARP)
			if arpLayer != nil {
				arp := arpLayer.(*layers.ARP)
				if arp.Operation == layers.ARPReply {
					srcIP := net.IP(arp.SourceProtAddress)
					if ipNet.Contains(srcIP) {
						hostsMap[srcIP.String()] = true
					}
				}
			}
		}
	}
}

// scanPortsParallel выполняет параллельное сканирование портов
func (ns *NetworkScanner) scanPortsParallel(ctx context.Context, hosts []string) []*models.Device {
	var devices []*models.Device
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Семафор для ограничения параллельности
	semaphore := make(chan struct{}, ns.config.MaxConcurrency)

	// Канал для результатов
	deviceChan := make(chan *models.Device, len(hosts))

	// Запускаем сканирование для каждого хоста
	for _, host := range hosts {
		select {
		case <-ctx.Done():
			ns.logger.Warnf("Scan cancelled")
			return devices
		default:
		}

		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			
			// Ограничиваем параллельность
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			device := ns.scanHost(ctx, ip)
			if device != nil {
				deviceChan <- device
			}
		}(host)
	}

	// Закрываем канал после завершения всех горутин
	go func() {
		wg.Wait()
		close(deviceChan)
	}()

	// Собираем результаты
	for device := range deviceChan {
		mu.Lock()
		devices = append(devices, device)
		mu.Unlock()
	}

	return devices
}

// scanHost сканирует один хост на наличие открытых портов
func (ns *NetworkScanner) scanHost(ctx context.Context, ip string) *models.Device {
	device := &models.Device{
		IP:           ip,
		Protocols:    []models.Protocol{},
		DiscoveredAt: time.Now(),
	}

	// Пытаемся получить hostname
	if hostname, err := utils.ResolveHostname(ip); err == nil {
		device.Hostname = hostname
	}

	// Сканируем порты
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, port := range ns.config.Ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()

			if ns.isPortOpen(ctx, ip, p) {
				protocol := ns.detectProtocol(ip, p)
				mu.Lock()
				device.Protocols = append(device.Protocols, protocol)
				mu.Unlock()
			}
		}(port)
	}

	wg.Wait()

	// Если не найдено ни одного протокола, возвращаем nil
	if len(device.Protocols) == 0 {
		return nil
	}

	return device
}

// isPortOpen проверяет, открыт ли порт на хосте
func (ns *NetworkScanner) isPortOpen(ctx context.Context, ip string, port int) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	
	// Создаем контекст с таймаутом
	timeout := ns.config.PortTimeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}

	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// detectProtocol определяет протокол по порту
func (ns *NetworkScanner) detectProtocol(ip string, port int) models.Protocol {
	protocol := models.Protocol{
		Port:      port,
		Available: true,
		DetectedAt: time.Now(),
	}

	switch port {
	case 554, 8554:
		protocol.Type = "RTSP"
		protocol.URL = fmt.Sprintf("rtsp://%s:%d", ip, port)
	case 1935:
		protocol.Type = "RTMP"
		protocol.URL = fmt.Sprintf("rtmp://%s:%d", ip, port)
	case 80, 8080:
		protocol.Type = "HTTP"
		protocol.URL = fmt.Sprintf("http://%s:%d", ip, port)
	default:
		protocol.Type = "UNKNOWN"
		protocol.URL = fmt.Sprintf("tcp://%s:%d", ip, port)
	}

	return protocol
}

// ScanPorts сканирует указанные порты на хосте
func (ns *NetworkScanner) ScanPorts(ctx context.Context, ip string, ports []int) []models.Protocol {
	var protocols []models.Protocol
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()

			if ns.isPortOpen(ctx, ip, p) {
				protocol := ns.detectProtocol(ip, p)
				mu.Lock()
				protocols = append(protocols, protocol)
				mu.Unlock()
			}
		}(port)
	}

	wg.Wait()
	return protocols
}
