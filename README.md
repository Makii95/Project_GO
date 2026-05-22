# PulsePay Payment Platform

Учебная платформа обработки платежей на Go. Проект реализует основной HTTP-сервер, регистрацию и авторизацию пользователей, создание платежей, обработку платежей через отдельный сервис, RabbitMQ, PostgreSQL, Prometheus, Grafana, Docker Compose и сервис нагрузки.

## Что Реализовано

- Графический сайт продукта: `http://localhost:8080`
- Регистрация пользователя
- Логин пользователя и получение JWT-токена
- Создание платежа
- Получение списка платежей текущего пользователя
- Middleware, который получает `userID` из JWT
- PostgreSQL для хранения пользователей, тестовых сообщений и платежей
- RabbitMQ для обмена событиями между сервисами
- Отдельный сервис `order-processor`, который имитирует обработку платежа
- Prometheus-метрики
- Grafana dashboard
- Docker Compose deployment
- Сервис нагрузки `load-generator`
- Graceful shutdown основного сервера
- JSON-логирование через `log/slog`

## Быстрый Запуск Через Docker

Перейди в папку проекта:

```powershell
cd "C:\Users\maki\Desktop\все лабы по номерам го\готовый проект\payment-platform-server"
```

Запусти весь стек:

```powershell
docker compose up --build
```

Если команда `docker` не находится, используй полный путь:

```powershell
& "C:\Program Files\Docker\Docker\resources\bin\docker.exe" compose up --build
```

Запуск в фоне:

```powershell
& "C:\Program Files\Docker\Docker\resources\bin\docker.exe" compose up -d --build
```

Проверить контейнеры:

```powershell
& "C:\Program Files\Docker\Docker\resources\bin\docker.exe" compose ps
```

## Адреса Сервисов

Основной сайт:

```text
http://localhost:8080
```

Grafana:

```text
http://localhost:3000
```

Логин и пароль Grafana:

```text
admin / admin
```

Prometheus:

```text
http://localhost:9090
```

RabbitMQ Management:

```text
http://localhost:15672
```

Логин и пароль RabbitMQ:

```text
guest / guest
```

Метрики основного сервера:

```text
http://localhost:8080/metrics
```

Метрики сервиса обработки:

```text
http://localhost:9091/metrics
```

## Как Проверить Через Сайт

Открой:

```text
http://localhost:8080
```

Дальше выполни сценарий:

1. Нажми `Демо`, чтобы заполнить тестовые данные.
2. Нажми `Зарегистрироваться`.
3. Нажми `Войти`.
4. Создай платеж.
5. Нажми `Обновить список`.
6. Сначала статус платежа будет `created`.
7. Через несколько секунд снова нажми `Обновить список`.
8. Статус должен стать `processed`, если RabbitMQ и `order-processor` работают.

## Как Работает Обработка Платежей

Схема обработки:

```text
payment-server -> RabbitMQ -> order-processor -> RabbitMQ -> payment-server -> PostgreSQL
```

Пошагово:

- пользователь создает платеж через сайт или API;
- основной сервер сохраняет платеж в PostgreSQL со статусом `created`;
- основной сервер отправляет событие о новом платеже в RabbitMQ;
- сервис `order-processor` получает событие;
- `order-processor` имитирует обработку через небольшую паузу;
- `order-processor` отправляет событие изменения статуса обратно в RabbitMQ;
- основной сервер получает событие и обновляет статус платежа в PostgreSQL на `processed`.

Если RabbitMQ не запущен, платеж создастся, но статус останется `created`.

## API

Тестовый endpoint:

```http
GET /test
```

Ответ:

```text
Hello!
```

Запись тестовой строки в БД:

```http
POST /dbtest
```

Регистрация:

```http
POST /register
POST /auth/register
```

Тело запроса:

```json
{
  "username": "demo",
  "password": "secret123"
}
```

Логин:

```http
POST /login
POST /auth/login
```

Ответ:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Создание платежа:

```http
POST /orders
POST /orders/create
```

Заголовок:

```text
Authorization: Bearer <JWT>
```

Тело запроса:

```json
{
  "description": "Оплата подписки Premium. Сумма: 1490 ₽"
}
```

Получение списка платежей:

```http
GET /orders/statuses
GET /orders/list
```

Заголовок:

```text
Authorization: Bearer <JWT>
```

## Ручная Проверка Через PowerShell

Зарегистрировать пользователя:

```powershell
$body = '{"username":"demo_user","password":"secret123"}'
Invoke-WebRequest -Uri "http://localhost:8080/register" -Method POST -ContentType "application/json" -Body $body -UseBasicParsing
```

Получить JWT:

```powershell
$login = Invoke-WebRequest -Uri "http://localhost:8080/login" -Method POST -ContentType "application/json" -Body $body -UseBasicParsing
$token = ($login.Content | ConvertFrom-Json).token
```

Создать платеж:

```powershell
Invoke-WebRequest -Uri "http://localhost:8080/orders" -Method POST -Headers @{Authorization="Bearer $token"} -ContentType "application/json" -Body '{"description":"Платеж 1500 ₽"}' -UseBasicParsing
```

