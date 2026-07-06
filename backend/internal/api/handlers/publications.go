package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/openpost/backend/internal/api/middleware"
	"github.com/openpost/backend/internal/capabilities"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/services/entitlements"
	"github.com/openpost/backend/internal/services/lifecycle"
	"github.com/uptrace/bun"
)

const (
	publicationsPath      = "/publications"
	publicationPathByID   = "/publications/{id}"
	publicationPathValid  = "/publications/{id}/validate"
	publicationEventsPath = "/publications/{id}/events"
)

type PublicationHandler struct {
	db          *bun.DB
	auth        middleware.Authenticator
	entitlement entitlements.Service
}

func NewPublicationHandler(db *bun.DB, authenticator middleware.Authenticator, entitlement entitlements.Service) *PublicationHandler {
	if entitlement == nil {
		entitlement = entitlements.NewSelfHostedService()
	}
	return &PublicationHandler{db: db, auth: authenticator, entitlement: entitlement}
}

type PublicationMediaInput struct {
	MediaID              string `json:"media_id" doc:"Media attachment ID"`
	Role                 string `json:"role,omitempty" doc:"Media role: attachment, cover, thumbnail"`
	AltText              string `json:"alt_text,omitempty" doc:"Alt text override"`
	ThumbnailTimestampMS int    `json:"thumbnail_timestamp_ms,omitempty" doc:"Video thumbnail timestamp"`
}

type RenditionInput struct {
	ID              string                  `json:"id,omitempty" doc:"Existing rendition ID for upsert"`
	SocialAccountID string                  `json:"social_account_id" doc:"Social account ID"`
	Profile         string                  `json:"profile,omitempty" doc:"Content profile override"`
	Body            string                  `json:"body,omitempty" doc:"Platform-specific body"`
	Title           string                  `json:"title,omitempty" doc:"Platform-specific title"`
	Description     string                  `json:"description,omitempty" doc:"Platform-specific description"`
	Settings        map[string]interface{}  `json:"settings,omitempty" doc:"Provider-specific settings"`
	Media           []PublicationMediaInput `json:"media,omitempty" doc:"Rendition-specific ordered media"`
}

type CreatePublicationInput struct {
	Body struct {
		WorkspaceID      string                  `json:"workspace_id" doc:"Workspace ID"`
		Title            string                  `json:"title" doc:"Internal publication title"`
		ContentProfile   string                  `json:"content_profile" doc:"Content profile"`
		SourceText       string                  `json:"source_text" doc:"Canonical source text"`
		SourceURL        string                  `json:"source_url,omitempty" doc:"Source URL for link shares"`
		Goal             string                  `json:"goal,omitempty" doc:"Publication goal"`
		Audience         string                  `json:"audience,omitempty" doc:"Target audience"`
		ScheduledAt      *time.Time              `json:"scheduled_at,omitempty" doc:"Optional schedule time"`
		Metadata         map[string]interface{}  `json:"metadata,omitempty" doc:"Publication metadata"`
		SocialAccountIDs []string                `json:"social_account_ids,omitempty" doc:"Accounts to create default renditions for"`
		Media            []PublicationMediaInput `json:"media,omitempty" doc:"Default ordered media"`
		Renditions       []RenditionInput        `json:"renditions,omitempty" doc:"Explicit platform/account renditions"`
	}
}

type UpdatePublicationInput struct {
	PathID string `path:"id" doc:"Publication ID"`
	Body   struct {
		Title          string                 `json:"title,omitempty" doc:"Internal publication title"`
		ContentProfile string                 `json:"content_profile,omitempty" doc:"Content profile"`
		SourceText     string                 `json:"source_text,omitempty" doc:"Canonical source text"`
		SourceURL      string                 `json:"source_url,omitempty" doc:"Source URL"`
		Goal           string                 `json:"goal,omitempty" doc:"Publication goal"`
		Audience       string                 `json:"audience,omitempty" doc:"Target audience"`
		ScheduledAt    *time.Time             `json:"scheduled_at,omitempty" doc:"Optional schedule time"`
		Metadata       map[string]interface{} `json:"metadata,omitempty" doc:"Publication metadata"`
	}
}

type UpsertRenditionsInput struct {
	PathID string `path:"id" doc:"Publication ID"`
	Body   struct {
		Renditions []RenditionInput `json:"renditions" doc:"Renditions to replace or upsert"`
	}
}

type ListPublicationsInput struct {
	WorkspaceID    string `query:"workspace_id" required:"true" doc:"Workspace ID"`
	Status         string `query:"status" doc:"Optional status filter"`
	ContentProfile string `query:"content_profile" doc:"Optional content profile filter"`
	Limit          int    `query:"limit" doc:"Limit, default 50"`
	Offset         int    `query:"offset" doc:"Offset"`
}

type GetPublicationInput struct {
	PathID string `path:"id" doc:"Publication ID"`
}

type ListPublicationEventsInput struct {
	PathID string `path:"id" doc:"Publication ID"`
	Limit  int    `query:"limit" doc:"Limit, default 100"`
}

type PublicationActionInput struct {
	PathID string `path:"id" doc:"Publication ID"`
}

type ReplyInput struct {
	PathID string `path:"id" doc:"Rendition ID"`
	Body   struct {
		Body     string                  `json:"body" doc:"Reply body"`
		Settings map[string]interface{}  `json:"settings,omitempty" doc:"Provider-specific reply settings"`
		Media    []PublicationMediaInput `json:"media,omitempty" doc:"Reply media"`
		ParentID string                  `json:"parent_id,omitempty" doc:"External comment or post ID to reply to"`
		RunAt    *time.Time              `json:"run_at,omitempty" doc:"Optional scheduled reply time"`
	}
}

