# 0006 Client Config UX and Caddy On-Demand Ask

Date: 2026-03-03

## Context

После перехода на запуск через `docker-compose` с Caddy пользователям нужно:
- проще запускать клиент (минимум обязательных переменных),
- сразу видеть полный публичный URL туннеля в логах,
- корректно запускать Caddy с `on_demand TLS` (без падения из-за отсутствия anti-abuse проверки).

## Decision

- Добавить упрощенный режим клиента через `SERVER_ADDR`:
  - `SERVER_CONTROL_URL` по умолчанию: `https://control.<SERVER_ADDR>/register`,
  - `SERVER_TUNNEL_ADDR` по умолчанию: `<SERVER_ADDR>:8081`.
- Сделать `CONN_POOL_SIZE` необязательным с дефолтом `4`.
- В лог регистрации клиента добавить полный публичный адрес:
  - `url=https://<subdomain>.<domain>`.
- Для Caddy `on_demand TLS` добавить `on_demand_tls.ask` и endpoint на сервере:
  - `GET /caddy/allow?domain=...`,
  - выдавать `200` только для `DOMAIN`, `control.DOMAIN` и одноуровневых поддоменов `*.DOMAIN`.

## Consequences

- Клиент можно запускать короче: достаточно `SERVER_ADDR` и `LOCAL_FORWARD_ADDR` (остальное по умолчанию).
- Уменьшается число ошибок конфигурации и ускоряется копирование рабочего URL из логов.
- Caddy перестает падать на старте при `on_demand TLS`.
- Появляется явный серверный контроль допустимых доменов для выпуска сертификатов.
