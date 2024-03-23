# Readme
## Локальная компиляция на WINDOWS:
Выполнить последовательно команды в терминале:<br>
1. set GOARCH=amd64
2. set GOOS=linux
3. go build -o build/tg_bot cmd/app/main.go
## Удаленный запуск в Docker
Для запуска используется файл настройки docker-compose.yml
Запускается 3 контейнера: бот, БД и Adminer
## Удаленный запуск в Docker с отладкой
Для запуска используется файл настройки docker-compose.debug.yml
Запускается 3 контейнера: бот, БД и Adminer