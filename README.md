Вот пример описания для вашего `README.md` на английском и русском языках.

---

### English

# Process Network Exporter

`Process Network Exporter` is a simple tool designed to monitor the network connections of specified processes and export metrics to Prometheus. It allows you to track the number of active connections for each process and the existence of these processes, making it easy to integrate with your monitoring stack.

## Features

- **TOML Configuration**: Configure the list of processes to monitor using a simple TOML file.
- **Prometheus Integration**: Exports metrics in a format compatible with Prometheus.
- **Customizable Logging**: Enable debug logging to track incoming requests, including timestamp, client IP, and requested URI.
- **Versioning**: Display the version, build time, and Git commit hash of the build.

## Installation

To build the exporter, use the following command:

```bash
go build -ldflags "-X main.version=1.0.0 -X main.buildTime=$(date -u '+%Y-%m-%dT%H:%M:%SZ') -X main.commitHash=$(git rev-parse --short HEAD)" -o process_net_exporter main.go
```

### Using Docker

You can also run the exporter using Docker. A pre-built Docker image is available in the Docker Hub repository.

To pull and run the Docker image:

```bash
docker pull denisitpro/process-net-exporter:latest
docker run -d -p 9042:9042 -v /path/to/config.toml:/app/config.toml denisitpro/process-net-exporter:latest
```

- `-p 9042:9042` maps the default exporter port to your host.
- `-v /path/to/config.toml:/app/config.toml` mounts your configuration file into the container.

## Usage

### Basic Usage

To run the exporter with a specified configuration file:

```bash
./process_net_exporter -c /path/to/config.toml
```

### Debug Mode

Enable debug logging by using the `-v2` flag:

```bash
./process_net_exporter -v2 -c /path/to/config.toml
```

### Display Version Information

To display the version, build time, and commit hash:

```bash
./process_net_exporter --version
```

## Configuration

The configuration is done via a TOML file. Below is an example `config.toml` file:

```toml
# List of processes to monitor
processes = ["nginx", "mysql", "redis"]
```

## Metrics

The following metrics are exported by the exporter:

- `process_network_connections`: Number of active network connections for specified processes, labeled by `process_name` and `state`.
- `process_exists`: Indicates if the process is running (1) or not (0), labeled by `process_name`.

## License

This project is licensed under the MIT License.

---

### Русский

# Process Network Exporter

`Process Network Exporter` — это простой инструмент для мониторинга сетевых подключений заданных процессов и экспорта метрик в Prometheus. Он позволяет отслеживать количество активных подключений для каждого процесса и их наличие, что упрощает интеграцию с вашей системой мониторинга.

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

### Использование Docker

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

- `process_network_connections`: Количество активных сетевых подключений для указанных процессов с метками `process_name` и `state`.
- `process_exists`: Показывает, запущен процесс (1) или нет (0), с меткой `process_name`.

## Лицензия

Этот проект лицензирован по лицензии MIT.

---
