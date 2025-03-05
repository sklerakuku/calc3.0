
<div align="center">
  <h1>🧮 Распределённый вычислитель 🧮
<br> арифметических выражений </h1>
</div>


## <calc3.0>

### Содержание 📌

- [ℹ️ О проекте](#ℹ️-о-проекте)
- [⬇️ Установка](#установка)
- [▶️ Запуск](#запуск)
- [⁉️ Примеры использования](#примеры) 
- [📖 Документация "Как это всё работает?"](#документация)

## ℹ️ О проекте 
Это веб-сервис, который позволяет пользователям быстро, а иногда и не очень вычислять арифметические выражения. Пользователь отправляет арифметическое выражение через HTTP-запрос, и в ответ он получает результат вычисления. 

За основу взято старое решение веб-сервиса, но его функционал полностью переписан. Новое решение вычисляет части арифметического выражения параллельно, что значительно ускоряет процесс.

## ⬇️ Установка
Запустить проект можно несколькими способами. [Далее](#запуск) мы разберем несколько из них.

Но для начала, нужно установить сервис на ваше устройство. Для этого клонируйте репозиторий или скачайте и распакуйте [zip-архив](https://github.com/sklerakuku/calc3.0/archive/refs/heads/main.zip).
```bash
git clone https://github.com/sklerakuku/calc3.0.git
```
Перейдите в директорию проекта
```bash
cd calc3.0
```
## ▶️ Запуск
Запустите сервисы оркестратора и агента поочередно в разных терминалах.
```bash
go run ./cmd/orchestrator
```
```bash
go run ./cmd/agent
```
Или же одной командой
```bash
go run ./cmd/orchestrator && go run ./cmd/agent
```
Всё готово!! Теперь вы сервис доступен по адресу http://localhost:8080/ 

## ⁉️ Примеры использования
Вы можете воспользоваться для запросов как curl-ом, так и Postman-ом. Или использовать пользовательский интерфейс веб-сервиса.

 **Переменные  среды сервиса**
COMPUTING_POWER - количество горутин
TIME_ADDITION_MS - время выполнения операции сложения в миллисекундах  
TIME_SUBTRACTION_MS - время выполнения операции вычитания в миллисекундах  
TIME_MULTIPLICATIONS_MS - время выполнения операции умножения в миллисекундах  
TIME_DIVISIONS_MS - время выполнения операции деления в миллисекундах*


### Добавление вычисления арифметического выражения
*localhost:8080/api/v1/calculate'*
    
```bash
curl --location 'localhost/api/v1/calculate' \ --header 'Content-Type: application/json' \ --data '{ "expression": "2+2*6" }'
```
 Коды ответа: 201 - выражение принято для вычисления, 422 - невалидные данные, 500 - что-то пошло не так

Тело ответа
```json
{
    "id": "0"
}
```
<br>

### Получение списка выражений
*localhost:8080/api/v1/expressions'*
    
```bash
curl --location 'localhost/api/v1/expressions' 
```
Тело ответа
```json
{
    "expressions": [
        {
            "id": <идентификатор выражения>,
            "status": <статус вычисления выражения>,
            "result": <результат выражения>
        },
        {
            "id": <идентификатор выражения>,
            "status": <статус вычисления выражения>,
            "result": <результат выражения>
        }
    ]
}

```

Коды ответа:

-   200 - успешно получен список выражений
-   500 - что-то пошло не так
<br>
### Получение выражения по его идентификатору
*localhost:8080/api/v1/expressions/:id'*
 ```bash
curl --location 'localhost/api/v1/expressions/0'
```

Коды ответа:

-   200 - успешно получено выражение
-   404 - нет такого выражения
-   500 - что-то пошло не так

Тело ответа

```json
{
    "expression":
        {
            "id": <идентификатор выражения>,
            "status": <статус вычисления выражения>,
            "result": <результат выражения>
        }
}

```

## 📖  Документация
