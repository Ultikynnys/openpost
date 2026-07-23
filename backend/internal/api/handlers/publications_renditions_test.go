package handlers

import (
	"bytes"
	"context"
	"database/sql"
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
	"github.com/uptrace/bun"
)

func TestPublicationPathIDDecodesLegacyPublicationIDs(t *testing.T) {
	require.Equal(t, "legacy-publication:post-1", publicationPathID("legacy-publication%3Apost-1"))
	require.Equal(t, "publication-1", publicationPathID("publication-1"))
}

func TestUpsertPublicationRenditionsPreservesOmittedRenditionsUntilExplicitDelete(t *testing.T) {
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
	require.Len(t, out.Renditions, 2)
	require.Equal(t, "tiktok-account", out.Renditions[0].SocialAccountID)
	require.Equal(t, "youtube-account", out.Renditions[1].SocialAccountID)
	require.Equal(t, "new youtube", out.Renditions[1].Body)

	var persisted []models.Rendition
	require.NoError(t, db.NewSelect().Model(&persisted).Order("social_account_id ASC").Scan(ctx))
	require.Len(t, persisted, 2)
	require.Equal(t, "tiktok-account", persisted[0].SocialAccountID)
	require.Equal(t, "old tiktok", persisted[0].Body)
	require.Equal(t, "youtube-account", persisted[1].SocialAccountID)

	unconfirmedReq := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/api/v1/publications/publication-1/renditions/youtube-account", nil)
	unconfirmedReq.Header.Set("Authorization", "Bearer web-token")
	unconfirmedRec := httptest.NewRecorder()
	e.ServeHTTP(unconfirmedRec, unconfirmedReq)
	require.Equal(t, http.StatusBadRequest, unconfirmedRec.Code, unconfirmedRec.Body.String())

	deleteReq := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/api/v1/publications/publication-1/renditions/youtube-account?confirm=true", nil)
	deleteReq.Header.Set("Authorization", "Bearer web-token")
	deleteRec := httptest.NewRecorder()
	e.ServeHTTP(deleteRec, deleteReq)
	require.Equal(t, http.StatusOK, deleteRec.Code, deleteRec.Body.String())

	persisted = nil
	require.NoError(t, db.NewSelect().Model(&persisted).Order("social_account_id ASC").Scan(ctx))
	require.Len(t, persisted, 1)
	require.Equal(t, "tiktok-account", persisted[0].SocialAccountID)
}

func TestReplacePublicationSegmentsKeepsDestinationOverridesForStableSegmentIDs(t *testing.T) {
	db := createHandlerTestDB(t,
		(*models.Publication)(nil),
		(*models.PublicationSegment)(nil),
		(*models.PublicationSegmentMedia)(nil),
		(*models.Rendition)(nil),
		(*models.RenditionSegment)(nil),
		(*models.RenditionSegmentMedia)(nil),
	)
	ctx := context.Background()
	now := time.Date(2026, time.July, 1, 9, 0, 0, 0, time.UTC)
	publication := &models.Publication{
		ID:              "publication-1",
		WorkspaceID:     "workspace-1",
		CreatedByID:     "user-1",
		Title:           "Launch",
		Intent:          "thread",
		ContentProfile:  models.ContentProfileThread,
		SourceText:      "First",
		SourceContent:   "First",
		Status:          models.PublicationStatusDraft,
		MetadataJSON:    "{}",
		ReleasePlanJSON: "{}",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_, err := db.NewInsert().Model(publication).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.PublicationSegment{
		ID:            "segment-1",
		PublicationID: publication.ID,
		Position:      0,
		Body:          "First",
		SettingsJSON:  "{}",
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.Rendition{
		ID:              "rendition-1",
		PublicationID:   publication.ID,
		SocialAccountID: "account-1",
		Platform:        "x",
		Profile:         models.ContentProfileThread,
		OutputProfile:   "x.thread",
		Body:            "Destination first",
		SettingsJSON:    "{}",
		Status:          models.RenditionStatusDraft,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.RenditionSegment{
		ID:                   "rendition-segment-1",
		RenditionID:          "rendition-1",
		PublicationSegmentID: "segment-1",
		Position:             0,
		Body:                 "Destination first",
		SettingsJSON:         `{"poll_options":"One\nTwo"}`,
		Status:               models.RenditionStatusDraft,
		CreatedAt:            now,
		UpdatedAt:            now,
	}).Exec(ctx)
	require.NoError(t, err)

	handler := NewPublicationHandler(db, testAuthenticator{}, nil)
	err = db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
		return handler.replacePublicationSegments(txCtx, tx, publication, []PublicationSegmentInput{{
			ID:   "segment-1",
			Body: "Updated first",
		}})
	})
	require.NoError(t, err)

	var canonical models.PublicationSegment
	require.NoError(t, db.NewSelect().Model(&canonical).Where("id = ?", "segment-1").Scan(ctx))
	require.Equal(t, "Updated first", canonical.Body)
	var destination models.RenditionSegment
	require.NoError(t, db.NewSelect().Model(&destination).Where("id = ?", "rendition-segment-1").Scan(ctx))
	require.Equal(t, "Destination first", destination.Body)
	require.JSONEq(t, `{"poll_options":"One\nTwo"}`, destination.SettingsJSON)
}
