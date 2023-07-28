# Как запускать
Запустить Docker, после чего в терминале в корневой директории проекта выполнить команду
```
docker-compose up -d
```
После этого в терминале в той же директории
```
go run .\cmd\gophermart\main.go -d="host=localhost dbname=gophermart-postgres user=gophermart-postgres password=gophermart-postgres port=3000 sslmode=disable"
```
После этого должен заработать gophermart

# Как запустить тесты
В корневой директории проекта в терминале прописываем
```
go generate ./...
```
Запускаем тесты также в терминале корневой директории
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
