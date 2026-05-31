# cli-assistant

Персональный CLI - ассистент, для DevOps - инженера, которая объединяет операции GitOps и Observability.

Утилита помогает управлять деплоями, отслеживать состояние приложений, проверять здоровье инфраструктуры и работать с GitOps инструментами (ArgoCD, Flux).

## Требования

- Go 1.26+
- [Task](https://taskfile.dev) (опционально)
- Для Argo CD: URL API и токен в переменной окружения

## Быстрый запуск

```bash
# Сборка
go build -o bin/assistant ./cmd/app

# Task
task build

# Справка
./bin/assistant --help

# Демо-данные (провайдер noop по умолчнию)
./bin/assistant deploy list
./bin/assistant observe health