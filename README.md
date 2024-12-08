# Monlog

Monlog - мониторинг логов прямо в браузере в реальном времени, имеющий достаточно гибкую настройку и неплохой функционал!

### Перед началом
Изначально приложение задумывалось как self-hosted без возможности закрыть доступ извне,
но в последней версии `v0.9.1` был добавлен способ аутентификации пользователя. 
Чтобы узнать подробнее, перейдите к разделу [Авторизация и аутентификация](#авторизация-и-аутентификация)

### Установка

Клонируем репозиторий
```shell
git clone https://github.com/algrvvv/monlog.git
cd monlog
```

Запускаем проверку на наличие нужных утилит в системе
> ВАЖНАЯ ДЕТАЛЬ! Удаленные или локальные машины с которых идет чтение логов должны быть unix системами.
> Windows не поддерживается.

```shell
# проверяем на наличие нужных утилит без которых невозможна работа приложения
make install
```

Настраиваем конфигурацию после команды
```shell
cp config.example.yml config.yml
```

В файле с примером конфигурации уже лежат хорошие варианты использования и настройки под ваши нужды.
Можете ознакомиться с ним, а также с разделом [Настройка конфигурации](#настройка-конфигурации)

Делаем билд
```shell
make build
```

### Использование

Для использования вам достаточно настроить корректно `config.yml` и запустить приложение
```shell
./bin/monlog
```

Приложение откроется на указанном вами порте в конфигурации. После чего вы можете выбрать интересующий вас 
сервер и смотреть за его логами в реальном времени.

А если вы хотите отлучиться, то при срабатывании триггера на нужный уровень лога - вы получите,
если настроили, уведомление в телеграмм (или любое другое при использовании [кастомного драйвера]()).

### Настройка конфигурации

В этом разделе я попытаюсь показать вам все тонкости настройки приложения.
Для начала вам стоит просто глянуть файл `config.example.yml`, чтобы увидеть в нем комментарии
к каждому полю и примерно уже понимать, что мы будем обсуждать здесь.

### Настройка уведовлений

Для уведомления можно настроить несколько полей:
- "notify" - в этом поле можно указать название драйвера для уведомления
- "recipients" - массив получателей. яв-ся дополнительной настройкой, которая нужна для более тонкой работы. К примеру указать айди чатов в телеграмм получателей.

По дефолту есть несколько вариантов уведомлений:
- "none" или просто пустая строка "" - уведомления выключены.
- "telegram" - драйвер уведомления, который требует дополнительной настройки в виде колонки recipients с указанием айди чатов.
- "terminal" - драйвер уведомлений, который подходит для локальной разработки. Требует установленной утилиты `terminal-notifier`

В любой момент вы можете написать свой драйвер для уведомления. Для этого требуется реализовать интерфейс `notify.NotificationSender`, который получает уже настроенный `notify.Notification`:
```go
type NotificationSender interface {
	// Send метод для отправки уведомления. Принимает сервер типа config.ServerConfig и само сообщение.
	// Сервер нужен для того, чтобы оттуда достать данные для отправки, к примеру айди пользователей в тг.
	Send(notification *Notification) error
}
```

`notification` - параметр, который в себе уже содержит все информацию о логе. Ваше задача - отформатировать и настроить ее отправку.

Полный пример настройки драйвера: `/internal/notify/drivers`

Подробнее: `go doc notify.NotificationSender` и `go doc notify.Notification`, а также доступные параметры для сервера: `go doc config.ServerConfig`


#### Подключение к удаленному серверу

В данный момент для подключения к удаленному серверу, для большей безопасности, используется 
только способ с использованием ключей SSH. Если у вас они не настроены давайте быстренько сделаем это!

Для начала давайте создадим пару ключей SSH
```shell
ssh-keygen -t rsa -b 4096 -C "some comment"
```

После этого закинем на ваш удаленный сервер
```shell
ssh-copy-ide [remote_username]@[server_ip_address]
```

Вуаля! Теперь нам остается лишь указать путь до нашего `id_rsa` файла в конфигурации.
Обычно это `~/.ssh/id_rsa`

#### Настройка шаблона для парсинга строк логов

Для каждого сервера в файле конфигурации предусмотрено поле `log_layout`, которое отвечает за 
шаблон строки лога. По этому шаблону строка парситься и получает нужные данные, к примеру, для
уведомления. Для большего удобства в этой строке уже зарезервированы такие подстроки:
- `%LEVEL%` - та часть строки, в которой указывается уровень лога, к примеру INFO, ERROR и тд.
- `%TIME%` - подстрока со временем лога
- `%MESSAGE%` - часть строки, где идет непосредственно сообщение лога
- `...` - специальная вспомогательная конструкция равнозначная `*.?` в регулярных выражениях, которая подразумевает любые символы. Нужно для более детальной настройки вашего шаблона
А в остальном этот шаблон работает как обычные регулярные выражения в го.
Стоит также упомянуть, что формат времени в логах естественно будет отличаться от проекта к проекту, поэтому
`%TIME%` будет основываться на поле конфигурации `log_time_format`, которая подробнее рассмотрена в разделе [ниже](#форматы-дат)

А теперь я вам покажу для большей наглядности примеры таких строк под разные случаи жизни.
Для большего удобства я буду показывать пример самой строки лога и шаблон к ней.

```text
--- log ---
[2024.10.12 18:00:49] [INFO] database connected

--- layout ---
\\[%TIME%\\] \\[%LEVEL%\\] %MESSAGE%
```
В примере выше используются дополнительные `\` для правильного экранирования вашего шаблона.
Два их потому, что если мы оставим один `\`, то мы не сможем корректно распарсить наш `yaml` файл.

```text
--- log ---
LOG ERROR [24.09.27 15:12:14] Connection refused

--- layout ---
^\\[LOG\\] ... %LEVEL% \\[%TIME%\\] %MESSAGE%
```
* `LOG` - ваша зарезервированная строка, которая есть в ваших логах

```text
--- log ---
{"time":"03.10.24 11:29:16","level":"INF","msg":"200 GET /home","program_info":{"pid":54605,"go_version":"go1.23.1"}}

--- layout ---
^{\"time\":\"%TIME%\",\"level\":\"%LEVEL%\",\"msg\":\"%MESSAGE%\",\"program_info\":...}$
```

##### Использование драйверов

С последним обновлением появились драйвера для обработки строк логов. Это будет очень полезно и удобно
особенно для json логгирования. Вы можете использовать приведенный выше способ или использовать драйвера. Во втором
случае вам нужно будет самостоятельно написать драйвер для обработки строки лога, которая не будет перезаписывать
основную строку, а лишь изменять ее на момент отображения. Для начала разберем пример:

Допустим, у нас есть json строка: 
```json
{"message":"200 GET / ~0.0077750682830811","context":{"ip":"127.0.0.1"},"level":200,"level_name":"INFO","channel":"local","datetime":"2024-11-29T12:39:09.354013+03:00","extra":{}}
```

Для того, чтобы наш драйвер корректно использовался в нашей системе необходимо реализовать интерфейс `LineHandleDriver`:
```go
type LineHandleDriver interface {
	// GetName метод для получения названия драйвера
	GetName() (name string)
	// Handle метод для вашей кастомной обработки строки лога
	Handle(line string) (result string)
}
```

Пример полной реализации и загрузки драйвера вы можете увидеть в `internal/drivers/custom/`.
В файле `json_laravel_driver.go` можно увидеть пример конктрентной реализации драйвера.
Результатом его работы будет следующая строка из ранее показаного json:
```text
INFO [2024-11-29 12:30:23] 200 GET / ~0.0077750682830811
```

Преимущества такого способа заключается в гибкости. Вы в любой момент можете изменить логику и формат 
вывода сообщения из полученного джысона. Также драйвера можно переиспользовать, если структура и логика вашего
драйвера соответствует другому формату данных. Драйвера могут использоваться не только для json формата данных, но и для любого другого, 
тут уже дело вашей фантазии и реализации драйвера :D

После того, как вы написал реализацию собественного драйвера и загрузили ее в общий список через функцию
`init`, как в примере `json_laravel_driver.go`, то вам остается лишь в конфигурации к нужному серверу
указать параметр `log_driver` со значением, которое будет равно названию вашего драйвера (все примеры также есть в `config.example.yml`).

#### Форматы дат

> Важно. Если вы используете драйвера для отображения данных, то формат времени нужно подстраивать под итоговую
> строку, то есть строку, которая выдет после работы драйвера.

Для форматирования дат используется библиотека [metakeule/fmtdate](https://gitlab.com/metakeule/fmtdate), 
чтобы шаблоны дат было настроить удобнее и понятнее.

Примеры кодов, взятые из репозитория библиотеки:
```text
M    - month (1)
MM   - month (01)
MMM  - month (Jan)
MMMM - month (January)
D    - day (2)
DD   - day (02)
DDD  - day (Mon)
DDDD - day (Monday)
YY   - year (06)
YYYY - year (2006)
hh   - hours (15)
mm   - minutes (04)
ss   - seconds (05)

AM/PM hours: 'h' followed by optional 'mm' and 'ss' followed by 'pm', e.g.

hpm        - hours (03PM)
h:mmpm     - hours:minutes (03:04PM)
h:mm:sspm  - hours:minutes:seconds (03:04:05PM)

Time zones: a time format followed by 'ZZZZ', 'ZZZ' or 'ZZ', e.g.

hh:mm:ss ZZZZ (16:05:06 +0100)
hh:mm:ss ZZZ  (16:05:06 CET)
hh:mm:ss ZZ   (16:05:06 +01:00)
```

#### Наследование настроек

В примере ниже вы увидите как можно вынести повторяющиеся поля для настроек.

```yaml
default_servers_setting:
  start_line: 0
  notify: true
  log_levels: "ERROR"
  log_time_format: "YYYY-MM-DD hh:mm"
  chat_ids: [ "YOUR_TG_CHAT_ID", "ANOTHER_TG_CHAT_ID" ]

servers:
  - id: 84
    enabled: true
    name: Server 1
    log_dir: /path/to/server1/logs
    # наследует параметры из default_server_config
    
  - id: 98
    enabled: true
    name: Server 2
    log_dir: /path/to/server2/logs
    log_levels: "WARN|ERROR" 
    # переназначаем для этого сервера настройки `log_levels`
    # наследует параметры из default_server_config
```

### Авторизация и аутентификация
В этом разделе речь пойдет про возможность разместить приложение и не прибегать к каким то
способам закрытия доступа к вашему приложению от других. По умолчанию эта опция выключена, так что
если вам она не нужна - вы можете смело пропускать этот пункт и в документации, и в конфигурации :D
<br>

Включив эту опцию для правильной работы вам нужно будет использовать команду:
```shell
# если вы запускаете напрямую с использованием го
go run cmd/monlog/main.go --create-user
# либо используя бинарник (также можно использовать шортнейм: -c)
./bin/monlog -c
```

После этого вас попросят ввести логин и пароль для вашего пользователя, под которым
вы будете входить.

> **ВАЖНО:** Допускается иметь только одного пользователя, так как разделение по ролям никакого нет.

Время жизни сессии выставлено на одни сутки, в дальнейшем я вынесу это в конфигурацию.
Также стоит иметь в виду, что перезапустив сервер сессия станет неактуальной.

Все, теперь при переходе на сайт вам нужно будет обязательно залогиниться, прежде, чем получить 
доступ к своим серверам с логами.

Если вы нашли какие то баги или дырки в безопасности - смело бейте в ишьюсы!

### Есть ошибки или предложения?

Если вы заметили, что ваш шаблон строки лога почему то не работает, можете настроить
поле `app.debug` на `true`, чтобы увидеть, какой вид принимает ваш шаблон и убедиться подходит ли
составленное вами регулярное выражением подходим.
<br>

Если вы заметили, что на некоторые ошибки вам не приходит уведомление в телеграмм, возможно дело в 
том, что возможно логи пришли в одно и тоже время, которое имеет указанный вами шаблон, к примеру 
`20.09.24 14:00`. В таком случае оно будет записано в `state.yml` и при проверке следующего лога
программа посчитает его неактуальным, так как время либо равно, либо меньше сохраненного в "состоянии" (state.yml)

В таких случая рекомендуется использовать формат времени, в котором есть секунды.

<br>
С нетерпением жду ваши ишьюсы и реквесты :D

### TODO

- [x] в логгировании вебсокетов зачесался "!BADKEY"
- [x] сделать настраиваемое колво подгружаемых логгов до момента подключения
- [x] разобраться с багом вечного отключения сервера
- [x] исправить баг при повторном подключении к серверу логгирование останавливается
- [x] добавить возможность догружать логги до подключения пользователя к просмотру логгов
- [x] добавить местное сохранение логгов и ограничить по размерам с перезаписью
- [x] улучшить визуал страницы с логами. как минимум исправить баг с невозможностью скролла
- [x] разделить подключение к локальным или удаленным серверам
- [x] сделать возможность просмотра сразу нескольких файлов с одного сервера в одной вкладке
- [x] добавить валидацию структуры конфигурации
- [x] можно попробовать добавить ужесточенную валидацию. к примеру, для id_rsa - filepath
- [x] добавить в будущем дефолтные конфигурации для серверов
- [x] сделать тг бота с уведомлениями
- [x] добавить разные шаблоны логов для мульти серверов. Оставляем ограниченный функционал и менее гибкую настройку
- [x] парсер для джысон формата строки логов (слишком ситуативный момент, скорее всего будет удобнее просто смотреть на json строку, чтобы видеть всю полноту картины)
- [x] добавить возможность вынести айди чатов для тг в общие настройки (в данный момент все ломается)
- [x] исправить ошибку с рередерингом страницы с ошибкой
- [ ] можно попробовать добавить хот релоад конфига
