# Payment Platform Server

HTTP-сервер на Go для лабораторных работ.

В проекте реализовано:

- `GET /test` - тестовый endpoint, возвращает `Hello!`
- `POST /dbtest` - записывает строку из тела запроса в PostgreSQL
- `POST /register` - создает пользователя в PostgreSQL
- `POST /login` - авторизует пользователя и возвращает JWT-токен
- подключение к PostgreSQL
- автоматическая инициализация таблиц
- graceful shutdown при остановке сервера

## Требования

Перед запуском должны быть установлены:

- Go
- PostgreSQL

Проверить Go:

```powershell
go version
```

Проверить, что PostgreSQL работает:

```powershell
Get-Service | Where-Object { $_.Name -like "postgres*" }
```

Если служба PostgreSQL запущена, в колонке `Status` будет `Running`.

## Папка проекта

Перейдите в папку проекта:

```powershell
cd "C:\Users\maki\Desktop\все лабы по номера го\лаба 3\payment-platform-server"
```

## Настройка базы данных

В проекте используется база данных:

```text
payment_platform
```

Пользователь PostgreSQL:

```text
postgres
```

Пароль:

```text
ТВОЙ_ПАРОЛЬ
```

Здесь нужно указать пароль от пользователя `postgres`, который вы сами задали при установке PostgreSQL.

Если база `payment_platform` еще не создана, создайте ее:

```powershell
$env:PGPASSWORD="ТВОЙ_ПАРОЛЬ"
& "C:\Program Files\PostgreSQL\18\bin\createdb.exe" -h localhost -p 5432 -U postgres payment_platform
```

Если база уже существует, команда может вывести ошибку, что такая база уже есть. Это нормально.

## Переменные окружения

Перед запуском сервера задайте переменные окружения:

```powershell
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASSWORD="ТВОЙ_ПАРОЛЬ"
$env:DB_NAME="payment_platform"
$env:DB_SSLMODE="disable"
$env:JWT_SECRET="my-secret-key"
```

Если порт `8080` занят, можно запустить сервер на другом порту:

```powershell
$env:SERVER_ADDRESS=":18082"
```

Если порт `8080` свободен, эту переменную можно не задавать.

## Запуск сервера

Запустите сервер из корня проекта:

```powershell
go run ./cmd/server
```

Если сервер запустился успешно, в терминале будет сообщение примерно такого вида:

```text
payment-platform-server: server started on :18082
```

Остановить сервер можно сочетанием клавиш:

```text
Ctrl+C
```

## Проверка GET /test

Откройте второй терминал и выполните:

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/test" -UseBasicParsing
```

Ожидаемый ответ:

```text
Hello!
```

Если вы запускаете сервер на порту `8080`, используйте:

```powershell
Invoke-WebRequest -Uri "http://localhost:8080/test" -UseBasicParsing
```

## Проверка POST /dbtest

Этот endpoint записывает строку из тела запроса в таблицу `test_messages`.

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/dbtest" -Method POST -Body "hello database" -UseBasicParsing
```

Ожидаемый результат:

```text
StatusCode: 201
Content: hello database
```

Проверить запись в PostgreSQL:

```powershell
$env:PGPASSWORD="ТВОЙ_ПАРОЛЬ"
& "C:\Program Files\PostgreSQL\18\bin\psql.exe" -h localhost -p 5432 -U postgres -d payment_platform -c "SELECT id, message, created_at FROM test_messages ORDER BY id DESC LIMIT 5;"
```

## Проверка POST /register

Этот endpoint создает нового пользователя в таблице `users`.

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/register" -Method POST -ContentType "application/json" -Body '{"username":"maks","password":"secret123"}' -UseBasicParsing
```

Ожидаемый результат:

```text
StatusCode: 201
```

В ответе будет JSON:

```json
{"id":1,"username":"maks"}
```

Если пользователь уже существует, сервер вернет:

```text
409 Conflict
```

В таком случае можно использовать другое имя пользователя:

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/register" -Method POST -ContentType "application/json" -Body '{"username":"maks2","password":"secret123"}' -UseBasicParsing
```

