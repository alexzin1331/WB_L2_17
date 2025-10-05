# Calendar Events

Простой HTTP-сервер для управления событиями календаря. Поддерживает создание, обновление, удаление и получение событий, логирование запросов и сбор метрик для Prometheus. Есть Swagger-документация.

### Запуск

`docker-compose up --build`

Сервер стартует на http://localhost:8080.

Swagger-документация
Доступна по адресу:
http://localhost:8080/swagger/index.html

### Эндпоинты

#### 1) _Создать событие_

**POST /create_event**

`{
  "user_id": 1,
  "date": "2025-07-10",
  "title": "Встреча",
  "event": "Обсуждение проекта"
}`

Ответ:

`{
  "result": "event created"
}`

#### 2) _Обновить событие_

**POST /update_event**


`{
  "user_id": 1,
  "date": "2025-07-10",
  "title": "Обновлённая встреча",
  "event": "Обновлённое описание"
}`

Ответ:

`{
  "result": "event updated"
}`

#### 3) _Удалить событие_

**POST /delete_event**

`{
  "user_id": 1,
  "date": "2025-07-10"
}`

Ответ:

`{
  "result": "event deleted"
}`

#### 4) _Получить события за день_

**GET /events_for_day?user_id=1&date=2025-07-10**

Ответ:

`{
  "result": [
    {
      "user_id": 1,
      "date": "2025-07-10T00:00:00Z",
      "title": "Встреча",
      "text": "Обсуждение проекта"
    }
  ]
}`
#### 5) Получить события за неделю

**GET /events_for_week?user_id=1&date=2025-07-10**

#### 6) Получить события за месяц

**GET /events_for_month?user_id=1&date=2025-07-10**

### Логирование

- Все HTTP-запросы логируются в stdout и файл /var/log/calendar/calendar.log

- Запросы отдельно записываются в http_requests.log

### Метрики

Метрики Prometheus доступны по адресу: 
http://localhost:8080/metrics

