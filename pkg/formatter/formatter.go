package formatter

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/local-video-server/internal/models"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

// Formatter - форматирование вывода
type Formatter struct {
	useColors bool
	detailed  bool
}

// NewFormatter создает новый форматтер
func NewFormatter(useColors bool, detailed bool) *Formatter {
	// Проверяем, поддерживает ли терминал цвета
	if useColors {
		useColors = isColorTerminal()
	}

	return &Formatter{
		useColors: useColors,
		detailed:  detailed,
	}
}

// isColorTerminal проверяет, поддерживает ли терминал цвета
func isColorTerminal() bool {
	// Проверяем переменные окружения
	term := os.Getenv("TERM")
	if term == "" {
		// Windows
		term = os.Getenv("TERM_PROGRAM")
	}

	// Проверяем, что это не перенаправление в файл
	fileInfo, _ := os.Stdout.Stat()
	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	// Проверяем известные терминалы, которые поддерживают цвета
	colorTerms := []string{"xterm", "xterm-256color", "screen", "screen-256color", "tmux", "tmux-256color"}
	for _, ct := range colorTerms {
		if strings.Contains(term, ct) {
			return true
		}
	}

	// Windows 10+ поддерживает ANSI цвета
	if os.Getenv("WT_SESSION") != "" || os.Getenv("ConEmuANSI") == "ON" {
		return true
	}

	return false
}

// color применяет цвет к тексту, если цвета включены
func (f *Formatter) color(text string, colorCode string) string {
	if !f.useColors {
		return text
	}
	return colorCode + text + ColorReset
}

// PrintDevices выводит список устройств в табличном формате
func (f *Formatter) PrintDevices(devices []*models.Device) {
	if len(devices) == 0 {
		fmt.Println(f.color("No devices found", ColorYellow))
		return
	}

	if f.detailed {
		f.printDevicesDetailed(devices)
	} else {
		f.printDevicesTable(devices)
	}
}

// printDevicesTable выводит устройства в виде таблицы
func (f *Formatter) printDevicesTable(devices []*models.Device) {
	// Заголовок
	header := fmt.Sprintf("%-15s %-18s %-20s %-15s %-15s %-10s",
		"IP", "MAC", "Hostname", "Manufacturer", "Model", "Protocols")
	fmt.Println(f.color(header, ColorBold+ColorCyan))
	fmt.Println(strings.Repeat("-", 95))

	// Данные
	for _, device := range devices {
		protocols := f.formatProtocols(device.Protocols)
		if len(protocols) > 10 {
			protocols = protocols[:7] + "..."
		}

		row := fmt.Sprintf("%-15s %-18s %-20s %-15s %-15s %-10s",
			f.color(device.IP, ColorGreen),
			device.MAC,
			device.Hostname,
			device.Manufacturer,
			device.Model,
			protocols,
		)
		fmt.Println(row)
	}

	fmt.Printf("\n%s: %d\n", f.color("Total devices", ColorBold), len(devices))
}

// printDevicesDetailed выводит детальную информацию об устройствах
func (f *Formatter) printDevicesDetailed(devices []*models.Device) {
	for i, device := range devices {
		if i > 0 {
			fmt.Println()
		}

		fmt.Println(f.color(fmt.Sprintf("Device #%d", i+1), ColorBold+ColorCyan))
		fmt.Println(strings.Repeat("=", 60))

		f.printDeviceDetails(device)
	}

	fmt.Printf("\n%s: %d\n", f.color("Total devices", ColorBold), len(devices))
}

