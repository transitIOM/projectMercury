package tools

import (
	"bytes"
	"time"

	"github.com/simonfrey/jsonl"
	log "github.com/sirupsen/logrus"
)

var (
	CurrentMessageLog       bytes.Buffer
	CurrentMessageVersionID string
)

type MessageLog struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

func NewMessage(msg string) MessageLog {
	return MessageLog{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Message:   msg,
	}
}

func PushMessageToStorage(message string) (err error) {

	messageObj := NewMessage(message)

	b := bytes.Buffer{}

	w := jsonl.NewWriter(&b)

	err = w.Write(messageObj)
	if err != nil {
		return err
	}
	w.Close()

	size := int64(b.Len())

	// append message to s3 store
	AppendMessage(&b, size)
	log.Infof("Added message: %v", b)

	err = pullDataFromStorage()
	if err != nil {
		log.Errorf("failed to copy message log locally: %v", err)
	}
	return nil
}

func pullDataFromStorage() error {
	b, err := GetLatestMessageLog()
	if err != nil {
		return err
	}

	CurrentMessageLog = b

	v, err := GetLatestMessageLogVersionID()
	if err != nil {
		return err
	}

	CurrentMessageVersionID = v

	return nil
}

func GetLastNLines(n int) *bytes.Buffer {
	data := CurrentMessageLog.Bytes()
	lines := bytes.Split(data, []byte("\n"))

	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}

	result := new(bytes.Buffer)
	for i := start; i < len(lines); i++ {
		result.Write(lines[i])
		if i < len(lines)-1 {
			result.WriteByte('\n')
		}
	}

	if len(data) > 0 && data[len(data)-1] == '\n' {
		result.WriteByte('\n')
	}

	return result
}

func GetLatestMessageVersion() string {
	return CurrentMessageVersionID
}
