# app-tunnel

Минимальный ngrok-подобный HTTP/HTTPS туннель на Go (stdlib). Два бинарника: сервер и клиент.

## Быстрый старт

1. Подготовьте переменные окружения.
2. Запустите сервер.
3. Запустите клиент и укажите локальный адрес сервиса.

Смотрите пример в `.env.example`.

## Ограничения текущей версии

- Только HTTP/HTTPS (HTTP/1.1), без TCP/UDP raw.
- Нет авторизации.
- TLS: только с уже подготовленным сертификатом. ACME/Let's Encrypt внутри не реализован (stdlib).
- Один запрос на один туннельный канал одновременно (конн-пулом).

## Запуск

Сервер:

```bash
CONTROL_ADDR=":8080" \
TUNNEL_ADDR=":8081" \
HTTP_ADDR=":8082" \
HTTPS_ADDR=":8443" \
TLS_CERT_FILE="/path/to/fullchain.pem" \
TLS_KEY_FILE="/path/to/privkey.pem" \
DOMAIN="example.com" \
SUBDOMAIN_STORE_PATH="./data/subdomains.txt" \
SUBDOMAIN_LENGTH="6" \
TUNNEL_TIMEOUT="30s" \
LOG_LEVEL="info" \
./bin/server
```

Клиент:

```bash
SERVER_CONTROL_URL="http://example.com:8080/register" \
SERVER_TUNNEL_ADDR="example.com:8081" \
LOCAL_FORWARD_ADDR="127.0.0.1:3000" \
REQUESTED_SUBDOMAIN="" \
CONN_POOL_SIZE="4" \
DIAL_TIMEOUT="10s" \
LOG_LEVEL="info" \
./bin/client
```

`CONN_POOL_SIZE` необязателен, по умолчанию используется `4`.

## Сборка из исходников

```bash
mkdir -p bin
GOCACHE=/tmp/go-build go build -o bin/server ./cmd/server
GOCACHE=/tmp/go-build go build -o bin/client ./cmd/client
```

## Быстрая проверка (smoke test)

1. Запустите простой локальный сервер на `3000`.
```bash
python3 -m http.server 3000
```

2. Запустите `server` и `client` как в разделе “Запуск”.

3. Возьмите поддомен из лога клиента (например `k5anw6.example.local`) и сделайте запрос:
```bash
curl -H "Host: k5anw6.example.local" http://127.0.0.1:8082/ -v
```

## Локальная проверка

1. Поднимите любой HTTP сервис локально (например на 3000).
2. Запустите сервер и клиент.
3. Сделайте запрос с нужным `Host` заголовком.

Пример:

```bash
curl -H "Host: <subdomain>.example.com" http://127.0.0.1:8082/
```

Если хотите обращаться без `Host` заголовка, добавьте поддомен в `hosts`.

## Docker

### Docker Compose (домен + SSL одной командой)

1. Укажите DNS:
- `A` запись для `example.com` на IP сервера.
- `A` wildcard запись для `*.example.com` на тот же IP.

2. Создайте `.env.compose`:

```bash
DOMAIN="example.com"
ACME_EMAIL="admin@example.com"
TUNNEL_PORT="8081"
SUBDOMAIN_LENGTH="6"
TUNNEL_TIMEOUT="30s"
LOG_LEVEL="info"
```

3. Запустите:

```bash
docker compose --env-file .env.compose up -d --build
```

4. Клиент подключается так:

```bash
SERVER_ADDR="example.com"
LOCAL_FORWARD_ADDR="127.0.0.1:3000"
```

По умолчанию клиент сам соберёт:
- `SERVER_CONTROL_URL=https://control.<DOMAIN>/register`
- `SERVER_TUNNEL_ADDR=<DOMAIN>:8081`

Опционально можно переопределить:
- `SERVER_CONTROL_URL` или `SERVER_CONTROL_HOST`/`SERVER_CONTROL_SCHEME`
- `SERVER_TUNNEL_ADDR` или `SERVER_TUNNEL_PORT`

Примечание:
- SSL завершается в Caddy (Let's Encrypt).
- Для `https://<subdomain>.example.com` используется on-demand TLS (сертификат выпускается при первом HTTPS запросе к новому поддомену).

### Сервер

```bash
docker build -f Dockerfile.server -t app-tunnel-server .
docker run --rm \
  -e CONTROL_ADDR=":8080" \
  -e TUNNEL_ADDR=":8081" \
  -e HTTP_ADDR=":8082" \
  -e HTTPS_ADDR="" \
  -e TLS_CERT_FILE="" \
  -e TLS_KEY_FILE="" \
  -e DOMAIN="example.com" \
  -e SUBDOMAIN_STORE_PATH="/app/subdomains.txt" \
  -e SUBDOMAIN_LENGTH="6" \
  -e TUNNEL_TIMEOUT="30s" \
  -e LOG_LEVEL="info" \
  -p 8080:8080 -p 8081:8081 -p 8082:8082 \
  app-tunnel-server
```

### Клиент

```bash
docker build -f Dockerfile.client -t app-tunnel-client .
docker run --rm \
  -e SERVER_ADDR="example.com" \
  -e LOCAL_FORWARD_ADDR="host.docker.internal:3000" \
  -e REQUESTED_SUBDOMAIN="" \
  -e DIAL_TIMEOUT="10s" \
  -e LOG_LEVEL="info" \
  app-tunnel-client
```

## Продакшен запуск

### DNS

- Создайте запись `*.example.com -> публичный IP сервера`.
- При необходимости добавьте `example.com -> публичный IP`.

### Порты и доступ

- `HTTP_ADDR` для внешнего HTTP трафика.
- `HTTPS_ADDR` для внешнего HTTPS трафика (если используете встроенный TLS).
- `CONTROL_ADDR` и `TUNNEL_ADDR` лучше закрыть фаерволом и открыть только доверенным источникам.

### TLS и сертификаты

Эта версия не реализует ACME. В проде рекомендуемый вариант:

- TLS завершать во внешнем прокси (Caddy/Traefik/nginx+certbot).
- До `app-tunnel` проксировать уже расшифрованный HTTP.

Если TLS завершается внутри:

- Обновляйте сертификаты отдельно и указывайте `TLS_CERT_FILE` и `TLS_KEY_FILE`.

### Системный сервис (systemd)

Рекомендуется оформить `server` и `client` как сервисы. Если используете Docker/compose, можно добавить соответствующие файлы позже.

## Поддержка и сопровождение

### Логи

Сервер и клиент пишут в stdout/stderr. Рекомендуется:

- направлять в journald или отдельные файлы,
- настроить ротацию.
Для подробных логов используйте `LOG_LEVEL="debug"`.

### Обновление поддоменов

Файл `SUBDOMAIN_STORE_PATH` хранит уже выданные поддомены и используется для переиспользования. Его нужно бэкапить вместе с окружением.

### Мониторинг

- Следите за числом 503 (нет доступного туннеля).
- Проверяйте ошибки `tunnel error` на клиенте.

### Безопасность (минимум)

- Ограничьте доступ к `CONTROL_ADDR` и `TUNNEL_ADDR`.
- При необходимости добавьте токены (можно реализовать простую auth-обертку).

## TLS и Let's Encrypt

Эта версия не использует ACME из-за ограничения на stdlib. Рекомендуемый вариант:
- Вынести выпуск/обновление сертификатов во внешний слой (например, Caddy/Traefik/nginx+certbot).
- Либо передавать готовые `TLS_CERT_FILE` и `TLS_KEY_FILE`.

## Решения

История решений хранится в `docs/adr/`.
