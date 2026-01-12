package handlers_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/transitIOM/projectMercury/internal/handlers"
	"github.com/transitIOM/projectMercury/test/mocks"
)

func TestPutGTFSSchedule(t *testing.T) {
	t.Run("valid zip file", func(t *testing.T) {
		mockSM := new(mocks.ObjectStorageManagerMock)
		versionID := "test-version-123"
		mockSM.On("PutSchedule", mock.Anything, mock.AnythingOfType("int64")).Return(versionID, nil)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="GTFSSchedule"; filename="schedule.zip"`)
		h.Set("Content-Type", "application/zip")
		part, _ := writer.CreatePart(h)
		part.Write([]byte("fake-zip-content"))
		writer.Close()

		req := httptest.NewRequest("PUT", "/schedule/", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		handler := handlers.PutGTFSSchedule(mockSM)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusAccepted, rr.Code)
		assert.Contains(t, rr.Body.String(), versionID)
		mockSM.AssertExpectations(t)
	})

	t.Run("invalid file type", func(t *testing.T) {
		mockSM := new(mocks.ObjectStorageManagerMock)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("GTFSSchedule", "schedule.txt")
		part.Write([]byte("fake-text-content"))
		writer.Close()

		req := httptest.NewRequest("PUT", "/schedule/", body)
		// Manually setting content type for the part in the handler logic is what matters
		// However, r.FormFile doesn't expose the part's content type easily unless we set it in the header.
		// Handlers use fileHeader.Header.Get("Content-Type")

		// Let's refine the test to set the content type of the part
		body = &bytes.Buffer{}
		writer = multipart.NewWriter(body)
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="GTFSSchedule"; filename="schedule.txt"`)
		h.Set("Content-Type", "text/plain")
		part, _ = writer.CreatePart(h)
		part.Write([]byte("fake-text-content"))
		writer.Close()

		req = httptest.NewRequest("PUT", "/schedule/", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		handler := handlers.PutGTFSSchedule(mockSM)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "unsupported file type")
	})

	t.Run("missing form file", func(t *testing.T) {
		mockSM := new(mocks.ObjectStorageManagerMock)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest("PUT", "/schedule/", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		handler := handlers.PutGTFSSchedule(mockSM)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
