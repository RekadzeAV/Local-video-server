package protocols

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/pkg/utils"
	"github.com/sirupsen/logrus"
)

// RTSPDetector - детектор RTSP протокола
type RTSPDetector struct {
	logger *logrus.Logger
}

// NewRTSPDetector создает новый RTSP детектор
func NewRTSPDetector() *RTSPDetector {
	return &RTSPDetector{
		logger: utils.GetLogger(),
	}
}

// GetName возвращает название протокола
func (d *RTSPDetector) GetName() string {
	return "RTSP"
}

// GetDefaultPort возвращает порт по умолчанию
func (d *RTSPDetector) GetDefaultPort() int {
	return 554
}

// Detect проверяет наличие RTSP протокола на устройстве
func (d *RTSPDetector) Detect(ip string, port int, timeout time.Duration) (*models.Protocol, error) {
	protocol := &models.Protocol{
		Type:       "RTSP",
		Port:       port,
		Available:  false,
		DetectedAt: time.Now(),
	}

	// Подключение к RTSP порту
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return protocol, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Устанавливаем таймаут для операций чтения/записи
	conn.SetDeadline(time.Now().Add(timeout))

	// Отправка OPTIONS запроса
	optionsReq := "OPTIONS * RTSP/1.0\r\n" +
		"CSeq: 1\r\n" +
		"User-Agent: Local-Video-Server/1.0\r\n" +
		"\r\n"

	if _, err := conn.Write([]byte(optionsReq)); err != nil {
		return protocol, fmt.Errorf("failed to send OPTIONS: %w", err)
	}

	// Чтение ответа
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return protocol, fmt.Errorf("failed to read response: %w", err)
	}

	// Проверка статуса ответа
	if !strings.Contains(response, "RTSP/1.0") {
		return protocol, fmt.Errorf("invalid RTSP response: %s", response)
	}

	// Проверка успешного ответа (200 OK или 401 Unauthorized - тоже признак RTSP)
	if strings.Contains(response, "200") || strings.Contains(response, "401") {
		protocol.Available = true
		protocol.URL = fmt.Sprintf("rtsp://%s:%d", ip, port)

		// Попытка получить DESCRIBE для определения потоков
		streams, err := d.getStreams(conn, ip, port, timeout)
		if err == nil && len(streams) > 0 {
			// Если удалось получить потоки, можно добавить дополнительную информацию
			d.logger.Debugf("Found %d RTSP streams on %s:%d", len(streams), ip, port)
		}
	}

	return protocol, nil
}

// getStreams пытается получить информацию о потоках через DESCRIBE
func (d *RTSPDetector) getStreams(conn net.Conn, ip string, port int, timeout time.Duration) ([]string, error) {
	// Отправка DESCRIBE запроса
	describeReq := fmt.Sprintf("DESCRIBE rtsp://%s:%d/ RTSP/1.0\r\n", ip, port) +
		"CSeq: 2\r\n" +
		"Accept: application/sdp\r\n" +
		"User-Agent: Local-Video-Server/1.0\r\n" +
		"\r\n"

	conn.SetDeadline(time.Now().Add(timeout))
	if _, err := conn.Write([]byte(describeReq)); err != nil {
		return nil, err
	}

	// Чтение SDP ответа
	reader := bufio.NewReader(conn)
	var sdp strings.Builder
	var streams []string

	// Читаем заголовки
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if line == "\r\n" || line == "\n" {
			break
		}
	}

	// Читаем SDP тело
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		sdp.WriteString(line)
		
		// Парсим SDP для поиска медиа потоков
		if strings.HasPrefix(line, "m=") {
			// m=video или m=audio указывает на поток
			parts := strings.Fields(line)
			if len(parts) > 0 {
				streams = append(streams, parts[0])
			}
		}
	}

	return streams, nil
}

// ParseSDP парсит SDP ответ и извлекает информацию о потоках
func (d *RTSPDetector) ParseSDP(sdp string) ([]models.RTSPStreamInfo, error) {
	var streams []models.RTSPStreamInfo
	
	lines := strings.Split(sdp, "\n")
	var currentStream *models.RTSPStreamInfo
	
	// Регулярные выражения для парсинга SDP
	codecRegex := regexp.MustCompile(`a=rtpmap:(\d+)\s+(\w+)/(\d+)`)
	resolutionRegex := regexp.MustCompile(`a=framesize:(\d+)\s+(\d+)x(\d+)`)
	fpsRegex := regexp.MustCompile(`a=framerate:([\d.]+)`)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Начало медиа описания
		if strings.HasPrefix(line, "m=video") {
			if currentStream != nil {
				streams = append(streams, *currentStream)
			}
			currentStream = &models.RTSPStreamInfo{
				Available: true,
				CheckedAt: time.Now(),
			}
		}
		
		if currentStream == nil {
			continue
		}
		
		// Парсинг кодека
		if matches := codecRegex.FindStringSubmatch(line); len(matches) > 0 {
			codec := strings.ToUpper(matches[2])
			if codec == "H264" || codec == "H265" || codec == "MPEG4" {
				currentStream.Codec = codec
			} else if codec == "JPEG" {
				currentStream.Codec = "MJPEG"
			}
		}
		
		// Парсинг разрешения
		if matches := resolutionRegex.FindStringSubmatch(line); len(matches) > 0 {
			currentStream.Resolution = fmt.Sprintf("%sx%s", matches[2], matches[3])
		}
		
		// Парсинг FPS
		if matches := fpsRegex.FindStringSubmatch(line); len(matches) > 0 {
			var fps float64
			fmt.Sscanf(matches[1], "%f", &fps)
			currentStream.FPS = fps
		}
	}
	
	if currentStream != nil {
		streams = append(streams, *currentStream)
	}
	
	return streams, nil
}
