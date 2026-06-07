# cli-assistant

Персональный CLI-ассистент для DevOps-инженера: GitOps (Argo CD) и Observability (Prometheus).

Помогает управлять деплоями, отслеживать состояние приложений и работать с observability-стеком из терминала.

## Быстрый старт

```bash
git clone https://github.com/itGeek-rus/cli-assistant.git
cd cli-assistant

# Сборка
task build
# или: go build -o bin/assistant ./cmd/app

./bin/assistant version
./bin/assistant deploy list
./bin/assistant observe health
```

> Команда `assistant` не появляется в PATH автоматически.  
> Используй `./bin/assistant` или установи бинарь в `GOPATH/bin` (см. ниже).

По умолчанию провайдеры `noop` — демо-данные без реального кластера.

## Установка

### Локально (из исходников)

```bash
task build
./bin/assistant --help
```

### Глобально

```bash
go build -o "$(go env GOPATH)/bin/assistant" ./cmd/app
assistant version
```

`task install` и `go install ./cmd/app` создают бинарь **`app`**, не `assistant`. Для глобальной команды `assistant` используй `go build` с `-o`, как выше.

## Требования

- Go 1.26+
- [Task](https://taskfile.dev/installation/) (опционально)

Для реальных интеграций:

- Argo CD — URL API и токен (`gitops.provider: argocd`)
- Prometheus — endpoint метрик (`observability.provider: prometheus`)

## Конфигурация

Файл конфигурации:

```
~/.config/cli_assistant/config.yaml
```

Пример (в репозитории: `.config/cli-assistant/config.yml` — скопируй и переименуй):

```yaml
log_level: info
output: human          # human | json
profile: default

gitops:
  provider: noop       # noop | argocd

observability:
  provider: noop       # noop | prometheus

profiles:
  default:
    kube_context: minikube
    gitops_url: https://argocd.example.com
    gitops_token_env: ARGOCD_AUTH_TOKEN
    gitops_insecure: false
    metrics_url: http://localhost:9090
    metrics_insecure: false
    logs_url: http://localhost:3100
```

## Переменные окружения

| Переменная | Описание |
|------------|----------|
| `CLI_ASSISTANT_LOG_LEVEL` | `debug`, `info`, `warn`, `error` |
| `CLI_ASSISTANT_OUTPUT` | `human` или `json` |
| `CLI_ASSISTANT_PROFILE` | активный профиль |
| `CLI_ASSISTANT_GITOPS_PROVIDER` | `noop`, `argocd` |
| `CLI_ASSISTANT_OBSERVABILITY_PROVIDER` | `noop`, `prometheus` |
| `ARGOCD_AUTH_TOKEN` | токен Argo CD (имя задаётся в `gitops_token_env`) |

## Команды

Корневая команда: **`assistant`** (после сборки — `./bin/assistant`).

### Общие

| Команда | Описание |
|---------|----------|
| `assistant version` | версия и активный профиль |
| `assistant --help` | справка |

### GitOps (`deploy`)

| Команда | Описание |
|---------|----------|
| `deploy list` | список приложений |
| `deploy list --namespace X --name-prefix Y` | с фильтрами |
| `deploy status` | все приложения (таблица) |
| `deploy status <name>` | одно приложение |
| `deploy diff <name>` | diff |
| `deploy sync <name>` | sync (подтверждение `[y/N]`) |
| `deploy sync <name> --dry-run` | только diff, без sync |
| `deploy sync <name> --yes` | без подтверждения |
| `deploy sync <name> --prune --force` | флаги Argo CD |

### Observability (`observe` / `obs`)

| Команда | Описание |
|---------|----------|
| `observe health` | здоровье Prometheus / Loki |
| `observe query <expr>` | PromQL instant query |
| `observe alerts` | список алертов (`noop`: demo-данные) |
| `observe alerts --state Firing` | фильтр по состоянию |

> При `observability.provider: prometheus` алерты пока не реализованы (пустой список). Метрики и health — через Prometheus API.

### Примеры

```bash
./bin/assistant deploy list
./bin/assistant deploy status demo-app
./bin/assistant deploy sync demo-app --yes

CLI_ASSISTANT_OUTPUT=json ./bin/assistant observe health
./bin/assistant observe query 'up'
./bin/assistant observe alerts --state Firing

# Argo CD
export ARGOCD_AUTH_TOKEN="your-token"
CLI_ASSISTANT_GITOPS_PROVIDER=argocd ./bin/assistant deploy list
```

## Архитектура

Clean Architecture: домен → use case → delivery, адаптеры снаружи.

```
cmd/app/                    # точка входа, DI
internal/
  delivery/cli/             # Cobra-команды
  usecase/                  # оркестрация
  domain/                   # модели и порты (deploy, observe)
  adapter/                  # argocd, prometheus, noop, factory
  config/                   # загрузка конфига
pkg/
  output/                   # human / json вывод
  log/                      # логгер
```

Поток данных:

```
CLI → UseCase → Port (interface) → Adapter (Argo CD / Prometheus / noop)
```

Провайдеры в конфиге: `gitops.provider`, `observability.provider`.

## Разработка

### Task

```bash
task build          # bin/assistant
task run -- deploy list
task check          # vet + test
task lint           # golangci-lint
task arch           # go-arch-lint
task cleancode      # fmt, vet, lint, gosec, test
task demo:list
task demo:health
task demo:query
task demo:alerts
```

### Инструменты (опционально)

| Утилита | Установка |
|---------|-----------|
| Task | `brew install go-task` |
| golangci-lint | `brew install golangci-lint` |
| go-arch-lint | `go install github.com/fe3dback/go-arch-lint@latest` |
| lefthook | `brew install lefthook && lefthook install` |
| govulncheck | `go install golang.org/x/vuln/cmd/govulncheck@latest` |

### CI

GitHub Actions: [`.github/workflows/ci.yml`](.github/workflows/ci.yml) — `go vet`, `go test`, `go build`, `golangci-lint`.

## Troubleshooting

| Проблема | Решение |
|----------|---------|
| `assistant: command not found` | Используй `./bin/assistant` или установи в `GOPATH/bin` |
| Конфиг не подхватывается | Путь: `~/.config/cli_assistant/config.yaml` (не `.yml`) |
| `argocd: set token in ...` | `export ARGOCD_AUTH_TOKEN=...` и `gitops.provider: argocd` |
| `prometheus: metrics_url is required` | Укажи `metrics_url` в профиле |
| `go install` даёт `app` | Собери явно: `go build -o .../bin/assistant ./cmd/app` |
