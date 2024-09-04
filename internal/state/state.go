package state

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"gitlab.com/metakeule/fmtdate"
	"gopkg.in/yaml.v3"

	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
)

type ServerData struct {
	ID             int    `yaml:"id"`
	LastNotifyTime string `yaml:"last_notify_time"`
}

type ServersState struct {
	Servers []ServerData `yaml:"servers"`
}

type ServerState struct {
	StateFile *os.File
	FileMutex sync.Mutex
	Srvs      ServersState
}

var ServState = &ServerState{}

func InitializeState() error {
	var err error
	ServState.StateFile, err = os.OpenFile("state.yml", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errors.New("Failed to open state.yml" + err.Error())
	}

	configHash := ServState.getConfigHash()
	size, err := ServState.StateFile.Stat()
	if err != nil {
		return errors.New("Failed to stat state.yml" + err.Error())
	}
	if size.Size() == 0 {
		// создаем новый файл
		ServState.generateNewStates()
		var marshalled []byte
		marshalled, err = yaml.Marshal(&ServState.Srvs)
		if err != nil {
			return errors.New("Failed to marshal state.yml" + err.Error())
		}
		comment := getComment(configHash)
		finalData := append([]byte(comment), marshalled...)
		_, err = ServState.StateFile.Write(finalData)
		return err
	}

	currentHast, err := ServState.getCurrentHashFromState()
	if err != nil {
		return err
	}

	_, err = ServState.StateFile.Seek(0, 0)
	if err != nil {
		return errors.New("Failed to seek state.yml" + err.Error())
	}
	var lines []string
	scanner := bufio.NewScanner(ServState.StateFile)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		return errors.New("Failed to scan state.yml at updating" + err.Error())
	}

	if err = yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &ServState.Srvs); err != nil {
		return errors.New("Failed to unmarshal state.yml" + err.Error())
	}

	if currentHast != configHash {
		var tmp *os.File
		tmp, err = os.CreateTemp("", "temp_state_*.yml")
		if err != nil {
			return errors.New("Failed to create temporary state.yml" + err.Error())
		}

		var statesMap = make(map[int]struct {
			status         bool
			lastTimeNotify string
		})
		for _, ss := range ServState.Srvs.Servers {
			statesMap[ss.ID] = struct {
				status         bool
				lastTimeNotify string
			}{
				status:         false,
				lastTimeNotify: ss.LastNotifyTime,
			}
		}

		var (
			newStates  []ServerData
			marshalled []byte
		)
		for _, cs := range config.Cfg.Servers {
			if state, ok := statesMap[cs.ID]; ok {
				state.status = true
				statesMap[cs.ID] = state
				continue
			}
			newStates = append(newStates, ServerData{
				ID:             cs.ID,
				LastNotifyTime: fmtdate.Format(cs.LogTimeFormat, time.Now()),
			})
		}

		for id, state := range statesMap {
			if state.status == true {
				newStates = append(newStates, ServerData{
					ID:             id,
					LastNotifyTime: state.lastTimeNotify,
				})
			}
		}

		comment := getComment(configHash)
		ServState.Srvs = ServersState{
			Servers: newStates,
		}
		marshalled, err = yaml.Marshal(&ServState.Srvs)
		if err != nil {
			return errors.New("Failed to marshal state.yml at final update" + err.Error())
		}
		finalData := append([]byte(comment), marshalled...)
		if _, err = ServState.StateFile.Seek(0, 0); err != nil {
			return errors.New("Failed to seek state.yml at state" + err.Error())
		}
		_, _ = tmp.Write(finalData)

		if err = ServState.StateFile.Close(); err != nil {
			return errors.New("Failed to close state.yml" + err.Error())
		}
		if err = os.Remove("state.yml"); err != nil {
			return errors.New("Failed to remove state.yml" + err.Error())
		}

		if err = os.Rename(tmp.Name(), "state.yml"); err != nil {
			return errors.New("Failed to rename state.yml" + err.Error())
		}
		ServState.StateFile, err = os.OpenFile("state.yml", os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return errors.New("Failed to reopen state.yml" + err.Error())
		}

		logger.Info("state successfully updated")
		return nil
	}

	logger.Info("State successfully initialized")
	return nil
}

