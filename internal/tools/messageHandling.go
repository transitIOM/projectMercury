package tools

import (
	"bytes"
	"sync"
	"time"

	"github.com/simonfrey/jsonl"
	log "github.com/sirupsen/logrus"
)

var (
	messageMu               sync.RWMutex
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

	// append message to s3 store
	_, err = AppendMessage(&b)
	if err != nil {
		return err
	}
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

	v, err := GetLatestMessageLogVersionID()
	if err != nil {
		return err
	}

	messageMu.Lock()
	defer messageMu.Unlock()
	CurrentMessageLog = b
	CurrentMessageVersionID = v

	return nil
}

func GetLastNLines(n int) (*bytes.Buffer, error) {
	messageMu.RLock()
	if CurrentMessageLog.Len() == 0 {
		messageMu.RUnlock()
		if err := pullDataFromStorage(); err != nil {
			return nil, err
		}
		messageMu.RLock()
	}
	defer messageMu.RUnlock()

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

	return result, nil
}

func GetLatestMessageVersion() (string, error) {
	messageMu.RLock()
	if CurrentMessageVersionID == "" {
		messageMu.RUnlock()
		if err := pullDataFromStorage(); err != nil {
			return "", err
		}
		messageMu.RLock()
	}
	defer messageMu.RUnlock()

	return CurrentMessageVersionID, nil
}
