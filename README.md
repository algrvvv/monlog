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
если настроили, уведомление в телеграмм.


### Настройка конфигурации

В этом разделе я попытаюсь показать вам все тонкости настройки приложения.
Для начала вам стоит просто глянуть файл `config.example.yml`, чтобы увидеть в нем комментарии
к каждому полю и примерно уже понимать, что мы будем обсуждать здесь.

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

#### Форматы дат

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
