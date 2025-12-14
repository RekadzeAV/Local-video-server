# Monitoring: Мониторинг и логирование

## Общие принципы

### 1. Уровни логирования

Приложение использует структурированное логирование с следующими уровнями:

- **DEBUG** — детальная информация для отладки
- **INFO** — общая информация о работе приложения
- **WARN** — предупреждения о потенциальных проблемах
- **ERROR** — ошибки, требующие внимания
- **FATAL** — критические ошибки, приводящие к остановке приложения

### 2. Форматы логов

#### JSON формат (рекомендуется для production):
```json
{
  "level": "info",
  "time": "2024-01-15T14:30:25Z",
  "message": "Начало сканирования сети",
  "subnet": "192.168.1.0/24",
  "component": "scanner"
}
```

#### Текстовый формат (для разработки):
```
2024-01-15T14:30:25Z [INFO] Начало сканирования сети subnet=192.168.1.0/24 component=scanner
```

---

## Конфигурация логирования

### 1. Через конфигурационный файл

```yaml
logging:
  level: "info"           # debug, info, warn, error
  format: "json"          # json, text
  output: "stdout"        # stdout, file
  file: "./logs/local-video-server.log"
  max_size: 100           # Максимальный размер файла в MB
  max_backups: 5          # Количество резервных копий
  max_age: 30             # Хранение логов в днях
  compress: true          # Сжатие старых логов
```

### 2. Через переменные окружения

```bash
# Уровень логирования
export LOG_LEVEL=debug

# Формат логов
export LOG_FORMAT=json

# Путь к файлу логов
export LOG_FILE=./logs/app.log

# Вывод в консоль
export LOG_OUTPUT=stdout
```

### 3. Через флаги командной строки

```bash
# Детальный вывод (DEBUG уровень)
./local-video-server scan --verbose

# Указание уровня логирования
./local-video-server scan --log-level warn

# Указание файла логов
./local-video-server scan --log-file ./logs/scan.log
```

---

## Расположение логов

### 1. По умолчанию

#### Windows:
```
C:\ProgramData\local-video-server\logs\local-video-server.log
```

#### Linux:
```
/var/log/local-video-server/local-video-server.log
```

#### macOS:
```
~/Library/Logs/local-video-server/local-video-server.log
```

### 2. Настраиваемое расположение

Логи могут быть сохранены в любую директорию, указанную в конфигурации:
```yaml
logging:
  file: "./logs/local-video-server.log"
```

### 3. Ротация логов

Логи автоматически ротируются при достижении максимального размера:
```
logs/
├── local-video-server.log          # Текущий лог
├── local-video-server.log.1        # Предыдущий лог
├── local-video-server.log.2.gz     # Сжатый старый лог
└── local-video-server.log.3.gz     # Еще более старый лог
```

---

## Структура логов

### 1. Поля логов

Каждая запись лога содержит следующие поля:

- **level** — уровень логирования
- **time** — временная метка (RFC3339)
- **message** — основное сообщение
- **component** — компонент системы (scanner, rtsp, onvif, etc.)
- **error** — детали ошибки (если есть)
- **context** — дополнительный контекст (IP адрес, порт, etc.)

### 2. Примеры логов

#### Успешное сканирование:
```json
{
  "level": "info",
  "time": "2024-01-15T14:30:25Z",
  "message": "Устройство обнаружено",
  "component": "scanner",
  "ip": "192.168.1.10",
  "manufacturer": "Hikvision",
  "model": "DS-2CD2342WD-I"
}
```

#### Ошибка подключения:
```json
{
  "level": "warn",
  "time": "2024-01-15T14:30:26Z",
  "message": "Не удалось подключиться к устройству",
  "component": "scanner",
  "ip": "192.168.1.11",
  "port": 554,
  "error": "Connection timeout",
  "timeout": "10s"
}
```

#### Критическая ошибка:
```json
{
  "level": "error",
  "time": "2024-01-15T14:30:27Z",
  "message": "Ошибка при сканировании сети",
  "component": "scanner",
  "error": "Network unreachable",
  "subnet": "192.168.1.0/24"
}
```

---

## Мониторинг работы приложения

### 1. Проверка статуса процесса

#### Windows:
```powershell
# Проверка запущенного процесса
Get-Process -Name "local-video-server" -ErrorAction SilentlyContinue

# Проверка через службу (если установлена)
Get-Service -Name "LocalVideoServer"
```

