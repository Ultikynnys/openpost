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

func TestUpsertPublicationRenditionsReplacesOmittedRenditions(t *testing.T) {
	db := createHandlerTestDB(t,
		(*models.WorkspaceMember)(nil),
		(*models.SocialAccount)(nil),
		(*models.Publication)(nil),
		(*models.Rendition)(nil),
		(*models.RenditionMedia)(nil),
		(*models.MediaAttachment)(nil),
		(*models.Job)(nil),
	)
	ctx := context.Background()
	now := time.Date(2026, time.July, 1, 9, 0, 0, 0, time.UTC)

	_, err := db.NewInsert().Model(&models.WorkspaceMember{
		WorkspaceID: "workspace-1",
		UserID:      "user-1",
		Role:        models.WorkspaceRoleAdmin,
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&[]models.SocialAccount{
		{
			ID:              "youtube-account",
			WorkspaceID:     "workspace-1",
			Slug:            "youtube",
			Platform:        "youtube",
			AccountID:       "channel-1",
			AccountUsername: "channel",
			AccessTokenEnc:  []byte("token"),
			IsActive:        true,
		},
		{
			ID:              "tiktok-account",
			WorkspaceID:     "workspace-1",
			Slug:            "tiktok",
			Platform:        "tiktok",
			AccountID:       "open-id",
			AccountUsername: "creator",
			AccessTokenEnc:  []byte("token"),
			IsActive:        true,
		},
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.Publication{
		ID:              "publication-1",
		WorkspaceID:     "workspace-1",
		CreatedByID:     "user-1",
		Title:           "Launch",
		ContentProfile:  models.ContentProfileShortVideo,
		SourceText:      "Launch",
		SourceContent:   "Launch",
		Status:          models.PublicationStatusDraft,
		MetadataJSON:    "{}",
		ReleasePlanJSON: "{}",
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&[]models.Rendition{
		{
			ID:              "youtube-rendition",
			PublicationID:   "publication-1",
			SocialAccountID: "youtube-account",
			Platform:        "youtube",
			Profile:         models.ContentProfileShortVideo,
			Body:            "old youtube",
			Title:           "Old title",
			SettingsJSON:    "{}",
			Status:          models.RenditionStatusDraft,
		},
		{
			ID:              "tiktok-rendition",
			PublicationID:   "publication-1",
			SocialAccountID: "tiktok-account",
			Platform:        "tiktok",
			Profile:         models.ContentProfileShortVideo,
			Body:            "old tiktok",
			Title:           "Old TikTok title",
			SettingsJSON:    "{}",
			Status:          models.RenditionStatusDraft,
		},
	}).Exec(ctx)
	require.NoError(t, err)

	e := echo.New()
	api := humaecho.NewWithGroup(e, e.Group("/api/v1"), huma.DefaultConfig("Test", "1.0.0"))
	NewPublicationHandler(db, testAuthenticator{}, nil).RegisterRoutes(api)

	body := bytes.NewBufferString(`{"renditions":[{"social_account_id":"youtube-account","profile":"short_video","body":"new youtube","title":"New title","description":"New description","settings":{"privacy":"private"}}]}`)
	req := httptest.NewRequestWithContext(ctx, http.MethodPut, "/api/v1/publications/publication-1/renditions", body)
	req.Header.Set("Authorization", "Bearer web-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var out PublicationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out.Renditions, 1)
	require.Equal(t, "youtube-account", out.Renditions[0].SocialAccountID)
	require.Equal(t, "new youtube", out.Renditions[0].Body)

	var persisted []models.Rendition
	require.NoError(t, db.NewSelect().Model(&persisted).Order("social_account_id ASC").Scan(ctx))
	require.Len(t, persisted, 1)
	require.Equal(t, "youtube-account", persisted[0].SocialAccountID)
}
