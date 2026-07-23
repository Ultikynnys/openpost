package publisher

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/services/crypto"
	"github.com/openpost/backend/internal/services/lifecycle"
	"github.com/openpost/backend/internal/services/tokenmanager"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

func TestHandlePublishPublicationJobRecordsLifecycleEvents(t *testing.T) {
	t.Parallel()

	srv := newPublisherLifecycleTestServer(t, &fakePublisherAdapter{externalID: "external-1"})

	require.NoError(t, srv.publishPublication(t))

	events := srv.lifecycleEvents(t)
	requireLifecycleTypes(t, events, lifecycle.EventProviderProcessing, lifecycle.EventPublished)
}

func TestHandlePublishPublicationJobRecordsRetryAndFailureEvents(t *testing.T) {
	t.Parallel()

	srv := newPublisherLifecycleTestServer(t, &fakePublisherAdapter{publishErr: errors.New("provider rejected post")})
	_, err := srv.db.NewUpdate().Model((*models.Rendition)(nil)).
		Set("status = ?", models.RenditionStatusFailed).
		Where("id = ?", "rendition-1").
		Exec(context.Background())
	require.NoError(t, err)

	err = srv.publishPublication(t)

	require.Error(t, err)
	events := srv.lifecycleEvents(t)
	requireLifecycleTypes(t, events, lifecycle.EventRetried, lifecycle.EventProviderProcessing, lifecycle.EventFailed)
	require.Contains(t, events[len(events)-1].MetadataJSON, "provider rejected post")
}

func TestSegmentedRenditionRetryResumesWithoutDuplicatingPublishedPrefix(t *testing.T) {
	t.Parallel()

	adapter := &fakePublisherAdapter{
		publishErrors: []error{nil, errors.New("second segment failed"), nil},
		externalIDs:   []string{"external-root", "", "external-reply"},
	}
	srv := newPublisherLifecycleTestServer(t, adapter)
	ctx := context.Background()
	segments := []models.PublicationSegment{
		{ID: "segment-1", PublicationID: "publication-1", Position: 0, Body: "Root", SettingsJSON: "{}"},
		{ID: "segment-2", PublicationID: "publication-1", Position: 1, Body: "Reply", SettingsJSON: "{}"},
	}
	_, err := srv.db.NewInsert().Model(&segments).Exec(ctx)
	require.NoError(t, err)
	renditionSegments := []models.RenditionSegment{
		{ID: "rendition-segment-1", RenditionID: "rendition-1", PublicationSegmentID: "segment-1", Position: 0, Body: "Root", SettingsJSON: "{}", Status: models.RenditionStatusReady},
		{ID: "rendition-segment-2", RenditionID: "rendition-1", PublicationSegmentID: "segment-2", Position: 1, Body: "Reply", SettingsJSON: "{}", Status: models.RenditionStatusReady},
	}
	_, err = srv.db.NewInsert().Model(&renditionSegments).Exec(ctx)
	require.NoError(t, err)

	require.ErrorContains(t, srv.publishPublication(t), "second segment failed")
	var first models.RenditionSegment
	require.NoError(t, srv.db.NewSelect().Model(&first).Where("id = ?", "rendition-segment-1").Scan(ctx))
	require.Equal(t, models.RenditionStatusPublished, first.Status)
	require.Equal(t, "external-root", first.ExternalID)

	require.NoError(t, srv.publishPublication(t))
	require.Equal(t, 3, adapter.publishCalls)
	require.Len(t, adapter.publishRequests, 3)
	require.Equal(t, "", adapter.publishRequests[0].ReplyToID)
	require.Equal(t, "external-root", adapter.publishRequests[1].ReplyToID)
	require.Equal(t, "external-root", adapter.publishRequests[2].ReplyToID)
	require.Equal(t, "Root", adapter.publishRequests[0].Content)
	require.Equal(t, "Reply", adapter.publishRequests[1].Content)
	require.Equal(t, "Reply", adapter.publishRequests[2].Content)

	var second models.RenditionSegment
	require.NoError(t, srv.db.NewSelect().Model(&second).Where("id = ?", "rendition-segment-2").Scan(ctx))
	require.Equal(t, models.RenditionStatusPublished, second.Status)
	require.Equal(t, "external-reply", second.ExternalID)
}

