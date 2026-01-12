package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/transitIOM/projectMercury/internal/handlers"
)

type MockReportManager struct {
	mock.Mock
}

func (m *MockReportManager) CreateIssueFromReport(ctx context.Context, title string, description string, email string, tags []string) error {
	args := m.Called(ctx, title, description, email, tags)
	return args.Error(0)
}

func TestPostReport(t *testing.T) {
	mockRM := new(MockReportManager)
	mockRM.On("CreateIssueFromReport", mock.Anything, "Bug", "Crash", "test@example.com", []string{"user-report", "bug"}).Return(nil)

	formData := url.Values{}
	formData.Set("title", "Bug")
	formData.Set("description", "Crash")
	formData.Set("email", "test@example.com")
	formData.Set("category", "bug")
	req := httptest.NewRequest("POST", "/report/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler := handlers.PostReport(mockRM)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockRM.AssertExpectations(t)
}
