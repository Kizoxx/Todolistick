# Используем официальный образ Golang
FROM golang:1.24-alpine

# Создаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальные файлы проекта
COPY . .

# Собираем бинарник
RUN go build -o main .

# Команда запуска
CMD ["/app/main"]
