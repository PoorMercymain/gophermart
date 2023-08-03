[![gophermart](https://github.com/PoorMercymain/gophermart/actions/workflows/gophermart.yml/badge.svg?branch=graceful-shutdown-accrual)](https://github.com/PoorMercymain/gophermart/actions/workflows/gophermart.yml) [![go vet test](https://github.com/PoorMercymain/gophermart/actions/workflows/statictest.yml/badge.svg?branch=graceful-shutdown-accrual)](https://github.com/PoorMercymain/gophermart/actions/workflows/statictest.yml) [![CI](https://github.com/PoorMercymain/gophermart/actions/workflows/blank.yml/badge.svg?branch=graceful-shutdown-accrual)](https://github.com/PoorMercymain/gophermart/actions/workflows/blank.yml)
[![Go Coverage](https://github.com/PoorMercyman/gophermart/wiki/coverage.svg)](https://raw.githack.com/wiki/PoorMercyman/gophermart/coverage.html)
# Как запускать (docker-compose)
Запустить Docker, после чего в терминале в корневой директории проекта выполнить команду
```
docker-compose up
```
# Как запускать (не через docker-compose)
В терминале в корневой директории запустить gophermart
```
go run .\cmd\gophermart\main.go
```
и accrual
```
go run .\cmd\accrual\main.go
```
Для запуска могут быть использованы флаги и переменные окружения:

```
URI для подключения к postgres:
флаг -d
переменная окружения DATABASE_URI
пример: -d="host=localhost dbname=gophermart-postgres user=gophermart-postgres password=gophermart-postgres port=3000 sslmode=disable"

Адрес запуска:
флаг -a
переменная окружения RUN_ADDRESS
пример: -a="localhost:8081"
```
У gophermart есть дополнительные флаги и переменные окружения:
```
URI mongo (для локального запуска без docker-compose ОБЯЗАТЕЛЬНО нужно указать его, т.к. значение по умолчанию предусматривает либо запуск через docker-compose, либо через actions на github):
флаг -m
переменная окружения MONGO_URI
пример: -m="mongodb://localhost:27017"

Адрес accrual:
флаг -c
переменная окружения ACCRUAL_ADDRESS
пример: -c="http://localhost:8085"
```

# Как обновить моки слоя для работы с БД gophermart`а
В корневой директории проекта в терминале прописываем
```
go generate ./...
```
# Как запустить тесты
Запускаем тесты в терминале корневой директории
```
go test ./... -v --count=1
```

# go-musthave-group-diploma-tpl

Шаблон репозитория для группового дипломного проекта курса "Go-разработчик"

# Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без
   префикса `https://`) для создания модуля

# Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-group-diploma-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/master .github
```

Затем добавьте полученные изменения в свой репозиторий.
