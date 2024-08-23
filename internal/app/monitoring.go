package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"

	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
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
	file, err := NewLogFile(config.Host)
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
}

// RemoveWSConnection метод для удаления конкретнного вебсокет соединения
func (s *ServerLogger) RemoveWSConnection(conn *websocket.Conn) {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()
	for i, c := range s.wsConns {
		if c == conn {
			s.wsConns = append(s.wsConns[:i], s.wsConns[i:+1]...)
			break
		}
	}
	if err := conn.Close(); err != nil {
		logger.Error("Failed to close websocket connection: "+err.Error(), err)
	}
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

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return errors.New("Create session error: " + err.Error())
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		client.Close()
		session.Close()
		return errors.New("Create stdout pipe error: " + err.Error())
	}

	s.session = session
	s.client = client
	s.pipe = stdout

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

// StartLogging основной метод для работы чтения логгирования
func (s *ServerLogger) StartLogging(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	err := s.connect()
	if err != nil {
		logger.Error("Connection failed: "+err.Error(), err, slog.Any("server", s.config.Host))
		return
	}
	defer s.client.Close()

	startLine := "1"
	if s.config.StartLine != "0" {
		startLine = s.config.StartLine
	}

	cmd := fmt.Sprintf("tail -n +%s %s && tail -f %s", startLine, s.config.LogDir, s.config.LogDir)
	if err = s.session.Start(cmd); err != nil {
		logger.Error("Command start failed: "+err.Error(), err, slog.Any("server", s.config.Host))
		return
	}
	defer s.session.Close()

	scanner := bufio.NewScanner(s.pipe)
	currentLine, err := strconv.Atoi(s.config.StartLine)
	if err != nil {
		logger.Error("Parse start line failed: "+err.Error(), err, slog.Any("server", s.config.Host))
		return
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("Context cancelled", slog.Any("server", s.config.Host))
			return
		default:
			if scanner.Scan() {
				line := scanner.Text()
				s.broadcastLine(line, currentLine)
				err = s.File.PushLineWithLimit(line, config.Cfg.App.MaxLocalLogSizeMB)
				if err != nil {
					logger.Error("Push line failed: "+err.Error(), err, slog.Any("server", s.config.Host))
				}
				currentLine++
			} else {
				if scanner.Err() != nil {
					logger.Error("Scanner error", err, slog.Any("server", s.config.Host))
				}

				// reconnect here:
				s.Close()
				if err = s.reconnect(ctx); err != nil {
					logger.Error(err.Error(), err, slog.Any("server", s.config.Host))
					return
				}

				if err = s.session.Start(cmd); err != nil {
					logger.Error("Command start from reconnect failed: "+err.Error(), err, slog.Any("server", s.config.Host))
					return
				}
				output, _ := s.session.StdoutPipe()
				scanner = bufio.NewScanner(output)
			}
		}
	}
}

// broadcastLine метод для обработки новой строки при чтении с лог файла
func (s *ServerLogger) broadcastLine(line string, currentLine int) {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	fmt.Printf("[COPY_%d] %d. %s\n", s.ID, currentLine, line)

	for _, conn := range s.wsConns {
		err := conn.WriteMessage(websocket.TextMessage, []byte(line))
		if err != nil {
			logger.Warn(err.Error(), slog.Any("warn", err))
			conn.Close()
			s.RemoveWSConnection(conn)
		}
	}
}

// MultiLog метод для одновременного логирования и отправки сообщения пользователю, который смотрит логи в вебе
func (s *ServerLogger) MultiLog(message string, args ...any) {
	args = append(args, slog.Any("server", s.config.Host))
	logger.Info(message, args...)
	s.broadcastLine(message, -1)
}
