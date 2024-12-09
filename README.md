## Кэширующий веб-сервер

Для того чтобы запустить проект, необходимо выполнить следующее:

1) Сначала, если необходимо, настроить переменные среды: [.env](.env).
   В базовом варианте, если необходимо, поменять только прокидываемые порты, чтобы не возникало конфликтов.
2) Выполнить "*make docker-run*" в корне проекта.

Для удобства тестирования преведён файл с запросами и ожидаемым результатом. [test.http](test.http).

P.S. Потенциально возможна ситуация, что миграции не успеют накатиться при первом запуске
(чего в ходе разработки не случалось). Ранее похожая ситуация решалась перезапуском, то есть выполнением пункта №2.

- Поскольку в ТЗ не указано, в каком виде передаётся токен: **Удаление документа [DELETE] /api/docs/<id>>**, то было
  принято решение передавать в качестве параметра.
- **Загрузка нового документа [POST] /api/docs**. Тут не было указано, чем именно является json в ответе, поэтому было
  принято решение, что это "данные документа", которые передавались в форме.
- Если не было чётко указано, в каком виде передаются параметры, то они передаются как праметры в URI
- Поскольку документы хранятся в ФС docker, то при повторной сборке "*make docker-run*" файлы удалятся.