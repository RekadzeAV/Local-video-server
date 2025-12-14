# Database: Работа с данными

## Обзор

**Local-video-server** на текущем этапе не использует традиционную базу данных. Данные хранятся в памяти во время выполнения и экспортируются в файлы (JSON, CSV, XML, YAML).

В будущем планируется добавление поддержки базы данных для хранения истории сканирований и состояния устройств.

---

## Текущая модель хранения данных

### В памяти (во время выполнения)

Данные хранятся в структурах Go в памяти приложения:

```go
// Реестр устройств
type DeviceRegistry struct {
    devices map[string]*Device
    mu      sync.RWMutex
}

// Устройство
type Device struct {
    IP           string            `json:"ip" yaml:"ip"`
    MAC          string            `json:"mac,omitempty" yaml:"mac,omitempty"`
    Hostname     string            `json:"hostname,omitempty" yaml:"hostname,omitempty"`
    Manufacturer string            `json:"manufacturer,omitempty" yaml:"manufacturer,omitempty"`
    Model        string            `json:"model,omitempty" yaml:"model,omitempty"`
    Protocols    []Protocol        `json:"protocols" yaml:"protocols"`
    RTSPStreams []RTSPStreamInfo  `json:"rtsp_streams,omitempty" yaml:"rtsp_streams,omitempty"`
    DiscoveredAt time.Time         `json:"discovered_at" yaml:"discovered_at"`
    LastSeen     time.Time         `json:"last_seen" yaml:"last_seen"`
}

// Протокол
type Protocol struct {
    Type      string   `json:"type" yaml:"type"`           // RTSP, RTMP, HLS, etc.
    Port      int      `json:"port" yaml:"port"`
    URL       string   `json:"url,omitempty" yaml:"url,omitempty"`
    Available bool     `json:"available" yaml:"available"`
    DetectedAt time.Time `json:"detected_at" yaml:"detected_at"`
}

// RTSP поток
type RTSPStreamInfo struct {
    URL         string   `json:"url" yaml:"url"`
    Codec       string   `json:"codec" yaml:"codec"`              // H.264, H.265, MJPEG
    Resolution  string   `json:"resolution" yaml:"resolution"`   // 1920x1080
    FPS         float64  `json:"fps" yaml:"fps"`
    Bitrate     int      `json:"bitrate,omitempty" yaml:"bitrate,omitempty"`
    AudioCodec  string   `json:"audio_codec,omitempty" yaml:"audio_codec,omitempty"`
    Channels    int      `json:"channels,omitempty" yaml:"channels,omitempty"`
    Available   bool     `json:"available" yaml:"available"`
    CheckedAt   time.Time `json:"checked_at" yaml:"checked_at"`
}
```

### Экспорт в файлы

Результаты сканирования экспортируются в различные форматы:

- **JSON** — структурированный формат для интеграций
- **CSV** — табличный формат для анализа в Excel/Google Sheets
- **XML** — формат для интеграций с другими системами
- **YAML** — читаемый формат для конфигураций

---

## Схема данных (концептуальная)

### ER диаграмма (для будущей БД)

```
┌─────────────┐
│   Device    │
├─────────────┤
│ id (PK)     │
│ ip          │
│ mac         │
│ hostname    │
│ manufacturer│
│ model       │
│ created_at  │
│ updated_at  │
└──────┬──────┘
       │
       │ 1:N
       │
┌──────▼──────────┐
│   Protocol      │
├─────────────────┤
│ id (PK)         │
│ device_id (FK)  │
│ type            │
│ port            │
│ url             │
│ available       │
│ detected_at     │
└─────────────────┘

┌─────────────┐
│   Device    │
└──────┬──────┘
       │
       │ 1:N
       │
┌──────▼──────────────┐
│   RTSPStream        │
├─────────────────────┤
│ id (PK)             │
│ device_id (FK)      │
│ url                 │
│ codec               │
│ resolution          │
│ fps                 │
│ bitrate             │
│ audio_codec         │
│ channels            │
│ available           │
│ checked_at          │
└─────────────────────┘

┌─────────────┐
│   Scan      │
├─────────────┤
│ id (PK)     │
│ subnet      │
│ started_at  │
│ finished_at │
│ devices_found│
│ status      │
└──────┬──────┘
       │
       │ 1:N
       │
┌──────▼──────────┐
│ ScanDevice      │
├─────────────────┤
│ scan_id (FK)    │
│ device_id (FK)  │
│ discovered_at   │
└─────────────────┘
```

