const loader = document.getElementById('loader');
const mainContent = document.getElementById('content');
let lastLine = 0;

function switchLoaderAndContent() {
    const loaderDisplay = loader.style.display;
    if (loaderDisplay === 'none') {
        loader.style.display = 'block';
        mainContent.style.display = 'none';
    } else {
        loader.style.display = 'none';
        mainContent.style.display = 'block'
    }
}

CodeMirror.defineMode("log", function (config, parserConfig) {
    function tokenBase(stream, state) {
        if (stream.eat("[") && stream.skipTo("]")) {
            let level = stream.current().slice(1, -1); // Получаем текст внутри скобок
            stream.eat("]");
            return tokenLog(level)
        }

        if (stream.match(/\d{4}\/\d{2}\/\d{2}/)) { // Обрабатываем дату в формате 2024/08/21
            return "date";
        }

        if (stream.match(/\d{2}\.\d{2}\.\d{2}/)) { // Обрабатываем дату в формате 21.08.24
            return "date";
        }

        if (stream.match(/\d{2}:\d{2}:\d{2}/)) { // Обрабатываем время в формате 14:03:17
            return "time";
        }

        if (stream.match(/INFO|DEBUG|WARN|ERROR/)) { // Обрабатываем уровни логирования
            return tokenLog(stream.current())
        }

        if (stream.match(/[\w:.,\-]+/)) { // Обрабатываем сообщения и другие токены
            return "string";
        }

        stream.next();
        return null;
    }

    function tokenLog(level) {
        if (["ERROR", "FATAL"].includes(level)) {
            return "error"; // Подсвечиваем ошибки
        } else if (["WARN", "WARNING"].includes(level)) {
            return "warn"; // Подсвечиваем предупреждения
        } else if (["INFO", "DEBUG"].includes(level)) {
            return "info"; // Подсвечиваем информационные и отладочные сообщения
        }

        return "log-level"; // Для всех остальных уровней логирования
    }

    return {
        startState: function () {
            return {tokenize: tokenBase};
        },
        token: function (stream, state) {
            return state.tokenize(stream, state);
        }
    };
});

function getLogMode(text) {
    if (text.trim().startsWith("{")) {
        return "application/json"
    } else {
        return "log"
    }
}

const editor = CodeMirror.fromTextArea(document.getElementById("logs"), {
    // mode: getLogMode(document.getElementById("code").value),
    mode: "log",
    lineNumbers: true,
    theme: "default",
    fontFamily: 'JetBrains Mono, monospace',
    readOnly: true,
    viewportMargin: Infinity
});

function autoScrollBottom() {
    editor.scrollTo(null, editor.getScrollInfo().height);
}

editor.on('change', function () {
    autoScrollBottom()
})

function addLogLinesToStart(lines) {
    const currentContent = editor.getValue();
    const newContent = lines.join('\n') + "\n" + currentContent;
    editor.setValue(newContent);
}

function addLogLine(newLog) {
    editor.replaceRange("\n" + newLog, CodeMirror.Pos(editor.lineCount()));
}

let highlights = [];
const m = document.getElementById('matches');

function searchAndHighlight(query) {
    // Очистка предыдущих подсветок
    highlights.forEach(mark => mark.clear());
    highlights = [];
    m.innerHTML = 0;

    if (!query) return; // Выход если запрос пуст

    const cursor = editor.getSearchCursor(query, null, {caseFold: true});

    // Поиск и подсветка всех совпадений
    while (cursor.findNext()) {
        const mark = editor.markText(cursor.from(), cursor.to(), {className: "highlight"});
        highlights.push(mark);
    }

    m.innerHTML = highlights.length;
}

const searchInput = document.getElementById('search')
searchInput.oninput = function () {
    searchAndHighlight(searchInput.value)
}

const urlParts = window.location.href.split("/");
const servID = urlParts[urlParts.length - 1];

const axiosURL = `http://${location.host}/api/v1/logs/prev/${servID}`
console.log(axiosURL)

async function getPrevRows() {
    try {
        const response = await axios.get(axiosURL);
        console.log(response);
        lastLine = response.data.lastLine;
        addLogLinesToStart(response.data.lines);
        autoScrollBottom()
        switchLoaderAndContent();
    } catch (error) {
        console.error("Error getting prev rows: ", error)
    }
}

getPrevRows();

function getPrevRowsByCount(lineCount) {
    const url = `http://${location.host}/api/v1/logs/prev/count/${servID}?start=${lastLine - lineCount}&end=${lastLine}`
    axios.get(url).then(res => {
        console.log(res)
        lastLine = res.data.lastLine
        addLogLinesToStart(res.data.lines)
        editor.scrollTo(0, 0);
    }).catch(err => {
        console.error(err)
    })
}

const wsURL = `ws://${location.host}/ws/${servID}`;
console.log(wsURL)
const socket = new WebSocket(wsURL);

socket.onopen = function (e) {
    console.log("connection established");
}

socket.onmessage = function (event) {
    console.log(event.data)
    addLogLine(event.data)
}

socket.onerror = function (error) {
    console.error("connection error", error)
}

socket.onclose = function (event) {
    console.log("connection closed")
}
