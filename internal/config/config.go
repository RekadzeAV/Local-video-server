package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/local-video-server/internal/models"
	"github.com/spf13/viper"
)

// LoadConfig загружает конфигурацию из файла или использует значения по умолчанию
func LoadConfig(configPath string) (*models.Config, error) {
	cfg := models.DefaultConfig()

	// Если путь к конфигу не указан, ищем в стандартных местах
	if configPath == "" {
		configPath = findConfigFile()
	}

	// Если конфиг найден, загружаем его
	if configPath != "" {
		viper.SetConfigFile(configPath)
		viper.SetConfigType("yaml")

		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		// Загружаем конфигурацию с правильной структурой
		if err := viper.Unmarshal(cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	// Устанавливаем значения по умолчанию для пустых полей
	setDefaults(cfg)

	return cfg, nil
}

// SaveConfig сохраняет конфигурацию в файл
func SaveConfig(cfg *models.Config, configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Преобразуем конфигурацию в map для viper
	viper.Set("scan", cfg.Scan)
	viper.Set("log", cfg.Log)
	viper.Set("network", cfg.Network)
	viper.Set("rtsp", cfg.RTSP)

	// Создаем директорию, если её нет
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return viper.WriteConfigAs(configPath)
}

// findConfigFile ищет конфигурационный файл в стандартных местах
func findConfigFile() string {
	// Список возможных путей к конфигу
	possiblePaths := []string{
		"configs/config.yaml",
		"config.yaml",
		"./configs/config.yaml",
		"./config.yaml",
	}

	// Также проверяем домашнюю директорию пользователя
	homeDir, err := os.UserHomeDir()
	if err == nil {
		possiblePaths = append(possiblePaths,
			filepath.Join(homeDir, ".local-video-server", "config.yaml"),
		)
	}

	// Ищем первый существующий файл
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// CreateDefaultConfigFile создает файл конфигурации по умолчанию
func CreateDefaultConfigFile(configPath string) error {
	cfg := models.DefaultConfig()
	return SaveConfig(cfg, configPath)
}

// setDefaults устанавливает значения по умолчанию для пустых полей
func setDefaults(cfg *models.Config) {
	// Если подсеть не указана и автоопределение включено, оставляем пустой
	// (будет определено позже)
	
	// Устанавливаем значения по умолчанию для портов, если список пуст
	if len(cfg.Scan.Ports) == 0 {
		cfg.Scan.Ports = []int{554, 1935, 80, 8080, 8554}
	}

	// Устанавливаем значения по умолчанию для RTSP путей
	if len(cfg.RTSP.DefaultPaths) == 0 {
		cfg.RTSP.DefaultPaths = []string{
			"/Streaming/Channels/101",
			"/Streaming/Channels/1",
			"/live/main_stream",
			"/live",
			"/cam/realmonitor",
		}
	}
}