---

## Будущая схема базы данных

### Варианты БД

#### SQLite (рекомендуется для начала)
- **Преимущества:**
  - Не требует отдельного сервера
  - Файловая БД, легко переносить
  - Хорошая производительность для небольших объемов
  - Поддержка SQL
- **Недостатки:**
  - Ограничения для больших объемов данных
  - Нет одновременной записи

#### PostgreSQL (для продакшена)
- **Преимущества:**
  - Мощная реляционная БД
  - Отличная производительность
  - Поддержка JSON полей
  - Масштабируемость
- **Недостатки:**
  - Требует отдельный сервер
  - Более сложная настройка

### SQL схема (PostgreSQL/SQLite)

```sql
-- Таблица устройств
CREATE TABLE devices (
    id SERIAL PRIMARY KEY,
    ip VARCHAR(45) NOT NULL UNIQUE,
    mac VARCHAR(17),
    hostname VARCHAR(255),
    manufacturer VARCHAR(100),
    model VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP
);

CREATE INDEX idx_devices_ip ON devices(ip);
CREATE INDEX idx_devices_mac ON devices(mac);

-- Таблица протоколов
CREATE TABLE protocols (
    id SERIAL PRIMARY KEY,
    device_id INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    port INTEGER,
    url TEXT,
    available BOOLEAN DEFAULT FALSE,
    detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, type, port)
);

CREATE INDEX idx_protocols_device_id ON protocols(device_id);
CREATE INDEX idx_protocols_type ON protocols(type);

-- Таблица RTSP потоков
CREATE TABLE rtsp_streams (
    id SERIAL PRIMARY KEY,
    device_id INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    codec VARCHAR(50),
    resolution VARCHAR(20),
    fps DECIMAL(5,2),
    bitrate INTEGER,
    audio_codec VARCHAR(50),
    channels INTEGER,
    available BOOLEAN DEFAULT FALSE,
    checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rtsp_streams_device_id ON rtsp_streams(device_id);
CREATE INDEX idx_rtsp_streams_available ON rtsp_streams(available);

-- Таблица сканирований
CREATE TABLE scans (
    id SERIAL PRIMARY KEY,
    subnet VARCHAR(50) NOT NULL,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP,
    devices_found INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'running',
    config JSONB
);

CREATE INDEX idx_scans_started_at ON scans(started_at);
CREATE INDEX idx_scans_status ON scans(status);

-- Связь сканирований и устройств
CREATE TABLE scan_devices (
    scan_id INTEGER NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    device_id INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (scan_id, device_id)
);

CREATE INDEX idx_scan_devices_scan_id ON scan_devices(scan_id);
CREATE INDEX idx_scan_devices_device_id ON scan_devices(device_id);
```

---

## Работа с данными

### Текущая реализация

#### Device Registry (в памяти)

```go
type DeviceRegistry struct {
    devices map[string]*Device
    mu      sync.RWMutex
}

func NewDeviceRegistry() *DeviceRegistry {
    return &DeviceRegistry{
        devices: make(map[string]*Device),
    }
}

func (r *DeviceRegistry) Add(device *Device) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.devices[device.IP] = device
}

func (r *DeviceRegistry) Get(ip string) (*Device, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    device, exists := r.devices[ip]
    return device, exists
}

func (r *DeviceRegistry) GetAll() []Device {
    r.mu.RLock()
    defer r.mu.RUnlock()
    devices := make([]Device, 0, len(r.devices))
    for _, device := range r.devices {
        devices = append(devices, *device)
    }
    return devices
}
```