type publisherLifecycleTestServer struct {
	db      *bun.DB
	service *Service
}

func newPublisherLifecycleTestServer(t *testing.T, adapter *fakePublisherAdapter) *publisherLifecycleTestServer {
	t.Helper()

	sqldb, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString()))
	require.NoError(t, err)
	sqldb.SetMaxOpenConns(1)

	db := bun.NewDB(sqldb, sqlitedialect.New())
	for _, model := range []interface{}{
		(*models.Workspace)(nil),
		(*models.SocialAccount)(nil),
		(*models.Publication)(nil),
		(*models.Rendition)(nil),
		(*models.PublicationSegment)(nil),
		(*models.RenditionSegment)(nil),
		(*models.RenditionSegmentMedia)(nil),
		(*models.RenditionMedia)(nil),
		(*models.MediaAttachment)(nil),
		(*models.PublicationLifecycleEvent)(nil),
		(*models.UsageCounter)(nil),
	} {
		_, err = db.NewCreateTable().Model(model).IfNotExists().Exec(context.Background())
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	encryptor := crypto.NewTokenEncryptor("test-secret-key")
	encAccess, err := encryptor.Encrypt("access-token")
	require.NoError(t, err)

	ctx := context.Background()
	_, err = db.NewInsert().Model(&models.Workspace{ID: "ws-1", Name: "Launch"}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.SocialAccount{
		ID:             "account-1",
		WorkspaceID:    "ws-1",
		Platform:       "x",
		AccountID:      "x-account",
		Slug:           "x-account",
		AccessTokenEnc: encAccess,
		IsActive:       true,
		CreatedAt:      time.Now().UTC(),
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.Publication{
		ID:             "publication-1",
		WorkspaceID:    "ws-1",
		CreatedByID:    "user-1",
		Title:          "Launch",
		ContentProfile: models.ContentProfileShortText,
		SourceText:     "Launch update",
		SourceContent:  "Launch update",
		Status:         models.PublicationStatusReady,
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.Rendition{
		ID:              "rendition-1",
		PublicationID:   "publication-1",
		SocialAccountID: "account-1",
		Platform:        "x",
		Profile:         models.ContentProfileShortText,
		Body:            "Launch update",
		Status:          models.RenditionStatusReady,
	}).Exec(ctx)
	require.NoError(t, err)

	manager := tokenmanager.NewTokenManager(db, encryptor)
	manager.SetProvider("x", adapter)
	service := NewService(db, manager)
	service.SetProvider("x", adapter)

	return &publisherLifecycleTestServer{db: db, service: service}
}

func (s *publisherLifecycleTestServer) publishPublication(t *testing.T) error {
	t.Helper()
	payload, err := json.Marshal(map[string]string{"publication_id": "publication-1"})
	require.NoError(t, err)
	return s.service.HandlePublishPublicationJob(context.Background(), string(payload))
}

func (s *publisherLifecycleTestServer) lifecycleEvents(t *testing.T) []models.PublicationLifecycleEvent {
	t.Helper()
	var events []models.PublicationLifecycleEvent
	require.NoError(t, s.db.NewSelect().Model(&events).Order("created_at ASC").Scan(context.Background()))
	return events
}

func requireLifecycleTypes(t *testing.T, events []models.PublicationLifecycleEvent, expected ...string) {
	t.Helper()
	require.Len(t, events, len(expected))
	for i, eventType := range expected {
		require.Equal(t, eventType, events[i].Type)
	}
}