Подождать обработку:

```powershell
Start-Sleep -Seconds 3
```

Получить список платежей:

```powershell
Invoke-WebRequest -Uri "http://localhost:8080/orders/statuses" -Method GET -Headers @{Authorization="Bearer $token"} -UseBasicParsing
```

Ожидаемый результат после обработки:

```json
{
  "orders": [
    {
      "id": 1,
      "description": "Платеж 1500 ₽",
      "status": "processed"
    }
  ]
}
```

## Grafana

Открой:

```text
http://localhost:3000
```

Логин:

```text
admin
```

Пароль:

```text
admin
```

Если Grafana попросит сменить пароль, можно нажать `Skip` или задать новый пароль.

Дашборд находится в разделе `Dashboards`. Название:

```text
Payment Platform
```

Данные на графиках появляются благодаря:

- запросам с сайта;
- ручным API-запросам;
- сервису нагрузки `load-generator`.

## Сервис Нагрузки

Сервис `load-generator` запускается автоматически через Docker Compose.

Он делает следующее:

- регистрирует тестового пользователя;
- логинится и получает JWT;
- периодически создает платежи;
- запрашивает список платежей.

Это нужно, чтобы в Prometheus и Grafana появлялись данные для демонстрации.

Файл сервиса:

```text
cmd/load-generator/main.go
```

## Структура Проекта

```text
cmd/server
```

Основной HTTP-сервер и встроенный сайт.

```text
cmd/server/web
```

HTML, CSS и JavaScript сайта. Файлы встраиваются в Go-бинарник через `embed`.

```text
cmd/order-processor
```

Отдельный сервис обработки платежей.

```text
cmd/load-generator
```

Сервис нагрузки для демонстрации Grafana.

```text
internal/delivery/http
```

HTTP handler'ы и middleware авторизации.

```text
internal/usecase
```

Бизнес-логика: регистрация, логин, JWT, платежи.

```text
internal/repository
```

Работа с PostgreSQL и инициализация таблиц.

```text
internal/broker
```

Работа с RabbitMQ.

```text
internal/metrics
```

Prometheus-метрики.

```text
deploy/prometheus
```

Конфигурация Prometheus.

```text
deploy/grafana
```

Provisioning и dashboard Grafana.

## Таблицы PostgreSQL

При запуске основной сервер автоматически создает таблицы:

- `test_messages`
- `users`
- `orders`

Инициализация находится здесь:

```text
internal/repository/message_repository.go
```

## Запуск Без Docker

Без Docker можно запустить только основной сервер с локальным PostgreSQL. RabbitMQ, Grafana и Prometheus в таком варианте нужно запускать отдельно.

Переменные окружения:

```powershell
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASSWORD="ТВОЙ_ПАРОЛЬ"
$env:DB_NAME="payment_platform"
$env:DB_SSLMODE="disable"
$env:JWT_SECRET="my-secret-key"
```

Если порт `8080` занят:

```powershell
$env:SERVER_ADDRESS=":18082"
```

Запуск:

```powershell
go run ./cmd/server
```

Сайт будет доступен по адресу:

```text
http://localhost:8080
```

или, если задан порт `18082`:

```text
http://localhost:18082
```

## Тесты

Запустить все тесты:

```powershell
go test ./...
```

Если на компьютере мало памяти или возникают проблемы с кэшем:

```powershell
$env:GOCACHE=(Join-Path (Get-Location) ".gocache")
$env:GOMODCACHE=(Join-Path (Get-Location) ".gomodcache")
$env:GOTELEMETRY="off"
$env:GOMAXPROCS="1"
go test -p 1 ./...
```

Проверить только HTTP handler'ы:

```powershell
go test ./internal/delivery/http
```

## Остановка

Остановить контейнеры:

```powershell
docker compose down
```

Если `docker` не находится:

```powershell
& "C:\Program Files\Docker\Docker\resources\bin\docker.exe" compose down
```

Остановить контейнеры и удалить данные PostgreSQL:

```powershell
docker compose down -v
```

## Частые Ошибки

### Docker не находится

Используй полный путь:

```powershell
& "C:\Program Files\Docker\Docker\resources\bin\docker.exe" compose up --build
```

### Сайт не обновился после изменения HTML/CSS/JS

Сайт встроен в Go через `embed`, поэтому нужно пересобрать сервер:

```powershell
docker compose up -d --build payment-server
```

Затем в браузере нажать:

```text
Ctrl+F5
```

### Статус платежа не меняется на processed

Проверь, что запущены:

- `rabbitmq`
- `order-processor`
- `payment-server`

Команда:

```powershell
docker compose ps
```

### Ошибка PostgreSQL при локальном запуске

Проверь:

- запущен ли PostgreSQL;
- правильный ли `DB_PASSWORD`;
- существует ли база `payment_platform`;
- свободен ли порт `5432`.

### 409 Conflict при регистрации

Пользователь с таким `username` уже существует. Используй другой логин или выполни вход.
