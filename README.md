# Merch Store

## Описание

Сервис для управления покупками мерча и передачи монет между сотрудниками.

## Запуск

1. Убедитесь, что у вас установлены Docker и Docker Compose.
2. Клонируйте репозиторий.
3. Запустите сервис:
   ```bash
   docker-compose up --build
   ```
4. Сервис будет доступен по адресу http://localhost:8080

## Тестирование

Для запуска тестов сначала запустите тестовую бд:
```bash
docker-compose -f docker-compose.test.yml --project-name merch_test up -d
```
А затем запустите сами тесты:
```bash
go test -v ./... -p 1
```

Тесты лежат либо в необходимых папках в `/internal`, либо в `/test`

### Покрытие

При запуске 
```bash
go test ./... -cover -p 1
```
получаем следующий отчет:
```
        merch_store/cmd/server          coverage: 0.0% of statements
ok      merch_store/internal/auth       0.230s  coverage: 78.6% of statements
ok      merch_store/internal/db 0.733s  coverage: 71.1% of statements
ok      merch_store/internal/handlers   0.789s  coverage: 53.5% of statements
?       merch_store/internal/models     [no test files]
ok      merch_store/test        0.940s  coverage: [no statements]
```

## Вопросы и решения

**Вопрос:** Как обрабатывать ошибки при передаче монет?  
**Решение:** Мы добавили проверку на достаточность монет у отправителя и валидацию суммы перевода.

**Вопрос:** Тесты за счет зависимости от тестовой бд не экранированы друг от друга  
**Решение:** Добавляем параметр `-p 1`, запрещающий запускаться тестам параллельно и отчищаем бд каждый раз после завершения теста
