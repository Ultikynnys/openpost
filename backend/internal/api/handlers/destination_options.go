package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/openpost/backend/internal/api/middleware"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/platform"
	"github.com/uptrace/bun"
)

type AccessTokenSource interface {
	GetValidAccessToken(ctx context.Context, accountID string) (string, error)
}

type DestinationOptionsHandler struct {
	db          *bun.DB
	auth        middleware.Authenticator
	providers   map[string]platform.Adapter
	tokenSource AccessTokenSource
}

func NewDestinationOptionsHandler(db *bun.DB, auth middleware.Authenticator, providers map[string]platform.Adapter, tokenSource AccessTokenSource) *DestinationOptionsHandler {
	return &DestinationOptionsHandler{
		db:          db,
		auth:        auth,
		providers:   providers,
		tokenSource: tokenSource,
	}
}

type DestinationOptionsInput struct {
	AccountID  string `path:"account_id" doc:"Connected social account ID"`
	RegionCode string `query:"region_code" default:"US" doc:"ISO 3166-1 alpha-2 region code"`
	Language   string `query:"language" default:"en" doc:"Language code for provider labels"`
}

type DestinationOptionsOutput struct {
	Body struct {
		Options map[string][]platform.DestinationOption `json:"options"`
	}
}

func (h *DestinationOptionsHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-account-destination-options",
		Method:      http.MethodGet,
		Path:        "/accounts/{account_id}/destination-options",
		Summary:     "List account-specific publishing options",
		Tags:        []string{tagAccounts},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{400, 403, 404, 502},
	}, func(ctx context.Context, input *DestinationOptionsInput) (*DestinationOptionsOutput, error) {
		var account models.SocialAccount
		if err := h.db.NewSelect().
			Model(&account).
			Where("id = ? AND is_active = ?", input.AccountID, true).
			Scan(ctx); err != nil {
			return nil, huma.Error404NotFound("connected account not found")
		}
		if err := providerReadinessWorkspaceAccess(ctx, h.db, account.WorkspaceID, middleware.GetUserID(ctx)); err != nil {
			return nil, err
		}

		adapter, ok := h.providers[account.Platform]
		if !ok || adapter == nil {
			return nil, huma.Error400BadRequest("provider is not configured")
		}
		optionProvider, ok := adapter.(platform.DestinationOptionsProvider)
		if !ok {
			return nil, huma.Error400BadRequest("provider does not expose account-specific publishing options")
		}
		if h.tokenSource == nil {
			return nil, huma.Error502BadGateway("provider access token service is unavailable")
		}
		accessToken, err := h.tokenSource.GetValidAccessToken(ctx, account.ID)
		if err != nil {
			return nil, huma.Error502BadGateway("failed to authorize provider options")
		}
		options, err := optionProvider.ListDestinationOptions(ctx, accessToken, platform.DestinationOptionsInput{
			RegionCode: input.RegionCode,
			Language:   input.Language,
		})
		if err != nil {
			return nil, huma.Error502BadGateway("failed to load provider publishing options")
		}

		output := &DestinationOptionsOutput{}
		output.Body.Options = options
		return output, nil
	})
}
