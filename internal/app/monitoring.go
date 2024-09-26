package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"

	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/algrvvv/monlog/internal/notify"
)

type ServerLogger struct {
	ID      int                 // айди сервера, его индекс в списке всех серверов, включая 0
	config  config.ServerConfig // конфигурация сервера
	client  *ssh.Client         // ссш клиент
	session *ssh.Session        // ссш сессия
	pipe    io.Reader           // канал для чтения новых логов
	wsConns []*websocket.Conn   // массив вебсокет соединений, которые должны получить данные
	wsMutex sync.Mutex          // мьютекс для массива с вебсокетами
	File    *LogFile            // локальный файл, в котором сохраняется часть логов с сервера
}

// NewServerLogger метод для создания нового сервера для чтения логов
func NewServerLogger(id int, config config.ServerConfig) *ServerLogger {
	file, err := NewLogFile(config.Name, config.Enabled)
	if err != nil {
		logger.Error("Failed to create and open local clone log file: %v", err)
		return nil
	}
	return &ServerLogger{
		ID:     id,
		config: config,
		File:   file,
	}
}

// AppendWSConnection метод для добавления вебсокет соединения, которому будут отправлятся новые данные
func (s *ServerLogger) AppendWSConnection(conn *websocket.Conn) {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	s.wsConns = append(s.wsConns, conn)
	logger.Info("new ws connection saved", slog.Any("connections", s.wsConns))
}

// RemoveWSConnection метод для удаления конкретнного вебсокет соединения
func (s *ServerLogger) RemoveWSConnection(conn *websocket.Conn) {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	for i, c := range s.wsConns {
		if c == conn {
			// ХЫХВАХЫВХА НУ БЛЯТЬ... СУКА ИСКАЛ НЕСКОЛЬКО ДНЕЙ БЛЯДСКУЮ ОШИБКУ...
			// НАДО ЖЕ БЫЛО НАПИСАТЬ s.wsConns[i:+1] вмето s.wsConns[i+1:] D:
			// СПАСИБО ЗА ЧАСЫ ЭТОГО БЛЯДСКОГО ДЕБАГА ╭∩╮( •̀_•́ )╭∩╮
			s.wsConns = append(s.wsConns[:i], s.wsConns[i+1:]...)
			break
		}
	}
	logger.Info("ws connection removed", slog.Any("connections", s.wsConns))
}

// Close метод, который закрывает соединение с удаленным сервером
func (s *ServerLogger) Close() {
	s.client.Close()
	s.session.Close()
}

// connect метод для подключения к удаленному серверу
func (s *ServerLogger) connect() error {
	sshConfig, err := NewSSHConfig(config.Cfg.App.PathToIDRSA, s.config.User)
	if err != nil {
		return errors.New("Create client error: " + err.Error())
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return errors.New("Connect to remote server error: " + err.Error())
	}
	s.client = client

	return nil
}

