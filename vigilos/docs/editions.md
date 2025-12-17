# Редакции и ограничения

## Таблица редакций
| Функция | Lite | Standard | Full | Expert | Package (NAS) | RegViD |
|---------|------|----------|------|--------|---------------|--------|
| Макс. камер | 2 | 8 | 24 | ∞ | 8 | 32 |
| База данных | SQLite | PostgreSQL | PostgreSQL | PostgreSQL | SQLite | PostgreSQL + TimescaleDB |
| Детекция движения | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| ANPR | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| Сетевое хранилище | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| Active Directory | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| Десктоп-клиент | ❌ | ❌ | ✅ | ✅ | ❌ | ✅ |
| Веб-карты | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| NAS оптимизация | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ |
| Водяной знак | ❌ | ❌ | ❌ | ⚠️ | ❌ | ❌ |

## Требования к железу (минимум)
- **Lite**: 2 CPU, 2 ГБ RAM, 100 ГБ HDD.
- **Standard**: 4 CPU, 4 ГБ RAM, 1 ТБ HDD.
- **Package (NAS)**: 2+ CPU (x86_64/aarch64/armv7l), 2+ ГБ RAM, DSM 7+/QTS 5+/ADM 4+.
- **RegViD**: 8+ CPU, 16+ ГБ RAM, SSD 256 ГБ (система), HDD 4+ ТБ RAID.

## Лицензирование и контроль
- AGPLv3 + CLA + Non-Commercial оговорка.
- Конфиг `editions.yaml`: лимиты камер, БД, функции.
- `internal/licensing`: validator, feature gates, watermark (для Expert).

