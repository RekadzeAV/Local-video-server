# Архитектура Vigilos CE 2.0

## Компоненты
- **Бэкенд (Go 1.21+)**
  - `cmd/vigilos-core` — основной сервис.
  - `cmd/vigilos-validator` — проверка редакций/ключей.
  - `cmd/vigilos-migrate` — миграции БД.
  - `internal/camera` — менеджер камер, RTSP клиент, ONVIF discovery.
  - `internal/media` — прокси/транскод, HLS генератор, WebRTC relay.
  - `internal/storage` — локальная запись, SMB/NFS клиенты.
  - `internal/events` — движок правил, процессор, webhook sender.
  - `internal/api` — REST v1, WebSocket server, middleware валидации.
  - `internal/licensing` — validator, feature gates, watermark.

- **Аналитика (C++)**
  - `analytics/include` — motion_detector, anpr_processor, frame_buffer.
  - `analytics/src/motion` — MOG2/CNT, ROI manager, event generator.
  - `analytics/src/anpr` — YOLOv8n ONNX, OCR (Tesseract/CNN), regional matcher.
  - `analytics/src/bridge` — CGO bridge, gRPC service.
  - `analytics/models` — ONNX веса, региональные паттерны.
  - `analytics/test` — unit/integration.

- **Веб (React + TS, Vite)**
  - `web/src/app` — store (RTK), router, providers.
  - `web/src/features` — live-view, archive, admin, maps.
  - `web/src/themes` — default, NAS-стили (synology/qnap/asustor).
  - `web/src/shared` — базовые компоненты, hooks, utils.

- **Десктоп (Qt 6)**
  - gRPC клиент, видеостена, пресеты раскладок, платформенные адаптеры.

- **Мобильные (KMM)**
  - Shared слой (API, модели, репозитории, usecase, DI).
  - Android (Compose), iOS (SwiftUI) клиенты.

- **Infra**
  - `infra/scripts/build` — build-go (static), build-cpp, build-web, package-nas.
  - `infra/scripts/docker` — dev/prod образы.
  - `infra/packaging` — deb/rpm/windows/nas (synology/qnap/asustor).
  - `infra/ci/workflows` — pr-checks, build-develop, release.
  - `infra/scripts/tools` — license-generator, config-validator.

## Ключевые потоки
1. **Live/Record**: Camera → RTSP client → Media proxy (HLS/WebRTC) → Storage (локальный/NAS) → Web/Qt/KMM.
2. **Analytics**: Media proxy → CGO bridge → C++ motion/ANPR → Events engine → API/WebSocket/Web UI overlays.
3. **Licensing/Edition**: editions.yaml → validator → feature gates (лимит камер, БД, функции).
4. **Events**: motion/anpr → rules engine → actions (record clip, notify, webhook).

## Технологии
- Go: net/http, WebSocket, CGO bridge, sqlite/postgres, goose/migrate.
- C++: OpenCV (MOG2/CNT), ONNX Runtime, Tesseract, gRPC.
- Web: React 18, Vite, TS, RTK Query, MUI, Video.js, WebSocket.
- Desktop: Qt 6, gRPC client.
- Mobile: KMM, Koin, Compose, SwiftUI, ExoPlayer/AVPlayer.
- CI/CD: GitHub Actions, Docker multi-arch, static builds, Hadolint/ESLint/golangci-lint/clang-tidy.

## Диаграмма каталогов (сокр.)
```
vigilos/
├── cmd/
├── internal/
├── analytics/
├── web/
├── desktop/
├── mobile/
├── infra/
└── docs/
```

