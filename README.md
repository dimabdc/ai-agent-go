# AI Agent Go

AI агенты для автоматизации code review и генерации QA отчетов с использованием LLM.

## Структура проекта

```
cmd/codereview/       # AI Code Review Agent для Gitea
cmd/qareport/         # QA Report Agent для Jira
internal/codereview/  # Code Review Agent логика
internal/codereview/tools/  # Инструменты Code Review
internal/qareport/    # QA Report Agent логика
internal/qareport/tools/    # Инструменты QA Report
internal/gitea/       # Интеграция с Gitea (официальный SDK)
internal/jira/        # Интеграция с Jira (go-jira SDK)
```

## Основные модели
- Chat: `qwen3.5-35b-a3b` (LM Studio)

## Сборка и установка

### Требования
- Go 1.21 или выше
- Git
- Файл `.env` с необходимыми переменными окружения

### Сборка из исходного кода

#### Code Review Agent
```bash
# Собрать бинарный файл
go build -o codereview ./cmd/codereview

# Или установить в $GOPATH/bin
go install ./cmd/codereview
```

#### QA Report Agent
```bash
# Собрать бинарный файл
go build -o qareport ./cmd/qareport

# Или установить в $GOPATH/bin
go install ./cmd/qareport
```

### Установка через Docker (опционально)

Создайте `Dockerfile` в корне проекта:

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o codereview ./cmd/codereview
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o qareport ./cmd/qareport

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/codereview /app/qareport .
COPY .env.example .env

ENTRYPOINT ["/app/codereview"]
```

### Проверка установки

```bash
# Справка по использованию
./codereview --help
./qareport --help

# Или если установлено в GOPATH
codereview --help
qareport --help
```

## AI Code Review Agent

Агент для автоматического code review Pull Request'ов в Gitea с использованием LLM.

### Архитектура агента

Code Review Agent реализует многопроходный pipeline анализа кода с использованием ADK (Agent Development Kit):

```
Planner → [Explorer ↔ Reviewer ↔ LoopBreak] → Reflector
          └──── Loop (max 6 итераций) ────┘
```

#### Подзадачи агентов

**1. Planner Agent** — планирование анализа
- Извлекает Jira задачу из названия PR (regex `#(\\w+-\\d+)`)
- Формирует план анализа diff на основе описания задачи
- Использует инструмент `get_pr_diff` для получения изменений

**2. Explorer Agent** — сбор контекста
- Рекурсивно обходит файлы в Head ветке PR через `list_files`
- Читает содержимое файлов через `get_file_content`
- Запрашивает недостающий контекст у LLM при необходимости

**3. Reviewer Agent** — анализ кода
- Проводит глубокий аудит изменений:
  - Безопасность (уязвимости, инъекции, XSS)
  - Архитектура (нарушение паттернов, разделения ответственности)
  - Бизнес-логика (корректность реализации)
  - Регрессии (требует конкретного доказательства из кода)
- Оценивает уверенность изменений (Высокая/Средняя/Низкая)

**4. LoopBreak Agent** — контроль итераций
- Определяет завершённость анализа
- Прерывает цикл при достижении стабильности результатов

**5. Reflector Agent** — валидация и публикация
- Проверяет качество отчёта
- Публикует отчет комментарием к PR через `post_comment`

### Возможности

#### Анализ кода
- **Полный анализ diff PR**: получение изменений через Gitea SDK (`GetPullRequestDiff`)
- **Чтение файлов**: содержимое файлов через `GetFile`
- **Обход директорий**: рекурсивный поиск файлов в Head ветке PR
- **Контекстный обмен**: динамический запрос недостающих данных у LLM

#### Глубокий аудит
- **Безопасность**: выявление уязвимостей (SQL injection, XSS, CSRF)
- **Архитектура**: проверка паттернов проектирования, разделения ответственности
- **Бизнес-логика**: валидность реализации требованиям из Jira
- **Регрессии**: обнаружение потенциальных проблем с существующим кодом

#### Интеграция
- **Jira**: автоматическое извлечение задачи по ключу из названия PR
- **Gitea**: официальный SDK (`code.gitea.io/sdk/gitea`) для работы с PR
- **Автоматическая публикация**: комментарии в PR через `CreateIssueComment`

### Запуск Code Review Agent

```bash
go run cmd/codereview/main.go <PR_URL>
```

### Параметры запуска

