package mocks

import (
	"bytes"
	"io"
	"net/url"

	"github.com/stretchr/testify/mock"
)

type ObjectStorageManagerMock struct {
	mock.Mock
}

func (m *ObjectStorageManagerMock) GetLatestGTFSVersionID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *ObjectStorageManagerMock) GetLatestURL() (*url.URL, string, error) {
	args := m.Called()
	return args.Get(0).(*url.URL), args.String(1), args.Error(2)
}

func (m *ObjectStorageManagerMock) PutSchedule(reader io.Reader, fileSize int64) (string, error) {
	args := m.Called(reader, fileSize)
	return args.String(0), args.Error(1)
}

func (m *ObjectStorageManagerMock) AppendMessage(message *bytes.Buffer) (string, error) {
	args := m.Called(message)
	return args.String(0), args.Error(1)
}

func (m *ObjectStorageManagerMock) GetLatestLog() (*bytes.Buffer, error) {
	args := m.Called()
	return args.Get(0).(*bytes.Buffer), args.Error(1)
}

func (m *ObjectStorageManagerMock) GetLatestMessageVersionID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *ObjectStorageManagerMock) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *ObjectStorageManagerMock) Close() error {
	args := m.Called()
	return args.Error(0)
}
