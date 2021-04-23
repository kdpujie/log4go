package log4go

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	"github.com/kdpujie/log4go/util"
)

// GlobalLevel global level
var GlobalLevel = DEBUG

// ConfFileWriter file writer config
type ConfFileWriter struct {
	Level       string `json:"level" mapstructure:"level"`
	PathPattern string `json:"path_pattern" mapstructure:"path_pattern"`
	Enable      bool   `json:"enable" mapstructure:"enable"`
}

// ConfConsoleWriter console writer config
type ConfConsoleWriter struct {
	Level  string `json:"level" mapstructure:"level"`
	Enable bool   `json:"enable" mapstructure:"enable"`
	Color  bool   `json:"color" mapstructure:"color"`
}

// KafKaMSGFields kafka msg fields
type KafKaMSGFields struct {
	ESIndex     string                 `json:"es_index" mapstructure:"es_index"`         // required, init field
	Level       string                 `json:"level"`                                    // dynamic, set by logger
	Code        string                 `json:"file"`                                     // dynamic, source code file:line_number
	Message     string                 `json:"message"`                                  // dynamic, message
	ServerIP    string                 `json:"server_ip" mapstructure:"server_ip"`       // required, init field, set by app
	PublicIP    string                 `json:"public_ip" mapstructure:"public_ip"`       // required, init field, set by app
	Timestamp   string                 `json:"timestamp" mapstructure:"timestamp"`       // required, dynamic, set by logger
	Now         int64                  `json:"now" mapstructure:"now"`                   // choice, unix timestamp, second
	ExtraFields map[string]interface{} `json:"extra_fields" mapstructure:"extra_fields"` // extra fields will be added
}

// ConfKafKaWriter kafka writer conf
type ConfKafKaWriter struct {
	Level          string `json:"level" mapstructure:"level"`
	Enable         bool   `json:"enable" mapstructure:"enable"`
	BufferSize     int    `json:"buffer_size" mapstructure:"buffer_size"`
	Debug          bool   `json:"debug" mapstructure:"debug"`                     // if true, will output the send msg
	SpecifyVersion bool   `json:"specify_version" mapstructure:"specify_version"` // if use the input version, default false
	Version        string `json:"version" mapstructure:"version"`                 // used to specify the kafka version, ex: 0.10.0.1 or 1.1.1

	Key string `json:"key" mapstructure:"key"` // kafka producer key, temp set, choice field

	ProducerTopic           string        `json:"producer_topic" mapstructure:"producer_topic"`
	ProducerReturnSuccesses bool          `json:"producer_return_successes" mapstructure:"producer_return_successes"`
	ProducerTimeout         time.Duration `json:"producer_timeout" mapstructure:"producer_timeout"` // ms
	Brokers                 []string      `json:"brokers" mapstructure:"brokers"`

	MSG KafKaMSGFields
}

// ConfAliLogHubWriter ali log hub writer config
type ConfAliLogHubWriter struct {
	Level           string `json:"level" mapstructure:"level"`
	Enable          bool   `json:"enable" mapstructure:"enable"`
	Topic           string `json:"topic" mapstructure:"topic"`
	Source          string `json:"source" mapstructure:"source"`
	ProjectName     string `json:"project_name" mapstructure:"project_name"`
	Endpoint        string `json:"endpoint" mapstructure:"endpoint"`
	AccessKeyId     string `json:"access_key_id" mapstructure:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret" mapstructure:"access_key_secret"`
	LogStoreName    string `json:"log_store_name" mapstructure:"log_store_name"`
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
	// global level
	GlobalLevel = getLevel(lc.Level)

	fullPath := lc.FullPath
	ShowFullPath(fullPath)

	if lc.FileWriter.Enable { // file
		if selfLevel := getLevel(lc.FileWriter.Level); selfLevel > -1 {
			Register(NewFileWriter(&lc.FileWriter))
		} else {
			Register(NewFileWriterWithLevel(GlobalLevel, &lc.FileWriter))
		}
	}

	if lc.ConsoleWriter.Enable { // Console
		if selfLevel := getLevel(lc.FileWriter.Level); selfLevel > -1 {
			Register(NewConsoleWriter(&lc.ConsoleWriter))
		} else {
			Register(NewConsoleWriterWithLevel(GlobalLevel, &lc.ConsoleWriter))
		}
	}

	if lc.AliLogHubWriter.Enable { // Ali Loghub
		if lc.AliLogHubWriter.Source == "" {
			lc.AliLogHubWriter.Source = util.GetLocalIpByTcp()
		}
		if selfLevel := getLevel(lc.AliLogHubWriter.Level); selfLevel > -1 {
			Register(NewAliLogHubWriter(&lc.AliLogHubWriter))
		} else {
			Register(NewAliLogHubWriterWithLevel(GlobalLevel, &lc.AliLogHubWriter))
		}
	}

	if lc.KafKaWriter.Enable { // kafka
		if selfLevel := getLevel(lc.KafKaWriter.Level); selfLevel > -1 {
			Register(NewKafKaWriter(&lc.KafKaWriter))
		} else {
			Register(NewKafKaWriterWithWriter(GlobalLevel, &lc.KafKaWriter))
		}
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

// 通过文本形式的日志级别，转换问数字型的日志级别
func getLevel(flag string) int {
	for i, f := range LevelFlags {
		if strings.TrimSpace(strings.ToUpper(flag)) == f {
			return i
		}
	}
	return -1
}
