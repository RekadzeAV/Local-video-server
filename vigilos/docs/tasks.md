# Декомпозиция задач

## Фаза 0 — Инфраструктура и прототип (3 недели)
- Монорепо структура, .gitignore, git hooks.
- CI/CD: PR checks (lint-go, build-core, test-unit), build-develop, release.
- Прототип CGO моста Go↔C++ (передача кадров, бенчмарки).
- MVP валидации редакций: editions.yaml, validator, тестовые ключи.
- Документация разработчика: CONTRIBUTING, ARCHITECTURE, DEVELOPMENT.

## Фаза 1 — Базовый жизненный цикл видео (8 недель)
**Backend**
- Sprint 1.1: RTSP клиент с переподключением; H.264/H.265 профили; HLS 2s сегменты; REST /api/v1/cameras; WS статус камер.
- Sprint 1.2: Циклическая запись; SQLite (Lite), PostgreSQL (Standard); миграции; архив по диапазону.
- Sprint 1.3: JWT auth; RBAC (viewer/operator/admin); middleware прав; /api/v1/auth/login, /refresh.
- Sprint 1.4: Парсинг editions; лимитер камер; статистика использования; /api/v1/system/edition.

**Web**
- Sprint 1.5: Vite+React+TS; RTK store; router; базовые MUI; light/dark.
- Sprint 1.6: VideoPlayer (Video.js); сетка 1/4/9/16; PTZ ONVIF; WS обновления.
- Sprint 1.7: Таймлайн; календарь; HLS playback; экспорт MP4.
- Sprint 1.8: Добавление камер; системные настройки; responsive; интеграция API.

**Критерии**: 2 камеры, запись 7 дней, live+архив в вебе, различимы Lite(2/SQLite) и Standard(8/Postgres), базовые тесты.

## Фаза 2 — Сетевые функции и аналитика (8 недель)
- Sprint 2.1: SMB/NFS клиенты, мониторинг доступности, стратегия запись cache→NAS.
- Sprint 2.2: Модель событий; правила if/then; действия: клип/уведомление/webhook; API правил.
- Sprint 2.3: C++ детекция (MOG2/CNT), ROI/masks, SSE/AVX, события с bbox.
- Sprint 2.4: CGO мост событий, протокол, конфиг через REST, визуализация ROI в вебе.
- Sprint 2.5: Оптимизация — downsampling, worker pool, CPU мониторинг, RPi4 бенчмарки.
- Sprint 2.6: KMM shared (API, модели, repo, usecase, DI).
- Sprint 2.7: Android (Compose, ExoPlayer, события, push).
- Sprint 2.8: iOS (SwiftUI, AVPlayer, события, push).
- Sprint 2.9: Docker multi-arch, static Go, static C++, сервисные скрипты.
- Sprint 2.10: NAS пакеты (Synology/QNAP/Asustor), GitHub Actions, эмуляторные тесты.

**Критерии**: запись на SMB/NFS, motion с ROI, мобильные live+events, авто сборка NAS пакетов, Full редакция (24 камеры, детекция, NAS).

## Фаза 3 — Профессиональные функции (8 недель)
- Sprint 3.1: YOLOv8n ONNX детектор номеров, препроцесс, фильтрация, регионы RU/EU/US/CN.
- Sprint 3.2: OCR (Tesseract 5), валидация регионов, геопривязка, экспорт CSV/API.
- Sprint 3.3: Очередь и кеш, метрики точности/производительности, конфиг через веб.
- Sprint 3.4: AD/LDAP (go-ldap-client), синхронизация, маппинг ролей, опц. Kerberos SSO.
- Sprint 3.5: Webhooks: шаблоны payload, ретри, лог ошибок, интеграция со СКУД.
- Sprint 3.6: Qt каркас, MDI, gRPC клиент, менеджер раскладок, нативный вид.
- Sprint 3.7: Видеостена: drag-n-drop, multi-monitor, пресеты, хоткеи.
- Sprint 3.8: Карты (Leaflet/OSM), координаты камер, coverage, экспорт KML/GeoJSON.
- Sprint 3.9: Отчеты: события, хранилище, PDF/Excel, дашборды.

**Критерии**: ANPR >85%, AD интеграция, Qt видеостены, камеры на карте, Expert/RegViD оформлены.

## Фаза 4 — Оптимизация, безопасность, релиз (6 недель)
- Sprint 4.1: Профилирование (pprof/valgrind), hot paths, пулы соединений, Redis cache (опц).
- Sprint 4.2: Нагрузка 100+ RTSP, слабое железо, NAS, минимальные требования.
- Sprint 4.3: Безопасность — инъекции, trivy/snyk, hardening сервисов, сетевые рекомендации.
- Sprint 4.4: Защита данных — шифрование архивов, маскирование PII, secure defaults, security.txt.
- Sprint 4.5: Документация — админ, пользователь, API (OpenAPI), разработчик.
- Sprint 4.6: Финал — инсталляторы всех платформ, NAS пакеты, RegViD ISO, Docker Hub.

**Критерии релиза**: 6 редакций собраны и протестированы, документация полна, CI/CD автоматизирован, лицензии готовы, анонс для сообщества.

