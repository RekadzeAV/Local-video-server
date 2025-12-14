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

// MJPEGDetector - детектор MJPEG протокола
type MJPEGDetector struct {
	logger     *logrus.Logger
	httpClient *http.Client
}

// NewMJPEGDetector создает новый MJPEG детектор
func NewMJPEGDetector() *MJPEGDetector {
	return &MJPEGDetector{
		logger: utils.GetLogger(),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetName возвращает название протокола
func (d *MJPEGDetector) GetName() string {
	return "MJPEG"
}

// GetDefaultPort возвращает порт по умолчанию
func (d *MJPEGDetector) GetDefaultPort() int {
	return 80
}

// Detect проверяет наличие MJPEG протокола на устройстве
func (d *MJPEGDetector) Detect(ip string, port int, timeout time.Duration) (*models.Protocol, error) {
	protocol := &models.Protocol{
		Type:       "MJPEG",
		Port:       port,
		Available:  false,
		DetectedAt: time.Now(),
	}

	// Создаем HTTP клиент с таймаутом
	client := &http.Client{
		Timeout: timeout,
	}

	// Стандартные пути для MJPEG потоков
	mjpegPaths := []string{
		"/mjpeg",
		"/mjpg",
		"/video",
		"/stream",
		"/cam",
		"/camera",
		"/live",
		"/img/video.mjpeg",
		"/axis-cgi/mjpg/video.cgi",
		"/cgi-bin/mjpeg",
		"/snapshot.cgi",
	}

	// Пробуем HTTP и HTTPS
	schemes := []string{"http", "https"}

	for _, scheme := range schemes {
		for _, path := range mjpegPaths {
			url := fmt.Sprintf("%s://%s:%d%s", scheme, ip, port, path)
			
			if d.checkMJPEGStream(client, url) {
				protocol.Available = true
				protocol.URL = url
				return protocol, nil
			}
		}
	}

	// Также проверяем веб-интерфейс на наличие MJPEG ссылок
	if d.checkWebInterfaceForMJPEG(client, ip, port) {
		protocol.Available = true
		protocol.URL = fmt.Sprintf("http://%s:%d", ip, port)
		return protocol, nil
	}

	return protocol, fmt.Errorf("MJPEG not found")
}

// checkMJPEGStream проверяет наличие валидного MJPEG потока
func (d *MJPEGDetector) checkMJPEGStream(client *http.Client, url string) bool {
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
	
	// MJPEG потоки обычно имеют один из этих типов:
	// multipart/x-mixed-replace; boundary=...
	// image/jpeg (для snapshot)
	// video/x-motion-jpeg
	if strings.Contains(contentType, "multipart/x-mixed-replace") ||
		strings.Contains(contentType, "image/jpeg") ||
		strings.Contains(contentType, "video/x-motion-jpeg") {
		
		// Для multipart/x-mixed-replace читаем немного данных
		if strings.Contains(contentType, "multipart") {
			body := make([]byte, 512)
			n, err := resp.Body.Read(body)
			if err != nil && err != io.EOF {
				return false
			}
			
			// Проверяем наличие JPEG маркеров (FF D8 FF)
			if n >= 3 && body[0] == 0xFF && body[1] == 0xD8 && body[2] == 0xFF {
				return true
			}
		} else {
			// Для обычных JPEG изображений проверяем заголовок
			body := make([]byte, 4)
			n, err := resp.Body.Read(body)
			if err != nil && err != io.EOF {
				return false
			}
			
			if n >= 2 && body[0] == 0xFF && body[1] == 0xD8 {
				return true
			}
		}
	}

	return false
}

// checkWebInterfaceForMJPEG проверяет веб-интерфейс на наличие MJPEG ссылок
func (d *MJPEGDetector) checkWebInterfaceForMJPEG(client *http.Client, ip string, port int) bool {
	schemes := []string{"http", "https"}
	paths := []string{"", "/", "/index.html", "/video.html", "/stream.html"}

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
			
			// Ищем упоминания MJPEG в HTML
			keywords := []string{"mjpeg", "mjpg", "multipart/x-mixed-replace", "motion-jpeg"}
			lowerContent := strings.ToLower(content)
			
			for _, keyword := range keywords {
				if strings.Contains(lowerContent, keyword) {
					return true
				}
			}
		}
	}

	return false
}
