package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/models"
	"github.com/stretchr/testify/require"
)

func TestUpdateScheduledPublicationReschedulesPublishJob(t *testing.T) {
	db := createHandlerTestDB(t,
		(*models.WorkspaceMember)(nil),
		(*models.Publication)(nil),
		(*models.Rendition)(nil),
		(*models.Job)(nil),
	)
	ctx := context.Background()
	now := time.Date(2026, time.July, 1, 9, 0, 0, 0, time.UTC)
	oldRunAt := time.Date(2026, time.July, 6, 12, 0, 0, 0, time.UTC)
	newRunAt := time.Date(2026, time.July, 8, 13, 30, 0, 0, time.UTC)

	_, err := db.NewInsert().Model(&models.WorkspaceMember{
		WorkspaceID: "workspace-1",
		UserID:      "user-1",
		Role:        "admin",
	}).Exec(ctx)
	require.NoError(t, err)

	_, err = db.NewInsert().Model(&models.Publication{
		ID:              "publication-1",
		WorkspaceID:     "workspace-1",
		CreatedByID:     "user-1",
		Title:           "Launch notes",
		ContentProfile:  "short_text",
		SourceText:      "Initial post text",
		SourceContent:   "Initial post text",
		Status:          models.PublicationStatusScheduled,
		ScheduledAt:     oldRunAt,
		MetadataJSON:    "{}",
		ReleasePlanJSON: "{}",
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Exec(ctx)
	require.NoError(t, err)

	jobs := []models.Job{
		{
			ID:          "old-publication-job",
			Type:        jobTypePublishPublication,
			Payload:     `{"publication_id":"publication-1"}`,
			Status:      "pending",
			RunAt:       oldRunAt,
			MaxAttempts: 3,
		},
		{
			ID:          "other-publication-job",
			Type:        jobTypePublishPublication,
			Payload:     `{"publication_id":"publication-2"}`,
			Status:      "pending",
			RunAt:       oldRunAt,
			MaxAttempts: 3,
		},
	}
	_, err = db.NewInsert().Model(&jobs).Exec(ctx)
	require.NoError(t, err)

	e := echo.New()
	api := humaecho.NewWithGroup(e, e.Group("/api/v1"), huma.DefaultConfig("Test", "1.0.0"))
	NewPublicationHandler(db, testAuthenticator{}, nil).RegisterRoutes(api)

	body := bytes.NewBufferString(`{"scheduled_at":"` + newRunAt.Format(time.RFC3339) + `"}`)
	req := httptest.NewRequestWithContext(ctx, http.MethodPut, "/api/v1/publications/publication-1", body)
	req.Header.Set("Authorization", "Bearer web-token")
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	e.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
	var out PublicationResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	require.Equal(t, newRunAt.Format(time.RFC3339), out.ScheduledAt)

	var remaining []models.Job
	err = db.NewSelect().Model(&remaining).Order("id ASC").Scan(ctx)
	require.NoError(t, err)
	require.Len(t, remaining, 2)

	var rescheduled *models.Job
	for index := range remaining {
		job := &remaining[index]
		require.NotEqual(t, "old-publication-job", job.ID)
		if job.Payload == `{"publication_id":"publication-1"}` {
			rescheduled = job
		}
	}
	require.NotNil(t, rescheduled)
	require.Equal(t, jobTypePublishPublication, rescheduled.Type)
	require.True(t, rescheduled.RunAt.Equal(newRunAt), "expected run_at %s, got %s", newRunAt, rescheduled.RunAt)
	require.Contains(t, jobIDs(remaining), "other-publication-job")
}

func jobIDs(jobs []models.Job) []string {
	ids := make([]string, 0, len(jobs))
	for _, job := range jobs {
		ids = append(ids, job.ID)
	}
	return ids
}
