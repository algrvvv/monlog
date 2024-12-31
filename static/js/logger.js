const nativeLog = console.log;
const nativeWarn = console.warn;
const nativeError = console.error;

function info(...args) {
  nativeLog(
    "%cINFO",
    "background:#4CAF50;color: #111;padding:0 5px;border-radius:10px",
    ...args
  );
}

function warn(...args) {
  nativeWarn(
    "%cWRN",
    "background:#FFC107;color: #111;padding:0 5px;border-radius:10px",
    ...args
  );
}

function error(...args) {
  nativeError(
    "%cERR",
    "background:#F44336;color: #111;padding:0 5px;border-radius:10px",
    ...args
  );
}

function debug(...args) {
  if (!debugLogs) return;
  nativeLog(
    "%cDEBG",
    "background:#2196F3;color: #111;padding:0 5px;border-radius:10px",
    ...args
  );
}

console.log = info;
console.warn = warn;
console.error = error;
console.debug = debug;