// printDeviceDetails выводит детальную информацию об одном устройстве
func (f *Formatter) printDeviceDetails(device *models.Device) {
	fmt.Printf("%s: %s\n", f.color("IP Address", ColorBold), f.color(device.IP, ColorGreen))

	if device.MAC != "" {
		fmt.Printf("%s: %s\n", f.color("MAC Address", ColorBold), device.MAC)
	}

	if device.Hostname != "" {
		fmt.Printf("%s: %s\n", f.color("Hostname", ColorBold), device.Hostname)
	}

	if device.Manufacturer != "" {
		fmt.Printf("%s: %s\n", f.color("Manufacturer", ColorBold), device.Manufacturer)
	}

	if device.Model != "" {
		fmt.Printf("%s: %s\n", f.color("Model", ColorBold), device.Model)
	}

	// Протоколы
	if len(device.Protocols) > 0 {
		fmt.Printf("%s:\n", f.color("Protocols", ColorBold))
		for _, protocol := range device.Protocols {
			status := f.color("✗", ColorRed)
			if protocol.Available {
				status = f.color("✓", ColorGreen)
			}
			fmt.Printf("  %s %s:%d (%s)\n",
				status,
				protocol.Type,
				protocol.Port,
				protocol.URL,
			)
		}
	}

	// RTSP потоки
	if len(device.RTSPStreams) > 0 {
		fmt.Printf("%s:\n", f.color("RTSP Streams", ColorBold))
		for _, stream := range device.RTSPStreams {
			status := f.color("✗", ColorRed)
			if stream.Available {
				status = f.color("✓", ColorGreen)
			}
			fmt.Printf("  %s %s\n", status, stream.URL)
			if stream.Codec != "" {
				fmt.Printf("    Codec: %s, Resolution: %s, FPS: %.2f\n",
					stream.Codec, stream.Resolution, stream.FPS)
			}
		}
	}

	// Временные метки
	fmt.Printf("%s: %s\n", f.color("Discovered At", ColorBold),
		device.DiscoveredAt.Format("2006-01-02 15:04:05"))
	if !device.LastSeen.IsZero() {
		fmt.Printf("%s: %s\n", f.color("Last Seen", ColorBold),
			device.LastSeen.Format("2006-01-02 15:04:05"))
	}
}

// formatProtocols форматирует список протоколов
func (f *Formatter) formatProtocols(protocols []models.Protocol) string {
	if len(protocols) == 0 {
		return "none"
	}

	var parts []string
	for _, p := range protocols {
		status := "✗"
		if p.Available {
			status = "✓"
		}
		parts = append(parts, fmt.Sprintf("%s%s:%d", status, p.Type, p.Port))
	}

	return strings.Join(parts, ", ")
}

// PrintSummary выводит краткую сводку
func (f *Formatter) PrintSummary(devices []*models.Device) {
	total := len(devices)
	withRTSP := 0
	withProtocols := 0
	protocolCounts := make(map[string]int)

	for _, device := range devices {
		if len(device.RTSPStreams) > 0 {
			withRTSP++
		}
		if len(device.Protocols) > 0 {
			withProtocols++
		}
		for _, protocol := range device.Protocols {
			protocolCounts[protocol.Type]++
		}
	}

	fmt.Println(f.color("\n=== Scan Summary ===", ColorBold+ColorCyan))
	fmt.Printf("%s: %d\n", f.color("Total Devices", ColorBold), total)
	fmt.Printf("%s: %d\n", f.color("Devices with Protocols", ColorBold), withProtocols)
	fmt.Printf("%s: %d\n", f.color("Devices with RTSP Streams", ColorBold), withRTSP)

	if len(protocolCounts) > 0 {
		fmt.Printf("\n%s:\n", f.color("Protocol Distribution", ColorBold))
		for protocol, count := range protocolCounts {
			fmt.Printf("  %s: %d\n", protocol, count)
		}
	}
}

// PrintError выводит ошибку с цветом
func (f *Formatter) PrintError(message string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", f.color("ERROR:", ColorBold+ColorRed), message)
}

// PrintWarning выводит предупреждение с цветом
func (f *Formatter) PrintWarning(message string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", f.color("WARNING:", ColorBold+ColorYellow), message)
}

// PrintSuccess выводит успешное сообщение с цветом
func (f *Formatter) PrintSuccess(message string) {
	fmt.Printf("%s %s\n", f.color("SUCCESS:", ColorBold+ColorGreen), message)
}

// PrintInfo выводит информационное сообщение
func (f *Formatter) PrintInfo(message string) {
	fmt.Printf("%s %s\n", f.color("INFO:", ColorBold+ColorBlue), message)
}

// PrintProgress выводит прогресс сканирования
func (f *Formatter) PrintProgress(current, total int, message string) {
	percent := float64(current) / float64(total) * 100
	fmt.Printf("\r%s [%d/%d] %.1f%% - %s",
		f.color("Progress:", ColorBold+ColorCyan),
		current, total, percent, message)
	if current == total {
		fmt.Println() // Новая строка после завершения
	}
}
