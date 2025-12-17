# Процессы и стандарты разработки

## Git workflow
- Ветки: `main` (релизы), `develop` (интеграция), `release/*`, `hotfix/*`, `feature/*`, `bugfix/*`, `refactor/*`.
- PR требования: все тесты, покрытия не падает, линтеры: golangci-lint, clang-tidy, ESLint+Prettier, Hadolint; документация и CHANGELOG обновлены.

## Тестирование
1. Юнит (≥75%): Go testify, C++ GTest, JS/TS Jest+RTL.
2. Интеграция: docker-compose среды, API тесты (Postman/Newman), БД.
3. E2E: Playwright (web), Squish/Qt Test (desktop), Maestro/Appium (mobile).
4. Нагрузка: k6 для API, RTSP симуляторы.

## Документация
- `docs/api/openapi.yaml`, примеры.
- `docs/user` — getting-started, administration, troubleshooting.
- `docs/developer` — architecture, building, contributing.
- `docs/legal` — LICENSE, CLA, EULA.
- Инструменты: Docusaurus, Swagger UI, Mermaid, MkDocs.

## Code quality gates
- Линтеры: golangci-lint, clang-tidy, ESLint+Prettier, Hadolint.
- Безопасность: trivy/snyk для зависимостей.
- CI: PR checks + build-develop + release (GitHub Actions).

## Поддерживаемые платформы
- Бэкенд: Linux (Ubuntu 20.04+, Debian 11+, CentOS 8+), Windows Server 2019+/10/11, macOS 11+ (dev).
- Клиенты: Web (Chrome 90+/Firefox 88+/Safari 14+), Desktop (Win10+/macOS11+/Linux Qt), Mobile (Android 8+/iOS14+).

## Порты и сеть
- HTTP 80 → redirect HTTPS 443 (web).
- RTSP 554 (ingest), gRPC 50051 (внутри), PostgreSQL 5432, SMB 445, NFS 2049.
- Firewall: разрешить исходящие SMB/NFS и входящие основные сервисы.

