# Deployment: Развертывание проекта

## Текущий статус

🚧 **Проект в разработке** — развертывание будет доступно после реализации базовой функциональности.

---

## Требования к системе

### Минимальные требования

- **ОС:** Windows 10+, Linux (Ubuntu 20.04+, Debian 11+), macOS 11+
- **Архитектура:** x86_64, ARM64
- **Память:** 100 MB RAM
- **Диск:** 50 MB свободного места
- **Сеть:** Доступ к локальной сети для сканирования

### Дополнительные требования

- **FFmpeg** (опционально) — для детального анализа RTSP потоков
  - Windows: скачать с [ffmpeg.org](https://ffmpeg.org/download.html)
  - Linux: `sudo apt install ffmpeg` или `sudo yum install ffmpeg`
  - macOS: `brew install ffmpeg`

---

## Сборка проекта

### Из исходного кода

#### Предварительные требования
- Go 1.21 или выше
- Git

#### Шаги сборки

```bash
# 1. Клонировать репозиторий
git clone https://github.com/yourusername/local-video-server.git
cd local-video-server

# 2. Установить зависимости
go mod download

# 3. Собрать проект
go build -o local-video-server ./cmd/server

# Или использовать Makefile
make build
```

### Кроссплатформенная сборка

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o local-video-server-linux-amd64 ./cmd/server

# Windows
GOOS=windows GOARCH=amd64 go build -o local-video-server-windows-amd64.exe ./cmd/server

# macOS
GOOS=darwin GOARCH=amd64 go build -o local-video-server-darwin-amd64 ./cmd/server

# ARM64 (Linux)
GOOS=linux GOARCH=arm64 go build -o local-video-server-linux-arm64 ./cmd/server
```

### Использование Makefile

```bash
# Сборка для текущей платформы
make build

# Сборка для всех платформ
make build-all

# Очистка
make clean

# Установка зависимостей
make deps

# Запуск тестов
make test
```

---

## Развертывание

### Локальное развертывание (разработка)

#### Windows

```powershell
# 1. Скачать или собрать бинарный файл
# 2. Поместить в папку (например, C:\local-video-server\)
# 3. Добавить в PATH (опционально)
# 4. Запустить
.\local-video-server.exe scan
```

#### Linux

```bash
# 1. Скачать или собрать бинарный файл
# 2. Сделать исполняемым
chmod +x local-video-server

# 3. Переместить в /usr/local/bin (опционально)
sudo mv local-video-server /usr/local/bin/

# 4. Запустить
local-video-server scan
```

#### macOS

```bash
# 1. Скачать или собрать бинарный файл
# 2. Сделать исполняемым
chmod +x local-video-server

# 3. Переместить в /usr/local/bin (опционально)
sudo mv local-video-server /usr/local/bin/

# 4. Запустить
local-video-server scan
```

### Продакшен развертывание

#### Вариант 1: Системный сервис (Linux)

Создать systemd service:

```ini
# /etc/systemd/system/local-video-server.service
[Unit]
Description=Local Video Server - Camera Discovery Tool
After=network.target

[Service]
Type=simple
User=video-server
WorkingDirectory=/opt/local-video-server
ExecStart=/opt/local-video-server/local-video-server serve
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Активация:

```bash
sudo systemctl daemon-reload
sudo systemctl enable local-video-server
sudo systemctl start local-video-server
sudo systemctl status local-video-server
```

#### Вариант 2: Docker контейнер

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o local-video-server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates ffmpeg
WORKDIR /root/

COPY --from=builder /app/local-video-server .
COPY --from=builder /app/configs/config.yaml ./configs/

CMD ["./local-video-server"]
```

Сборка и запуск:

```bash
docker build -t local-video-server .
docker run --network=host local-video-server scan
```

#### Вариант 3: Windows Service

Использовать NSSM (Non-Sucking Service Manager):

```powershell
# Установить NSSM
# Скачать с https://nssm.cc/download

# Создать сервис
nssm install LocalVideoServer "C:\local-video-server\local-video-server.exe"
nssm set LocalVideoServer AppDirectory "C:\local-video-server"
nssm set LocalVideoServer AppParameters "serve"

# Запустить сервис
nssm start LocalVideoServer
```

---

## Конфигурация

### Конфигурационный файл (config.yaml)

```yaml
# configs/config.yaml
server:
  host: "0.0.0.0"
  port: 8080

scanner:
  subnet: "auto"  # "auto" для автоматического определения
  timeout: 5s
  max_concurrent: 50
  ports:
    - 554   # RTSP
    - 1935  # RTMP
    - 80    # HTTP
    - 8080  # HTTP альтернативный

protocols:
  rtsp:
    enabled: true
    timeout: 10s
  rtmp:
    enabled: true
    timeout: 5s
  hls:
    enabled: true
    timeout: 5s
  onvif:
    enabled: true
    timeout: 10s
  upnp:
    enabled: true
    timeout: 5s

rtsp:
  check_streams: true
  use_ffmpeg: true
  ffmpeg_path: "ffmpeg"  # или полный путь
  ffprobe_path: "ffprobe"

export:
  default_format: "json"
  output_dir: "./exports"

logging:
  level: "info"  # debug, info, warn, error
  format: "json"  # json, text
  output: "stdout"  # stdout, file
  file: "./logs/local-video-server.log"
```

### Переменные окружения

```bash
# Переопределение конфигурации через переменные окружения
export LOCAL_VIDEO_SERVER_LOG_LEVEL=debug
export LOCAL_VIDEO_SERVER_SCANNER_TIMEOUT=10s
export LOCAL_VIDEO_SERVER_EXPORT_FORMAT=csv
```

---

## Сетевая конфигурация

### Требования к сети

- **Доступ к локальной сети** — приложение должно иметь доступ к подсети для сканирования
- **Права администратора** (опционально) — для ARP сканирования на некоторых системах
- **Firewall** — может потребоваться разрешение для исходящих соединений

### Настройка Firewall

#### Linux (iptables)

```bash
# Разрешить исходящие соединения
sudo iptables -A OUTPUT -p tcp --dport 554 -j ACCEPT
sudo iptables -A OUTPUT -p tcp --dport 1935 -j ACCEPT
sudo iptables -A OUTPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A OUTPUT -p tcp --dport 8080 -j ACCEPT
sudo iptables -A OUTPUT -p udp --dport 3702 -j ACCEPT  # ONVIF
```

#### Windows Firewall

```powershell
# Разрешить приложение через Windows Firewall
New-NetFirewallRule -DisplayName "Local Video Server" -Direction Outbound -Program "C:\local-video-server\local-video-server.exe" -Action Allow
```

---

## Мониторинг и логи

### Логирование

Логи сохраняются в:
- **Консоль** — по умолчанию
- **Файл** — если настроено в конфигурации

```yaml
logging:
  level: "info"
  format: "json"
  output: "file"
  file: "/var/log/local-video-server.log"
```

### Проверка работы

```bash
# Проверить статус (если запущен как сервис)
sudo systemctl status local-video-server

# Посмотреть логи
tail -f /var/log/local-video-server.log

# Проверить процесс
ps aux | grep local-video-server
```

---

## Обновление

### Обновление бинарного файла

```bash
# 1. Остановить сервис (если запущен)
sudo systemctl stop local-video-server

# 2. Скачать новую версию
wget https://github.com/yourusername/local-video-server/releases/latest/local-video-server

# 3. Заменить старый файл
sudo mv local-video-server /opt/local-video-server/

# 4. Запустить сервис
sudo systemctl start local-video-server
```

### Обновление через Git

```bash
# 1. Остановить сервис
sudo systemctl stop local-video-server

# 2. Обновить код
cd /opt/local-video-server
git pull

# 3. Пересобрать
go build -o local-video-server ./cmd/server

# 4. Запустить сервис
sudo systemctl start local-video-server
```

---

## Откат (Rollback)

### Восстановление предыдущей версии

```bash
# 1. Остановить сервис
sudo systemctl stop local-video-server

# 2. Восстановить предыдущую версию из бэкапа
sudo cp /opt/local-video-server/backup/local-video-server-previous /opt/local-video-server/local-video-server

# 3. Запустить сервис
sudo systemctl start local-video-server
```

---

## Безопасность

### Рекомендации

1. **Запуск от непривилегированного пользователя** — не запускать от root
2. **Ограничение доступа к конфигурации** — права доступа только для владельца
3. **Регулярные обновления** — обновлять до последней версии
4. **Мониторинг логов** — следить за подозрительной активностью
5. **Ограничение сетевого доступа** — только к локальной сети

### Создание пользователя (Linux)

```bash
# Создать пользователя
sudo useradd -r -s /bin/false local-video-server

# Установить права
sudo chown -R local-video-server:local-video-server /opt/local-video-server
```

---

## Troubleshooting

### Проблемы с запуском

**Проблема:** Приложение не запускается
- Проверить наличие бинарного файла
- Проверить права на выполнение: `chmod +x local-video-server`
- Проверить зависимости: `ldd local-video-server` (Linux)

**Проблема:** Ошибка "permission denied"
- Проверить права доступа к файлам
- Запустить с правами администратора (если требуется для сетевого сканирования)

### Проблемы с сетью

**Проблема:** Не находит устройства в сети
- Проверить доступность сети: `ping 192.168.1.1`
- Проверить firewall настройки
- Проверить, что приложение запущено на правильном интерфейсе

**Проблема:** Медленное сканирование
- Увеличить `max_concurrent` в конфигурации
- Уменьшить `timeout` для быстрых сетей
- Проверить загрузку сети

### Проблемы с FFmpeg

**Проблема:** FFmpeg не найден
- Установить FFmpeg: `sudo apt install ffmpeg`
- Указать полный путь в конфигурации: `ffmpeg_path: "/usr/bin/ffmpeg"`

---

## CI/CD (будущее)

### GitHub Actions

```yaml
# .github/workflows/build.yml
name: Build and Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go build -o local-video-server ./cmd/server
      - uses: actions/upload-artifact@v3
        with:
          name: local-video-server
          path: local-video-server
```

---

## Контакты и поддержка

- **Issues:** [GitHub Issues](https://github.com/yourusername/local-video-server/issues)
- **Документация:** [docs/](docs/)
- **Email:** support@example.com