// reconnect метод, который пытается переподключиться
func (s *ServerLogger) reconnect(ctx context.Context) error {
	s.MultiLog("Attempting to reconnect to remote server...")
	for {
		select {
		case <-ctx.Done():
			s.MultiLog("Reconnect stopped")
			return ctx.Err()
		default:
			s.MultiLog("Trying to reconnect...")
			if err := s.connect(); err == nil {
				s.MultiLog("Successfully reconnected")
				return nil
			} else {
				s.MultiLog("Reconnect failed: " + err.Error())
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (s *ServerLogger) StartLogging(ctx context.Context, wg *sync.WaitGroup) {
	if !s.config.Enabled {
		wg.Done()
		return
	}

	if s.config.IsLocal {
		s.startLocalLogging(ctx, wg)
	} else {
		s.startRemoteLogging(ctx, wg)
	}
}

// startLocalLogging метод для чтения логов с локальных файлов.
// Почему для локальных логов используется еще один локальный файл для сохранения? зачем дублировать?
// Это сделано из за того, что владелец файла - программа, которая туда записывает логи, не будет синхронизированна
// с потоком, который читает оттуда логи. Соответственно, это может повлечь непредсказуемые последствия.
func (s *ServerLogger) startLocalLogging(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer s.deleteLocalLogs()

	files := strings.Fields(s.config.LogDir)
	names := strings.Split(s.config.Name, ",")
	var lwg sync.WaitGroup

	startLine := "1"
	if s.config.StartLine != "0" {
		startLine = s.config.StartLine
	}

	for i, file := range files {
		var name string
		if len(names) != len(files) {
			name = names[0]
		} else {
			name = names[i]
		}

		lwg.Add(1)
		go func() {
			defer lwg.Done()
			command := fmt.Sprintf("tail -n +%s -f %s", startLine, file)
			cmd := exec.CommandContext(ctx, "sh", "-c", command)
			defer func() {
				if err := cmd.Wait(); err != nil {
					logger.Error("Failed to wait command: "+err.Error(), err, slog.String("server", s.config.Name))
				}
			}()

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				logger.Error("Failed to get stdout pipe: "+err.Error(), err, slog.String("server", s.config.Name))
				return
			}

			if err = cmd.Start(); err != nil {
				logger.Error("Failed to start command: "+err.Error(), err, slog.String("server", s.config.Name))
				return
			}

			scanner := bufio.NewScanner(stdout)
			currentLine, err := strconv.Atoi(s.config.StartLine)
			if err != nil {
				logger.Error("Failed to parse start line: "+err.Error(), err, slog.String("server", s.config.Name))
				return
			}

			for scanner.Scan() {
				var line string
				if len(names) == 1 {
					line = scanner.Text()
				} else {
					line = fmt.Sprintf("[%s] %s", name, scanner.Text())
				}
				s.broadcastLine(line, currentLine)
				err = s.File.PushLineWithLimit(line, config.Cfg.App.MaxLocalLogSizeMB)
				currentLine++
			}

			if err = scanner.Err(); err != nil {
				logger.Error("Failed to scan stdout: "+err.Error(), err, slog.String("server", s.config.Name))
			}
		}()
	}

	lwg.Wait()
}

// StartLogging основной метод для работы чтения логгирования
func (s *ServerLogger) startRemoteLogging(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer s.deleteLocalLogs()

	err := s.connect()
	if err != nil {
		logger.Error("Connection failed: "+err.Error(), err, slog.Any("server", s.config.Host))
		return
	}
	defer s.client.Close()

	rwg := sync.WaitGroup{}
	files := strings.Fields(s.config.LogDir)
	names := strings.Split(s.config.Name, ",")

	for i, file := range files {
		rwg.Add(1)
		go func() {
			defer rwg.Done()
			var (
				name        string
				currentLine int
				stdout      io.Reader
				session     *ssh.Session
			)

			if len(names) != len(files) {
				name = names[0]
			} else {
				name = names[i]
			}

			session, err = s.client.NewSession()
			if err != nil {
				s.client.Close()
				logger.Error("Failed to create session: "+err.Error(), err, slog.String("server", s.config.Host))
				return
			}

			stdout, err = session.StdoutPipe()
			if err != nil {
				s.client.Close()
				session.Close()
				logger.Error("Failed to get stdout pipe: "+err.Error(), err, slog.String("server", s.config.Name))
				return
			}

			startLine := "1"
			if s.config.StartLine != "0" {
				startLine = s.config.StartLine
			}
			cmd := fmt.Sprintf("tail -n +%s -f %s", startLine, file)

			if err = session.Start(cmd); err != nil {
				logger.Error("Command start failed: "+err.Error(), err, slog.Any("server", s.config.Host))
				return
			}

			scanner := bufio.NewScanner(stdout)
			done := make(chan struct{})
			currentLine, err = strconv.Atoi(s.config.StartLine)
			if err != nil {
				logger.Error("Parse start line failed: "+err.Error(), err, slog.Any("server", s.config.Host))
				return
			}

			go func() {
				defer close(done)
				for scanner.Scan() {
					var line string
					if len(names) == 1 {
						line = scanner.Text()
					} else {
						line = fmt.Sprintf("[%s] %s", name, scanner.Text())
					}
					s.broadcastLine(line, currentLine)
					if err = s.File.PushLineWithLimit(line, config.Cfg.App.MaxLocalLogSizeMB); err != nil {
						logger.Error("Push line failed: "+err.Error(), err, slog.Any("server", s.config.Host))
					}
					currentLine++
				}
				if err = scanner.Err(); err != nil {
					logger.Error("Scanner failed: "+err.Error(), err, slog.Any("server", s.config.Host))
				} else {
					logger.Info("Scanner finished", slog.Any("server", s.config.Host))
				}
			}()

			select {
			case <-ctx.Done():
				logger.Info("Getting signal, stopping session...", slog.Any("server", s.config.Host))
				if err = session.Signal(ssh.SIGTERM); err != nil {
					logger.Warn("Failed to send SIGTERM to server", slog.Any("server", s.config.Host))
				}

				select {
				case <-done:
					logger.Info("Session closed by SIGTERM", slog.Any("server", s.config.Host))
				case <-time.After(5 * time.Second):
					logger.Warn("Session dont closed by SIGTERM, trying SIGKILL...", slog.Any("server", s.config.Host))
					if err = session.Signal(ssh.SIGKILL); err != nil {
						logger.Error("Failed to send SIGKILL to server", err, slog.Any("server", s.config.Host))
					}

					// код ниже можно оставить, хотя по итогу на одном из серверов даже через килл процесс не завершался
					// после завершения программа и проверки этого процесса на сервере - его уже не было
					// так что оставлять это ожидание я не буду

					// logger.Info("Waiting for session to close", slog.Any("server", s.config.Host))
					// select {
					// case <-done:
					// 	 logger.Info("Session closed by SIGKILL", slog.Any("server", s.config.Host))
					// case <-time.After(10 * time.Second):
					//	 logger.Error("Failed to kill session", err, slog.Any("server", s.config.Host))
					// }
				}
			case <-done:
				logger.Info("Command has been finished successfully", slog.Any("server", s.config.Host))
			}
		}()
	}
	rwg.Wait()
}

// broadcastLine метод для обработки новой строки при чтении с лог файла
func (s *ServerLogger) broadcastLine(line string, currentLine int) {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	// logger.Info("broadcasting line", slog.Any("connections", s.wsConns))
	// go state.ParseLineAndSendNotify(s.ID, line)
	notify.Notifier.Nchan <- &notify.NQueueItem{ID: s.ID, Line: line}
	for _, conn := range s.wsConns {
		err := conn.WriteMessage(websocket.TextMessage, []byte(line))
		if err != nil {
			logger.Warn(err.Error(), slog.Any("warn", err))
		}
	}
}

// MultiLog метод для одновременного логирования и отправки сообщения пользователю, который смотрит логи в вебе
func (s *ServerLogger) MultiLog(message string, args ...any) {
	args = append(args, slog.Any("server", s.config.Host))
	logger.Info(message, args...)
	s.broadcastLine(message, -1)
}

func (s *ServerLogger) deleteLocalLogs() {
	logger.Info("Start delete local log file", slog.Any("server", s.config.Host))
	err := s.File.CLoseAndRemove()
	if err != nil {
		logger.Error("Failed to close log file: "+err.Error(), err)
	} else {
		logger.Info("Log file closed and removed successfully")
	}
}
