# cli-assistant

Персональный CLI-ассистент для DevOps-инженера: GitOps (Argo CD) и Observability (Prometheus, Loki, Alertmanager).

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
task install
# или:
go build -ldflags "-s -w" -o "$(go env GOPATH)/bin/assistant" ./cmd/app

assistant version
```

`go install ./cmd/app` без `-o` создаёт бинарь **`app`**, не `assistant`. Для глобальной команды `assistant` используй `task install` или `go build` с `-o`.

### Docker

```bash
task docker:build
task docker:run -- version
task docker:run -- deploy list
```

Конфиг монтируется из `~/.config/cli_assistant` (read-only).

## Требования

- Go 1.26+
- [Task](https://taskfile.dev/installation/) (опционально)
- Docker (опционально, для `task docker:*`)

Для реальных интеграций:

| Сервис | Конфиг | Назначение |
|--------|--------|------------|
| Argo CD | `gitops.provider: argocd` | список, статус, sync, diff |
| Prometheus | `observability.provider: prometheus` | health, PromQL |
| Loki | `observability.logs_provider: loki` | `observe logs` |
| Alertmanager | `observability.alerts_provider: alertmanager` | `observe alerts`, `deploy inspect` |

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
  provider: noop       # noop | prometheus — метрики и health
  logs_provider: noop  # noop | loki
  alerts_provider: noop # noop | alertmanager

profiles:
  default:
    kube_context: minikube
    gitops_url: https://argocd.example.com
    gitops_token_env: ARGOCD_AUTH_TOKEN
    gitops_insecure: false
    metrics_url: http://localhost:9090
    metrics_insecure: false
    logs_url: http://localhost:3100
    logs_insecure: false
    alerts_url: http://localhost:9093
    alerts_insecure: false
```

При `observability.provider: prometheus` метрики и health идут через Prometheus API.  
Алерты и логи подключаются отдельными провайдерами (`alerts_provider`, `logs_provider`).

## Переменные окружения

| Переменная | Описание |
|------------|----------|
| `CLI_ASSISTANT_LOG_LEVEL` | `debug`, `info`, `warn`, `error` |
| `CLI_ASSISTANT_OUTPUT` | `human` или `json` |
| `CLI_ASSISTANT_PROFILE` | активный профиль |
| `CLI_ASSISTANT_GITOPS_PROVIDER` | `noop`, `argocd` |
| `CLI_ASSISTANT_OBSERVABILITY_PROVIDER` | `noop`, `prometheus` |
| `CLI_ASSISTANT_LOGS_PROVIDER` | `noop`, `loki` |
| `CLI_ASSISTANT_ALERTS_PROVIDER` | `noop`, `alertmanager` |
| `ARGOCD_AUTH_TOKEN` | токен Argo CD (имя задаётся в `gitops_token_env`) |

## Команды

Корневая команда: **`assistant`** (после сборки — `./bin/assistant`).

### Общие

| Команда | Описание |
|---------|----------|
| `assistant version` | версия, commit, дата сборки, активный профиль |
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
| `deploy inspect <name>` | GitOps + alerts + metrics + logs |

### Observability (`observe` / `obs`)

| Команда | Описание |
|---------|----------|
| `observe health` | здоровье Prometheus / Loki |
| `observe query <expr>` | PromQL instant query |
| `observe alerts` | список алертов (`noop`: demo; `alertmanager`: реальные) |
| `observe alerts --state Firing` | фильтр по состоянию |
| `observe logs <query>` | LogQL (`--since`, `--limit`) |

### Примеры

```bash
# noop (по умолчанию)
./bin/assistant deploy list
./bin/assistant deploy status demo-app
./bin/assistant deploy inspect demo-app
./bin/assistant observe query 'up'
./bin/assistant observe alerts --state Firing
./bin/assistant observe logs '{app="demo-app"}' --since 30m --limit 50

# JSON-вывод
CLI_ASSISTANT_OUTPUT=json ./bin/assistant observe health

# Argo CD
export ARGOCD_AUTH_TOKEN="your-token"
CLI_ASSISTANT_GITOPS_PROVIDER=argocd ./bin/assistant deploy list

# Prometheus + Loki + Alertmanager
CLI_ASSISTANT_OBSERVABILITY_PROVIDER=prometheus \
CLI_ASSISTANT_LOGS_PROVIDER=loki \
CLI_ASSISTANT_ALERTS_PROVIDER=alertmanager \
./bin/assistant observe alerts
```

## Архитектура

Clean Architecture: домен → use case → delivery, адаптеры снаружи.

```
cmd/app/                         # точка входа, DI
internal/
  delivery/cli/                  # Cobra-команды
  usecase/                       # оркестрация
  domain/                        # модели и порты (deploy, observe)
  adapter/
    argocd, prometheus, loki, alertmanager, noop
    composite                    # сборка metrics + alerts + logs
    factory                      # выбор провайдеров по конфигу
  config/                        # загрузка конфига
  platform/scope/                # config.Profile → domain.Scope
pkg/
  output/                        # human / json вывод
  log/                           # логгер
  version/                       # version, commit, build date (ldflags)
```

Поток данных:

```
CLI → UseCase → Port (interface) → Adapter (Argo CD / Prometheus / Loki / Alertmanager / noop)
```

Провайдеры в конфиге:

- `gitops.provider` — GitOps
- `observability.provider` — метрики и health
- `observability.logs_provider` — логи
- `observability.alerts_provider` — алерты

## Разработка

### Task

```bash
task build              # bin/assistant (с ldflags version)
task run -- deploy list
task install            # GOPATH/bin/assistant
task check              # vet + test
task lint               # golangci-lint
task arch               # go-arch-lint
task cleancode          # fmt, vet, lint, gosec, test, arch
task docker:build       # Docker-образ
task docker:run -- version

# демо на noop
task demo:list
task demo:health
task demo:query
task demo:alerts
task demo:logs
task demo:inspect
task demo:version
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

GitHub Actions: [`.github/workflows/ci.yml`](.github/workflows/ci.yml) — `go vet`, `go test`, `go build` с version ldflags, `golangci-lint`, `go-arch-lint`, сборка Docker-образа.

## Troubleshooting

| Проблема | Решение |
|----------|---------|
| `assistant: command not found` | Используй `./bin/assistant` или `task install` |
| `unknown flag: --since` | Пересобери бинарь: `task build` |
| Конфиг не подхватывается | Путь: `~/.config/cli_assistant/config.yaml` (не `.yml`) |
| `argocd: set token in ...` | `export ARGOCD_AUTH_TOKEN=...` и `gitops.provider: argocd` |
| `prometheus: metrics_url is required` | Укажи `metrics_url` в профиле |
| `loki: logs_url is required` | Укажи `logs_url` или `logs_provider: noop` |
| `alertmanager: alerts_url is required` | Укажи `alerts_url` или `alerts_provider: noop` |
| `unsupported alerts provider` | Проверь опечатку: `alertmanager`, не `alertmanger` |
| `go install` даёт `app` | Используй `task install` или `go build -o .../bin/assistant` |
