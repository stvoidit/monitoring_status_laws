package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
)

// Config - ...
type Config struct {
	Debug              bool    `json:"debug"`
	DepartmentsCanEdit []int64 `json:"admin_departments"`
	DB                 struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Dbname   string `json:"dbname"`
		Login    string `json:"login"`
		Password string `json:"password"`
	} `json:"db"`
	Srv struct {
		Domain  string `json:"domain"`
		Address string `json:"address"`
		SSL     bool   `json:"ssl"`
	} `json:"srv"`
	Megaplan struct {
		Domain             string `json:"domain"`
		UUID               string `json:"uuid"`
		Token              string `json:"token"`
		InsecureSkipVerify bool   `json:"insecure"`
		ProviderHost       string `toml:"provider_api"`
	} `json:"megaplan"`
	TGBOT struct {
		Domain   string `json:"domain"`
		Name     string `json:"name"`
		TokenAPI string `json:"token"`
	} `json:"tgbot"`
}

func (cnf Config) String() string {
	buf, err := json.MarshalIndent(cnf, "", "  ")
	checkErrUtil("cnf.String", err)
	return string(buf)
}

// LoadConfig - ...
func LoadConfig(file io.Reader) (cnf *Config, err error) {
	cnf = new(Config)
	err = json.NewDecoder(file).Decode(cnf)
	if cnf.Srv.Address == "" {
		cnf.Srv.Address = "0.0.0.0:8080"
	}
	if value, ok := os.LookupEnv("DEBUG"); ok && value == "1" {
		cnf.Debug = true
	}
	if value, ok := os.LookupEnv("TG_BOT_TOKEN"); ok {
		cnf.TGBOT.TokenAPI = value
	}
	if value, ok := os.LookupEnv("SERVER_DOMAIN"); ok {
		cnf.Srv.Domain = value
	}
	if prefix, ok := os.LookupEnv("DB_PREFIX"); ok {
		cnf.DB.Dbname = fmt.Sprintf("%s.%s", prefix, cnf.DB.Dbname)
	}
	return
}

// LoadConfigFromFile - ...
func LoadConfigFromFile(filename string) (cnf *Config, err error) {
	file, err := os.Open(filename)
	if err != nil {
		createDefaultConfig()
		return nil, fmt.Errorf("файл конфигурации не найден: %w, сгенерирован шаблон файла конфигурации \"config_example.json\"", err)
	}
	defer func() {
		if e := file.Close(); e != nil {
			err = e
		}
	}()
	return LoadConfig(file)
}

func createDefaultConfig() {
	w, err := os.Create("config_example.json")
	if err != nil {
		panic(err)
	}
	defer checkErrUtil("createDefaultConfig.Close", w.Close())
	var cnf = new(Config)
	cnf.Srv.Address = "0.0.0.0:8080"
	e := json.NewEncoder(w)
	e.SetIndent("", "\t")
	checkErrUtil("createDefaultConfig.Encode", e.Encode(cnf))
}

// ContainsInDepartment - переданный int64 содержится в списке []int64
func (cnf Config) ContainsInDepartment(depID int64) bool {
	if cnf.DepartmentsCanEdit == nil {
		return false
	}
	for i := 0; i < len(cnf.DepartmentsCanEdit); i++ {
		if cnf.DepartmentsCanEdit[i] == depID {
			return true
		}
	}
	return false
}

func (cnf Config) ConnStringDB() string {
	var connUrl = url.URL{
		Scheme: "postgresql",
		Host:   net.JoinHostPort(cnf.DB.Host, cnf.DB.Port),
		Path:   cnf.DB.Dbname,
		User:   url.UserPassword(cnf.DB.Login, cnf.DB.Password),
		RawQuery: (url.Values{
			"search_path": []string{"public,monitoring_draft_laws"},
		}).Encode(),
	}
	return connUrl.String()
}

func (cnf Config) ServiceAddr() string {
	var schema = "http"
	if cnf.Srv.SSL {
		schema = "https"
	}
	var addr = url.URL{
		Scheme: schema,
		Host:   cnf.Srv.Domain,
	}
	return addr.String()
}

func checkErrUtil(label string, err error) {
	if err != nil {
		slog.Warn(label, slog.String("error", err.Error()))
	}
}
