[![Contributors](https://img.shields.io/github/contributors/bryack/obsidian_rag.svg?style=flat)](https://github.com/bryack/obsidian_rag/graphs/contributors)
[![Stargazers](https://img.shields.io/github/stars/bryack/obsidian_rag.svg?style=flat)](https://github.com/bryack/obsidian_rag/stargazers)
[![Issues](https://img.shields.io/github/issues/bryack/obsidian_rag.svg?style=flat)](https://github.com/bryack/obsidian_rag/issues)
[![LinkedIn](https://img.shields.io/badge/-LinkedIn-black.svg?style=flat&logo=linkedin&colorB=0077B5)](https://linkedin.com/in/anna-nurgaleeva-ba9a6338)

# Obsidian RAG: Персональный ИИ-ассистент для глубокого анализа знаний

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white) ![Qdrant](https://img.shields.io/badge/Qdrant-E94E1B?style=flat&logo=qdrant&logoColor=white) ![Ollama](https://img.shields.io/badge/Ollama-000000?style=flat) ![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat&logo=docker&logoColor=white) ![License](https://img.shields.io/badge/License-MIT-green.svg)

Obsidian RAG — это высокопроизводительная локальная система поиска и генерации ответов (Retrieval-Augmented Generation), разработанная специально для работы с объемными базами заметок в Obsidian.

## Мотивация проекта

При работе с большими базами знаний (9500+ заметок) стандартные решения часто дают сбой:
1. **Контекстное окно**: Простое подключение LLM к папке с заметками быстро переполняет память модели.
2. **Потеря информации**: Поиск по ключевым словам не находит данные, записанные другими словами.
3. **Приватность**: Личные заметки не должны покидать локальную машину.

Этот проект решает данные проблемы, используя **гибридный поиск BM25 + Dense** и Clean Architecture.

## Ключевые особенности

- **BM25 Hybrid Search**: Комбинация семантического поиска (BGE-M3) с BM25 для точного поиска по ключевым словам. BM25 учитывает редкость термов (IDF) и нормализует по длине документа.
- **Retrieval Quality Evaluation**: Автоматическая оценка качества поиска через acceptance tests с метриками Precision@K, Recall@K, MRR.
- **Scoped Search**: Ограничение области поиска конкретными папками без потери производительности.
- **Clean Architecture**: Полное разделение бизнес-логики и инфраструктуры (Qdrant, Ollama, FileSystem).
- **Stats Persistence**: BM25-статистика сохраняется между перезапусками (`.stats/bm25_stats.json`), что позволяет искать без повторной индексации.

## Технологический стек

- **Язык**: Go 1.25+ с Clean Architecture (Hexagonal)
- **Векторная БД**: Qdrant (gRPC) — хранение эмбеддингов и sparse vectors
- **LLM**: Ollama (bge-m3 для эмбеддингов, qwen3.5 для генерации)
- **Парсинг**: Goldmark с поддержкой Wikilinks и Frontmatter
- **Тестирование**: Testcontainers-go для интеграционных тестов

## Архитектура системы

```
internal/domain/     # Ядро: RagEngine, BM25Stats, интерфейсы
adapters/           # Реализации: Qdrant, Ollama, FileSystem, StatsRepo
cmd/cli/           # CLI интерфейс
testcases/         # Acceptance tests с ground truth
```

## Быстрый старт

### Предварительные требования
- Go 1.25+
- Docker (для Qdrant)
- Ollama с моделями `bge-m3` и `qwen3.5:9b`

### Индексация
```bash
go run cmd/cli/main.go index /path/to/obsidian
```
При первой индексации:
- **Pass 1**: Сбор BM25-статистики по всем документам
- **Pass 2**: Индексация с BM25-весами (только изменённые файлы)

### Поиск
```bash
# Поиск фрагментов
go run cmd/cli/main.go ask /path/to/obsidian "Что такое каналы в Go?"

# Поиск в конкретной папке (scoped search)
go run cmd/cli/main.go -folder "Work/Projects" ask /path/to/obsidian "Project status"

# RAG с генерацией ответа
go run cmd/cli/main.go -llm ask /path/to/obsidian "Сформулируй саммари проекта X"
```

### Тестирование

**Unit и Integration тесты** (быстрые):
```bash
go test ./... -short
```

**Acceptance тесты** (требуют запущенного Qdrant и Ollama):
```bash
go test ./testcases/... -run TestRetrievalQualityEvaluation
```

> 💡 **Fedora / SELinux**: Если Testcontainers не запускаются (проблемы с RYUK), используйте:
> ```bash
> export TESTCONTAINERS_RYUK_DISABLED=true
> go test ./testcases/... -run TestRetrievalQualityEvaluation
> ```
> На большинстве систем (Ubuntu, macOS, Windows) RYUK работает стабильно и отключать не нужно.

> ⚠️ **Важно**: Acceptance tests используют `testcases/fixtures/ground_truth.yaml` с ручной разметкой релевантных документов. Для своего vault необходимо создать собственный ground truth, иначе тесты будут падать.

## Текущие результаты Retrieval Quality

| Запрос | Precision@5 | Статус |
|--------|-------------|--------|
| North Star (редкие термины) | 0.60 | ✅ |
| Hebset project | 0.60 | ✅ |
| Каналы в Go | 0.80 | ✅ |
| Docker удалить контейнер | 0.80 | ✅ |
| **Average** | **0.70** | ✅ (target: ≥0.60) |

## Контакты

- Telegram: [@bryacka](https://t.me/bryacka)
- LinkedIn: [Anna Nurgaleeva](https://www.linkedin.com/in/anna-nurgaleeva-ba9a6338)

---
License: MIT
