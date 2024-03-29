---Описание---
Приложение состоит из сервера и агентов которых можно запустить любое количество, для теста создано 3 агента.
Агенты работают с операциями: +, -, *, /, (, ) и целыми числа (так проще вводить в строке браузера)
Важно! Так как выражения добавляются прямо из строки браузера, операцию "+" необходимо заменить на соответствующий код %2B, так как
браузер знак "+" интерпритирует как пробел. Остальные символы -, *, /, (, ) заменять не нужно.
Например, что бы добавить для просчета выражение "2+2" нужно ввести "2%2B2"
Агент по умолчанию может забрать до 2 выражений на просчет

---Установка---
Необходимо установить дополнительные пакеты. Версия go - 1.21.
go get github.com/mattn/go-sqlite3
go get github.com/Knetic/govaluate

Можно удалить ./calc.db что бы она пересоздалась при запуске сервера, что бы не осталось тестовых решенных выражений.

---Сервер---
Запускается сервер из файла /server.go - на порту 8080, при первоначальной инициализации создается база данных sqlite ./calc.db
По умолчанию время работы каждой операции устанавливается в 10 секунд, в дальнейшем задержки можно изменить вводя команды.
После запуска становятся доступны команды для работы с сервером. Смотреть ниже.
Выражения и время выполнения операций записывается в базу данных, по этому после перезапуска сервера выражения остануться.

---Агент---
Агент запускается файлом /agent.go, для примера создано 3 агента. Каждый агент по умолчанию может забрать до 2 выражений на просчет
из списка выражений (можно изменить). Агент каждую секунду обращается к серверу, во время обращения его статус будет активным. Если в момент выполнения
агент был выключен и результат не был получен, то его статус изменится на неактивный, а сервер изменит статус задачи на "ожидание".

Для изменения характеристик агентов файл /agent.go
Изменение количество выражений которое агент может взять в работу: изменить строки 72 и 77 на больше чем 2.
При создании нового агента, в коде изменить "agent1" на другое имя, в строках: 81, 94, 104, 153.

---Задачи---
После добавления задачи, каждая имеет свой статус. Можно посмотреть как весь лист задач, так и конкретную задачу по ID.
Задачи имеют статусы: ожидание, в работе, завершено, ошибка. В моменте можно увидеть какая задача обрабатывается каким агентом.
По завершению будет получен результат вычисления.

---Операции---
Агенты работают с операциями: +, -, *, /, (, ) по умолчанию устанавливается время выполнение каждой операции в 10 секунд.
Что бы изменить время выполнения используйте команды.

---Список команд---
http://localhost:8080/list - выводит список добавленых выражений
http://localhost:8080/oper - выводит список операций с таймаутами
http://localhost:8080/oper?plus=20 - изменяет время выполнения операции сложения на 20 секунды, можно установить свое любое
http://localhost:8080/oper?minus=30
http://localhost:8080/oper?multiply=40
http://localhost:8080/oper?divide=60
http://localhost:8080/?calc=2%2B2 - добавляет выражение "2+2" = 4 в список задач %2B - это +, операцию "+" всегда нужно менять на код
http://localhost:8080/?calc=2*3 - выражение "2*2" = 6
http://localhost:8080/?calc=1-3*2 - выражение "1-3*2" = -5
http://localhost:8080/?calc=2%2B(2-2)/2*2 - выражение 2+(2-2)/2*2 = 2
http://localhost:8080/id?num=1 - выводит задачу по указаному ID
http://localhost:8080/agents - статусы агентов

---Рекомендации---
1. Запустите сервер, проверьте его работу выполнив команды:
http://localhost:8080/list - список должен быть пустым, если вы удалили тестовую базу
http://localhost:8080/oper - выводит список операций с таймаутами
2. Измените таймауты соответствующими командами на свое усмотрение
3. Накидайте выражений в очередь, штук 10, так как каждый агент заберет на выполнение по умолчанию по 2 штуки
4. Запустить поочередно 3 агентов
5. Проверьте их статусы командой
http://localhost:8080/agents
6. Обновляйте http://localhost:8080/list - выражения разбираются агентами
7. Добавьте еще выражений
8. Остановите одного из агентов, и проверьте статусы http://localhost:8080/agents, агент должен быть "false"
9. Проверьте http://localhost:8080/list выражение обрабатываемые отключенным агентов, должны перейти в статус "ожидание" через 10 секунд
10. Дождитесь обработки всех задач

* Можно сначала запустить сервер и всех агентов, а потом добавлять задачи, тогда агенты сразу будут брать их в работу.

---PS---
Где-то что-то упростил, прошу понять и простить (^_^)