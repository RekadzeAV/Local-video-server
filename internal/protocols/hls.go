package protocols

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/pkg/utils"
	"github.com/sirupsen/logrus"
)

// HLSDetector - детектор HLS протокола
type HLSDetector struct {
	logger     *logrus.Logger
	httpClient *http.Client
}

// NewHLSDetector создает новый HLS детектор
func NewHLSDetector() *HLSDetector {
	return &HLSDetector{
		logger: utils.GetLogger(),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetName возвращает название протокола
func (d *HLSDetector) GetName() string {
	return "HLS"
}

// GetDefaultPort возвращает порт по умолчанию
func (d *HLSDetector) GetDefaultPort() int {
	return 80
}

// Detect проверяет наличие HLS протокола на устройстве
func (d *HLSDetector) Detect(ip string, port int, timeout time.Duration) (*models.Protocol, error) {
	protocol := &models.Protocol{
		Type:       "HLS",
		Port:       port,
		Available:  false,
		DetectedAt: time.Now(),
	}

	// Создаем HTTP клиент с таймаутом
	client := &http.Client{
		Timeout: timeout,
	}

	// Стандартные пути для HLS манифестов
	hlsPaths := []string{
		"/hls/stream.m3u8",
		"/live/stream.m3u8",
		"/stream.m3u8",
		"/index.m3u8",
		"/playlist.m3u8",
		"/video.m3u8",
		"/live.m3u8",
	}

	// Пробуем HTTP и HTTPS
	schemes := []string{"http", "https"}

	for _, scheme := range schemes {
		for _, path := range hlsPaths {
			url := fmt.Sprintf("%s://%s:%d%s", scheme, ip, port, path)
			
			if d.checkHLSManifest(client, url) {
				protocol.Available = true
				protocol.URL = url
				return protocol, nil
			}
		}
	}

	// Также проверяем веб-интерфейс на наличие ссылок на .m3u8 файлы
	if d.checkWebInterfaceForHLS(client, ip, port) {
		protocol.Available = true
		protocol.URL = fmt.Sprintf("http://%s:%d", ip, port)
		return protocol, nil
	}

	return protocol, fmt.Errorf("HLS not found")
}

// checkHLSManifest проверяет наличие валидного HLS манифеста
func (d *HLSDetector) checkHLSManifest(client *http.Client, url string) bool {
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	// Проверяем Content-Type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/vnd.apple.mpegurl") &&
		!strings.Contains(contentType, "application/x-mpegURL") &&
		!strings.Contains(contentType, "text/plain") {
		return false
	}

	// Читаем первые байты для проверки формата
	body := make([]byte, 1024)
	n, err := resp.Body.Read(body)
	if err != nil && err != io.EOF {
		return false
	}

	content := string(body[:n])
	
	// Проверяем наличие ключевых слов HLS манифеста
	return strings.Contains(content, "#EXTM3U") || 
		   strings.Contains(content, "#EXT-X-VERSION") ||
		   strings.Contains(content, "#EXTINF")
}

// checkWebInterfaceForHLS проверяет веб-интерфейс на наличие ссылок на HLS
func (d *HLSDetector) checkWebInterfaceForHLS(client *http.Client, ip string, port int) bool {
	schemes := []string{"http", "https"}
	paths := []string{"", "/", "/index.html", "/live.html", "/stream.html"}

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

			body := make([]byte, 8192)
			n, err := resp.Body.Read(body)
			if err != nil && err != io.EOF {
				continue
			}

			content := string(body[:n])
			
			// Ищем ссылки на .m3u8 файлы
			if strings.Contains(content, ".m3u8") {
				return true
			}
		}
	}

	return false
}
