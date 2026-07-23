package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/platform"
	"github.com/stretchr/testify/require"
)

type destinationOptionsTestAdapter struct {
	platform.Adapter
	input platform.DestinationOptionsInput
	token string
}

func (a *destinationOptionsTestAdapter) ListDestinationOptions(_ context.Context, accessToken string, input platform.DestinationOptionsInput) (map[string][]platform.DestinationOption, error) {
	a.token = accessToken
	a.input = input
	return map[string][]platform.DestinationOption{
		"youtube_playlists": {{Value: "playlist-1", Label: "Product videos"}},
	}, nil
}

type destinationOptionsTokenSource struct {
	accountID string
}

func (s *destinationOptionsTokenSource) GetValidAccessToken(_ context.Context, accountID string) (string, error) {
	s.accountID = accountID
	return "valid-access-token", nil
}

func TestDestinationOptionsUsesConnectedAccountAndFreshToken(t *testing.T) {
	db := createHandlerTestDB(t, (*models.WorkspaceMember)(nil), (*models.SocialAccount)(nil))
	ctx := context.Background()
	_, err := db.NewInsert().Model(&models.WorkspaceMember{
		WorkspaceID: "ws-1",
		UserID:      "user-1",
		Role:        models.WorkspaceRoleAdmin,
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.SocialAccount{
		ID:             "youtube-1",
		WorkspaceID:    "ws-1",
		Slug:           "youtube-main",
		Platform:       "youtube",
		AccountID:      "channel-1",
		AccessTokenEnc: []byte("encrypted"),
		IsActive:       true,
	}).Exec(ctx)
	require.NoError(t, err)

	adapter := &destinationOptionsTestAdapter{}
	tokenSource := &destinationOptionsTokenSource{}
	e := echo.New()
	api := humaecho.NewWithGroup(e, e.Group("/api/v1"), huma.DefaultConfig("Test", "1.0.0"))
	NewDestinationOptionsHandler(db, testAuthenticator{}, map[string]platform.Adapter{
		"youtube": adapter,
	}, tokenSource).RegisterRoutes(api)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/accounts/youtube-1/destination-options?region_code=PT&language=pt", nil)
	req.Header.Set("Authorization", "Bearer web-token")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var output struct {
		Options map[string][]platform.DestinationOption `json:"options"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &output))
	require.Equal(t, []platform.DestinationOption{{Value: "playlist-1", Label: "Product videos"}}, output.Options["youtube_playlists"])
	require.Equal(t, "youtube-1", tokenSource.accountID)
	require.Equal(t, "valid-access-token", adapter.token)
	require.Equal(t, platform.DestinationOptionsInput{RegionCode: "PT", Language: "pt"}, adapter.input)
}
