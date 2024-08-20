package main

import (
	"bufio"
	"fmt"
	"github.com/algrvvv/monlog/internal/app"
	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/mdobak/go-xerrors"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"log/slog"
	"strconv"
)

func main() {
	err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatal(err)
	}

	err = logger.NewLogger("monlog.log")
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range config.Cfg.Servers {
		newSession(s)
	}
}

func newSession(s config.ServerConfig) {
	clientConfig, err := app.NewSSHConfig(config.Cfg.App.PathToIDRSA, s.User)
	if err != nil {
		logger.Logger.Error(err.Error(), slog.Any("error", xerrors.New(err)))
	}

	logger.Logger.Info("Connection to server...")
	var client *ssh.Client
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	client, err = ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		logger.Logger.Error(err.Error(), slog.Any("error", xerrors.New(err)))
	}
	defer client.Close()

	logger.Logger.Info("Connected to server", slog.Any("server", addr))

	session, err := client.NewSession()
	if err != nil {
		logger.Logger.Error(err.Error(), slog.Any("error", xerrors.New(err)))
	}
	defer session.Close()
	logger.Logger.Info("Created new session to server", slog.Any("server", addr))

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
		logger.Logger.Error(err.Error(), slog.Any("error", xerrors.New(err)))
	}

	go func() {
		scanner := bufio.NewScanner(reader)
		currentLine, _ := strconv.Atoi(startLine)
		for scanner.Scan() {
			line := scanner.Text()
			processLogLine(line, currentLine)
			currentLine++
		}

		if err = scanner.Err(); err != nil && err != io.EOF {
			log.Fatalf("Error reading output: %v", err)
		}
	}()

	err = session.Wait()
	if err != nil && err != io.EOF {
		logger.Logger.Error(err.Error(), slog.Any("error", xerrors.New(err)))
	}
}

func processLogLine(line string, currentLine int) {
	fmt.Printf("COPY FROM SERVER: [%d] - %s\n", currentLine, line)
}