type PublicationOutput struct {
	Body PublicationResponse
}

type PublicationListOutput struct {
	TotalCount int  `header:"X-Total-Count"`
	Limit      int  `header:"X-Limit"`
	Offset     int  `header:"X-Offset"`
	NextOffset int  `header:"X-Next-Offset"`
	HasMore    bool `header:"X-Has-More"`
	Body       []PublicationResponse
}

type PublicationValidationOutput struct {
	Body struct {
		Valid  bool                           `json:"valid"`
		Issues []capabilities.ValidationIssue `json:"issues"`
	}
}

type PublicationEventsOutput struct {
	Body []PublicationLifecycleEventResponse
}

type ActionOutput struct {
	Body struct {
		Message string `json:"message"`
		JobID   string `json:"job_id,omitempty"`
	}
}

type PublicationResponse struct {
	ID             string              `json:"id"`
	WorkspaceID    string              `json:"workspace_id"`
	CreatedByID    string              `json:"created_by"`
	Title          string              `json:"title"`
	ContentProfile string              `json:"content_profile"`
	SourceText     string              `json:"source_text"`
	SourceURL      string              `json:"source_url,omitempty"`
	Goal           string              `json:"goal,omitempty"`
	Audience       string              `json:"audience,omitempty"`
	Status         string              `json:"status"`
	ScheduledAt    string              `json:"scheduled_at,omitempty"`
	ActualRunAt    string              `json:"actual_run_at,omitempty"`
	Metadata       map[string]any      `json:"metadata"`
	CreatedAt      string              `json:"created_at"`
	UpdatedAt      string              `json:"updated_at"`
	Renditions     []RenditionResponse `json:"renditions"`
	Media          []MediaSummary      `json:"media"`
}

type RenditionResponse struct {
	ID              string                 `json:"id"`
	PublicationID   string                 `json:"publication_id"`
	SocialAccountID string                 `json:"social_account_id"`
	Platform        string                 `json:"platform"`
	Profile         string                 `json:"profile"`
	Body            string                 `json:"body"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Settings        map[string]interface{} `json:"settings"`
	Status          string                 `json:"status"`
	ExternalID      string                 `json:"external_id,omitempty"`
	ExternalURL     string                 `json:"external_url,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	Media           []MediaSummary         `json:"media"`
}

type MediaSummary struct {
	ID                   string  `json:"id"`
	MimeType             string  `json:"mime_type"`
	Size                 int64   `json:"size"`
	OriginalFilename     string  `json:"original_filename"`
	Width                int     `json:"width"`
	Height               int     `json:"height"`
	DurationMS           int64   `json:"duration_ms"`
	FrameRate            float64 `json:"frame_rate"`
	AspectRatio          string  `json:"aspect_ratio"`
	DominantType         string  `json:"dominant_type"`
	PosterThumbnailURL   string  `json:"poster_thumbnail_url,omitempty"`
	AnalysisStatus       string  `json:"analysis_status"`
	AnalysisError        string  `json:"analysis_error,omitempty"`
	PublicURLReady       bool    `json:"public_url_ready"`
	PublicURLCheckedAt   string  `json:"public_url_checked_at,omitempty"`
	PublicURLStatus      int     `json:"public_url_status"`
	PublicURLError       string  `json:"public_url_error,omitempty"`
	URL                  string  `json:"url"`
	Role                 string  `json:"role,omitempty"`
	DisplayOrder         int     `json:"display_order,omitempty"`
	AltText              string  `json:"alt_text,omitempty"`
	ThumbnailTimestampMS int     `json:"thumbnail_timestamp_ms,omitempty"`
}

