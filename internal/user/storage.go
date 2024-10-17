package user

import (
	"errors"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

const path = "user.yml"

var u user

type user struct {
	Login    string `yaml:"login"`
	Password string `yaml:"password"`
}

func CreateUserConfigFile() error {
	var username string
	fmt.Print("input username: ")
	if _, err := fmt.Scanf("%s", &username); err != nil {
		return err
	}

	var password string
	fmt.Print("input password: ")
	if _, err := fmt.Scanf("%s", &password); err != nil {
		return err
	}

	pass, err := GenerateHashForPass(password)
	if err != nil {
		return err
	}

	u = user{Login: username, Password: pass}

	data, err := yaml.Marshal(u)
	if err != nil {
		return nil
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Println("[WARN] failed to close file:", err)
		}
	}()

	if _, err = file.Write(data); err != nil {
		return nil
	}

	return nil
}

func LoadUser() error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("user data not found.\nto create a user use the --create-user flag")
		}
		return err
	}

	if err = yaml.Unmarshal(data, &u); err != nil {
		return err
	}

	return nil
}

// GetUserData возвращает логин и пароль пользователя
func GetUserData() (string, string) {
	return u.Login, u.Password
}