func (s *ServerState) getConfigHash() string {
	hasher := sha256.New()
	configServerData, _ := json.Marshal(config.Cfg.Servers)
	hasher.Write(configServerData)
	hash := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	re := regexp.MustCompile(`[[:punct:]]|[[:space:]]`)
	return re.ReplaceAllString(hash, "")
}

func (s *ServerState) getCurrentHashFromState() (string, error) {
	scanner := bufio.NewScanner(s.StateFile)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", errors.New("scanner error happened: " + err.Error())
	}

	secondLineWithHash := lines[2]
	secondLineWithHash = strings.ReplaceAll(secondLineWithHash, "#", "")
	return strings.TrimSpace(secondLineWithHash), nil
}

func (s *ServerState) generateNewStates() {
	var serverStates []ServerData
	for _, server := range config.Cfg.Servers {
		serverStates = append(serverStates, ServerData{
			ID:             server.ID,
			LastNotifyTime: fmtdate.Format(server.LogTimeFormat, time.Now()),
		})
	}
	s.Srvs = ServersState{
		Servers: serverStates,
	}
}

func GetLastNotifyTimeById(id int) string {
	for _, server := range ServState.Srvs.Servers {
		if server.ID == id {
			return server.LastNotifyTime
		}
	}
	return ""
}

func CompareLastNotifyTime(id int, newTimeStr string) bool {
	lastTime := GetLastNotifyTimeById(id)
	if lastTime == "" {
		logger.Error("Got empty last time", nil)
		return false
	}
	var serv config.ServerConfig
	for _, server := range config.Cfg.Servers {
		if id == server.ID {
			serv = server
		}
	}
	oldTime, err := fmtdate.Parse(serv.LogTimeFormat, lastTime)
	if err != nil {
		logger.Error("Failed to parse last time from server "+serv.LogTimeFormat+" "+err.Error(), err)
		return false
	}
	newTime, err := fmtdate.Parse(serv.LogTimeFormat, newTimeStr)
	if err != nil {
		logger.Error("Failed to parse new time from server "+serv.LogTimeFormat+" "+err.Error(), err)
		return false
	}

	return oldTime.Before(newTime)
}

func UpdateLastNotifyTime(id int, newTimeStr string) error {
	ServState.FileMutex.Lock()
	defer ServState.FileMutex.Unlock()

	tmp, err := os.CreateTemp("", "temp_state_*.yml")
	if err != nil {
		return errors.New("Failed to create temporary state.yml" + err.Error())
	}

	for i, state := range ServState.Srvs.Servers {
		if id == state.ID {
			ServState.Srvs.Servers[i].LastNotifyTime = newTimeStr
		}
	}

	configHash := ServState.getConfigHash()
	marshalledData, err := yaml.Marshal(&ServState.Srvs)
	if err != nil {
		return errors.New("Failed to marshal state.yml" + err.Error())
	}
	comment := getComment(configHash)
	finalData := append([]byte(comment), marshalledData...)
	_, _ = tmp.Write(finalData)

	if err = ServState.StateFile.Close(); err != nil {
		return errors.New("Failed to close state.yml" + err.Error())
	}
	if err = os.Remove("state.yml"); err != nil {
		return errors.New("Failed to remove state.yml" + err.Error())
	}
	if err = os.Rename(tmp.Name(), "state.yml"); err != nil {
		return errors.New("Failed to rename state.yml" + err.Error())
	}

	ServState.StateFile, err = os.OpenFile("state.yml", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errors.New("Failed to reopen state.yml" + err.Error())
	}

	logger.Info("state successfully updated")
	return nil
}

func getComment(configHash string) string {
	return fmt.Sprintf("# do not change this file. it is used to save information about recent notifications\n# also don't touch the third line as it plays a very important role :D\n# %s\n\n", configHash)
}