#### CLI-аргументы (go-flags)
| Ключ | Переменная окружения | Описание | Значение по умолчанию |
|------|----------------------|----------|------------------------|
| `--jira-url` | `JIRA_BASE_URL` | Jira API base URL (*обязательный*) | * required * |
| `--jira-token` | `JIRA_TOKEN` | Jira API token (*обязательный*) | * required * |
| `--gitea-url` | `GITEA_BASE_URL` | Gitea API base URL (*обязательный*) | * required * |
| `--gitea-token` | `GITEA_TOKEN` | Gitea API token (*обязательный*) | * required * |
| `--openai-key` | `OPENAI_API_KEY` | OpenAI API key (*обязательный*) | * required * |
| `--openai-url` | `OPENAI_BASE_URL` | OpenAI-compatible API base URL (для LM Studio, Ollama и т.д.) | `http://localhost:1234/v1` |
| `--model` | `CHAT_MODEL` | Имя модели chat LLM | `qwen3.5-35b-a3b` |

#### Позиционный аргумент
- `<PR_URL>` — URL Pull Request'а в формате `https://gitea.example.com/owner/repo/pulls/123`

### Инструменты Code Review Agent
- `get_pr_diff` — получение diff Pull Request'а через `GetPullRequestDiff`
- `get_file_content` — чтение контента файла через `GetFile`
- `list_files` — получение списка файлов в Head ветке PR
- `post_comment` — публикация комментария в PR через `CreateIssueComment`

## QA Report Agent

Агент для автоматической генерации QA отчетов задач Jira с использованием LLM.

### Архитектура агента

QA Report Agent реализует двухэтапный pipeline:

```
QA Analysis Agent → Reflector Agent
```

#### Подзадачи агентов

**1. QA Analysis Agent** — генерация отчёта
- Получает задачу из Jira по ключу (например, `PRJ-123`)
- Извлекает ссылки на PR из кастомного поля `customfield_20702` (разделены новой строкой)
- Для каждого PR:
  - Получает diff через Gitea SDK (`GetPullRequestDiff`)
  - Получает список файлов в ветке `list_files`
  - Читает содержимое файлов через `GetFile`
- Генерирует отчёт на русском языке с анализом:
  - **Цепочка влияния**: оценка потенциального воздействия изменений на другие компоненты системы
  - **Сценарии для ручного тестирования**: набор тестовых сценариев для QA-инженера
  - **Оценка рисков**: выявление потенциальных проблем в коде
  - **Тестирование функциональности**: проверка корректности реализации требований

**2. Reflector Agent** — валидация отчёта
- Проверяет качество и полноту отчёта
- Публикует результат как комментарий в задаче Jira через `post_comment`

### Возможности

#### Интеграция с Jira
- **Получение задач**: через go-jira SDK
- **Извлечение PR**: чтение кастомного поля `customfield_20702` с ссылками на PR
- **Автоматическая публикация**: отчёт публикуется как комментарий в задаче Jira

#### Анализ изменений
- **Получение diff PR'ов**: официальный Gitea SDK (`code.gitea.io/sdk/gitea`)
- **Чтение файлов**: содержимое файлов через `GetFile`
- **Обход директорий**: рекурсивный поиск файлов в PR
- **Обработка ошибок**: пропущенные PR'ы, невалидные URL, неудачное получение diff

#### Генерация отчёта
- **Формат Jira Wiki Markup**: структурированный отчёт для Jira
- **Русский язык**: все тексты на русском с форматированием
- **Конкретика без вымысла**: не придумывает цифры без логики из кода
- **Итерационный анализ**: до 50 итераций для глубокого анализа (MaxIterations: 50)

### Запуск QA Report Agent

```bash
go run cmd/qareport/main.go <TASK_KEY>
```

### Параметры запуска

#### CLI-аргументы (go-flags)
| Ключ | Переменная окружения | Описание | Значение по умолчанию |
|------|----------------------|----------|------------------------|
| `--jira-url` | `JIRA_BASE_URL` | Jira API base URL (*обязательный*) | * required * |
| `--jira-token` | `JIRA_TOKEN` | Jira API token (*обязательный*) | * required * |
| `--gitea-url` | `GITEA_BASE_URL` | Gitea API base URL (*обязательный*) | * required * |
| `--gitea-token` | `GITEA_TOKEN` | Gitea API token (*обязательный*) | * required * |
| `--openai-key` | `OPENAI_API_KEY` | OpenAI API key (*обязательный*) | * required * |
| `--openai-url` | `OPENAI_BASE_URL` | OpenAI-compatible API base URL (для LM Studio, Ollama и т.д.) | `http://localhost:1234/v1` |
| `--model` | `CHAT_MODEL` | Имя модели chat LLM | `qwen3.5-35b-a3b` |

