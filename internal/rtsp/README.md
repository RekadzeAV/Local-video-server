# RTSP Module

Модуль для проверки RTSP потоков и получения параметров.

## Компоненты

### 1. Client (`client.go`)
RTSP клиент с поддержкой:
- Методов: OPTIONS, DESCRIBE, SETUP, PLAY
- Аутентификации: Basic и Digest
- Парсинга RTSP ответов

### 2. Parser (`parser.go`)
Парсер SDP (Session Description Protocol) для извлечения:
- Видео кодеков (H.264, H.265, MJPEG)
- Аудио кодеков (AAC, PCM, G.711)
- Разрешения (width × height)
- FPS (кадров в секунду)
- Битрейта
- Профилей кодирования

### 3. FFmpeg Integration (`ffmpeg.go`)
Интеграция с FFmpeg/ffprobe для:
- Детального анализа потоков
- Fallback при проблемах с встроенным клиентом
- Проверки доступности потоков

### 4. Checker (`checker.go`)
Основной модуль проверки RTSP каналов:
- Проверка отдельных потоков
- Параллельная проверка нескольких потоков
- Обнаружение доступных потоков на устройстве
- Быстрая проверка доступности

## Пример использования

```go
package main

import (
    "fmt"
    "github.com/local-video-server/internal/config"
    "github.com/local-video-server/internal/rtsp"
)

func main() {
    // Загружаем конфигурацию
    cfg, _ := config.LoadConfig("configs/config.yaml")
    
    // Создаем RTSP checker
    checker := rtsp.NewChecker(&cfg.RTSP)
    
    // Проверяем один поток
    streamInfo, err := checker.CheckStream(
        "rtsp://192.168.1.10:554/Streaming/Channels/101",
        "admin",
        "password",
    )
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Stream URL: %s\n", streamInfo.URL)
    fmt.Printf("Codec: %s\n", streamInfo.Codec)
    fmt.Printf("Resolution: %s\n", streamInfo.Resolution)
    fmt.Printf("FPS: %.2f\n", streamInfo.FPS)
    fmt.Printf("Audio: %s, %d channels\n", streamInfo.AudioCodec, streamInfo.Channels)
    
    // Обнаруживаем потоки на устройстве
    streams, err := checker.DiscoverStreams("192.168.1.10", 554, "admin", "password")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Found %d streams:\n", len(streams))
    for _, stream := range streams {
        fmt.Printf("  - %s: %s %s @ %.2f fps\n", 
            stream.URL, stream.Codec, stream.Resolution, stream.FPS)
    }
}
```

## Конфигурация

Настройки RTSP в `config.yaml`:

```yaml
rtsp:
  timeout: 5s              # Таймаут для RTSP запросов
  use_ffmpeg: true         # Использовать FFmpeg для проверки
  ffmpeg_path: ""          # Путь к FFmpeg (пусто = из PATH)
  default_paths:           # Стандартные пути для проверки
    - "/Streaming/Channels/101"
    - "/Streaming/Channels/1"
    - "/live/main_stream"
    - "/live"
    - "/cam/realmonitor"
```

## Требования

- Go 1.21+
- FFmpeg (опционально, для детального анализа)

## Поддерживаемые кодеки

### Видео:
- H.264 (AVC)
- H.265 (HEVC)
- MJPEG
- MPEG-4

### Аудио:
- AAC
- PCM
- G.711 (PCMU/PCMA)
- G.722
