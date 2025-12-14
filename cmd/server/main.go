package main

import (
	"fmt"
	"os"
	"time"

	"github.com/local-video-server/internal/config"
	"github.com/local-video-server/internal/models"
	"github.com/local-video-server/internal/scanner"
	"github.com/local-video-server/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	configPath string
	verbose    bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "local-video-server",
		Short: "Local Video Server - обнаружение видеокамер в локальной сети",
		Long: `Local-video-server - это кроссплатформенное приложение на Go,
которое сканирует локальную сеть на наличие видеокамер и определяет
поддерживаемые протоколы (RTSP, RTMP, HLS, WebRTC, ONVIF, MJPEG, etc.)`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Загружаем конфигурацию
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				fmt.Printf("Warning: failed to load config: %v\n", err)
				cfg = models.DefaultConfig()
			}

			// Инициализируем логирование
			logLevel := cfg.Log.Level
			if verbose {
				logLevel = "debug"
			}

			if err := utils.InitLogger(logLevel, cfg.Log.Format, cfg.Log.File); err != nil {
				fmt.Printf("Failed to initialize logger: %v\n", err)
				os.Exit(1)
			}

			utils.GetLogger().Info("Local-video-server started")
		},
	}

	// Флаги
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "путь к конфигурационному файлу")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "подробный вывод (debug уровень)")

	// Команды
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Инициализировать конфигурационный файл",
	Long:  "Создает файл конфигурации по умолчанию в configs/config.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := "configs/config.yaml"
		if err := config.CreateDefaultConfigFile(configPath); err != nil {
			fmt.Printf("Failed to create config file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Configuration file created: %s\n", configPath)
	},
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Сканировать сеть на наличие видеокамер",
	Long:  "Сканирует указанную подсеть и обнаруживает видеокамеры",
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetLogger()

		// Загружаем конфигурацию
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			logger.Errorf("Failed to load config: %v", err)
			os.Exit(1)
		}

		// Логируем информацию о сети
		utils.LogNetworkInfo()

		// Определяем подсеть для сканирования
		subnet := cfg.Scan.Subnet
		if subnet == "" && cfg.Network.AutoDetectSubnet {
			detectedSubnet, err := utils.GetDefaultSubnet()
			if err != nil {
				logger.Errorf("Failed to detect subnet: %v", err)
				os.Exit(1)
			}
			subnet = detectedSubnet
			logger.Infof("Auto-detected subnet: %s", subnet)
		}

		if subnet == "" {
			logger.Error("Subnet not specified and auto-detection failed")
			os.Exit(1)
		}

		logger.Infof("Starting network scan: %s", subnet)

		// Создаем детектор для сканирования
		detector := scanner.NewDetector(&cfg.Scan)

		// Выполняем сканирование с таймаутом
		timeout := cfg.Scan.DiscoveryTimeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}

		devices, err := detector.ScanWithTimeout(subnet, timeout)
		if err != nil {
			logger.Errorf("Scan failed: %v", err)
			os.Exit(1)
		}

		// Выводим результаты
		logger.Infof("Scan completed. Found %d device(s):", len(devices))
		for i, device := range devices {
			logger.Infof("\nDevice %d:", i+1)
			logger.Infof("  IP: %s", device.IP)
			if device.Hostname != "" {
				logger.Infof("  Hostname: %s", device.Hostname)
			}
			if device.Manufacturer != "" {
				logger.Infof("  Manufacturer: %s", device.Manufacturer)
			}
			if device.Model != "" {
				logger.Infof("  Model: %s", device.Model)
			}
			if len(device.Protocols) > 0 {
				logger.Infof("  Protocols:")
				for _, protocol := range device.Protocols {
					logger.Infof("    - %s (port %d): %s", protocol.Type, protocol.Port, protocol.URL)
				}
			}
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Показать версию приложения",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Local-video-server v0.1.0")
		fmt.Println("Build: development")
	},
}
