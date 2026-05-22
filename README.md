# PulsePay Payment Platform

Учебная платформа обработки платежей на Go.

В проекте реализованы: HTTP-сервер, регистрация и логин пользователей, JWT-авторизация, создание платежей, просмотр платежей пользователя, PostgreSQL, RabbitMQ, отдельный сервис обработки платежей, Prometheus, Grafana, Docker Compose и сервис нагрузки.

## Быстрый запуск

Перейти в папку проекта:

```powershell
cd "C:\path\to\payment-platform-server"
```

Запустить весь проект:

```powershell
docker compose up --build
```

Запуск в фоне:

```powershell
docker compose up -d --build
```

Проверить контейнеры:

```powershell
docker compose ps
```

Если команда `docker` не находится, можно использовать полный путь:

```powershell
& "C:\Program Files\Docker\Docker\resources\bin\docker.exe" compose up -d --build
```

## Адреса

- Сайт: `http://localhost:8080`
- Grafana: `http://localhost:3000`, логин `admin`, пароль `admin`
- Prometheus: `http://localhost:9090`
- RabbitMQ Management: `http://localhost:15672`, логин `guest`, пароль `guest`
- Метрики сервера: `http://localhost:8080/metrics`
- Метрики обработчика платежей: `http://localhost:9091/metrics`

## Проверка через сайт

1. Открыть `http://localhost:8080`.
2. Нажать `Демо`, чтобы заполнить тестовые данные.
3. Зарегистрировать пользователя.
4. Выполнить вход.
5. Создать платеж.
6. Обновить список платежей.
7. Через несколько секунд статус платежа должен измениться с `created` на `processed`.

## Как работает обработка платежа

```text
payment-server -> RabbitMQ -> order-processor -> RabbitMQ -> payment-server -> PostgreSQL
```

Пользователь создает платеж, основной сервер сохраняет его в PostgreSQL и отправляет событие в RabbitMQ. Сервис `order-processor` получает событие, имитирует обработку и отправляет обратно событие о новом статусе. После этого основной сервер обновляет статус платежа в базе данных.

## API

```http
GET /test
POST /dbtest
POST /register
POST /login
POST /orders
GET /orders/statuses
GET /metrics
```

Для `POST /orders` и `GET /orders/statuses` нужен заголовок:

```text
Authorization: Bearer <JWT>
```

Пример регистрации:

```json
{
  "username": "demo",
  "password": "secret123"
}
```

Пример создания платежа:

```json
{
  "description": "Оплата подписки Premium"
}
```

## Проверка через PowerShell

```powershell
$body = '{"username":"demo_user","password":"secret123"}'

Invoke-WebRequest -Uri "http://localhost:8080/register" -Method POST -ContentType "application/json" -Body $body -UseBasicParsing

$login = Invoke-WebRequest -Uri "http://localhost:8080/login" -Method POST -ContentType "application/json" -Body $body -UseBasicParsing
$token = ($login.Content | ConvertFrom-Json).token

Invoke-WebRequest -Uri "http://localhost:8080/orders" -Method POST -Headers @{Authorization="Bearer $token"} -ContentType "application/json" -Body '{"description":"Платеж 1500"}' -UseBasicParsing

Start-Sleep -Seconds 3

Invoke-WebRequest -Uri "http://localhost:8080/orders/statuses" -Method GET -Headers @{Authorization="Bearer $token"} -UseBasicParsing
```

## Структура проекта

- `cmd/server` - основной HTTP-сервер и сайт
- `cmd/server/web` - HTML, CSS и JavaScript интерфейса
- `cmd/order-processor` - сервис обработки платежей
- `cmd/load-generator` - сервис нагрузки для Grafana
- `internal/delivery/http` - handlers и middleware
- `internal/usecase` - бизнес-логика
- `internal/repository` - работа с PostgreSQL
- `internal/broker` - работа с RabbitMQ
- `internal/metrics` - Prometheus-метрики
- `deploy` - настройки Grafana и Prometheus

## База данных

При запуске сервер автоматически создает таблицы:

- `test_messages`
- `users`
- `orders`

## Тесты

Запустить все тесты:

```powershell
go test ./...
```

Запустить тесты HTTP-слоя:

```powershell
go test ./internal/delivery/http
```

## Остановка

Остановить контейнеры:

```powershell
docker compose down
```

Остановить контейнеры и удалить данные PostgreSQL:

```powershell
docker compose down -v
```

