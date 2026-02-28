# AI Agent Go

RAG-агент на Golang с поддержкой:
- Разбиение Go файлов на логические блоки
- Векторная база данных (Chroma)
- Анализ кода и изменений
- Code completion с RAG
- OpenAI совместимый протокол

## Структура проекта

```
cmd/server/           # Точка входа приложения
internal/agent/       # Агенты анализа
internal/chroma/      # Интеграция с Chroma DB
internal/codebase/    # Парсинг и разбиение Go файлов
internal/rag/         # RAG логика
internal/server/      # HTTP сервер и API
pkg/api/              # API интерфейсы
pkg/models/           # Модели данных
```

## Запуск

```bash
go run cmd/server/main.go
```