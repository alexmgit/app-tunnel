# 0002 TLS and ACME

Date: 2026-02-04

## Context

Требуется поддержка HTTPS и желательно Let's Encrypt. Ограничение: использовать только stdlib.

## Decision

- Поддержать TLS только при наличии заранее подготовленных `TLS_CERT_FILE` и `TLS_KEY_FILE`.
- ACME/Let's Encrypt не реализовывать в коде (stdlib не покрывает полноценный ACME клиент).
- Рекомендовать внешний TLS-терминатор (Caddy/Traefik/nginx+certbot).

## Consequences

- Первичная выдача/обновление сертификатов выполняется вне сервиса.
