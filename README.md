# app-tunnel

Минимальный ngrok-подобный HTTP/HTTPS туннель на Go (только stdlib).

Проект состоит из двух бинарников:
- `server` принимает публичный трафик и проксирует его в туннель.
- `client` держит туннель к серверу и форвардит запросы в локальный сервис.

## Быстрый запуск через Docker Compose

Это основной путь запуска в проде: домен + SSL + автозапуск сервисов.

### 1) DNS

Нужны записи на публичный IP сервера:
- `A` для корневого домена, например `example.com`.
- `A` wildcard для `*.example.com`.

### 2) Подготовьте `.env.compose`

Пример:

```bash
DOMAIN="example.com"
ACME_EMAIL="admin@example.com"
TUNNEL_PORT="8081"
SUBDOMAIN_LENGTH="6"
TUNNEL_TIMEOUT="30s"
LOG_LEVEL="info"
```

Можно взять шаблон:

```bash
cp .env.compose.example .env.compose
```

### 3) Запустите стек

Совместимый вариант (работает даже на средах без `--env-file`):

```bash
cp .env.compose .env
docker compose up -d --build
```

Если у вас поддерживается флаг `--env-file`:

```bash
docker compose --env-file .env.compose up -d --build
```

### 4) Запустите клиент

Минимальный запуск:

```bash
SERVER_ADDR="example.com" \
LOCAL_FORWARD_ADDR="127.0.0.1:3000" \
./bin/client
```

Клиент автоматически использует:
- `SERVER_CONTROL_URL=https://control.<SERVER_ADDR>/register`
- `SERVER_TUNNEL_ADDR=<SERVER_ADDR>:8081`

После регистрации в логах будет полный публичный URL:
- `url=https://<subdomain>.<domain>`

## Ограничения текущей версии

- Только HTTP/HTTPS (HTTP/1.1), без TCP/UDP raw.
- Нет авторизации.
- ACME не реализован внутри Go-сервера (TLS рекомендуется завершать внешним прокси, например Caddy).
- Один туннельный коннект обслуживает один запрос за раз (масштабирование через pool).

## Ручной запуск и сборка

### Сборка из исходников

```bash
mkdir -p bin
GOCACHE=/tmp/go-build go build -o bin/server ./cmd/server
GOCACHE=/tmp/go-build go build -o bin/client ./cmd/client
```

### Запуск сервера вручную

```bash
CONTROL_ADDR=":8080" \
TUNNEL_ADDR=":8081" \
HTTP_ADDR=":8082" \
HTTPS_ADDR="" \
TLS_CERT_FILE="" \
TLS_KEY_FILE="" \
DOMAIN="example.com" \
SUBDOMAIN_STORE_PATH="./data/subdomains.txt" \
SUBDOMAIN_LENGTH="6" \
TUNNEL_TIMEOUT="30s" \
LOG_LEVEL="info" \
./bin/server
```

### Запуск клиента вручную

```bash
SERVER_ADDR="example.com" \
LOCAL_FORWARD_ADDR="127.0.0.1:3000" \
REQUESTED_SUBDOMAIN="" \
CONN_POOL_SIZE="4" \
DIAL_TIMEOUT="10s" \
LOG_LEVEL="info" \
./bin/client
```

`CONN_POOL_SIZE` необязателен, по умолчанию `4`.

## Ручной запуск через Docker (без compose)

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

## Проверка работы

1. Поднимите любой локальный HTTP сервис на `3000`, например:
```bash
python3 -m http.server 3000
```

2. Запустите сервер и клиент.

3. Возьмите адрес из лога клиента и откройте его в браузере.

Для локальной проверки без DNS можно отправить `Host` вручную:

```bash
curl -H "Host: <subdomain>.example.com" http://127.0.0.1:8082/
```

## Эксплуатация

### Порты и доступ

- Публичные порты: `80`, `443`, `8081` (туннель).
- `CONTROL_ADDR` и `TUNNEL_ADDR` лучше ограничить firewall-правилами.

### Логи и мониторинг

- Сервер и клиент пишут в stdout/stderr.
- Для подробных логов используйте `LOG_LEVEL="debug"`.
- Следите за `503` (нет доступного туннеля) и `tunnel error` на клиенте.

### Хранение поддоменов

`SUBDOMAIN_STORE_PATH` хранит выданные поддомены для переиспользования. Этот файл стоит бэкапить.

## ADR

История решений: `docs/adr/`.
