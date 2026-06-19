# Security

- Токены только через env (`ARGOCD_AUTH_TOKEN`), не в config.yml
- `*_insecure: true` - только dev
- `deploy sync` требует подтверждение без `--yes`
- TLS: по умолчанию verify включен
- Сообщайте об уязвимостях: ...