#### Linux:
```bash
# Проверка процесса
ps aux | grep local-video-server

# Проверка через systemd
systemctl status local-video-server

# Проверка портов (если есть HTTP API)
netstat -tuln | grep 8080
```

#### macOS:
```bash
# Проверка процесса
ps aux | grep local-video-server

# Проверка через launchd (если установлена)
launchctl list | grep local-video-server
```

### 2. Проверка использования ресурсов

#### Windows (PowerShell):
```powershell
Get-Process -Name "local-video-server" | Select-Object CPU, WorkingSet, Id
```

#### Linux:
```bash
# Использование CPU и памяти
top -p $(pgrep local-video-server)

# Или через htop
htop -p $(pgrep local-video-server)

# Краткая информация
ps -p $(pgrep local-video-server) -o pid,%cpu,%mem,rss,vsz,etime,cmd
```

### 3. Проверка сетевой активности

#### Linux:
```bash
# Мониторинг сетевых соединений
netstat -anp | grep local-video-server

# Или через ss
ss -tunp | grep local-video-server

# Мониторинг трафика
iftop -i eth0 -f "port 554 or port 1935 or port 80"
```

---

## Анализ логов

### 1. Просмотр логов в реальном времени

#### Linux/macOS:
```bash
# Просмотр последних записей
tail -f /var/log/local-video-server/local-video-server.log

# С фильтрацией по уровню
tail -f /var/log/local-video-server/local-video-server.log | grep ERROR

# С цветным выводом (если установлен ccze)
tail -f /var/log/local-video-server/local-video-server.log | ccze -A
```

#### Windows (PowerShell):
```powershell
# Просмотр последних записей
Get-Content .\logs\local-video-server.log -Wait -Tail 50

# С фильтрацией
Get-Content .\logs\local-video-server.log -Wait | Select-String "ERROR"
```

### 2. Поиск в логах

#### Linux:
```bash
# Поиск ошибок
grep -i error /var/log/local-video-server/local-video-server.log

# Поиск по IP адресу
grep "192.168.1.10" /var/log/local-video-server/local-video-server.log

# Поиск за определенный период
grep "2024-01-15" /var/log/local-video-server/local-video-server.log

# Подсчет ошибок
grep -c "ERROR" /var/log/local-video-server/local-video-server.log
```

#### Windows (PowerShell):
```powershell
# Поиск ошибок
Select-String -Path .\logs\local-video-server.log -Pattern "ERROR"

# Поиск по IP адресу
Select-String -Path .\logs\local-video-server.log -Pattern "192.168.1.10"

# Подсчет ошибок
(Select-String -Path .\logs\local-video-server.log -Pattern "ERROR").Count
```

### 3. Анализ JSON логов

#### Использование jq (Linux/macOS):
```bash
# Все ошибки
cat local-video-server.log | jq 'select(.level == "error")'

# Ошибки за последний час
cat local-video-server.log | jq 'select(.level == "error" and .time > "2024-01-15T13:00:00Z")'

# Группировка по компонентам
cat local-video-server.log | jq -r '.component' | sort | uniq -c

# Статистика по уровням
cat local-video-server.log | jq -r '.level' | sort | uniq -c
```

---

## Метрики и статистика

### 1. Встроенные метрики

Приложение может выводить статистику работы:

```json
{
  "level": "info",
  "time": "2024-01-15T14:35:00Z",
  "message": "Статистика сканирования",
  "component": "scanner",
  "stats": {
    "total_hosts": 256,
    "scanned_hosts": 256,
    "devices_found": 5,
    "cameras_found": 5,
    "errors": 2,
    "duration_seconds": 45.3,
    "avg_time_per_host": 0.18
  }
}
```

### 2. Экспорт метрик

Метрики могут быть экспортированы в различные форматы:

#### Prometheus формат (если будет реализовано):
```
# HELP local_video_server_devices_total Total number of devices found
# TYPE local_video_server_devices_total counter
local_video_server_devices_total 5

# HELP local_video_server_scan_duration_seconds Duration of network scan
# TYPE local_video_server_scan_duration_seconds histogram
local_video_server_scan_duration_seconds_bucket{le="10"} 0
local_video_server_scan_duration_seconds_bucket{le="30"} 0
local_video_server_scan_duration_seconds_bucket{le="60"} 1
```

---

## Алерты и уведомления

### 1. Критические ошибки

Следующие события должны вызывать немедленное внимание:

