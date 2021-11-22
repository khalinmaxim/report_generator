Сервис генерации отчётов


- Запускается в docker. 
- Файл конфигурации:
```
{
  "loglevel" : "${RP_LOG_LEVEL}",
  "dry": ${RP_DRY},
  "test": ${RP_TEST},
  "database": {
    "host": "${RP_HOST}",
    "port": ${RP_PORT},
    "name": "${RP_NAME}",
    "username": "${RP_USER}",
    "password": "${RP_PASSWORD}",
    "channel": "${RP_CHANNEL}",
    "timeout": ${RP_TIMEOUT}
  },
  "outpath": "${RP_OUTPATH}"
}
```
- Переменные окружения
  - RP_LOG_LEVEL - уровень журналирования
  - RP_DRY - сухой режим (не записывает отчёт и не удаляет из очереди заданий)
  - RP_TEST - тестовый режим (сохраняет тестовый отчёт)
  - RP_HOST - адрес СУБД
  - RP_PORT - порт СУБД
  - RP_NAME - имя БД
  - RP_USER - пользователь СУБД
  - RP_PASSWORD - пароль к СУБД
  - RP_CHANNEL - канал ожидания уведомления
  - RP_TIMEOUT - период ожидания уведомления

- Ожидает задания в очереди (таблица reporter.queue) и выполняет каждое последовательно.
- Сохраняет результат - электронную таблицу в формате XLSX в путь указанный в шаблоне.