## Проверка POST /login

Этот endpoint проверяет логин и пароль существующего пользователя и возвращает JWT-токен.

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/login" -Method POST -ContentType "application/json" -Body '{"username":"maks","password":"secret123"}' -UseBasicParsing
```

Ожидаемый результат:

```text
StatusCode: 200
```

В ответе будет JSON с JWT-токеном:

```json
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```

JWT-токен состоит из трех частей, разделенных точками:

```text
header.payload.signature
```

Если пароль неправильный, сервер вернет:

```text
401 Unauthorized
```

## Проверка пользователей в БД

Посмотреть последних пользователей:

```powershell
$env:PGPASSWORD="ТВОЙ_ПАРОЛЬ"
& "C:\Program Files\PostgreSQL\18\bin\psql.exe" -h localhost -p 5432 -U postgres -d payment_platform -c "SELECT id, username, created_at FROM users ORDER BY id DESC LIMIT 5;"
```

Пароли в таблице не хранятся открытым текстом. Сервер сохраняет `password_hash`.

## Запуск тестов

```powershell
go test ./...
```

Если возникнет ошибка доступа к кэшу Go, задайте кэш внутри папки проекта:

```powershell
$env:GOCACHE=(Join-Path (Get-Location) ".gocache")
$env:GOMODCACHE=(Join-Path (Get-Location) ".gomodcache")
$env:GOTELEMETRY="off"
go test ./...
```

## Таблицы, которые создаются автоматически

При запуске сервера создаются таблицы:

```sql
CREATE TABLE IF NOT EXISTS test_messages (
    id SERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Инициализация таблиц находится в repository-слое:

```text
internal/repository/message_repository.go
```

## Архитектура проекта

Проект разделен на слои:

```text
cmd/server/main.go
```

Точка входа. Здесь запускается сервер, подключается PostgreSQL и собираются зависимости.

```text
internal/delivery/http/handler.go
```

HTTP-слой. Здесь находятся handler'ы:

- `/test`
- `/dbtest`
- `/register`
- `/login`

```text
internal/usecase`
```

Слой бизнес-логики. Здесь находятся сценарии:

- получить тестовое сообщение
- записать строку в БД
- зарегистрировать пользователя
- авторизовать пользователя
- сгенерировать JWT

```text
internal/repository/message_repository.go
```

Repository-слой. Здесь находится работа с PostgreSQL:

- создание таблиц
- запись сообщений
- создание пользователей
- поиск пользователя по username

```text
internal/domain/user.go
```

Модель пользователя.

## Частые ошибки

### Ошибка подключения к PostgreSQL

Пример:

```text
failed to connect to database
```

Проверьте:

- запущен ли PostgreSQL
- правильный ли пароль `DB_PASSWORD`
- существует ли база `payment_platform`
- свободен ли порт `5432`

### Ошибка 409 Conflict при регистрации

Это значит, что пользователь с таким `username` уже существует.

Используйте другое имя пользователя или выполните login с уже созданным пользователем.

### PowerShell спрашивает про Y

Если `Invoke-WebRequest` показывает предупреждение и спрашивает:

```text
Do you want to continue?
```

Можно написать:

```text
Y
```

и нажать Enter.

Чтобы предупреждение не появлялось, добавляйте:

```powershell
-UseBasicParsing
```

### Порт 8080 занят

Запустите сервер на другом порту:

```powershell
$env:SERVER_ADDRESS=":18082"
go run ./cmd/server
```

Тогда запросы нужно отправлять на:

```text
http://localhost:18082
```

## Быстрый запуск и проверка

Откройте терминал в папке проекта:

```powershell
cd "C:\Users\maki\Desktop\все лабы по номера го\лаба 3\payment-platform-server"
```

Задайте настройки БД и JWT:

```powershell
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASSWORD="ТВОЙ_ПАРОЛЬ"
$env:DB_NAME="payment_platform"
$env:DB_SSLMODE="disable"
$env:JWT_SECRET="my-secret-key"
```

Если порт `8080` занят, задайте другой порт:

```powershell
$env:SERVER_ADDRESS=":18082"
```

Запустите сервер:

```powershell
go run ./cmd/server
```

Теперь откройте второй терминал и проверьте регистрацию:

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/register" -Method POST -ContentType "application/json" -Body '{"username":"maks","password":"secret123"}' -UseBasicParsing
```

Потом проверьте логин:

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/login" -Method POST -ContentType "application/json" -Body '{"username":"maks","password":"secret123"}' -UseBasicParsing
```

Если все работает, в ответ на `/login` придет JSON примерно такого вида:

```json
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```

Это и есть JWT-токен.

Еще можно проверить старые endpoint:

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/test" -UseBasicParsing
```

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/dbtest" -Method POST -Body "hello db" -UseBasicParsing
```

Остановить сервер можно в первом терминале через:

```text
Ctrl+C
```

## Проверка заказов

Сначала зарегистрируйте пользователя и получите JWT-токен через `/login`.

В PowerShell можно сохранить токен в переменную:

```powershell
$loginResponse = Invoke-WebRequest -Uri "http://localhost:18082/login" -Method POST -ContentType "application/json" -Body '{"username":"maks","password":"secret123"}' -UseBasicParsing
$token = ($loginResponse.Content | ConvertFrom-Json).token
```

Создать новый заказ:

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/orders" -Method POST -Headers @{Authorization="Bearer $token"} -ContentType "application/json" -Body '{"description":"new payment order"}' -UseBasicParsing
```

Ожидаемый ответ:

```json
{"id":1,"description":"new payment order","status":"created"}
```

Проверить статусы всех заказов пользователя:

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/orders/statuses" -Method GET -Headers @{Authorization="Bearer $token"} -UseBasicParsing
```

Ожидаемый ответ:

```json
{"orders":[{"id":1,"description":"new payment order","status":"created"}]}
```

Заказы привязаны к пользователю по JWT-токену. Если не передать заголовок `Authorization`, сервер вернет `401 Unauthorized`.

## Middleware авторизации

В лабораторной 5 для заказов используется middleware:

```text
internal/delivery/http/middleware.go
```

Middleware делает следующее:

- читает JWT из заголовка `Authorization`
- проверяет формат `Bearer <token>`
- валидирует JWT через usecase
- получает `userID` владельца токена
- кладет `userID` в `context.Context`
- передает запрос дальше в handler заказов

Handler заказов уже не разбирает JWT сам. Он получает `userID` из контекста и передает его в usecase.

Проверка разделения заказов по пользователям:

```powershell
$login1 = Invoke-WebRequest -Uri "http://localhost:18082/login" -Method POST -ContentType "application/json" -Body '{"username":"maks","password":"secret123"}' -UseBasicParsing
$token1 = ($login1.Content | ConvertFrom-Json).token
```

```powershell
$login2 = Invoke-WebRequest -Uri "http://localhost:18082/login" -Method POST -ContentType "application/json" -Body '{"username":"maks2","password":"secret123"}' -UseBasicParsing
$token2 = ($login2.Content | ConvertFrom-Json).token
```

Создать заказ для первого пользователя:

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/orders" -Method POST -Headers @{Authorization="Bearer $token1"} -ContentType "application/json" -Body '{"description":"order for first user"}' -UseBasicParsing
```

Создать заказ для второго пользователя:

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/orders" -Method POST -Headers @{Authorization="Bearer $token2"} -ContentType "application/json" -Body '{"description":"order for second user"}' -UseBasicParsing
```

Проверить заказы первого пользователя:

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/orders/statuses" -Method GET -Headers @{Authorization="Bearer $token1"} -UseBasicParsing
```

Проверить заказы второго пользователя:

```powershell
Invoke-WebRequest -Uri "http://localhost:18082/orders/statuses" -Method GET -Headers @{Authorization="Bearer $token2"} -UseBasicParsing
```

У каждого пользователя должен быть свой список заказов.