- **FATAL** — критические ошибки, приводящие к остановке
- **ERROR** с компонентом "scanner" — проблемы со сканированием сети
- **ERROR** с компонентом "rtsp" — проблемы с проверкой RTSP потоков
- Множественные **WARN** за короткий период

### 2. Мониторинг через внешние системы

#### Интеграция с системными логами:

##### Linux (rsyslog):
```
# /etc/rsyslog.d/local-video-server.conf
$ModLoad imfile
$InputFilePollInterval 10
$InputFileName /var/log/local-video-server/local-video-server.log
$InputFileTag local-video-server:
$InputFileStateFile local-video-server-state
$InputFileSeverity error
$InputFileFacility local0
$InputRunFileMonitor
```

##### Windows (Event Log):
Логи могут быть интегрированы с Windows Event Log через соответствующие библиотеки.

---

## Отладка

### 1. Включение DEBUG режима

#### Через конфигурацию:
```yaml
logging:
  level: "debug"
```

#### Через переменную окружения:
```bash
export LOG_LEVEL=debug
```

#### Через флаг:
```bash
./local-video-server scan --verbose
```

### 2. Типичные проблемы и их диагностика

#### Проблема: Медленное сканирование

**Диагностика:**
```bash
# Проверка логов на таймауты
grep "timeout" local-video-server.log

# Проверка количества параллельных операций
grep "max_concurrent" local-video-server.log
```

**Решение:**
- Увеличить таймауты в конфигурации
- Уменьшить `max_concurrent` если сеть перегружена
- Проверить доступность устройств

#### Проблема: Устройства не обнаруживаются

**Диагностика:**
```bash
# Проверка ошибок сканирования
grep "ERROR\|WARN" local-video-server.log | grep scanner

# Проверка сетевых ошибок
grep "Connection\|Network" local-video-server.log
```

**Решение:**
- Проверить настройки firewall
- Убедиться, что подсеть указана правильно
- Проверить доступность устройств вручную

#### Проблема: RTSP потоки не проверяются

**Диагностика:**
```bash
# Проверка ошибок RTSP
grep "rtsp\|RTSP" local-video-server.log | grep ERROR

# Проверка наличия FFmpeg
grep "ffmpeg\|FFmpeg" local-video-server.log
```

**Решение:**
- Убедиться, что FFmpeg установлен и доступен
- Проверить учетные данные для RTSP
- Проверить доступность RTSP портов

---

## Рекомендации по мониторингу

### 1. Production окружение

- Используйте уровень логирования **INFO** или **WARN**
- Используйте **JSON формат** для удобного парсинга
- Настройте **ротацию логов** для предотвращения переполнения диска
- Мониторьте размер лог-файлов
- Настройте алерты на критические ошибки

### 2. Development окружение

- Используйте уровень логирования **DEBUG**
- Используйте **текстовый формат** для удобства чтения
- Выводите логи в **stdout** для просмотра в реальном времени

### 3. Тестирование

- Используйте уровень логирования **DEBUG**
- Сохраняйте логи тестовых запусков для анализа
- Используйте отдельные лог-файлы для каждого теста

---

## Интеграция с системами мониторинга

### 1. Prometheus (будущее)

Если будет реализован HTTP API с метриками:
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'local-video-server'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### 2. Grafana (будущее)

Дашборды для визуализации:
- Количество обнаруженных устройств
- Время сканирования
- Количество ошибок
- Использование ресурсов

### 3. ELK Stack (будущее)

Интеграция с Elasticsearch, Logstash, Kibana:
- Централизованное хранение логов
- Поиск и анализ логов
- Визуализация через Kibana

---

## Резюме

### Основные команды для мониторинга:

```bash
# Просмотр логов в реальном времени
tail -f /var/log/local-video-server/local-video-server.log

# Поиск ошибок
grep ERROR /var/log/local-video-server/local-video-server.log

# Проверка статуса (Linux)
systemctl status local-video-server

# Проверка использования ресурсов
ps aux | grep local-video-server
```

### Ключевые файлы:

- **Логи:** `/var/log/local-video-server/local-video-server.log`
- **Конфигурация:** `/etc/local-video-server/config.yaml`
- **Экспорты:** `./exports/` (настраивается)

### Рекомендации:

1. ✅ Регулярно проверяйте логи на наличие ошибок
2. ✅ Настройте ротацию логов
3. ✅ Мониторьте использование ресурсов
4. ✅ Используйте структурированное логирование (JSON) в production
5. ✅ Настройте алерты на критические ошибки
