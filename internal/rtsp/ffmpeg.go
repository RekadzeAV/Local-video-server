package rtsp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/local-video-server/pkg/utils"
)

// FFmpegProbeResult представляет результат работы ffprobe
type FFmpegProbeResult struct {
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
	} `json:"format"`
	Streams []struct {
		CodecName    string `json:"codec_name"`
		CodecType    string `json:"codec_type"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		RFrameRate   string `json:"r_frame_rate"`
		BitRate      string `json:"bit_rate"`
		SampleRate   string `json:"sample_rate"`
		Channels     int    `json:"channels"`
		Profile      string `json:"profile"`
		Level        int    `json:"level"`
	} `json:"streams"`
}

// CheckStreamWithFFmpeg проверяет RTSP поток используя FFmpeg/ffprobe
func CheckStreamWithFFmpeg(rtspURL string, username, password string, ffmpegPath string, timeout time.Duration) (*StreamInfo, error) {
	logger := utils.GetLogger()

	// Определяем путь к ffprobe
	ffprobePath := "ffprobe"
	if ffmpegPath != "" {
		// Если указан путь к ffmpeg, используем ffprobe из той же директории
		if strings.HasSuffix(ffmpegPath, "ffmpeg") || strings.HasSuffix(ffmpegPath, "ffmpeg.exe") {
			ffprobePath = strings.Replace(ffmpegPath, "ffmpeg", "ffprobe", 1)
			if strings.HasSuffix(ffprobePath, ".exe") {
				ffprobePath = strings.TrimSuffix(ffprobePath, ".exe") + ".exe"
			}
		} else {
			// Пытаемся найти ffprobe рядом с ffmpeg
			ffprobePath = ffmpegPath + "probe"
		}
	}

	// Формируем URL с аутентификацией, если нужно
	probeURL := rtspURL
	if username != "" && password != "" {
		// Вставляем credentials в URL
		if strings.HasPrefix(rtspURL, "rtsp://") {
			parts := strings.SplitN(rtspURL[7:], "@", 2)
			if len(parts) == 2 {
				// Уже есть credentials
				probeURL = rtspURL
			} else {
				// Добавляем credentials
				hostPart := strings.SplitN(rtspURL[7:], "/", 2)
				if len(hostPart) == 2 {
					probeURL = fmt.Sprintf("rtsp://%s:%s@%s/%s", username, password, hostPart[0], hostPart[1])
				} else {
					probeURL = fmt.Sprintf("rtsp://%s:%s@%s", username, password, hostPart[0])
				}
			}
		}
	}

	// Команда ffprobe
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		"-timeout", fmt.Sprintf("%.0f", timeout.Seconds()*1000000), // в микросекундах
		"-rtsp_transport", "tcp", // Используем TCP для надежности
		probeURL,
	}

	logger.Debugf("Running ffprobe: %s %v", ffprobePath, args)

	// Создаем команду
	cmd := exec.Command(ffprobePath, args...)
	
	// Устанавливаем таймаут
	ctx, cancel := time.WithTimeout(context.Background(), timeout*2)
	defer cancel()
	cmd = exec.CommandContext(ctx, ffprobePath, args...)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w, output: %s", err, string(output))
	}

	// Парсим JSON вывод
	var probeResult FFmpegProbeResult
	if err := json.Unmarshal(output, &probeResult); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Конвертируем результат в StreamInfo
	streamInfo := &StreamInfo{
		URL:        rtspURL,
		Available:  true,
		VideoTracks: []VideoTrack{},
		AudioTracks: []AudioTrack{},
	}

	// Обрабатываем потоки
	for _, stream := range probeResult.Streams {
		if stream.CodecType == "video" {
			videoTrack := VideoTrack{
				Codec: normalizeCodec(stream.CodecName),
			}

			if stream.Width > 0 && stream.Height > 0 {
				videoTrack.Resolution = fmt.Sprintf("%dx%d", stream.Width, stream.Height)
				streamInfo.Resolution = videoTrack.Resolution
			}

			// Парсим FPS из r_frame_rate (формат: "25/1")
			if stream.RFrameRate != "" {
				parts := strings.Split(stream.RFrameRate, "/")
				if len(parts) == 2 {
					if num, err1 := strconv.ParseFloat(parts[0], 64); err1 == nil {
						if den, err2 := strconv.ParseFloat(parts[1], 64); err2 == nil && den > 0 {
							videoTrack.FPS = num / den
							streamInfo.FPS = videoTrack.FPS
						}
					}
				}
			}

			// Парсим битрейт
			if stream.BitRate != "" {
				if bitrate, err := strconv.Atoi(stream.BitRate); err == nil {
					videoTrack.Bitrate = bitrate
					streamInfo.Bitrate = bitrate
				}
			}

			if stream.Profile != "" {
				videoTrack.Profile = stream.Profile
			}

			if stream.Level > 0 {
				videoTrack.Level = fmt.Sprintf("%d", stream.Level)
			}

			streamInfo.VideoTracks = append(streamInfo.VideoTracks, videoTrack)
			streamInfo.Codec = videoTrack.Codec
		} else if stream.CodecType == "audio" {
			audioTrack := AudioTrack{
				Codec: normalizeAudioCodec(stream.CodecName),
			}

			if stream.Channels > 0 {
				audioTrack.Channels = stream.Channels
				streamInfo.Channels = stream.Channels
			}

			if stream.SampleRate != "" {
				if sampleRate, err := strconv.Atoi(stream.SampleRate); err == nil {
					audioTrack.SampleRate = sampleRate
				}
			}

			if stream.BitRate != "" {
				if bitrate, err := strconv.Atoi(stream.BitRate); err == nil {
					audioTrack.Bitrate = bitrate
				}
			}

			streamInfo.AudioTracks = append(streamInfo.AudioTracks, audioTrack)
			streamInfo.AudioCodec = audioTrack.Codec
		}
	}

	// Если битрейт не найден в потоках, пытаемся взять из format
	if streamInfo.Bitrate == 0 && probeResult.Format.BitRate != "" {
		if bitrate, err := strconv.Atoi(probeResult.Format.BitRate); err == nil {
			streamInfo.Bitrate = bitrate
		}
	}

	return streamInfo, nil
}

// TestStreamWithFFmpeg проверяет доступность потока через FFmpeg
func TestStreamWithFFmpeg(rtspURL string, username, password string, ffmpegPath string, timeout time.Duration) (bool, error) {
	logger := utils.GetLogger()

	// Определяем путь к ffmpeg
	ffmpegCmd := "ffmpeg"
	if ffmpegPath != "" {
		ffmpegCmd = ffmpegPath
	}

	// Формируем URL с аутентификацией
	testURL := rtspURL
	if username != "" && password != "" {
		if strings.HasPrefix(rtspURL, "rtsp://") {
			parts := strings.SplitN(rtspURL[7:], "@", 2)
			if len(parts) == 2 {
				testURL = rtspURL
			} else {
				hostPart := strings.SplitN(rtspURL[7:], "/", 2)
				if len(hostPart) == 2 {
					testURL = fmt.Sprintf("rtsp://%s:%s@%s/%s", username, password, hostPart[0], hostPart[1])
				} else {
					testURL = fmt.Sprintf("rtsp://%s:%s@%s", username, password, hostPart[0])
				}
			}
		}
	}

	// Команда для проверки доступности (пытаемся получить несколько кадров)
	args := []string{
		"-rtsp_transport", "tcp",
		"-i", testURL,
		"-t", "1", // Получаем 1 секунду потока
		"-f", "null",
		"-", // Вывод в null
	}

	logger.Debugf("Testing stream with ffmpeg: %s %v", ffmpegCmd, args)

	// Создаем команду с таймаутом
	ctx, cancel := time.WithTimeout(context.Background(), timeout*2)
	defer cancel()
	cmd := exec.CommandContext(ctx, ffmpegCmd, args...)

	// Выполняем команду
	err := cmd.Run()
	if err != nil {
		// Проверяем, была ли это ошибка таймаута или реальная ошибка потока
		if ctx.Err() == context.DeadlineExceeded {
			return false, fmt.Errorf("timeout while testing stream")
		}
		return false, fmt.Errorf("stream test failed: %w", err)
	}

	return true, nil
}
