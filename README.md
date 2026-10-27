# WarpScout Web

Веб-интерфейс для поиска рабочих эндпоинтов Cloudflare WARP: запускается на Ubuntu-сервере или в Docker, управляется из браузера.
Сканирование идёт на сервере, аккаунт WARP и история хранятся там же. Отдельно можно проверить, работает ли через найденный
эндпоинт Telegram.

[English version](README_EN.md)

> Независимый проект, основанный на открытом коде других авторов (см. [Источники](#источники-и-благодарности)).
> Он не является официальным релизом и не связан с авторами исходных проектов.

## Возможности

- Пресеты Standard, Durable и Full; протоколы WireGuard, AmneziaWG, MASQUE H3 и MASQUE H2; IPv4/IPv6.
- Фильтры по странам и узлам (включая исключения), свои диапазоны, MTU, DNS, проверка скорости, перебор портов, WARP-in-WARP.
- Поиск параметров AWG junk / I1 и MASQUE SNI; результаты сохраняются, их можно применить к сканированию одной кнопкой.
- **Проверка Telegram:** через каждый туннель отправляется настоящий MTProto-запрос ко всем 5 дата-центрам Telegram.
  Адрес засчитывается, только когда ответили все пять; такие адреса идут первыми. Режим «Только с Telegram» оставляет лишь их.
- Живая таблица результатов, экспорт конфигов (WireGuard, AmneziaWG, usque, Mihomo), текстовый отчёт, CSV, копирование лучшего адреса.
- История сканирований с восстановлением всех параметров формы.
- SOCKS5-прокси для проверки одного эндпоинта (слушает только loopback сервера).
- Сканирование продолжается на сервере, если закрыть вкладку; прогресс и результат видны после возврата.
- Русский и английский интерфейс.

## Быстрый старт

### Docker

```bash
git clone https://github.com/smersh16/warpscout-web warpscout-web && cd warpscout-web
echo "WARPSCOUT_TOKEN=$(openssl rand -hex 24)" > .env
docker compose up -d --build
```

Откройте `http://127.0.0.1:8787/` и введите токен из `.env`. Данные лежат в томе `warpscout-data`.
По умолчанию порт опубликован только на `127.0.0.1`; для доступа с другого компьютера используйте SSH-туннель
(`ssh -L 8787:127.0.0.1:8787 user@сервер`) или HTTPS-прокси.

### Ubuntu (systemd)

```bash
sudo ./deploy/ubuntu/install.sh
```

Скрипт ставит Go, собирает бинарник и создаёт сервис `warpscout-web`. Подробности, HTTPS и настройки безопасности:
[deploy/ubuntu/README.md](deploy/ubuntu/README.md).

### Попробовать без сети и аккаунта

```bash
go run ./cmd/warpscout-web -listen 127.0.0.1:8787 -no-auth -demo
```

Демо-режим подменяет ядро имитацией, чтобы можно было посмотреть интерфейс.

## Безопасность

Слушает `127.0.0.1` по умолчанию; доступ по токену, серверные сессии на 7 дней, блокировка подбора токена по IP,
повторный ввод токена при экспорте аккаунта (там приватный ключ), защита от CSRF. По обычному HTTP токен идёт открытым текстом,
поэтому для внешнего доступа нужен HTTPS. Подробнее: [deploy/ubuntu/README.md](deploy/ubuntu/README.md#безопасность).

## Как устроено

```
браузер  ⇄  cmd/warpscout-web + internal/webserver   HTTP, SSE, веб-интерфейс, сессии, хранилище
                   ⇄  mobileapi.Start / Cancel        JSON-протокол ядра (тот же, что у Android-приложения)
                         ⇄  core.Engine, core.ScanSpec   операции, пресеты, ограничение «одна операция за раз»
                               ⇄  internal/warpscout      движок сканирования WARPSCOUT + проверка Telegram
```

| Путь | Назначение |
| --- | --- |
| `cmd/warpscout-web`, `internal/webserver` | веб-сервер и встроенный интерфейс (новый код) |
| `core` | типы, движок операций и `ScanSpec.Resolve` с правилами пресетов (из warpscout-android, расширено) |
| `mobileapi` | JSON-мост к ядру (из warpscout-android) |
| `internal/warpscout` | движок сканирования (из WARPSCOUT; `telegram.go` из warpscout-tg) |
| `deploy/ubuntu`, `Dockerfile.web`, `docker-compose.yml` | установка и запуск |
| `main.go`, `Dockerfile.cli` | консольная версия WARPSCOUT, как в исходном проекте |

Разработка: `go test ./...`. Подробности о происхождении кода и порядке синхронизации: [UPSTREAM.md](UPSTREAM.md).

## Источники и благодарности

Проект построен на открытом коде. Все три источника распространяются по лицензии MIT, их уведомления об авторских правах сохранены
в [LICENSE](LICENSE) и [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

| Проект | Автор | Что взято | Где в этом репозитории |
| --- | --- | --- | --- |
| [vernette/warpscout](https://github.com/vernette/warpscout) | Nikita S. (@vernette) | Исходный движок: регистрация WARP-аккаунта, сканирование, туннели, поиск junk/SNI, генерация конфигов, SOCKS | `internal/warpscout` (база v0.16.1 и точечные правки из master), `main.go` |
| [openwarpkit/warpscout-android](https://github.com/openwarpkit/warpscout-android) | OpenWarpKit | Основа этого репозитория: оболочка вокруг движка для нескольких клиентов (JSON-протокол, `core`, `mobileapi`, мобильный бэкенд, автоподбор I1) | `core`, `mobileapi`, `internal/warpscout/mobile_backend.go`, `scan_i1.go`, `junk_i1.go` |
| [niklzz/warpscout-tg](https://github.com/niklzz/warpscout-tg) | niklzz | Идея и код проверки доступности Telegram (MTProto `req_pq_multi` ко всем DC) | `internal/warpscout/telegram.go` |

Также используются: [amneziawg-go](https://github.com/amnezia-vpn/amneziawg-go), [usque](https://github.com/Diniboy1123/usque),
[WireGuard Go](https://github.com/WireGuard/wireguard-go), [gVisor](https://github.com/google/gvisor),
[Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss),
[quic-go](https://github.com/quic-go/quic-go) и другие модули из [go.mod](go.mod); полный список с лицензиями в
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

### Что изменено относительно источников

- Добавлено: веб-сервер и интерфейс, серверные сессии и защита входа, история и хранилище аккаунта, развёртывание в Docker и на Ubuntu.
- `core`: `ScanSpec` и `Resolve` (правила пресетов на сервере, один источник истины), опции и результат проверки Telegram.
- `internal/warpscout`: перенос проверки Telegram из warpscout-tg в более новую базу (форк построен на старой версии, поэтому вручную),
  порядок результатов с учётом Telegram, фильтр «Только с Telegram», точечные правки из master WARPSCOUT.
- Не перенесено: Android-приложение, флаги `-tg` и `-tg-only` консольной версии и колонки Telegram в её таблицах.

Android-приложение, на котором основана эта оболочка, живёт в [openwarpkit/warpscout-android](https://github.com/openwarpkit/warpscout-android);
консольные бинарники WARPSCOUT берите в [релизах оригинала](https://github.com/vernette/warpscout/releases).

## Ответственное использование

Инструмент подбирает эндпоинты Cloudflare WARP. Используйте его в соответствии с условиями Cloudflare и законами вашей страны.

## Лицензия

MIT, см. [LICENSE](LICENSE): уведомление об авторских правах оригинального проекта (Nikita S.) сохранено, добавлено имя владельца этого репозитория; собственные дополнения этого репозитория
распространяются на тех же условиях.
