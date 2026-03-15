package services_test

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/stretchr/testify/assert"
	"github.com/transitIOM/projectMercury/internal/domain/models"
	"github.com/transitIOM/projectMercury/internal/domain/services"
)

// Mocks
type mockScheduleProvider struct {
	reader   io.ReadCloser
	checksum string
	err      error
}

func (m *mockScheduleProvider) GetGTFSReader(ctx context.Context) (io.ReadCloser, error) {
	return m.reader, m.err
}
func (m *mockScheduleProvider) GetChecksum(ctx context.Context) (string, error) {
	return m.checksum, m.err
}

type mockScheduleFetcher struct {
	err error
}

func (m *mockScheduleFetcher) FetchLatestSchedule(ctx context.Context) error {
	return m.err
}

type mockGatherer struct {
	positions []models.VehiclePosition
	notifyCh  chan struct{}
}

func (m *mockGatherer) GetVehiclePositions(ctx context.Context) ([]models.VehiclePosition, error) {
	return m.positions, nil
}
func (m *mockGatherer) NotifyChannel() <-chan struct{} {
	return m.notifyCh
}

type mockReportManager struct {
	err error
}

func (m *mockReportManager) CreateIssueFromReport(ctx context.Context, report models.UserReport) error {
	return m.err
}

type mockMessageRepository struct {
	messages []models.Message
	version  string
	err      error
}

func (m *mockMessageRepository) GetMessages(ctx context.Context) ([]models.Message, string, error) {
	return m.messages, m.version, m.err
}
func (m *mockMessageRepository) SaveMessage(ctx context.Context, message models.Message) (string, error) {
	return m.version, m.err
}

func TestTransitService_Subscription(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mg := &mockGatherer{notifyCh: make(chan struct{}, 1)}
	s := services.NewTransitService(ctx, nil, nil, mg, nil, nil)

	sub, err := s.Subscribe(ctx, "all")
	assert.NoError(t, err)

	testPos := []models.VehiclePosition{{VehicleID: "bus-1", Latitude: 1.0, Longitude: 2.0}}
	mg.positions = testPos
	mg.notifyCh <- struct{}{}

	select {
	case received := <-sub:
		feed, ok := received.(*gtfs.FeedMessage)
		assert.True(t, ok)
		assert.Equal(t, 1, len(feed.Entity))
		assert.Equal(t, "bus-1", *feed.Entity[0].Vehicle.Vehicle.Id)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for update")
	}
}

func TestTransitService_Methods(t *testing.T) {
	ctx := context.Background()
	sp := &mockScheduleProvider{checksum: "hash123", reader: io.NopCloser(strings.NewReader("fake"))}
	sf := &mockScheduleFetcher{}
	mg := &mockGatherer{notifyCh: make(chan struct{})}
	rm := &mockReportManager{}
	mr := &mockMessageRepository{version: "v1", messages: []models.Message{{Content: "hello"}}}

	s := services.NewTransitService(ctx, sp, sf, mg, rm, mr)

	t.Run("GetGTFSChecksum", func(t *testing.T) {
		checksum, err := s.GetGTFSChecksum(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "hash123", checksum)
	})

	t.Run("GetGTFS", func(t *testing.T) {
		reader, err := s.GetGTFS(ctx)
		assert.NoError(t, err)
		content, _ := io.ReadAll(reader)
		assert.Equal(t, "fake", string(content))
	})

	t.Run("PostReport", func(t *testing.T) {
		err := s.PostReport(ctx, models.UserReport{Title: "Issue"})
		assert.NoError(t, err)
	})

	t.Run("GetMessages", func(t *testing.T) {
		msgs, v, err := s.GetMessages(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(msgs))
		assert.Equal(t, "v1", v)
	})

	t.Run("PostMessage", func(t *testing.T) {
		v, err := s.PostMessage(ctx, models.Message{Content: "new"})
		assert.NoError(t, err)
		assert.Equal(t, "v1", v)
	})
}
