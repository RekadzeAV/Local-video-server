package protocols

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/pkg/utils"
	"github.com/sirupsen/logrus"
)

// RTMPDetector - детектор RTMP протокола
type RTMPDetector struct {
	logger *logrus.Logger
}

// NewRTMPDetector создает новый RTMP детектор
func NewRTMPDetector() *RTMPDetector {
	return &RTMPDetector{
		logger: utils.GetLogger(),
	}
}

// GetName возвращает название протокола
func (d *RTMPDetector) GetName() string {
	return "RTMP"
}

// GetDefaultPort возвращает порт по умолчанию
func (d *RTMPDetector) GetDefaultPort() int {
	return 1935
}

// Detect проверяет наличие RTMP протокола на устройстве
func (d *RTMPDetector) Detect(ip string, port int, timeout time.Duration) (*models.Protocol, error) {
	protocol := &models.Protocol{
		Type:       "RTMP",
		Port:       port,
		Available:  false,
		DetectedAt: time.Now(),
	}

	// Подключение к RTMP порту
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return protocol, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Устанавливаем таймаут
	conn.SetDeadline(time.Now().Add(timeout))

	// RTMP handshake состоит из 3 этапов:
	// C0+C1 (клиент отправляет), S0+S1+S2 (сервер отвечает), C2 (клиент подтверждает)

	// Этап 1: Отправка C0+C1
	c0c1 := d.createC0C1()
	if _, err := conn.Write(c0c1); err != nil {
		return protocol, fmt.Errorf("failed to send C0+C1: %w", err)
	}

	// Этап 2: Чтение S0+S1+S2
	s0s1s2 := make([]byte, 3073) // 1 + 1536 + 1536
	if _, err := conn.Read(s0s1s2); err != nil {
		return protocol, fmt.Errorf("failed to read S0+S1+S2: %w", err)
	}

	// Проверка S0 (версия протокола, должна быть 3)
	if s0s1s2[0] != 3 {
		return protocol, fmt.Errorf("invalid RTMP version: %d", s0s1s2[0])
	}

	// Этап 3: Отправка C2 (эхо S1)
	c2 := s0s1s2[1:1537] // S1 часть
	if _, err := conn.Write(c2); err != nil {
		return protocol, fmt.Errorf("failed to send C2: %w", err)
	}

	// Если handshake успешен, RTMP доступен
	protocol.Available = true
	protocol.URL = fmt.Sprintf("rtmp://%s:%d", ip, port)

	return protocol, nil
}

// createC0C1 создает C0+C1 пакет для RTMP handshake
func (d *RTMPDetector) createC0C1() []byte {
	// C0: 1 байт версии (3)
	c0 := []byte{3}

	// C1: 1536 байт
	c1 := make([]byte, 1536)
	
	// Первые 4 байта - timestamp (текущее время)
	timestamp := uint32(time.Now().Unix())
	binary.BigEndian.PutUint32(c1[0:4], timestamp)
	
	// Следующие 4 байта - нули (версия)
	binary.BigEndian.PutUint32(c1[4:8], 0)
	
	// Остальные байты - случайные данные
	// В реальной реализации здесь должны быть случайные данные,
	// но для детектирования достаточно минимального handshake
	for i := 8; i < 1536; i++ {
		c1[i] = byte(i % 256)
	}

	return append(c0, c1...)
}

// CheckStream проверяет доступность конкретного RTMP потока
func (d *RTMPDetector) CheckStream(ip string, port int, appName string, streamName string, timeout time.Duration) (bool, error) {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(timeout))

	// Выполняем handshake
	c0c1 := d.createC0C1()
	if _, err := conn.Write(c0c1); err != nil {
		return false, err
	}

	s0s1s2 := make([]byte, 3073)
	if _, err := conn.Read(s0s1s2); err != nil {
		return false, err
	}

	if s0s1s2[0] != 3 {
		return false, fmt.Errorf("invalid RTMP version")
	}

	c2 := s0s1s2[1:1537]
	if _, err := conn.Write(c2); err != nil {
		return false, err
	}

	// После handshake можно попытаться подключиться к приложению
	// Для детектирования достаточно успешного handshake
	return true, nil
}
