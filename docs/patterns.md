# Patterns: Правила написания кода

## Общие принципы

### 1. Читаемость кода
- **Имена должны быть понятными** — избегайте сокращений, кроме общепринятых
- **Функции должны делать одну вещь** — Single Responsibility Principle
- **Комментарии объясняют "почему", а не "что"** — код должен быть самодокументируемым
- **Избегайте магических чисел** — используйте константы с понятными именами

### 2. Структура кода
- **Пакеты должны иметь четкую ответственность** — один пакет = одна область функциональности
- **Избегайте циклических зависимостей** — используйте интерфейсы для развязки
- **Внутренние пакеты (`internal/`) не экспортируются** — только для внутреннего использования
- **Публичные пакеты (`pkg/`) могут использоваться внешне** — тщательно продумывайте API

### 3. Обработка ошибок
- **Всегда проверяйте ошибки** — не игнорируйте возвращаемые ошибки
- **Оборачивайте ошибки с контекстом** — используйте `fmt.Errorf` с `%w`
- **Логируйте ошибки на соответствующем уровне** — не дублируйте логирование
- **Возвращайте ошибки, а не паникуйте** — panic только для критических ошибок

```go
// ✅ Правильно
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// ❌ Неправильно
result, _ := doSomething() // игнорирование ошибки
```

### 4. Конкурентность
- **Используйте goroutines для параллельных операций** — но не создавайте слишком много
- **Синхронизация через каналы или sync примитивы** — избегайте гонок данных
- **Используйте context для отмены операций** — поддерживайте graceful shutdown
- **Ограничивайте количество одновременных операций** — используйте worker pools

```go
// ✅ Правильно
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

var wg sync.WaitGroup
results := make(chan Device, len(hosts))

for _, host := range hosts {
    wg.Add(1)
    go func(ip string) {
        defer wg.Done()
        device := scanDevice(ctx, ip)
        if device != nil {
            results <- *device
        }
    }(host)
}

go func() {
    wg.Wait()
    close(results)
}()
```

## Стиль кода Go

### 1. Именование
- **Публичные функции/типы:** PascalCase (`ScanNetwork`, `Device`)
- **Приватные функции/типы:** camelCase (`scanDevice`, `deviceInfo`)
- **Константы:** PascalCase или camelCase в зависимости от видимости
- **Интерфейсы:** обычно заканчиваются на `-er` (`Scanner`, `Checker`)

### 2. Структуры
- **Используйте теги для JSON/XML/YAML** — для сериализации
- **Встраивание структур** — для композиции, не наследования
- **Инициализация через конструкторы** — `NewDevice()`, `NewScanner()`

```go
// ✅ Правильно
type Device struct {
    IP          string    `json:"ip" yaml:"ip"`
    MAC         string    `json:"mac,omitempty" yaml:"mac,omitempty"`
    Protocols   []Protocol `json:"protocols" yaml:"protocols"`
    DiscoveredAt time.Time `json:"discovered_at" yaml:"discovered_at"`
}

func NewDevice(ip string) *Device {
    return &Device{
        IP:          ip,
        Protocols:   make([]Protocol, 0),
        DiscoveredAt: time.Now(),
    }
}
```

### 3. Интерфейсы
- **Интерфейсы должны быть маленькими** — предпочтительнее несколько маленьких интерфейсов
- **Интерфейсы определяются там, где используются** — не там, где реализуются
- **Используйте интерфейсы для тестирования** — легко создавать моки

```go
// ✅ Правильно
type ProtocolDetector interface {
    Detect(ctx context.Context, device *Device) ([]Protocol, error)
}

type RTSPDetector struct {
    timeout time.Duration
}

func (d *RTSPDetector) Detect(ctx context.Context, device *Device) ([]Protocol, error) {
    // реализация
}
```

### 4. Обработка контекста
- **Первый параметр — context.Context** — для всех функций, которые могут быть отменены
- **Проверяйте ctx.Done()** — в долгих операциях
- **Передавайте context вниз** — не создавайте новый без необходимости

```go
// ✅ Правильно
func ScanNetwork(ctx context.Context, subnet string) ([]Device, error) {
    for _, host := range hosts {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
            device := scanDevice(ctx, host)
            // ...
        }
    }
}
```

## Паттерны проектирования

### 1. Dependency Injection
- **Передавайте зависимости через конструкторы** — не создавайте внутри функций
- **Используйте интерфейсы** — для развязки компонентов

```go
type Scanner struct {
    detector ProtocolDetector
    logger   Logger
    timeout  time.Duration
}

func NewScanner(detector ProtocolDetector, logger Logger, timeout time.Duration) *Scanner {
    return &Scanner{
        detector: detector,
        logger:   logger,
        timeout:  timeout,
    }
}
```

### 2. Factory Pattern
- **Используйте для создания сложных объектов** — с валидацией и инициализацией

```go
func NewRTSPChecker(config RTSPConfig) (*RTSPChecker, error) {
    if config.Timeout <= 0 {
        return nil, errors.New("timeout must be positive")
    }
    return &RTSPChecker{
        timeout: config.Timeout,
        client:  newRTSPClient(),
    }, nil
}
```

### 3. Strategy Pattern
- **Для различных алгоритмов обнаружения** — разные стратегии сканирования

