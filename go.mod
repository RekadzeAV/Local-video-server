module github.com/local-video-server

go 1.21

require (
	// ONVIF
	github.com/use-go/onvif v0.0.0-20231025082739-ff6e66b2e8f5

	// Сетевое сканирование
	github.com/google/gopacket v1.1.19

	// HTTP клиент
	github.com/go-resty/resty/v2 v2.11.0

	// Конфигурация
	gopkg.in/yaml.v3 v3.0.1

	// Логирование
	github.com/sirupsen/logrus v1.9.3

	// Утилиты
	github.com/spf13/cobra v1.8.0  // CLI
	github.com/spf13/viper v1.18.2 // Конфигурация

	// GUI
	fyne.io/fyne/v2 v2.4.5 // Графический интерфейс в стиле Windows 10/11
)
