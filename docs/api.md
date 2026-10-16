# NSU Assistant API

Base URL: http://localhost:8080

## 1. Health Check

Проверка работоспособности Go-сервера.

### Request

GET /health

### Response 200 OK

```json
{
  "status": "ok"
}
```

## 2. Получение расписания

Возвращает расписание указанной группы
на определённую дату.

### Request

GET /api/v1/schedule

### Query Parameters

| Параметр | Тип    | Обязательный | Описание        |
| -------- | ------ | ------------ | --------------- |
| group    | string | Да           | Номер группы    |
| date     | string | Да           | Дата YYYY-MM-DD |

### Example

GET /api/v1/schedule?group=24101&date=2026-10-19

### Response 200 OK

```json
{
  "group": "24101",
  "date": "2026-10-19",
  "lessons": [
    {
      "subject": "Математический анализ",
      "start_time": "09:00",
      "end_time": "10:35",
      "room": "312"
    }
  ]
}
```

Если занятий нет, возвращается пустой массив:

```json
{
  "group": "24101",
  "date": "2026-10-19",
  "lessons": []
}
```

### Response 400 Bad Request

Если параметры указаны неправильно:

```json
{
  "error": "invalid_date",
  "message": "Date must be in YYYY-MM-DD format"
}
```

### Response 502 Bad Gateway

Если источник расписания недоступен:

```json
{
  "error": "schedule_source_unavailable",
  "message": "Unable to fetch schedule"
}
```