### Будущая реализация с БД

#### Интерфейс хранилища

```go
type Storage interface {
    // Устройства
    SaveDevice(ctx context.Context, device *Device) error
    GetDevice(ctx context.Context, ip string) (*Device, error)
    ListDevices(ctx context.Context, filters DeviceFilters) ([]Device, error)
    UpdateDevice(ctx context.Context, device *Device) error
    DeleteDevice(ctx context.Context, ip string) error
    
    // Протоколы
    SaveProtocol(ctx context.Context, deviceIP string, protocol *Protocol) error
    GetProtocols(ctx context.Context, deviceIP string) ([]Protocol, error)
    
    // RTSP потоки
    SaveRTSPStream(ctx context.Context, deviceIP string, stream *RTSPStreamInfo) error
    GetRTSPStreams(ctx context.Context, deviceIP string) ([]RTSPStreamInfo, error)
    
    // Сканирования
    SaveScan(ctx context.Context, scan *Scan) error
    GetScan(ctx context.Context, id int) (*Scan, error)
    ListScans(ctx context.Context, limit, offset int) ([]Scan, error)
}
```

---

## Миграции (для будущей БД)

### Использование migrate или golang-migrate

```sql
-- migrations/001_initial_schema.up.sql
CREATE TABLE devices (...);
CREATE TABLE protocols (...);
CREATE TABLE rtsp_streams (...);
CREATE TABLE scans (...);
CREATE TABLE scan_devices (...);

-- migrations/001_initial_schema.down.sql
DROP TABLE scan_devices;
DROP TABLE scans;
DROP TABLE rtsp_streams;
DROP TABLE protocols;
DROP TABLE devices;
```

---

## Кэширование

### Текущая реализация

- **In-memory кэш** — результаты сканирования хранятся в памяти
- **Время жизни:** до завершения программы

### Будущая реализация

- **Redis** (опционально) — для кэширования часто запрашиваемых данных
- **TTL кэш** — автоматическое истечение старых данных
- **Кэш устройств** — кэширование информации об устройствах на 1 час

---

## Резервное копирование

### Текущая реализация

- **Экспорт в файлы** — пользователь может сохранить результаты сканирования
- **Формат:** JSON, CSV, XML, YAML

### Будущая реализация

- **Автоматический экспорт** — регулярный экспорт данных в файлы
- **Бэкапы БД** — автоматическое резервное копирование SQLite/PostgreSQL
- **Экспорт в облако** — опциональная отправка бэкапов в облачное хранилище

---

## Производительность

### Оптимизации

1. **Индексы** — на часто используемых полях (ip, mac, type)
2. **Пакетная вставка** — bulk insert для множественных записей
3. **Connection pooling** — пул соединений с БД
4. **Кэширование** — кэш часто запрашиваемых данных

### Ожидаемые объемы данных

- **Устройства:** до 1000 устройств в сети
- **Протоколы:** 3-5 протоколов на устройство
- **RTSP потоки:** 1-4 потока на устройство
- **Сканирования:** история до 1000 сканирований

---

## Безопасность данных

### Текущая реализация

- **Нет чувствительных данных** — только IP адреса и техническая информация
- **Локальное хранение** — данные не передаются в интернет

### Будущая реализация

- **Шифрование БД** — для SQLite с поддержкой шифрования
- **Аутентификация** — если будет веб-интерфейс
- **Очистка старых данных** — автоматическое удаление данных старше N дней

---

## Интеграция с внешними системами

### Экспорт данных

- **JSON API** — REST API для получения данных
- **Webhooks** — уведомления о новых устройствах
- **Экспорт в VMS** — интеграция с системами видеонаблюдения

### Импорт данных

- **Импорт из CSV** — загрузка списка устройств из файла
- **Импорт из других систем** — интеграция с существующими системами
