package log4go

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/kdpujie/log4go/util"
)

// GlobalLevel global level
var GlobalLevel = DEBUG

// ConfFileWriter file writer config
type ConfFileWriter struct {
	Level   string `json:"level" mapstructure:"level"`
	LogPath string `json:"log_path" mapstructure:"log_path"`
	Enable  bool   `json:"enable" mapstructure:"enable"`
}

// ConfConsoleWriter console writer config
type ConfConsoleWriter struct {
	Level  string `json:"level" mapstructure:"level"`
	Enable bool   `json:"enable" mapstructure:"enable"`
	Color  bool   `json:"color" mapstructure:"color"`
}

// ConfAliLogHubWriter ali log hub writer config
type ConfAliLogHubWriter struct {
	Level           string `json:"level" mapstructure:"level"`
	Enable          bool   `json:"enable" mapstructure:"enable"`
	LogName         string `json:"log_name" mapstructure:"log_name"`
	LogSource       string `json:"log_source" mapstructure:"log_source"`
	ProjectName     string `json:"project_name" mapstructure:"project_name"`
	Endpoint        string `json:"endpoint" mapstructure:"endpoint"`
	AccessKeyId     string `json:"access_key_id" mapstructure:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret" mapstructure:"access_key_secret"`
	StoreName       string `json:"store_name" mapstructure:"store_name"`
	BufSize         int    `json:"buf_size" mapstructure:"buf_size"`
}

// LogConfig log config
type LogConfig struct {
	Level           string              `json:"level" mapstructure:"level"`
	FullPath        bool                `json:"full_path" mapstructure:"full_path"`
	FileWriter      ConfFileWriter      `json:"file_writer" mapstructure:"file_writer"`
	ConsoleWriter   ConfConsoleWriter   `json:"console_writer" mapstructure:"console_writer"`
	AliLogHubWriter ConfAliLogHubWriter `json:"ali_log_hub_writer" mapstructure:"ali_log_hub_writer"`
	KafKaWriter     ConfKafKaWriter     `json:"kafka_writer" mapstructure:"kafka_writer"`
}

// SetupLog setup log
func SetupLog(lc LogConfig) (err error) {
	// 全局配置
	GlobalLevel = getLevel(lc.Level)

	fullPath := lc.FullPath
	ShowFullPath(fullPath)

	if lc.FileWriter.Enable {
		w := NewFileWriter()
		w.level = getLevel0(lc.FileWriter.Level, GlobalLevel)
		if err = w.SetPathPattern(lc.FileWriter.LogPath); err != nil {
			return err
		}
		Register(w)
	}

	if lc.ConsoleWriter.Enable {
		w := NewConsoleWriter()
		w.level = getLevel0(lc.ConsoleWriter.Level, GlobalLevel)
		w.SetColor(lc.ConsoleWriter.Color)
		Register(w)
	}

	if lc.AliLogHubWriter.Enable {
		w := NewAliLogHubWriter(lc.AliLogHubWriter.BufSize)
		if lc.AliLogHubWriter.LogSource == "" {
			lc.AliLogHubWriter.LogSource = util.GetLocalIpByTcp()
		}
		w.level = getLevel0(lc.AliLogHubWriter.Level, GlobalLevel)
		w.SetLog(lc.AliLogHubWriter.LogName, lc.AliLogHubWriter.LogSource)
		w.SetProject(lc.AliLogHubWriter.ProjectName, lc.AliLogHubWriter.StoreName)
		w.SetEndpoint(lc.AliLogHubWriter.Endpoint)
		w.SetAccessKey(lc.AliLogHubWriter.AccessKeyId, lc.AliLogHubWriter.AccessKeySecret)
		Register(w)
	}

	if lc.KafKaWriter.Enable {
		w := NewKafKaWriter(&lc.KafKaWriter)
		w.level = getLevel0(lc.KafKaWriter.Level, GlobalLevel)
		Register(w)
	}
	// 全局配置
	return nil
}

// SetupLogWithConf setup log with config file
func SetupLogWithConf(file string) (err error) {
	var lc LogConfig
	cnt, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	if err = json.Unmarshal(cnt, &lc); err != nil {
		return
	}
	return SetupLog(lc)
}

func getLevel(flag string) int {
	return getLevel0(flag, DEBUG)
}

// 默认为Debug模式
func getLevel0(flag string, defaultFlag int) int {
	for i, f := range LevelFlags {
		if strings.TrimSpace(strings.ToUpper(flag)) == f {
			return i
		}
	}
	fmt.Printf("[ERROR] 未找到合适的日志级别[%s]，使用默认值:%d \n", flag, defaultFlag)
	return defaultFlag
}
