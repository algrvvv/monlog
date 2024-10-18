package app

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/algrvvv/monlog/internal/logger"
)

type LogFile struct {
	*os.File
	mu sync.Mutex
}

func getHashedFilename(filename string) string {
	hasher := sha1.New()
	hasher.Write([]byte(filename))
	fullName := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	re := regexp.MustCompile(`[[:punct:]]|[[:space:]]`)
	return re.ReplaceAllString(fullName, "")
}

func NewLogFile(filename string, enabled bool) (*LogFile, error) {
	if !enabled {
		// nolint
		return nil, nil
	}

	hashedFilename := getHashedFilename(filename)
	path := fmt.Sprintf("logs/%s.log", hashedFilename)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, errors.New("Failed to create log file: " + err.Error())
	}
	return &LogFile{File: file}, nil
}

// GetLineCount метод, который использует команду `wc -l file` для получения колва строк файла.
// В будущем планируется добавить поддержку Windows
func (lf *LogFile) GetLineCount() (int, error) {
	cmd := exec.Command("wc", "-l", lf.Name())
	output, err := cmd.Output()
	if err != nil {
		return -1, err
	}
	fields := strings.Fields(string(output))
	return strconv.Atoi(fields[0])
}

// PushLineWithLimit метод для записи в конец строки.
// Если размер уже имеющегося файла превышает лимит, то мы его обраезам
func (lf *LogFile) PushLineWithLimit(line string, limitMB int) error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	maxSize := int64(limitMB * 1024 * 1024)
	filename := lf.Name()

	stat, err := lf.File.Stat()
	if err != nil {
		return err
	}

	if stat.Size()+int64(len(line)) > maxSize {
		var temp *os.File
		temp, err = os.CreateTemp("", lf.Name())
		if err != nil {
			return err
		}

		_, err = lf.Seek(maxSize, io.SeekEnd)
		if err != nil {
			return err
		}
		if err = os.Remove(filename); err != nil {
			return errors.New("Failed to remove old log file: " + err.Error())
		}
		if err = os.Rename(temp.Name(), filename); err != nil {
			return errors.New("Failed to rename temp log file: " + err.Error())
		}
		lf.File, err = os.OpenFile(filename, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
	}

	_, err = lf.File.WriteString(line + "\n")
	if err != nil {
		return err
	}

	return nil
}

// ReadFullFile метод для чтения всего файла кусками.
// Вторым параметром передается колбек для большей мобильности метода.
// К примеру, получения кусочка данных и отправка их по вебсокетам и тд
func (lf *LogFile) ReadFullFile(targetLine int, callback func([]byte) []string) {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	pos, err := lf.getCursorPositionAndGoToStart()
	if err != nil {
		logger.Error("Error getting cursor position: "+err.Error(), err)
		return
	}

	var (
		n       int
		lineNum int
		reader  = bufio.NewReader(lf.File)
		buffer  = make([]byte, 1024)
	)

	for {
		n, err = reader.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			logger.Error("error reading file: "+err.Error(), err)
			return
		}

		for i := 0; i < n; i++ {
			if buffer[i] == '\n' {
				lineNum++
			}

			if lineNum >= targetLine {
				err = lf.setCursorPosition(pos)
				if err != nil {
					logger.Error("Error setting cursor position: "+err.Error(), err)
				}
				return
			}
		}
		callback(buffer[:n])
	}

	err = lf.setCursorPosition(pos)
	if err != nil {
		logger.Error("Error setting cursor position: "+err.Error(), err)
	}
}

// ReadLines метод для чтения куска файла, начиная с `startLine` заканчивая `endLine`
func (lf *LogFile) ReadLines(startLine, endLine int) []string {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	pos, err := lf.getCursorPositionAndGoToStart()
	if err != nil {
		logger.Error("Error getting cursor position: "+err.Error(), err)
		return []string{}
	}

	if startLine < 0 {
		startLine = 0
	}

	logger.Info(fmt.Sprintf("Reading lines from file: %s; %d/%d", lf.Name(), startLine, endLine))
	scanner := bufio.NewScanner(lf.File)
	var lines []string
	line := 0

	for scanner.Scan() {
		line++
		if line >= startLine && line <= endLine {
			lines = append(lines, scanner.Text())
		}
		if line > endLine {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Error("error reading log file: "+err.Error(), err)
	}

	err = lf.setCursorPosition(pos)
	if err != nil {
		logger.Error("Error setting cursor position: "+err.Error(), err)
	}

	return lines
}

func (lf *LogFile) CLoseAndRemove() error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if err := lf.Close(); err != nil {
		return err
	}
	if err := os.Remove(lf.Name()); err != nil {
		return err
	}
	return nil
}

func (lf *LogFile) getCursorPositionAndGoToStart() (int64, error) {
	currentPos, err := lf.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	_, err = lf.Seek(0, 0)
	if err != nil {
		return 0, err
	}

	return currentPos, nil
}

func (lf *LogFile) setCursorPosition(pos int64) error {
	_, err := lf.Seek(pos, io.SeekStart)
	return err
}
