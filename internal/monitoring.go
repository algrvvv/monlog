package internal

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log/slog"
	"strconv"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"

	"github.com/algrvvv/monlog/internal/app"
	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/algrvvv/monlog/internal/server/handlers"
)

var (
	serverID int
	RSAPath  string
)

func ConnectAndReadLogs(id int, s config.ServerConfig) {
	serverID = id
	RSAPath = config.Cfg.App.PathToIDRSA
	newSession(s)
}

func newSession(s config.ServerConfig) {
	clientConfig, err := app.NewSSHConfig(RSAPath, s.User)
	if err != nil {
		logger.Error(err.Error(), err)
	}

	logger.Info("Connection to server...")
	var client *ssh.Client
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	client, err = ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		logger.Error(err.Error(), err)
	}
	defer client.Close()

	logger.Info("Connected to server", slog.Any("server", addr))

	session, err := client.NewSession()
	if err != nil {
		logger.Error(err.Error(), err)
	}
	defer session.Close()
	logger.Info("Created new session to server", slog.Any("server", addr))

	var startLine string
	if s.StartLine == "0" {
		startLine = "1"
	} else {
		startLine = s.StartLine
	}
	cmd := fmt.Sprintf("tail -n +%s %s && tail -f %s", startLine, s.LogDir, s.LogDir)
	reader, writer := io.Pipe()
	defer reader.Close()

	session.Stdout = writer
	session.Stderr = writer

	err = session.Start(cmd)
	if err != nil {
		logger.Error(err.Error(), err)
	}

	go func() {
		scanner := bufio.NewScanner(reader)
		currentLine, _ := strconv.Atoi(startLine)
		for scanner.Scan() {
			line := scanner.Text()
			broadcastLine(line, currentLine)
			currentLine++
		}

		if err = scanner.Err(); err != nil && err != io.EOF {
			log.Fatalf("Error reading output: %v", err)
		}
	}()

	err = session.Wait()
	if err != nil && err != io.EOF {
		logger.Error(err.Error(), err)
	}
}

func broadcastLine(line string, currentLine int) {
	fmt.Printf("%d - NEWLOG FROM SERVER\n", currentLine)
	handlers.Mu.Lock()
	defer handlers.Mu.Unlock()

	for client, servID := range handlers.Clients {
		if servID == serverID {
			logger.Info("Отправка новой строки лога", slog.Any("server_id", servID))
			err := client.WriteMessage(websocket.TextMessage, []byte(line))
			if err != nil {
				logger.Error(err.Error(), err)
				client.Close()
				delete(handlers.Clients, client)
			}
		}
	}
}
