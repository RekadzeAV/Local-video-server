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

// DASHDetector - детектор MPEG-DASH протокола
type DASHDetector struct {
	logger     *logrus.Logger
	httpClient *http.Client
}

// NewDASHDetector создает новый DASH детектор
func NewDASHDetector() *DASHDetector {
	return &DASHDetector{
		logger: utils.GetLogger(),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetName возвращает название протокола
func (d *DASHDetector) GetName() string {
	return "MPEG-DASH"
}

// GetDefaultPort возвращает порт по умолчанию
func (d *DASHDetector) GetDefaultPort() int {
	return 80
}

// Detect проверяет наличие MPEG-DASH протокола на устройстве
func (d *DASHDetector) Detect(ip string, port int, timeout time.Duration) (*models.Protocol, error) {
	protocol := &models.Protocol{
		Type:       "MPEG-DASH",
		Port:       port,
		Available:  false,
		DetectedAt: time.Now(),
	}

	// Создаем HTTP клиент с таймаутом
	client := &http.Client{
		Timeout: timeout,
	}

	// Стандартные пути для DASH манифестов
	dashPaths := []string{
		"/dash/stream.mpd",
		"/stream.mpd",
		"/manifest.mpd",
		"/playlist.mpd",
		"/video.mpd",
		"/live.mpd",
		"/dash/manifest.mpd",
	}

	// Пробуем HTTP и HTTPS
	schemes := []string{"http", "https"}

	for _, scheme := range schemes {
		for _, path := range dashPaths {
			url := fmt.Sprintf("%s://%s:%d%s", scheme, ip, port, path)
			
			if d.checkDASHManifest(client, url) {
				protocol.Available = true
				protocol.URL = url
				return protocol, nil
			}
		}
	}

	// Также проверяем веб-интерфейс на наличие ссылок на .mpd файлы
	if d.checkWebInterfaceForDASH(client, ip, port) {
		protocol.Available = true
		protocol.URL = fmt.Sprintf("http://%s:%d", ip, port)
		return protocol, nil
	}

	return protocol, fmt.Errorf("MPEG-DASH not found")
}

// checkDASHManifest проверяет наличие валидного DASH манифеста
func (d *DASHDetector) checkDASHManifest(client *http.Client, url string) bool {
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
	if !strings.Contains(contentType, "application/dash+xml") &&
		!strings.Contains(contentType, "application/xml") &&
		!strings.Contains(contentType, "text/xml") {
		return false
	}

	// Читаем первые байты для проверки формата
	body := make([]byte, 2048)
	n, err := resp.Body.Read(body)
	if err != nil && err != io.EOF {
		return false
	}

	content := string(body[:n])
	
	// Проверяем наличие ключевых элементов DASH манифеста
	// DASH манифесты - это XML файлы с определенной структурой
	return strings.Contains(content, "<?xml") &&
		   (strings.Contains(content, "<MPD") ||
		    strings.Contains(content, "<MediaPresentationDescription") ||
		    strings.Contains(content, "type=\"dynamic\"") ||
		    strings.Contains(content, "type=\"static\""))
}

// checkWebInterfaceForDASH проверяет веб-интерфейс на наличие ссылок на DASH
func (d *DASHDetector) checkWebInterfaceForDASH(client *http.Client, ip string, port int) bool {
	schemes := []string{"http", "https"}
	paths := []string{"", "/", "/index.html", "/live.html", "/stream.html", "/dash.html"}

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
			
			// Ищем ссылки на .mpd файлы или упоминания DASH
			if strings.Contains(content, ".mpd") ||
			   strings.Contains(strings.ToLower(content), "dash") ||
			   strings.Contains(strings.ToLower(content), "mpeg-dash") {
				return true
			}
		}
	}

	return false
}
