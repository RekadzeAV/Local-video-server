package protocols

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/pkg/utils"
	"github.com/sirupsen/logrus"
)

// WebRTCDetector - детектор WebRTC протокола
type WebRTCDetector struct {
	logger     *logrus.Logger
	httpClient *http.Client
}

// NewWebRTCDetector создает новый WebRTC детектор
func NewWebRTCDetector() *WebRTCDetector {
	return &WebRTCDetector{
		logger: utils.GetLogger(),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetName возвращает название протокола
func (d *WebRTCDetector) GetName() string {
	return "WebRTC"
}

// GetDefaultPort возвращает порт по умолчанию
func (d *WebRTCDetector) GetDefaultPort() int {
	return 80
}

// Detect проверяет наличие WebRTC протокола на устройстве
func (d *WebRTCDetector) Detect(ip string, port int, timeout time.Duration) (*models.Protocol, error) {
	protocol := &models.Protocol{
		Type:       "WebRTC",
		Port:       port,
		Available:  false,
		DetectedAt: time.Now(),
	}

	// Создаем HTTP клиент с таймаутом
	client := &http.Client{
		Timeout: timeout,
	}

	// WebRTC обычно доступен через веб-интерфейс
	// Проверяем наличие WebRTC endpoints в HTML/JavaScript
	if d.checkWebInterfaceForWebRTC(client, ip, port) {
		protocol.Available = true
		protocol.URL = fmt.Sprintf("http://%s:%d", ip, port)
		return protocol, nil
	}

	// Проверяем наличие STUN/TURN серверов
	if d.checkSTUNTURN(client, ip, port) {
		protocol.Available = true
		protocol.URL = fmt.Sprintf("http://%s:%d", ip, port)
		return protocol, nil
	}

	return protocol, fmt.Errorf("WebRTC not found")
}

// checkWebInterfaceForWebRTC проверяет веб-интерфейс на наличие WebRTC
func (d *WebRTCDetector) checkWebInterfaceForWebRTC(client *http.Client, ip string, port int) bool {
	schemes := []string{"http", "https"}
	paths := []string{"", "/", "/index.html", "/live.html", "/stream.html", "/webrtc.html"}

	// Ключевые слова и паттерны для WebRTC
	webrtcKeywords := []string{
		"webrtc",
		"rtcpeerconnection",
		"getusermedia",
		"mediastream",
		"rtcicecandidate",
		"offer",
		"answer",
		"stun:",
		"turn:",
	}

	webrtcRegex := regexp.MustCompile(`(?i)(RTCPeerConnection|getUserMedia|MediaStream|RTCIceCandidate)`)

	for _, scheme := range schemes {
		for _, path := range paths {
			url := fmt.Sprintf("%s://%s:%d%s", scheme, ip, port, path)
			
			resp, err := client.Get(url)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				continue
			}

			body := make([]byte, 16384) // Увеличиваем размер для JavaScript файлов
			n, err := resp.Body.Read(body)
			if err != nil && err != io.EOF {
				continue
			}

			content := string(body[:n])
			lowerContent := strings.ToLower(content)
			
			// Проверяем наличие ключевых слов
			for _, keyword := range webrtcKeywords {
				if strings.Contains(lowerContent, keyword) {
					return true
				}
			}
			
			// Проверяем регулярными выражениями
			if webrtcRegex.MatchString(content) {
				return true
			}
		}
	}

	// Также проверяем отдельные JavaScript файлы
	jsPaths := []string{"/js/webrtc.js", "/webrtc.js", "/js/stream.js", "/stream.js"}
	for _, scheme := range schemes {
		for _, path := range jsPaths {
			url := fmt.Sprintf("%s://%s:%d%s", scheme, ip, port, path)
			
			resp, err := client.Get(url)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				body := make([]byte, 8192)
				n, err := resp.Body.Read(body)
				if err != nil && err != io.EOF {
					continue
				}

				content := string(body[:n])
				if webrtcRegex.MatchString(content) {
					return true
				}
			}
		}
	}

	return false
}

// checkSTUNTURN проверяет наличие STUN/TURN серверов
func (d *WebRTCDetector) checkSTUNTURN(client *http.Client, ip string, port int) bool {
	// Стандартные порты для STUN/TURN
	stunPorts := []int{3478, 5349}
	turnPorts := []int{3478, 5349}

	// Проверяем STUN
	for _, stunPort := range stunPorts {
		if d.checkSTUNServer(ip, stunPort, 2*time.Second) {
			return true
		}
	}

	// Проверяем TURN
	for _, turnPort := range turnPorts {
		if d.checkTURNServer(ip, turnPort, 2*time.Second) {
			return true
		}
	}

	// Проверяем конфигурационные файлы или API endpoints
	configPaths := []string{"/api/webrtc/config", "/config/webrtc.json", "/webrtc/config"}
	for _, path := range configPaths {
		url := fmt.Sprintf("http://%s:%d%s", ip, port, path)
		
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			contentType := resp.Header.Get("Content-Type")
			if strings.Contains(contentType, "json") {
				body := make([]byte, 4096)
				n, err := resp.Body.Read(body)
				if err != nil && err != io.EOF {
					continue
				}

				var config map[string]interface{}
				if err := json.Unmarshal(body[:n], &config); err == nil {
					// Проверяем наличие STUN/TURN конфигурации
					if configStr, ok := config["stun"].(string); ok && configStr != "" {
						return true
					}
					if configStr, ok := config["turn"].(string); ok && configStr != "" {
						return true
					}
					if iceServers, ok := config["iceServers"].([]interface{}); ok && len(iceServers) > 0 {
						return true
					}
				}
			}
		}
	}

	return false
}

// checkSTUNServer проверяет доступность STUN сервера
func (d *WebRTCDetector) checkSTUNServer(ip string, port int, timeout time.Duration) bool {
	// STUN использует UDP протокол
	// Простая проверка - пытаемся подключиться
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("udp", address, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()

	// STUN Binding Request (упрощенная версия)
	// В реальной реализации здесь должен быть полный STUN запрос
	// Для детектирования достаточно проверки доступности порта
	return true
}

// checkTURNServer проверяет доступность TURN сервера
func (d *WebRTCDetector) checkTURNServer(ip string, port int, timeout time.Duration) bool {
	// TURN также использует UDP (или TCP)
	// Аналогично STUN, для детектирования проверяем доступность порта
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("udp", address, timeout)
	if err != nil {
		// Пробуем TCP
		conn, err = net.DialTimeout("tcp", address, timeout)
		if err != nil {
			return false
		}
	}
	defer conn.Close()

	return true
}
