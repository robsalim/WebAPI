package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"runtime"
)

type Config struct {
	Database DatabaseConfig `json:"database"`
	IServer  IServerConfig  `json:"iserver"`
	Delays   DelaysConfig   `json:"delays"`
	NTP      NTPConfig      `json:"ntp"`
	IsAdmin  bool
}

type DatabaseConfig struct {
	Base string `json:"base"`
	// Остальные параметры БД остаются в коде
	DataSource string
	User       string
	Password   string
}

type IServerConfig struct {
	Path string `json:"path"`
}

type DelaysConfig struct {
    DeltaHour float64 `json:"delta_hour"`  // заменил int на float64
}

type NTPConfig struct {
	Server string `json:"server"`
}

func LoadConfig() *Config {
	// Значения по умолчанию
	config := &Config{
		Database: DatabaseConfig{
			Base:       "BeeDotNet",
			DataSource: "tcp:10.96.30.62,10060",
			User:       "**",
			Password:   "**",
		},
		IServer: IServerConfig{
			Path: "C:\\BeeDotNet\\IServer\\IServer.exe",
		},
		Delays: DelaysConfig{
			DeltaHour: 0.01,
		},
		NTP: NTPConfig{
			Server: "pool.ntp.org",
		},
	}

	// Читаем файл конфигурации
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Printf("⚠️  config.json не найден, использую значения по умолчанию")
	} else {
		var fileConfig map[string]interface{}
		err = json.Unmarshal(configFile, &fileConfig)
		if err == nil {
			// Перезаписываем database.base
			if db, ok := fileConfig["database"].(map[string]interface{}); ok {
				if base, ok := db["base"].(string); ok && base != "" {
					config.Database.Base = base
					log.Printf("✅ Загружено database.base: %s", base)
				}
			}
			
			// Перезаписываем iserver.path
			if iserver, ok := fileConfig["iserver"].(map[string]interface{}); ok {
				if path, ok := iserver["path"].(string); ok && path != "" {
					config.IServer.Path = path
					log.Printf("✅ Загружено iserver.path: %s", path)
				}
			}
			
			// Перезаписываем delays.delta_hour
			if delays, ok := fileConfig["delays"].(map[string]interface{}); ok {
				if deltaHour, ok := delays["delta_hour"].(float64); ok && deltaHour > 0 {
					config.Delays.DeltaHour = deltaHour  // теперь float64
					log.Printf("✅ Загружено delays.delta_hour: %f", deltaHour)
				}
			}
			
			
			
			// Перезаписываем ntp.server
			if ntp, ok := fileConfig["ntp"].(map[string]interface{}); ok {
				if server, ok := ntp["server"].(string); ok && server != "" {
					config.NTP.Server = server
					log.Printf("✅ Загружено ntp.server: %s", server)
				}
			}
		}
	}

	// Проверка прав администратора
	config.IsAdmin = checkAdminRights()

	return config
}

func checkAdminRights() bool {
	if runtime.GOOS == "windows" {
		_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
		return err == nil
	}
	return os.Geteuid() == 0
}
