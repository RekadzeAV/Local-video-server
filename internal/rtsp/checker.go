package rtsp

import (
	"fmt"
	"time"

	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/pkg/utils"
)

// Checker представляет модуль проверки RTSP каналов
type Checker struct {
	config *models.RTSPConfig
}

// NewChecker создает новый RTSP checker
func NewChecker(config *models.RTSPConfig) *Checker {
	return &Checker{
		config: config,
	}
}

// CheckStream проверяет RTSP поток и возвращает информацию о нем
// Использует встроенный RTSP клиент, с fallback на FFmpeg при необходимости
func (c *Checker) CheckStream(rtspURL string, username, password string) (*models.RTSPStreamInfo, error) {
	logger := utils.GetLogger()

	// Сначала пытаемся использовать встроенный RTSP клиент
	logger.Debugf("Checking RTSP stream: %s", rtspURL)
	
	streamInfo, err := CheckStream(rtspURL, username, password, c.config.Timeout)
	if err != nil {
		logger.Debugf("RTSP client failed: %v, trying FFmpeg fallback", err)
		
		// Fallback на FFmpeg, если включен
		if c.config.UseFFmpeg {
			ffmpegInfo, ffmpegErr := CheckStreamWithFFmpeg(rtspURL, username, password, c.config.FFmpegPath, c.config.Timeout)
			if ffmpegErr != nil {
				return nil, fmt.Errorf("both RTSP client and FFmpeg failed: RTSP: %w, FFmpeg: %v", err, ffmpegErr)
			}
			streamInfo = ffmpegInfo
		} else {
			return nil, fmt.Errorf("RTSP client failed: %w", err)
		}
	}

	// Конвертируем в models.RTSPStreamInfo
	rtspStreamInfo := streamInfo.ToRTSPStreamInfo()
	rtspStreamInfo.CheckedAt = time.Now()

	return &rtspStreamInfo, nil
}

// CheckMultipleStreams проверяет несколько RTSP потоков параллельно
func (c *Checker) CheckMultipleStreams(streams []StreamCheckRequest) []StreamCheckResult {
	logger := utils.GetLogger()
	results := make([]StreamCheckResult, 0, len(streams))

	// Создаем канал для результатов
	resultChan := make(chan StreamCheckResult, len(streams))

	// Запускаем проверку потоков параллельно
	for _, stream := range streams {
		go func(req StreamCheckRequest) {
			result := StreamCheckResult{
				URL: req.URL,
			}

			streamInfo, err := c.CheckStream(req.URL, req.Username, req.Password)
			if err != nil {
				result.Error = err.Error()
				result.Available = false
				logger.Debugf("Stream check failed for %s: %v", req.URL, err)
			} else {
				result.StreamInfo = streamInfo
				result.Available = streamInfo.Available
				logger.Debugf("Stream check successful for %s: codec=%s, resolution=%s, fps=%.2f",
					req.URL, streamInfo.Codec, streamInfo.Resolution, streamInfo.FPS)
			}

			resultChan <- result
		}(stream)
	}

	// Собираем результаты
	for i := 0; i < len(streams); i++ {
		result := <-resultChan
		results = append(results, result)
	}

	return results
}

// TestStream проверяет доступность потока (быстрая проверка)
func (c *Checker) TestStream(rtspURL string, username, password string) (bool, error) {
	logger := utils.GetLogger()

	// Если FFmpeg доступен, используем его для быстрой проверки
	if c.config.UseFFmpeg {
		available, err := TestStreamWithFFmpeg(rtspURL, username, password, c.config.FFmpegPath, c.config.Timeout)
		if err == nil {
			return available, nil
		}
		logger.Debugf("FFmpeg test failed, trying RTSP client: %v", err)
	}

	// Fallback на RTSP клиент
	client, err := NewClient(rtspURL, username, password, c.config.Timeout)
	if err != nil {
		return false, fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Отправляем OPTIONS для быстрой проверки
	response, err := client.Options()
	if err != nil {
		return false, fmt.Errorf("OPTIONS failed: %w", err)
	}

	return response.StatusCode == 200, nil
}

// DiscoverStreams пытается обнаружить доступные RTSP потоки на устройстве
func (c *Checker) DiscoverStreams(deviceIP string, port int, username, password string) ([]models.RTSPStreamInfo, error) {
	logger := utils.GetLogger()
	discoveredStreams := []models.RTSPStreamInfo{}

	// Формируем базовый URL
	baseURL := fmt.Sprintf("rtsp://%s:%d", deviceIP, port)
	if port == 554 {
		baseURL = fmt.Sprintf("rtsp://%s", deviceIP)
	}

	// Проверяем стандартные пути из конфигурации
	pathsToCheck := c.config.DefaultPaths
	if len(pathsToCheck) == 0 {
		// Используем пути по умолчанию
		pathsToCheck = []string{
			"/Streaming/Channels/101",
			"/Streaming/Channels/1",
			"/live/main_stream",
			"/live",
			"/cam/realmonitor",
		}
	}

	logger.Debugf("Discovering RTSP streams on %s, checking %d paths", deviceIP, len(pathsToCheck))

	// Проверяем каждый путь
	for _, path := range pathsToCheck {
		streamURL := baseURL + path
		
		streamInfo, err := c.CheckStream(streamURL, username, password)
		if err != nil {
			logger.Debugf("Stream %s not available: %v", streamURL, err)
			continue
		}

		if streamInfo.Available {
			discoveredStreams = append(discoveredStreams, *streamInfo)
			logger.Infof("Discovered RTSP stream: %s (codec=%s, resolution=%s, fps=%.2f)",
				streamURL, streamInfo.Codec, streamInfo.Resolution, streamInfo.FPS)
		}
	}

	// Также пытаемся найти потоки через DESCRIBE на корневом пути
	rootStreamURL := baseURL + "/"
	rootStreamInfo, err := c.CheckStream(rootStreamURL, username, password)
	if err == nil && rootStreamInfo.Available {
		// Проверяем, не дублируется ли этот поток
		isDuplicate := false
		for _, existing := range discoveredStreams {
			if existing.URL == rootStreamInfo.URL {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			discoveredStreams = append(discoveredStreams, *rootStreamInfo)
		}
	}

	return discoveredStreams, nil
}

// StreamCheckRequest представляет запрос на проверку потока
type StreamCheckRequest struct {
	URL      string
	Username string
	Password string
}

// StreamCheckResult представляет результат проверки потока
type StreamCheckResult struct {
	URL        string
	StreamInfo *models.RTSPStreamInfo
	Available  bool
	Error      string
}
