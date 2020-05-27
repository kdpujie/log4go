package log4go

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Shopify/sarama"
)

const timestampFormat = "2006-01-02T15:04:05.000+0800"

// KafKaMSGFields kafka msg fields
type KafKaMSGFields struct {
	ESIndex     string                 `json:"es_index" mapstructure:"es_index"` // required, init field
	Level       string                 // dynamic, set by logger
	Code        string                 `json:"file" mapstructure:"file"`                 // source code file:line_number
	Message     string                 `json:"message" mapstructure:"message"`           // required, dynamic
	ServerIP    string                 `json:"server_ip" mapstructure:"server_ip"`       // required, init field, set by app
	Timestamp   string                 `json:"time_stamp" mapstructure:"time_stamp"`     // required, dynamic, set by logger
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

// KafKaWriter kafka writer
type KafKaWriter struct {
	level    int
	producer sarama.SyncProducer
	messages chan *sarama.ProducerMessage
	conf     *ConfKafKaWriter

	run  bool // avoid the block with no running kafka writer
	quit chan struct{}
	stop chan struct{}
}

// NewKafKaWriter new kafka writer
func NewKafKaWriter(conf *ConfKafKaWriter) *KafKaWriter {
	defaultLevel := 0
	if conf.Level != "" {
		defaultLevel = getLevel0(conf.Level, defaultLevel)
	}

	return &KafKaWriter{
		conf:  conf,
		quit:  make(chan struct{}),
		stop:  make(chan struct{}),
		level: defaultLevel,
	}
}

// NewKafKaWriterWithWriter new kafka writer with level
func NewKafKaWriterWithWriter(conf *ConfKafKaWriter, level int) *KafKaWriter {
	defaultLevel := DEBUG
	maxLevel := len(LevelFlags)
	// maxLevel >= 1 always true
	maxLevel = maxLevel - 1

	if level >= defaultLevel && level <= maxLevel {
		defaultLevel = level
	}

	return &KafKaWriter{
		conf:  conf,
		quit:  make(chan struct{}),
		stop:  make(chan struct{}),
		level: defaultLevel,
	}
}

// Init service for Record
func (k *KafKaWriter) Init() error {
	err := k.Start()
	if err != nil {
		Error("Init err=%s \n", err.Error())
	}
	return nil
}

// Write service for Record
func (k *KafKaWriter) Write(r *Record) error {
	if r.level < k.level {
		return nil
	}

	logMsg := r.info
	if logMsg == "" {
		return nil
	}
	data := k.conf.MSG
	// timestamp, level
	data.Level = LevelFlags[r.level]
	now := time.Now()
	data.Now = now.Unix()
	data.Timestamp = now.Format(timestampFormat)
	data.Message = logMsg
	data.Code = r.code

	byteData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var structData map[string]interface{}
	err = json.Unmarshal(byteData, &structData)
	if err != nil {
		return err
	}

	delete(structData, "extraFields")

	// not exist new fields will be added
	for k, v := range data.ExtraFields {
		if _, ok := structData[k]; !ok {
			structData[k] = v
		}
	}

	jsonStructDataByte, err := json.Marshal(structData)
	if err != nil {
		return err
	}

	jsonData := string(jsonStructDataByte)

	key := ""
	if k.conf.Key != "" {
		key = k.conf.Key
	}

	msg := &sarama.ProducerMessage{
		Topic: k.conf.ProducerTopic,
		// autofill or use specify timestamp, you must set Version >= sarama.V0_10_0_1
		// Timestamp: time.Now(),
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(jsonData),
	}

	if k.conf.Debug {
		log.Printf("kafka-writer msg [topic: %v, timestamp: %v, brokers: %v]\nkey:   %v\nvalue: %v\n", msg.Topic,
			msg.Timestamp, k.conf.Brokers, key, jsonData)
	}
	go k.asyncWriteMessages(msg)

	return nil
}

func (k *KafKaWriter) asyncWriteMessages(msg *sarama.ProducerMessage) {
	if msg != nil {
		k.messages <- msg
	}
}

// send kafka message to kafka
func (k *KafKaWriter) daemonProducer() {
	k.run = true

next:
	for {
		select {
		case mes, ok := <-k.messages:
			if ok {
				partition, offset, err := k.producer.SendMessage(mes)

				if err != nil {
					log.Printf("SendMessage(topic=%s, partition=%v, offset=%v, key=%s, value=%s,timstamp=%v) err=%s\n\n", mes.Topic,
						partition, offset, mes.Key, mes.Value, mes.Timestamp, err.Error())
					continue
				} else {
					if k.conf.Debug {
						log.Printf("SendMessage(topic=%s, partition=%v, offset=%v, key=%s, value=%s,timstamp=%v)\n\n", mes.Topic,
							partition, offset, mes.Key, mes.Value, mes.Timestamp)
					}
				}
			}
		case <-k.stop:
			break next
		}
	}
	k.quit <- struct{}{}
}

// Start start the kafka writer
func (k *KafKaWriter) Start() (err error) {
	log.Println("start kafka writer ...")
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = k.conf.ProducerReturnSuccesses
	cfg.Producer.Timeout = k.conf.ProducerTimeout

	// if want set timestamp for data should set version
	versionStr := k.conf.Version
	kafkaVer := sarama.V0_10_0_1

	if k.conf.SpecifyVersion {
		if versionStr != "" {
			if kafkaVersion, err := sarama.ParseKafkaVersion(versionStr); err == nil {
				// should be careful set the version, maybe occur EOF error
				kafkaVer = kafkaVersion
			}
		}
	}
	// if not specify the version, use the sarama.V0_10_0_1 to guarante the timestamp can be control
	cfg.Version = kafkaVer

	// NewHashPartitioner returns a Partitioner which behaves as follows. If the message's key is nil then a
	// random partition is chosen. Otherwise the FNV-1a hash of the encoded bytes of the message key is used,
	// modulus the number of partitions. This ensures that messages with the same key always end up on the
	// same partition.
	// cfg.Producer.Partitioner = sarama.NewHashPartitioner
	// cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	// cfg.Producer.Partitioner = sarama.NewReferenceHashPartitioner

	k.producer, err = sarama.NewSyncProducer(k.conf.Brokers, cfg)
	if err != nil {
		log.Printf("sarama.NewSyncProducer err, message=%s \n", err)
		return err
	}
	size := k.conf.BufferSize
	if size <= 1 {
		size = 1
	}
	k.messages = make(chan *sarama.ProducerMessage, size)

	go k.daemonProducer()
	log.Println("start kafka writer ok")
	return err
}

// Stop stop the kafka writer
func (k *KafKaWriter) Stop() {
	if k.run {
		close(k.messages)
		<-k.stop
		if err := k.producer.Close(); err != nil {
			log.Printf("stop kafka writer failed:%v\n", err)
		}
	}
}