type PublicationLifecycleEventResponse struct {
	ID             string         `json:"id"`
	WorkspaceID    string         `json:"workspace_id"`
	PublicationID  string         `json:"publication_id"`
	RenditionID    string         `json:"rendition_id,omitempty"`
	Type           string         `json:"type"`
	Status         string         `json:"status"`
	Message        string         `json:"message"`
	Metadata       map[string]any `json:"metadata"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
	CreatedAt      string         `json:"created_at"`
}

func (h *PublicationHandler) RegisterRoutes(api huma.API) {
	h.createPublication(api)
	h.listPublications(api)
	h.getPublication(api)
	h.listPublicationEvents(api)
	h.updatePublication(api)
	h.upsertRenditions(api)
	h.validatePublication(api)
	h.schedulePublication(api)
	h.publishNow(api)
	h.replyToRendition(api)
}

func (h *PublicationHandler) createPublication(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "create-publication",
		Method:      http.MethodPost,
		Path:        publicationsPath,
		Summary:     "Create a publication",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{400, 403},
	}, func(ctx context.Context, input *CreatePublicationInput) (*PublicationOutput, error) {
		userID := middleware.GetUserID(ctx)
		if input.Body.WorkspaceID == "" {
			return nil, huma.Error400BadRequest(errWorkspaceIDRequired)
		}
		if err := h.checkWorkspaceAccess(ctx, input.Body.WorkspaceID, userID); err != nil {
			return nil, err
		}
		if input.Body.ContentProfile == "" {
			input.Body.ContentProfile = models.ContentProfileShortText
		}
		if len(input.Body.Renditions) == 0 {
			input.Body.Renditions = h.defaultRenditionInputs(input.Body.SocialAccountIDs, input.Body.ContentProfile, input.Body.SourceText, input.Body.Title, input.Body.Media)
		}
		accountMap, err := h.loadAccounts(ctx, input.Body.WorkspaceID, renditionAccountIDs(input.Body.Renditions))
		if err != nil {
			return nil, err
		}
		if err := h.validateMediaBelongsToWorkspace(ctx, input.Body.WorkspaceID, allMediaIDs(input.Body.Media, input.Body.Renditions)); err != nil {
			return nil, err
		}

		now := time.Now().UTC()
		metadataJSON := mustJSON(input.Body.Metadata)
		publication := &models.Publication{
			ID:              uuid.New().String(),
			WorkspaceID:     input.Body.WorkspaceID,
			CreatedByID:     userID,
			Title:           publicationFirstNonEmpty(input.Body.Title, firstContentLine(input.Body.SourceText), "Untitled publication"),
			ContentProfile:  input.Body.ContentProfile,
			SourceText:      input.Body.SourceText,
			SourceContent:   input.Body.SourceText,
			SourceURL:       input.Body.SourceURL,
			Goal:            input.Body.Goal,
			Audience:        input.Body.Audience,
			Status:          models.PublicationStatusDraft,
			MetadataJSON:    metadataJSON,
			ReleasePlanJSON: metadataJSON,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if input.Body.ScheduledAt != nil {
			publication.ScheduledAt = *input.Body.ScheduledAt
			publication.Status = models.PublicationStatusScheduled
		}

		err = h.db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
			if _, err := tx.NewInsert().Model(publication).Exec(txCtx); err != nil {
				return err
			}
			return h.insertRenditions(txCtx, tx, publication, input.Body.Renditions, input.Body.Media, accountMap)
		})
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}

		resp, err := h.loadPublicationResponse(ctx, publication.ID, userID)
		if err != nil {
			return nil, err
		}
		return &PublicationOutput{Body: resp}, nil
	})
}

func (h *PublicationHandler) listPublications(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-publications",
		Method:      http.MethodGet,
		Path:        publicationsPath,
		Summary:     "List publications",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
	}, func(ctx context.Context, input *ListPublicationsInput) (*PublicationListOutput, error) {
		userID := middleware.GetUserID(ctx)
		if err := h.checkWorkspaceAccess(ctx, input.WorkspaceID, userID); err != nil {
			return nil, err
		}
		limit := input.Limit
		if limit <= 0 || limit > 200 {
			limit = 50
		}
		query := h.db.NewSelect().Model((*models.Publication)(nil)).Where("workspace_id = ?", input.WorkspaceID)
		if input.Status != "" {
			query = query.Where("status = ?", input.Status)
		}
		if input.ContentProfile != "" {
			query = query.Where("content_profile = ?", input.ContentProfile)
		}
		total, err := query.Count(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to count publications")
		}
		var publications []models.Publication
		if err := query.Order("created_at DESC").Limit(limit).Offset(input.Offset).Scan(ctx, &publications); err != nil {
			return nil, huma.Error500InternalServerError("failed to list publications")
		}
		body := make([]PublicationResponse, 0, len(publications))
		for _, publication := range publications {
			resp, err := h.loadPublicationResponse(ctx, publication.ID, userID)
			if err != nil {
				return nil, err
			}
			body = append(body, resp)
		}
		next := input.Offset + len(body)
		return &PublicationListOutput{
			TotalCount: total,
			Limit:      limit,
			Offset:     input.Offset,
			NextOffset: next,
			HasMore:    next < total,
			Body:       body,
		}, nil
	})
}

func (h *PublicationHandler) getPublication(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-publication",
		Method:      http.MethodGet,
		Path:        publicationPathByID,
		Summary:     "Get a publication",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{404},
	}, func(ctx context.Context, input *GetPublicationInput) (*PublicationOutput, error) {
		resp, err := h.loadPublicationResponse(ctx, input.PathID, middleware.GetUserID(ctx))
		if err != nil {
			return nil, err
		}
		return &PublicationOutput{Body: resp}, nil
	})
}

func (h *PublicationHandler) listPublicationEvents(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-publication-events",
		Method:      http.MethodGet,
		Path:        publicationEventsPath,
		Summary:     "List publication lifecycle events",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{404},
	}, func(ctx context.Context, input *ListPublicationEventsInput) (*PublicationEventsOutput, error) {
		publication, err := h.loadPublication(ctx, input.PathID, middleware.GetUserID(ctx))
		if err != nil {
			return nil, err
		}
		events, err := lifecycle.NewService(h.db).ListForPublication(ctx, publication.WorkspaceID, publication.ID, input.Limit)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to load publication events")
		}
		body := make([]PublicationLifecycleEventResponse, 0, len(events))
		for _, event := range events {
			body = append(body, publicationLifecycleEventResponse(event))
		}
		return &PublicationEventsOutput{Body: body}, nil
	})
}

func (h *PublicationHandler) updatePublication(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "update-publication",
		Method:      http.MethodPut,
		Path:        publicationPathByID,
		Summary:     "Update a publication",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
	}, func(ctx context.Context, input *UpdatePublicationInput) (*PublicationOutput, error) {
		publication, err := h.loadPublication(ctx, input.PathID, middleware.GetUserID(ctx))
		if err != nil {
			return nil, err
		}
		if input.Body.Title != "" {
			publication.Title = input.Body.Title
		}
		if input.Body.ContentProfile != "" {
			publication.ContentProfile = input.Body.ContentProfile
		}
		if input.Body.SourceText != "" {
			publication.SourceText = input.Body.SourceText
			publication.SourceContent = input.Body.SourceText
		}
		publication.SourceURL = input.Body.SourceURL
		publication.Goal = input.Body.Goal
		publication.Audience = input.Body.Audience
		rescheduleQueuedJob := input.Body.ScheduledAt != nil && publication.Status == models.PublicationStatusScheduled
		if input.Body.ScheduledAt != nil {
			publication.ScheduledAt = *input.Body.ScheduledAt
		}
		if input.Body.Metadata != nil {
			publication.MetadataJSON = mustJSON(input.Body.Metadata)
			publication.ReleasePlanJSON = publication.MetadataJSON
		}
		publication.UpdatedAt = time.Now().UTC()
		if err := h.db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
			if _, err := tx.NewUpdate().Model(publication).Where("id = ?", publication.ID).Exec(txCtx); err != nil {
				return err
			}
			if rescheduleQueuedJob {
				_, err := h.replacePublicationJobTx(txCtx, tx, publication.ID, publication.ScheduledAt)
				return err
			}
			return nil
		}); err != nil {
			return nil, huma.Error500InternalServerError("failed to update publication")
		}
		resp, err := h.loadPublicationResponse(ctx, publication.ID, middleware.GetUserID(ctx))
		if err != nil {
			return nil, err
		}
		return &PublicationOutput{Body: resp}, nil
	})
}

func (h *PublicationHandler) upsertRenditions(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "upsert-publication-renditions",
		Method:      http.MethodPut,
		Path:        "/publications/{id}/renditions",
		Summary:     "Replace or upsert publication renditions",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
	}, func(ctx context.Context, input *UpsertRenditionsInput) (*PublicationOutput, error) {
		publication, err := h.loadPublication(ctx, input.PathID, middleware.GetUserID(ctx))
		if err != nil {
			return nil, err
		}
		accountMap, err := h.loadAccounts(ctx, publication.WorkspaceID, renditionAccountIDs(input.Body.Renditions))
		if err != nil {
			return nil, err
		}
		if err := h.validateMediaBelongsToWorkspace(ctx, publication.WorkspaceID, allMediaIDs(nil, input.Body.Renditions)); err != nil {
			return nil, err
		}
		err = h.db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
			var existingIDs []string
			if err := tx.NewSelect().
				Model((*models.Rendition)(nil)).
				Column("id").
				Where("publication_id = ?", publication.ID).
				Scan(txCtx, &existingIDs); err != nil {
				return err
			}
			if len(existingIDs) > 0 {
				if _, err := tx.NewDelete().
					Model((*models.RenditionMedia)(nil)).
					Where("rendition_id IN (?)", bun.List(existingIDs)).
					Exec(txCtx); err != nil {
					return err
				}
				if _, err := tx.NewDelete().
					Model((*models.Rendition)(nil)).
					Where("publication_id = ?", publication.ID).
					Exec(txCtx); err != nil {
					return err
				}
			}
			return h.insertRenditions(txCtx, tx, publication, input.Body.Renditions, nil, accountMap)
		})
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		resp, err := h.loadPublicationResponse(ctx, publication.ID, middleware.GetUserID(ctx))
		if err != nil {
			return nil, err
		}
		return &PublicationOutput{Body: resp}, nil
	})
}

func (h *PublicationHandler) validatePublication(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "validate-publication",
		Method:      http.MethodPost,
		Path:        publicationPathValid,
		Summary:     "Validate publication renditions",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
	}, func(ctx context.Context, input *PublicationActionInput) (*PublicationValidationOutput, error) {
		if _, err := h.loadPublication(ctx, input.PathID, middleware.GetUserID(ctx)); err != nil {
			return nil, err
		}
		issues, err := h.validatePublicationByID(ctx, input.PathID)
		if err != nil {
			return nil, err
		}
		resp := &PublicationValidationOutput{}
		resp.Body.Issues = issues
		resp.Body.Valid = !hasBlockingIssues(issues)
		return resp, nil
	})
}

func (h *PublicationHandler) schedulePublication(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "schedule-publication",
		Method:      http.MethodPost,
		Path:        "/publications/{id}/schedule",
		Summary:     "Schedule a publication",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
	}, func(ctx context.Context, input *PublicationActionInput) (*ActionOutput, error) {
		publication, err := h.loadPublication(ctx, input.PathID, middleware.GetUserID(ctx))
		if err != nil {
			return nil, err
		}
		if publication.ScheduledAt.IsZero() {
			return nil, huma.Error400BadRequest("scheduled_at is required before scheduling")
		}
		issues, err := h.validatePublicationByID(ctx, publication.ID)
		if err != nil {
			return nil, err
		}
		if hasBlockingIssues(issues) {
			return nil, huma.Error400BadRequest("publication has blocking validation errors")
		}
		jobID, err := h.replacePublicationJob(ctx, publication.ID, publication.ScheduledAt)
		if err != nil {
			return nil, err
		}
		if err := h.markPublicationQueued(ctx, publication.ID); err != nil {
			return nil, err
		}
		return actionMessage("publication scheduled", jobID), nil
	})
}

func (h *PublicationHandler) publishNow(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "publish-publication-now",
		Method:      http.MethodPost,
		Path:        "/publications/{id}/publish-now",
		Summary:     "Publish a publication now",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
	}, func(ctx context.Context, input *PublicationActionInput) (*ActionOutput, error) {
		publication, err := h.loadPublication(ctx, input.PathID, middleware.GetUserID(ctx))
		if err != nil {
			return nil, err
		}
		issues, err := h.validatePublicationByID(ctx, publication.ID)
		if err != nil {
			return nil, err
		}
		if hasBlockingIssues(issues) {
			return nil, huma.Error400BadRequest("publication has blocking validation errors")
		}
		jobID, err := h.replacePublicationJob(ctx, publication.ID, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		if err := h.markPublicationQueued(ctx, publication.ID); err != nil {
			return nil, err
		}
		return actionMessage("publication queued", jobID), nil
	})
}

func (h *PublicationHandler) replyToRendition(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "reply-to-rendition",
		Method:      http.MethodPost,
		Path:        "/renditions/{id}/reply",
		Summary:     "Queue an explicit provider reply",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
	}, func(ctx context.Context, input *ReplyInput) (*ActionOutput, error) {
		rendition, publication, err := h.loadRenditionWithPublication(ctx, input.PathID, middleware.GetUserID(ctx))
		if err != nil {
			return nil, err
		}
		payload := map[string]interface{}{
			"rendition_id":   rendition.ID,
			"publication_id": publication.ID,
			"body":           input.Body.Body,
			"parent_id":      input.Body.ParentID,
			"settings":       input.Body.Settings,
			"media":          input.Body.Media,
			"action":         "reply",
		}
		payloadJSON := mustJSON(payload)
		runAt := time.Now().UTC()
		if input.Body.RunAt != nil {
			runAt = *input.Body.RunAt
		}
		job := &models.Job{ID: uuid.New().String(), Type: jobTypePublishPublication, Payload: payloadJSON, Status: "pending", RunAt: runAt, MaxAttempts: 3}
		if _, err := h.db.NewInsert().Model(job).Exec(ctx); err != nil {
			return nil, huma.Error500InternalServerError("failed to enqueue reply")
		}
		return actionMessage("reply queued", job.ID), nil
	})
}

func (h *PublicationHandler) insertRenditions(ctx context.Context, tx bun.Tx, publication *models.Publication, inputs []RenditionInput, defaultMedia []PublicationMediaInput, accounts map[string]models.SocialAccount) error {
	now := time.Now().UTC()
	for _, input := range inputs {
		account, ok := accounts[input.SocialAccountID]
		if !ok {
			return huma.Error400BadRequest("one or more social accounts are invalid, disconnected, or outside this workspace")
		}
		profile := publicationFirstNonEmpty(input.Profile, publication.ContentProfile)
		settingsJSON := mustJSON(input.Settings)
		rendition := &models.Rendition{
			ID:              publicationFirstNonEmpty(input.ID, uuid.New().String()),
			PublicationID:   publication.ID,
			SocialAccountID: input.SocialAccountID,
			Platform:        account.Platform,
			Profile:         profile,
			Body:            publicationFirstNonEmpty(input.Body, publication.SourceText),
			Title:           publicationFirstNonEmpty(input.Title, publication.Title),
			Description:     input.Description,
			SettingsJSON:    settingsJSON,
			Status:          models.RenditionStatusDraft,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if publication.Status == models.PublicationStatusScheduled {
			rendition.Status = models.RenditionStatusScheduled
		}
		if _, err := tx.NewInsert().Model(rendition).Exec(ctx); err != nil {
			return err
		}
		mediaInputs := input.Media
		if len(mediaInputs) == 0 {
			mediaInputs = defaultMedia
		}
		for order, mediaInput := range mediaInputs {
			role := publicationFirstNonEmpty(mediaInput.Role, "attachment")
			row := &models.RenditionMedia{
				RenditionID:          rendition.ID,
				MediaID:              mediaInput.MediaID,
				Role:                 role,
				DisplayOrder:         order,
				AltText:              mediaInput.AltText,
				ThumbnailTimestampMS: mediaInput.ThumbnailTimestampMS,
			}
			if _, err := tx.NewInsert().Model(row).Exec(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *PublicationHandler) loadPublicationResponse(ctx context.Context, publicationID, userID string) (PublicationResponse, error) {
	publication, err := h.loadPublication(ctx, publicationID, userID)
	if err != nil {
		return PublicationResponse{}, err
	}
	var renditions []models.Rendition
	if err := h.db.NewSelect().Model(&renditions).Where("publication_id = ?", publication.ID).Order("created_at ASC").Scan(ctx); err != nil {
		return PublicationResponse{}, huma.Error500InternalServerError("failed to load renditions")
	}
	mediaByRendition, publicationMedia, err := h.loadRenditionMedia(ctx, renditionIDs(renditions))
	if err != nil {
		return PublicationResponse{}, err
	}
	response := publicationResponse(publication, publicationMedia)
	response.Renditions = make([]RenditionResponse, 0, len(renditions))
	for _, rendition := range renditions {
		response.Renditions = append(response.Renditions, renditionResponse(rendition, mediaByRendition[rendition.ID]))
	}
	return response, nil
}

func (h *PublicationHandler) loadPublication(ctx context.Context, publicationID, userID string) (*models.Publication, error) {
	var publication models.Publication
	if err := h.db.NewSelect().Model(&publication).Where("id = ?", publicationID).Scan(ctx); err != nil {
		return nil, huma.Error404NotFound("publication not found")
	}
	if err := h.checkWorkspaceAccess(ctx, publication.WorkspaceID, userID); err != nil {
		return nil, err
	}
	return &publication, nil
}

func (h *PublicationHandler) loadRenditionWithPublication(ctx context.Context, renditionID, userID string) (*models.Rendition, *models.Publication, error) {
	var rendition models.Rendition
	if err := h.db.NewSelect().Model(&rendition).Where("id = ?", renditionID).Scan(ctx); err != nil {
		return nil, nil, huma.Error404NotFound("rendition not found")
	}
	publication, err := h.loadPublication(ctx, rendition.PublicationID, userID)
	if err != nil {
		return nil, nil, err
	}
	return &rendition, publication, nil
}

func (h *PublicationHandler) loadRenditionMedia(ctx context.Context, ids []string) (map[string][]MediaSummary, []MediaSummary, error) {
	out := map[string][]MediaSummary{}
	publicationMedia := []MediaSummary{}
	if len(ids) == 0 {
		return out, publicationMedia, nil
	}
	var rows []struct {
		RenditionID          string `bun:"rendition_id"`
		Role                 string `bun:"role"`
		DisplayOrder         int    `bun:"display_order"`
		AltText              string `bun:"alt_text"`
		ThumbnailTimestampMS int    `bun:"thumbnail_timestamp_ms"`
		models.MediaAttachment
	}
	if err := h.db.NewSelect().
		TableExpr("rendition_media AS rm").
		ColumnExpr("rm.rendition_id, rm.role, rm.display_order, rm.alt_text, rm.thumbnail_timestamp_ms").
		ColumnExpr("m.*").
		Join("JOIN media_attachments AS m ON m.id = rm.media_id").
		Where("rm.rendition_id IN (?)", bun.List(ids)).
		Order("rm.rendition_id ASC", "rm.display_order ASC").
		Scan(ctx, &rows); err != nil {
		return nil, nil, huma.Error500InternalServerError("failed to load rendition media")
	}
	seenPublicationMedia := map[string]struct{}{}
	for _, row := range rows {
		item := mediaSummary(row.MediaAttachment, row.Role, row.DisplayOrder, row.AltText, row.ThumbnailTimestampMS)
		out[row.RenditionID] = append(out[row.RenditionID], item)
		if _, ok := seenPublicationMedia[item.ID]; !ok {
			seenPublicationMedia[item.ID] = struct{}{}
			publicationMedia = append(publicationMedia, item)
		}
	}
	sort.Slice(publicationMedia, func(i, j int) bool { return publicationMedia[i].DisplayOrder < publicationMedia[j].DisplayOrder })
	return out, publicationMedia, nil
}

func (h *PublicationHandler) validatePublicationByID(ctx context.Context, publicationID string) ([]capabilities.ValidationIssue, error) {
	var renditions []models.Rendition
	if err := h.db.NewSelect().Model(&renditions).Where("publication_id = ?", publicationID).Scan(ctx); err != nil {
		return nil, huma.Error500InternalServerError("failed to load renditions")
	}
	mediaByRendition, _, err := h.loadRenditionMedia(ctx, renditionIDs(renditions))
	if err != nil {
		return nil, err
	}
	accountsByID, err := h.loadValidationAccounts(ctx, renditionAccountIDsFromModels(renditions))
	if err != nil {
		return nil, err
	}
	issues := []capabilities.ValidationIssue{}
	for _, rendition := range renditions {
		settings := map[string]interface{}{}
		_ = json.Unmarshal([]byte(rendition.SettingsJSON), &settings)
		mediaItems := make([]capabilities.MediaItem, 0, len(mediaByRendition[rendition.ID]))
		for _, item := range mediaByRendition[rendition.ID] {
			mediaItems = append(mediaItems, capabilities.MediaItem{
				ID:              item.ID,
				MimeType:        item.MimeType,
				Size:            item.Size,
				Width:           item.Width,
				Height:          item.Height,
				DurationMS:      item.DurationMS,
				AnalysisStatus:  item.AnalysisStatus,
				AnalysisError:   item.AnalysisError,
				PublicURLReady:  item.PublicURLReady,
				PublicURLStatus: item.PublicURLStatus,
				PublicURLError:  item.PublicURLError,
				URL:             item.URL,
			})
		}
		issues = append(issues, capabilities.Validate(rendition.Platform, rendition.Profile, rendition.Body, rendition.Title, rendition.Description, mediaItems, settings)...)
		if account, ok := accountsByID[rendition.SocialAccountID]; ok {
			issues = append(issues, renditionScopeIssues(rendition, account)...)
		}
		issues = append(issues, renditionProcessingIssues(rendition)...)
	}
	return issues, nil
}

func (h *PublicationHandler) loadValidationAccounts(ctx context.Context, accountIDs []string) (map[string]models.SocialAccount, error) {
	uniqueIDs := uniqueNonEmpty(accountIDs)
	if len(uniqueIDs) == 0 {
		return map[string]models.SocialAccount{}, nil
	}
	var accounts []models.SocialAccount
	if err := h.db.NewSelect().
		Model(&accounts).
		Where("id IN (?)", bun.List(uniqueIDs)).
		Where("is_active = ?", true).
		Scan(ctx); err != nil {
		return nil, huma.Error500InternalServerError("failed to load social account scopes")
	}
	out := make(map[string]models.SocialAccount, len(accounts))
	for _, account := range accounts {
		out[account.ID] = account
	}
	return out, nil
}

func renditionScopeIssues(rendition models.Rendition, account models.SocialAccount) []capabilities.ValidationIssue {
	granted := splitScopes(account.GrantedScopes)
	if len(granted) == 0 {
		return nil
	}
	missing := missingScopes(requiredScopes(rendition.Platform), granted)
	if len(missing) == 0 {
		return nil
	}
	return []capabilities.ValidationIssue{{
		Severity: "error",
		Code:     "missing_scope",
		Message:  "Connected account is missing required publishing scopes: " + strings.Join(missing, ", "),
		Provider: rendition.Platform,
		Profile:  rendition.Profile,
		Field:    "granted_scopes",
	}}
}

func renditionProcessingIssues(rendition models.Rendition) []capabilities.ValidationIssue {
	if rendition.Status != models.RenditionStatusFailed || strings.TrimSpace(rendition.ErrorMessage) == "" {
		return nil
	}
	return []capabilities.ValidationIssue{{
		Severity: "error",
		Code:     "native_processing_failed",
		Message:  rendition.ErrorMessage,
		Provider: rendition.Platform,
		Profile:  rendition.Profile,
		Field:    "status",
	}}
}

func (h *PublicationHandler) checkWorkspaceAccess(ctx context.Context, workspaceID, userID string) error {
	if workspaceID == "" {
		return huma.Error400BadRequest(errWorkspaceIDRequired)
	}
	if !middleware.WorkspaceScopeAllows(ctx, workspaceID) {
		return huma.Error403Forbidden(errWorkspaceAccessDenied)
	}
	var members []models.WorkspaceMember
	if err := h.db.NewSelect().Model(&members).Where("workspace_id = ? AND user_id = ?", workspaceID, userID).Scan(ctx); err != nil {
		return huma.Error500InternalServerError(errValidateWorkspaceAccess)
	}
	if len(members) == 0 {
		return huma.Error403Forbidden(errWorkspaceAccessDenied)
	}
	return nil
}

func (h *PublicationHandler) loadAccounts(ctx context.Context, workspaceID string, accountIDs []string) (map[string]models.SocialAccount, error) {
	uniqueIDs := uniqueNonEmpty(accountIDs)
	if len(uniqueIDs) == 0 {
		return map[string]models.SocialAccount{}, nil
	}
	var accounts []models.SocialAccount
	if err := h.db.NewSelect().Model(&accounts).
		Where("workspace_id = ?", workspaceID).
		Where("is_active = ?", true).
		Where("id IN (?)", bun.List(uniqueIDs)).
		Scan(ctx); err != nil {
		return nil, huma.Error500InternalServerError("failed to validate social accounts")
	}
	if len(accounts) != len(uniqueIDs) {
		return nil, huma.Error400BadRequest("one or more social accounts are invalid, disconnected, or outside this workspace")
	}
	out := make(map[string]models.SocialAccount, len(accounts))
	for _, account := range accounts {
		out[account.ID] = account
	}
	return out, nil
}

func (h *PublicationHandler) validateMediaBelongsToWorkspace(ctx context.Context, workspaceID string, mediaIDs []string) error {
	uniqueIDs := uniqueNonEmpty(mediaIDs)
	if len(uniqueIDs) == 0 {
		return nil
	}
	count, err := h.db.NewSelect().
		Model((*models.MediaAttachment)(nil)).
		Where("workspace_id = ?", workspaceID).
		Where("id IN (?)", bun.List(uniqueIDs)).
		Count(ctx)
	if err != nil {
		return huma.Error500InternalServerError("failed to validate media attachments")
	}
	if count != len(uniqueIDs) {
		return huma.Error400BadRequest("one or more media attachments are invalid or outside this workspace")
	}
	return nil
}

func (h *PublicationHandler) defaultRenditionInputs(accountIDs []string, profile, body, title string, media []PublicationMediaInput) []RenditionInput {
	out := make([]RenditionInput, 0, len(accountIDs))
	for _, accountID := range uniqueNonEmpty(accountIDs) {
		out = append(out, RenditionInput{SocialAccountID: accountID, Profile: profile, Body: body, Title: title, Media: media})
	}
	return out
}

func (h *PublicationHandler) replacePublicationJob(ctx context.Context, publicationID string, runAt time.Time) (string, error) {
	var jobID string
	err := h.db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
		var err error
		jobID, err = h.replacePublicationJobTx(txCtx, tx, publicationID, runAt)
		return err
	})
	if err != nil {
		return "", huma.Error500InternalServerError("failed to enqueue publication")
	}
	return jobID, nil
}

func (h *PublicationHandler) replacePublicationJobTx(ctx context.Context, tx bun.Tx, publicationID string, runAt time.Time) (string, error) {
	if _, err := tx.NewDelete().
		Model((*models.Job)(nil)).
		Where(publishPublicationJobPublicationIDWhere(h.db), jobTypePublishPublication, publicationID).
		Exec(ctx); err != nil {
		return "", err
	}
	payload := mustJSON(map[string]string{"publication_id": publicationID})
	job := &models.Job{ID: uuid.New().String(), Type: jobTypePublishPublication, Payload: payload, Status: "pending", RunAt: runAt, MaxAttempts: 3}
	if _, err := tx.NewInsert().Model(job).Exec(ctx); err != nil {
		return "", err
	}
	return job.ID, nil
}

func (h *PublicationHandler) markPublicationQueued(ctx context.Context, publicationID string) error {
	if _, err := h.db.NewUpdate().Model((*models.Publication)(nil)).
		Set("status = ?", models.PublicationStatusScheduled).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", publicationID).
		Exec(ctx); err != nil {
		return huma.Error500InternalServerError("failed to update publication")
	}
	if _, err := h.db.NewUpdate().Model((*models.Rendition)(nil)).
		Set("status = ?", models.RenditionStatusScheduled).
		Set("updated_at = ?", time.Now().UTC()).
		Where("publication_id = ?", publicationID).
		Where("status NOT IN (?)", bun.List([]string{models.RenditionStatusPublished, models.RenditionStatusPublishing})).
		Exec(ctx); err != nil {
		return huma.Error500InternalServerError("failed to update renditions")
	}
	return nil
}

func publicationResponse(publication *models.Publication, media []MediaSummary) PublicationResponse {
	metadata := map[string]any{}
	_ = json.Unmarshal([]byte(publication.MetadataJSON), &metadata)
	return PublicationResponse{
		ID:             publication.ID,
		WorkspaceID:    publication.WorkspaceID,
		CreatedByID:    publication.CreatedByID,
		Title:          publication.Title,
		ContentProfile: publication.ContentProfile,
		SourceText:     publication.SourceText,
		SourceURL:      publication.SourceURL,
		Goal:           publication.Goal,
		Audience:       publication.Audience,
		Status:         publication.Status,
		ScheduledAt:    formatOptionalTime(publication.ScheduledAt),
		ActualRunAt:    formatOptionalTime(publication.ActualRunAt),
		Metadata:       metadata,
		CreatedAt:      publication.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      publication.UpdatedAt.Format(time.RFC3339),
		Media:          media,
	}
}

func renditionResponse(rendition models.Rendition, media []MediaSummary) RenditionResponse {
	settings := map[string]interface{}{}
	_ = json.Unmarshal([]byte(rendition.SettingsJSON), &settings)
	return RenditionResponse{
		ID:              rendition.ID,
		PublicationID:   rendition.PublicationID,
		SocialAccountID: rendition.SocialAccountID,
		Platform:        rendition.Platform,
		Profile:         rendition.Profile,
		Body:            rendition.Body,
		Title:           rendition.Title,
		Description:     rendition.Description,
		Settings:        settings,
		Status:          rendition.Status,
		ExternalID:      rendition.ExternalID,
		ExternalURL:     rendition.ExternalURL,
		ErrorMessage:    rendition.ErrorMessage,
		Media:           media,
	}
}

func mediaSummary(media models.MediaAttachment, role string, order int, altText string, thumbnailTimestampMS int) MediaSummary {
	if altText == "" {
		altText = media.AltText
	}
	return MediaSummary{
		ID:                   media.ID,
		MimeType:             media.MimeType,
		Size:                 media.Size,
		OriginalFilename:     media.OriginalFilename,
		Width:                media.Width,
		Height:               media.Height,
		DurationMS:           media.DurationMS,
		FrameRate:            media.FrameRate,
		AspectRatio:          media.AspectRatio,
		DominantType:         media.DominantType,
		PosterThumbnailURL:   mediaPublicationPosterURL(media),
		AnalysisStatus:       media.AnalysisStatus,
		AnalysisError:        media.AnalysisError,
		PublicURLReady:       media.PublicURLReady,
		PublicURLCheckedAt:   formatOptionalTime(media.PublicURLCheckedAt),
		PublicURLStatus:      media.PublicURLStatus,
		PublicURLError:       media.PublicURLError,
		URL:                  "/media/" + media.ID,
		Role:                 role,
		DisplayOrder:         order,
		AltText:              altText,
		ThumbnailTimestampMS: thumbnailTimestampMS,
	}
}

func mediaPublicationPosterURL(media models.MediaAttachment) string {
	if media.ThumbnailObjectKey == "" {
		return ""
	}
	return "/media/" + media.ID + "/poster"
}

func publicationLifecycleEventResponse(event models.PublicationLifecycleEvent) PublicationLifecycleEventResponse {
	metadata := map[string]any{}
	_ = json.Unmarshal([]byte(event.MetadataJSON), &metadata)
	return PublicationLifecycleEventResponse{
		ID:             event.ID,
		WorkspaceID:    event.WorkspaceID,
		PublicationID:  event.PublicationID,
		RenditionID:    event.RenditionID,
		Type:           event.Type,
		Status:         event.Status,
		Message:        event.Message,
		Metadata:       metadata,
		IdempotencyKey: event.IdempotencyKey,
		CreatedAt:      event.CreatedAt.Format(time.RFC3339),
	}
}

func renditionAccountIDs(renditions []RenditionInput) []string {
	out := make([]string, 0, len(renditions))
	for _, rendition := range renditions {
		out = append(out, rendition.SocialAccountID)
	}
	return out
}

func allMediaIDs(defaultMedia []PublicationMediaInput, renditions []RenditionInput) []string {
	out := make([]string, 0, len(defaultMedia))
	for _, item := range defaultMedia {
		out = append(out, item.MediaID)
	}
	for _, rendition := range renditions {
		for _, item := range rendition.Media {
			out = append(out, item.MediaID)
		}
	}
	return out
}

func renditionIDs(renditions []models.Rendition) []string {
	out := make([]string, 0, len(renditions))
	for _, rendition := range renditions {
		out = append(out, rendition.ID)
	}
	return out
}

func renditionAccountIDsFromModels(renditions []models.Rendition) []string {
	out := make([]string, 0, len(renditions))
	for _, rendition := range renditions {
		out = append(out, rendition.SocialAccountID)
	}
	return out
}

func uniqueNonEmpty(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func mustJSON(value interface{}) string {
	if value == nil {
		return "{}"
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func publicationFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstContentLine(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

func formatOptionalTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func hasBlockingIssues(issues []capabilities.ValidationIssue) bool {
	for _, issue := range issues {
		if issue.Severity == "error" {
			return true
		}
	}
	return false
}

func actionMessage(message, jobID string) *ActionOutput {
	resp := &ActionOutput{}
	resp.Body.Message = message
	resp.Body.JobID = jobID
	return resp
}
