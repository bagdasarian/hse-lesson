#!/bin/bash
# Полное тестирование API сервиса аренды жилья
# Запуск: bash curl_requests.sh
# Требования: jq, curl, запущенный сервер на localhost:8080

BASE_URL="http://localhost:8080"
SEP="========================================"
TMP="/tmp/api_response.json"

# Выполнить запрос, вывести JSON и HTTP-статус
req() {
  local method="$1"; shift
  local url="$1"; shift
  local code
  code=$(curl -s -o "$TMP" -w "%{http_code}" -X "$method" "$@" "$url")
  jq . "$TMP" 2>/dev/null || cat "$TMP"
  echo "HTTP статус: $code"
}

# -------------------------------------------------------
# ЛР1: Тестовый эндпоинт
# Ожидаем: 200 OK, тело "Hello!"
# -------------------------------------------------------
echo "$SEP"
echo "ЛР1 | GET /test"
echo "Ожидаем: 200 OK, Hello!"
echo "$SEP"
req GET "$BASE_URL/test"

# -------------------------------------------------------
# ЛР2: Запись строки в БД
# Ожидаем: 201 Created, JSON {id, text, created_at}
# -------------------------------------------------------
echo ""
echo "$SEP"
echo "ЛР2 | POST /dbtest — запись строки в БД"
echo "Ожидаем: 201 Created, {id, text, created_at}"
echo "$SEP"
req POST "$BASE_URL/dbtest" \
  -H "Content-Type: application/json" \
  -d '{"text":"Тестовая строка из ЛР2"}'

# -------------------------------------------------------
# ЛР3: Регистрация пользователя
# Ожидаем: 201 Created, {id, login, created_at}
# Повтор с тем же логином → 409 Conflict
# -------------------------------------------------------
echo ""
echo "$SEP"
echo "ЛР3 | POST /auth/register — регистрация"
echo "Ожидаем: 201 Created, {id, login, created_at}"
echo "$SEP"
req POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"login":"john_doe","password":"secret123"}'

echo ""
echo "$SEP"
echo "ЛР3 | POST /auth/register — дубликат логина"
echo "Ожидаем: 409 Conflict"
echo "$SEP"
req POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"login":"john_doe","password":"other_pass"}'

# -------------------------------------------------------
# ЛР3: Авторизация — получение JWT
# Ожидаем: 200 OK, {token: "..."}
# -------------------------------------------------------
echo ""
echo "$SEP"
echo "ЛР3 | POST /auth/login — авторизация"
echo "Ожидаем: 200 OK, {token: \"<jwt>\"}"
echo "$SEP"
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"login":"john_doe","password":"secret123"}' | jq -r '.token')
echo "Получен токен: $TOKEN"

echo ""
echo "$SEP"
echo "ЛР3 | POST /auth/login — неверный пароль"
echo "Ожидаем: 401 Unauthorized"
echo "$SEP"
req POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"login":"john_doe","password":"wrongpass"}'

# -------------------------------------------------------
# ЛР4+5: Создание бронирования (требует JWT middleware)
# Ожидаем: 201 Created, {id, user_id, status:"new", ...}
# Без токена → 401 Unauthorized
# -------------------------------------------------------
echo ""
echo "$SEP"
echo "ЛР4+5 | POST /bookings/create — создание бронирования"
echo "Ожидаем: 201 Created, {id, user_id, status:\"new\", ...}"
echo "$SEP"
req POST "$BASE_URL/bookings/create" \
  -H "Authorization: Bearer $TOKEN"

echo ""
echo "$SEP"
echo "ЛР5 | POST /bookings/create — без токена (проверка middleware)"
echo "Ожидаем: 401 Unauthorized"
echo "$SEP"
req POST "$BASE_URL/bookings/create"

# -------------------------------------------------------
# ЛР4+5: Список бронирований пользователя
# Ожидаем: 200 OK, массив бронирований
# -------------------------------------------------------
echo ""
echo "$SEP"
echo "ЛР4+5 | GET /bookings/list — список бронирований"
echo "Ожидаем: 200 OK, [{id, user_id, status, ...}, ...]"
echo "$SEP"
req GET "$BASE_URL/bookings/list" \
  -H "Authorization: Bearer $TOKEN"

echo ""
echo "$SEP"
echo "Все тесты завершены."
echo "$SEP"