```go
type DiscoveryStrategy interface {
    Discover(ctx context.Context, subnet string) ([]Device, error)
}

type ARPDiscovery struct{}
type ONVIFDiscovery struct{}
type UPnPDiscovery struct{}
```

### 4. Builder Pattern
- **Для сложных конфигураций** — пошаговое построение объектов

```go
type ConfigBuilder struct {
    config Config
}

func NewConfigBuilder() *ConfigBuilder {
    return &ConfigBuilder{config: DefaultConfig()}
}

func (b *ConfigBuilder) WithTimeout(d time.Duration) *ConfigBuilder {
    b.config.Timeout = d
    return b
}

func (b *ConfigBuilder) Build() Config {
    return b.config
}
```

## Работа с ошибками

### 1. Типизированные ошибки
- **Используйте переменные ошибок** — для проверки типа ошибки

```go
var (
    ErrDeviceNotFound = errors.New("device not found")
    ErrTimeout        = errors.New("operation timeout")
)

func findDevice(ip string) (*Device, error) {
    // ...
    if notFound {
        return nil, ErrDeviceNotFound
    }
}
```

### 2. Оборачивание ошибок
- **Добавляйте контекст** — используйте `fmt.Errorf` с `%w`

```go
func scanDevice(ctx context.Context, ip string) (*Device, error) {
    device, err := connect(ip)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to %s: %w", ip, err)
    }
    return device, nil
}
```

### 3. Retry механизм
- **Для сетевых операций** — с экспоненциальной задержкой

```go
func retry(ctx context.Context, fn func() error, maxRetries int) error {
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        if err := fn(); err == nil {
            return nil
        }
        lastErr = err
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(time.Duration(i+1) * time.Second):
        }
    }
    return fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

## Логирование

### 1. Уровни логирования
- **DEBUG** — детальная информация для отладки
- **INFO** — общая информация о работе
- **WARN** — предупреждения, не критичные проблемы
- **ERROR** — ошибки, требующие внимания

### 2. Структурированное логирование
- **Используйте поля** — для структурированных данных

```go
logger.WithFields(logrus.Fields{
    "ip":       device.IP,
    "protocol": "RTSP",
    "port":     554,
}).Info("Scanning device")
```

### 3. Контекст в логах
- **Добавляйте контекст** — IP адрес, протокол, операция

```go
logger.WithField("device_ip", ip).Error("Failed to scan device")
```

## Тестирование

### 1. Unit тесты
- **Один тест = одна проверка** — не смешивайте несколько проверок
- **Используйте табличные тесты** — для множественных сценариев
- **Моки для внешних зависимостей** — используйте интерфейсы

```go
func TestScanDevice(t *testing.T) {
    tests := []struct {
        name    string
        ip      string
        wantErr bool
    }{
        {"valid ip", "192.168.1.1", false},
        {"invalid ip", "999.999.999.999", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            device, err := scanDevice(tt.ip)
            if (err != nil) != tt.wantErr {
                t.Errorf("scanDevice() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 2. Интеграционные тесты
- **Тестируйте взаимодействие компонентов** — но изолированно
- **Используйте тестовые данные** — не зависите от реальной сети

## Общие правила для всех проектов

### 1. Версионирование
- **Семантическое версионирование** — MAJOR.MINOR.PATCH
- **Git теги для релизов** — `v1.0.0`, `v1.1.0`

### 2. Документация
- **Godoc комментарии** — для всех публичных функций и типов
- **README с примерами** — как использовать проект
- **CHANGELOG** — история изменений

### 3. Безопасность
- **Не логируйте чувствительные данные** — пароли, токены
- **Валидация входных данных** — проверяйте все входные параметры
- **Ограничение ресурсов** — таймауты, лимиты соединений

### 4. Производительность
- **Профилирование перед оптимизацией** — измеряйте, не угадывайте
- **Избегайте преждевременной оптимизации** — сначала правильность, потом скорость
- **Используйте пулы объектов** — для часто создаваемых объектов

### 5. Git
- **Осмысленные коммиты** — один коммит = одна логическая единица
- **Понятные сообщения коммитов** — что и зачем изменено
- **Не коммитить временные файлы** — используйте .gitignore

## Примеры хорошего кода

### Хорошая структура функции:
```go
// ScanNetwork сканирует указанную подсеть и возвращает список обнаруженных устройств.
// Использует параллельное сканирование для ускорения процесса.
func ScanNetwork(ctx context.Context, subnet string, config ScanConfig) ([]Device, error) {
    // 1. Валидация входных данных
    if err := validateSubnet(subnet); err != nil {
        return nil, fmt.Errorf("invalid subnet: %w", err)
    }
    
    // 2. Получение списка хостов
    hosts, err := getActiveHosts(ctx, subnet)
    if err != nil {
        return nil, fmt.Errorf("failed to get hosts: %w", err)
    }
    
    // 3. Параллельное сканирование
    devices := make([]Device, 0)
    results := make(chan Device, len(hosts))
    // ... параллельная обработка
    
    return devices, nil
}
```

### Хорошая обработка ошибок:
```go
func checkRTSPStream(ctx context.Context, url string) (*StreamInfo, error) {
    conn, err := net.DialTimeout("tcp", extractHost(url), 5*time.Second)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to RTSP server: %w", err)
    }
    defer conn.Close()
    
    // ... остальная логика
}
```
