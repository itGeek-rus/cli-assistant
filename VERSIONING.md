## Версионирование

## SemVer
Формат тегов: `vMAJOR.MINOR.PATCH`.

Проект следует [Semantic Versioning](https://semver.org/): `MAJOR.MINOR.PATCH`.

| Версия | Когда повышать |
| ------ | ---------------|
| MAJOR | ломающие изменения CLI/конфига |
| MINOR | новые команды, провайдеры |
| PATCH | багфиксы, docs, CI | 

Источник версии при сборке:
1. `git describe --tags` -> тег `v1.2.3`
2. иначе `dev` + short commit

| Поле         | Источник                                  |
|--------------|-------------------------------------------|
| `version`    | git tag `vX.Y.Z` или `dev`                |
| `commit`     | git SHA при сборке                        |
| `build_date` | timestamp CI/локальной сборки             |
 | `go` | `runtime/debug.ReadBuildInfo().GoVersion` |

## Релиз
1. Обновить `CHANGELOG.md`
2. `git tag v1.0.0 && git push origin v1.0.0`
3. GitHub Release собирает бинарники (workflow `release.yml`)

Сборка: `task build` / CI / релиз - через ldflags в `pkg/version`.

## API versioning
REST: префикс `/v1/`. Breaking changes -> `/v2/`.

```bash
./bin/assistant version

`VERSION:`
`1.0.0`