#### Позиционный аргумент
- `<TASK_KEY>` — Jira task key в формате `PRJ-123`

### Инструменты QA Report Agent
- `get_pr_diff` — получение diff Pull Request'а через Gitea SDK
- `get_file_content` — чтение контента файла через `GetFile`
- `list_files` — получение списка файлов в Head ветке PR
- `post_comment` — публикация отчета как комментарий в задаче Jira

## Настройка окружения

### Переменные окружения (.env)

Файл `.env`:

```env
# Jira конфигурация
JIRA_BASE_URL=https://jira.example.com/rest/api/2
JIRA_TOKEN=your-jira-api-token

# Gitea конфигурация
GITEA_BASE_URL=https://gitea.example.com/api/v1
GITEA_TOKEN=your-gitea-api-token

# LLM конфигурация (OpenAI-compatible API)
OPENAI_BASE_URL=http://localhost:1234/v1
CHAT_MODEL=qwen3.5-35b-a3b
OPENAI_API_KEY=your-openai-api-key
```

### Описание переменных

#### Jira
| Переменная | Описание | Обязательная | Значение по умолчанию |
|------------|----------|--------------|----------------------|
| `JIRA_BASE_URL` | Базовый URL Jira API | Да | * required * |
| `JIRA_TOKEN` | API токен для аутентификации | Да | * required * |

#### Gitea
| Переменная | Описание | Обязательная | Значение по умолчанию |
|------------|----------|--------------|----------------------|
| `GITEA_BASE_URL` | Базовый URL Gitea API | Да | * required * |
| `GITEA_TOKEN` | API токен для аутентификации | Да | * required * |

#### LLM (OpenAI-compatible)
| Переменная | Описание | Обязательная | Значение по умолчанию |
|------------|----------|--------------|----------------------|
| `OPENAI_BASE_URL` | URL OpenAI-compatible API (LM Studio, Ollama и т.д.) | Нет | `http://localhost:1234/v1` |
| `CHAT_MODEL` | Имя модели chat LLM | Нет | `qwen3.5-35b-a3b` |
| `OPENAI_API_KEY` | API ключ (требуется даже для локальных моделей) | Да | * required * |

### Приоритет конфигурации

1. **CLI-аргументы** — максимальный приоритет (переопределяют все остальные)
2. **Переменные окружения** — средний приоритет
3. **Значения по умолчанию** — минимальный приоритет

### Примеры запуска

#### Code Review Agent с .env
```bash
# Файл .env настроен (все обязательные параметры)
go run cmd/codereview/main.go https://gitea.example.com/owner/repo/pulls/123
```

#### Code Review Agent с CLI-аргументами
```bash
go run cmd/codereview/main.go \
  --jira-url=https://jira.example.com/rest/api/2 \
  --jira-token=jira-api-token \
  --gitea-url=https://gitea.example.com/api/v1 \
  --gitea-token=gitea-api-token \
  --openai-key=openai-api-key \
  https://gitea.example.com/owner/repo/pulls/123
```

#### QA Report Agent с .env
```bash
# Файл .env настроен (все обязательные параметры)
go run cmd/qareport/main.go PRJ-123
```

#### QA Report Agent с CLI-аргументами
```bash
go run cmd/qareport/main.go \
  --jira-url=https://jira.example.com/rest/api/2 \
  --jira-token=jira-api-token \
  --gitea-url=https://gitea.example.com/api/v1 \
  --gitea-token=gitea-api-token \
  --openai-key=openai-api-key \
  PRJ-456
```

#### Переопределение переменной окружения через CLI
```bash
# .env содержит базовые настройки, но модель меняется через CLI
go run cmd/qareport/main.go \
  --model=qwen3.5-35b-a3b \
  PRJ-789
```

#### Использование LM Studio (локальная модель)
```bash
# .env с локальным endpoint
export OPENAI_BASE_URL=http://localhost:1234/v1
export CHAT_MODEL=qwen3.5-35b-a3b
export OPENAI_API_KEY=any-value # требуется, но не используется

go run cmd/codereview/main.go https://gitea.example.com/owner/repo/pulls/123
```
