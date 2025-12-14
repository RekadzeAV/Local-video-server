package utils

import (
	"fmt"
	"net"
	"strings"
)

// NetworkInterface представляет сетевой интерфейс
type NetworkInterface struct {
	Name      string
	IP        net.IP
	Mask      net.IPMask
	Subnet    *net.IPNet
	Broadcast net.IP
}

// GetLocalIP возвращает локальный IP адрес
func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// GetNetworkInterfaces возвращает список активных сетевых интерфейсов
func GetNetworkInterfaces() ([]NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var result []NetworkInterface
	for _, iface := range interfaces {
		// Пропускаем неактивные интерфейсы
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Пропускаем loopback интерфейсы
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			// Пропускаем IPv6 адреса
			if ipNet.IP.To4() == nil {
				continue
			}

			// Вычисляем broadcast адрес
			broadcast := make(net.IP, 4)
			for i := 0; i < 4; i++ {
				broadcast[i] = ipNet.IP[i] | ^ipNet.Mask[i]
			}

			result = append(result, NetworkInterface{
				Name:      iface.Name,
				IP:        ipNet.IP,
				Mask:      ipNet.Mask,
				Subnet:    ipNet,
				Broadcast: broadcast,
			})
		}
	}

	return result, nil
}

// GetDefaultSubnet возвращает подсеть по умолчанию (первая активная)
func GetDefaultSubnet() (string, error) {
	interfaces, err := GetNetworkInterfaces()
	if err != nil {
		return "", err
	}

	if len(interfaces) == 0 {
		return "", fmt.Errorf("no active network interfaces found")
	}

	// Возвращаем первую найденную подсеть
	subnet := interfaces[0].Subnet
	return subnet.String(), nil
}

// ParseSubnet парсит строку подсети (например, "192.168.1.0/24")
func ParseSubnet(subnet string) (*net.IPNet, error) {
	// Если подсеть не указана, пытаемся определить автоматически
	if subnet == "" {
		defaultSubnet, err := GetDefaultSubnet()
		if err != nil {
			return nil, err
		}
		subnet = defaultSubnet
	}

	// Парсим CIDR нотацию
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		// Если не CIDR, пытаемся добавить /24
		if strings.Contains(subnet, "/") {
			return nil, err
		}
		subnet = subnet + "/24"
		_, ipNet, err = net.ParseCIDR(subnet)
		if err != nil {
			return nil, err
		}
	}

	return ipNet, nil
}

// GetSubnetHosts возвращает список всех хостов в подсети
func GetSubnetHosts(subnet string) ([]string, error) {
	ipNet, err := ParseSubnet(subnet)
	if err != nil {
		return nil, err
	}

	var hosts []string
	ip := ipNet.IP.Mask(ipNet.Mask)
	broadcast := make(net.IP, len(ip))
	for i := range ip {
		broadcast[i] = ip[i] | ^ipNet.Mask[i]
	}

	// Генерируем все IP адреса в подсети
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		// Пропускаем network и broadcast адреса
		if !ip.Equal(ipNet.IP) && !ip.Equal(broadcast) {
			hosts = append(hosts, ip.String())
		}
	}

	return hosts, nil
}

// incrementIP увеличивает IP адрес на 1
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// IsIPInSubnet проверяет, находится ли IP адрес в указанной подсети
func IsIPInSubnet(ipStr, subnet string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	ipNet, err := ParseSubnet(subnet)
	if err != nil {
		return false, err
	}

	return ipNet.Contains(ip), nil
}

// GetSubnetMask возвращает маску подсети в виде строки (например, "255.255.255.0")
func GetSubnetMask(subnet string) (string, error) {
	ipNet, err := ParseSubnet(subnet)
	if err != nil {
		return "", err
	}

	mask := ipNet.Mask
	return fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]), nil
}

// GetNetworkAddress возвращает сетевой адрес подсети
func GetNetworkAddress(subnet string) (string, error) {
	ipNet, err := ParseSubnet(subnet)
	if err != nil {
		return "", err
	}

	return ipNet.IP.String(), nil
}

// GetBroadcastAddress возвращает broadcast адрес подсети
func GetBroadcastAddress(subnet string) (string, error) {
	ipNet, err := ParseSubnet(subnet)
	if err != nil {
		return "", err
	}

	broadcast := make(net.IP, len(ipNet.IP))
	for i := range ipNet.IP {
		broadcast[i] = ipNet.IP[i] | ^ipNet.Mask[i]
	}

	return broadcast.String(), nil
}

// ValidateIP проверяет корректность IP адреса
func ValidateIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// ValidateSubnet проверяет корректность подсети в CIDR нотации
func ValidateSubnet(subnet string) bool {
	_, _, err := net.ParseCIDR(subnet)
	return err == nil
}

// ResolveHostname пытается разрешить hostname для IP адреса
func ResolveHostname(ip string) (string, error) {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return "", err
	}

	// Убираем точку в конце, если есть
	hostname := names[0]
	if strings.HasSuffix(hostname, ".") {
		hostname = hostname[:len(hostname)-1]
	}

	return hostname, nil
}

// GetInterfaceByName возвращает информацию о сетевом интерфейсе по имени
func GetInterfaceByName(name string) (*NetworkInterface, error) {
	interfaces, err := GetNetworkInterfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Name == name {
			return &iface, nil
		}
	}

	return nil, fmt.Errorf("interface %s not found", name)
}

// LogNetworkInfo логирует информацию о сетевых интерфейсах
func LogNetworkInfo() {
	logger := GetLogger()
	interfaces, err := GetNetworkInterfaces()
	if err != nil {
		logger.Errorf("Failed to get network interfaces: %v", err)
		return
	}

	logger.Infof("Found %d active network interface(s):", len(interfaces))
	for _, iface := range interfaces {
		logger.Infof("  - %s: %s (%s)", iface.Name, iface.IP.String(), iface.Subnet.String())
	}
}
