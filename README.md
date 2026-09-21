# Кэширующий сервер файлшаринга

Сервис поддерживающий регистрацию/авторизацию по JWT,
хранение документов (файл + метаданные + произвольный JSON). Доступ
разграничен по владельцу, публичности (доступен всем) и списку `grant`.

## Запуск через Docker Compose

### 1. Переменные окружения

Скопируте шаблон:

```bash
cp .env.template .env
```

### 2. Запуск

```bash
docker compose up --build
```

Остановка:
```bash
docker compose down
```

### 3. Применение миграции

```bash
make migrate
make rollback # если нужно почистить базу
```

### 4. Проверка

Импортируйте Postman коллекцию [AstralTestTask.postman_collection.json](AstralTestTask.postman_collection.json)

## Роуты

### Аутентификация (/api/auth):

- POST /api/auth/register - регистрация
- POST /api/auth/login - вход, выдаёт JWT
- DELETE /api/auth/logout - выход (требует Bearer-токен)
- GET /api/auth/me - данные текущего пользователя (требует Bearer-токен)

Пароль передаётся один раз при регистрации/логине, хранится как bcrypt-хэш.
Авторизация везде дальше - заголовок `Authorization: Bearer <token>`.

### Документы (/api/docs, все требуют Bearer-токен):

- POST /api/docs — загрузка документа
- GET /api/docs — список документов
- HEAD /api/docs — то же самое, без тела ответа
- GET /api/docs/:id — один документ (файл или JSON)
- HEAD /api/docs/:id — то же самое, без тела ответа
- DELETE /api/docs/:id — удаление документа
