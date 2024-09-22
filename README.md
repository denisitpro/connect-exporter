# Process Network Exporter

`Process Network Exporter` — это простой инструмент для мониторинга сетевых подключений заданных процессов и экспорта метрик в Prometheus. Он позволяет отслеживать количество активных подключений для каждого процесса и их наличие, что упрощает интеграцию с вашей системой мониторинга.

## Требования

- **ss**, **awk**, **uniq**, **sort** : Для корректной работы убедитесь, что эти утилиты установлены.
- **Тестировалось на Ubuntu 24.04**: Код был протестирован и проверен на Ubuntu 24.04.

## Особенности

- **Конфигурация через TOML**: Настройте список процессов для мониторинга с помощью простого TOML-файла.
- **Интеграция с Prometheus**: Экспортирует метрики в формате, совместимом с Prometheus.
- **Настраиваемое логирование**: Включите отладочное логирование для отслеживания входящих запросов, включая временную метку, IP-адрес клиента и запрашиваемый URI.
- **Версионирование**: Отображение версии, времени сборки и хэша коммита Git.

## Установка

Чтобы собрать экспортер, используйте следующую команду:

```bash
go build -ldflags "-X main.version=1.0.0 -X main.buildTime=$(date -u '+%Y-%m-%dT%H:%M:%SZ') -X main.commitHash=$(git rev-parse --short HEAD)" -o process_net_exporter main.go
```

### Использование Docker - todo сделать

Вы также можете запустить экспортер с использованием Docker. Готовый образ доступен в репозитории Docker Hub.

Чтобы скачать и запустить Docker-образ:

```bash
docker pull denisitpro/process-net-exporter:latest
docker run -d -p 9042:9042 -v /path/to/config.toml:/app/config.toml denisitpro/process-net-exporter:latest
```

- `-p 9042:9042` сопоставляет порт экспортера по умолчанию с вашим хостом.
- `-v /path/to/config.toml:/app/config.toml` монтирует ваш конфигурационный файл в контейнер.



## Использование

### Основное использование

Для запуска экспортера с использованием указанного конфигурационного файла:

```bash
./process_net_exporter -c /path/to/config.toml
```

### Режим отладки

Включите отладочное логирование с помощью флага `-v2`:

```bash
./process_net_exporter -v2 -c /path/to/config.toml
```

### Отображение информации о версии

Для отображения версии, времени сборки и хэша коммита:

```bash
./process_net_exporter --version
```

## Конфигурация

Конфигурация выполняется с помощью TOML-файла. Пример файла `config.toml`:

```toml
# Список процессов для мониторинга
processes = ["nginx", "mysql", "redis"]
```

## Метрики

Экспортер экспортирует следующие метрики:

- `process_connections_state_total`: Количество активных сетевых подключений для указанных процессов с метками `process_name` и `state`.

## Лицензия

Этот проект лицензирован по лицензии MIT.
