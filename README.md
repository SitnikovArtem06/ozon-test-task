## Запуск

1. Скопировать пример конфига:

```bash
cp .env.example .env
```


2. Запуск с `in-memory`:


```bash
STORAGE=memory docker compose up --build
```

3. Запуск с `postgres`:

```bash
STORAGE=postgres docker compose up --build
```

миграции поднимаются автоматически
