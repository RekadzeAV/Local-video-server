package models

import "time"

// Config - конфигурация приложения
type Config struct {
	// Настройки сканирования
	Scan ScanConfig `yaml:"scan" json:"scan"`

	// Настройки логирования
	Log LogConfig `yaml:"log" json:"log"`

	// Настройки сети
	Network NetworkConfig `yaml:"network" json:"network"`

	// Настройки RTSP
	RTSP RTSPConfig `yaml:"rtsp" json:"rtsp"`
}

// ScanConfig - настройки сканирования
type ScanConfig struct {
	// Подсеть для сканирования (например, "192.168.1.0/24")
	Subnet string `yaml:"subnet" json:"subnet"`

	// Таймаут для сканирования порта (в секундах)
	PortTimeout time.Duration `yaml:"port_timeout" json:"port_timeout"`

	// Таймаут для обнаружения устройств (в секундах)
	DiscoveryTimeout time.Duration `yaml:"discovery_timeout" json:"discovery_timeout"`

	// Максимальное количество параллельных сканирований
	MaxConcurrency int `yaml:"max_concurrency" json:"max_concurrency"`

	// Порты для сканирования
	Ports []int `yaml:"ports" json:"ports"`

	// Включить ONVIF Discovery
	EnableONVIF bool `yaml:"enable_onvif" json:"enable_onvif"`

	// Включить UPnP/SSDP Discovery
	EnableUPnP bool `yaml:"enable_upnp" json:"enable_upnp"`

	// Проверять RTSP потоки
	CheckRTSP bool `yaml:"check_rtsp" json:"check_rtsp"`
}

// LogConfig - настройки логирования
type LogConfig struct {
	// Уровень логирования (debug, info, warn, error)
	Level string `yaml:"level" json:"level"`

	// Формат вывода (text, json)
	Format string `yaml:"format" json:"format"`

	// Путь к файлу логов (пусто = stdout)
	File string `yaml:"file" json:"file"`
}

// NetworkConfig - настройки сети
type NetworkConfig struct {
	// Автоматически определять подсеть
	AutoDetectSubnet bool `yaml:"auto_detect_subnet" json:"auto_detect_subnet"`

	// Интерфейс для сканирования (пусто = все интерфейсы)
	Interface string `yaml:"interface" json:"interface"`
}

// RTSPConfig - настройки RTSP
type RTSPConfig struct {
	// Таймаут для RTSP запросов (в секундах)
	Timeout time.Duration `yaml:"timeout" json:"timeout"`

	// Использовать FFmpeg для проверки потоков
	UseFFmpeg bool `yaml:"use_ffmpeg" json:"use_ffmpeg"`

	// Путь к FFmpeg (пусто = из PATH)
	FFmpegPath string `yaml:"ffmpeg_path" json:"ffmpeg_path"`

	// Стандартные пути RTSP потоков для проверки
	DefaultPaths []string `yaml:"default_paths" json:"default_paths"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Scan: ScanConfig{
			Subnet:            "",
			PortTimeout:        2 * time.Second,
			DiscoveryTimeout:   10 * time.Second,
			MaxConcurrency:     50,
			Ports:              []int{554, 1935, 80, 8080, 8554},
			EnableONVIF:        true,
			EnableUPnP:         true,
			CheckRTSP:          false,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
			File:   "",
		},
		Network: NetworkConfig{
			AutoDetectSubnet: true,
			Interface:        "",
		},
		RTSP: RTSPConfig{
			Timeout:      5 * time.Second,
			UseFFmpeg:    true,
			FFmpegPath:   "",
			DefaultPaths: []string{
				"/Streaming/Channels/101",
				"/Streaming/Channels/1",
				"/live/main_stream",
				"/live",
				"/cam/realmonitor",
			},
		},
	}
}
