package platform

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const redditAuthBase = "https://www.reddit.com"
const redditAPIBase = "https://oauth.reddit.com"
const redditUserAgent = "OpenPost/1.0"

type RedditAdapter struct {
	clientID     string
	clientSecret string
	redirectURI  string
	userAgent    string
}

func NewRedditAdapter(clientID, clientSecret, redirectURI string) *RedditAdapter {
	return &RedditAdapter{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		userAgent:    redditUserAgent,
	}
}

func (r *RedditAdapter) redditHeaders(accessToken string) map[string]string {
	headers := map[string]string{
		"User-Agent": r.userAgent,
	}
	if accessToken != "" {
		if cookie := r.redditSessionCookie(accessToken); cookie != "" {
			headers["Cookie"] = "reddit_session=" + cookie
		} else {
			headers[headerAuthorization] = bearerPrefix + accessToken
		}
	}
	return headers
}

// redditSessionCookie extracts the session cookie from a cookie:-prefixed access token.
// The token format is "cookie:modhash:sessionvalue" or "cookie:sessionvalue".
func (r *RedditAdapter) redditSessionCookie(accessToken string) string {
	if !strings.HasPrefix(accessToken, "cookie:") {
		return ""
	}
	parts := strings.SplitN(strings.TrimPrefix(accessToken, "cookie:"), ":", 2)
	if len(parts) == 2 {
		return parts[1] // "modhash:sessionvalue" → sessionvalue is after second colon
	}
	return parts[0]
}

// redditModhash extracts the modhash from a cookie:-prefixed access token.
func (r *RedditAdapter) redditModhash(accessToken string) string {
	if !strings.HasPrefix(accessToken, "cookie:") {
		return ""
	}
	parts := strings.SplitN(strings.TrimPrefix(accessToken, "cookie:"), ":", 3)
	if len(parts) >= 2 && parts[1] != "" {
		return parts[0]
	}
	return ""
}

func (r *RedditAdapter) basicAuthHeader() string {
	auth := base64.StdEncoding.EncodeToString([]byte(r.clientID + ":" + r.clientSecret))
	return "Basic " + auth
}

// CreateSession validates a Reddit session cookie and returns the modhash for
// subsequent API calls. No OAuth app or username/password required — the user
// copies their reddit_session cookie from their browser's dev tools.
func (r *RedditAdapter) CreateSession(ctx context.Context, username, sessionCookie string) (string, string, int, error) {
	// Fetch the modhash from the profile endpoint
	respBody, err := DoRequest(ctx, http.MethodGet, redditAuthBase+"/api/me.json", nil, map[string]string{
		"User-Agent": r.userAgent,
		"Cookie":     "reddit_session=" + sessionCookie,
	})
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid reddit session cookie: %w", err)
	}

	var profile struct {
		Data struct {
			Name    string `json:"name"`
			Modhash string `json:"modhash"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &profile); err != nil {
		return "", "", 0, fmt.Errorf("decoding reddit profile: %w", err)
	}

	if profile.Data.Name == "" {
		return "", "", 0, fmt.Errorf("reddit session cookie did not return a valid user")
	}

	modhash := profile.Data.Modhash
	_ = username

	// Cookies are long-lived (up to 1 year).
	return modhash, sessionCookie, 365 * 24 * 3600, nil
}

// GenerateAuthURL builds the Reddit OAuth 2.0 authorization URL.
// duration=permanent is used so a refresh token is returned.
func (r *RedditAdapter) GenerateAuthURL(state string) (string, map[string]string) {
	params := map[string]string{
		"response_type":       oauthResponseType,
		oauthParamClientID:    r.clientID,
		oauthParamRedirectURI: r.redirectURI,
		"scope":               "identity submit read mysubreddits",
		"state":               state,
		"duration":            "permanent",
	}
	return redditAuthBase + "/api/v1/authorize?" + encodeRedditAuthQuery(params), nil
}

func encodeRedditAuthQuery(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return values.Encode()
}

// ExchangeCode exchanges an authorization code for access + refresh tokens.
func (r *RedditAdapter) ExchangeCode(ctx context.Context, code string, _ map[string]string) (*TokenResult, error) {
	values := map[string]string{
		grantType:             oauthGrantAuthCode,
		oauthParamCode:        code,
		oauthParamRedirectURI: r.redirectURI,
	}

	headers := map[string]string{
		headerAuthorization: r.basicAuthHeader(),
		"User-Agent":        r.userAgent,
	}

	respBody, err := DoFormURLEncoded(ctx, http.MethodPost, redditAuthBase+"/api/v1/access_token", values, headers)
	if err != nil {
		return nil, fmt.Errorf("reddit token exchange: %w", err)
	}

	return parseRedditTokenResponse(respBody)
}

func (r *RedditAdapter) RefreshCapability() RefreshCapability {
	return RefreshCapability{
		Supported:        true,
		CredentialSource: RefreshCredentialRefreshToken,
	}
}

func (r *RedditAdapter) RefreshToken(ctx context.Context, input RefreshTokenInput) (*TokenResult, error) {
	if input.RefreshToken == "" {
		return nil, fmt.Errorf("reddit refresh requires a refresh token")
	}

	values := map[string]string{
		grantType:                             oauthGrantRefresh,
		string(RefreshCredentialRefreshToken): input.RefreshToken,
	}

	headers := map[string]string{
		headerAuthorization: r.basicAuthHeader(),
		"User-Agent":        r.userAgent,
	}

	respBody, err := DoFormURLEncoded(ctx, http.MethodPost, redditAuthBase+"/api/v1/access_token", values, headers)
	if err != nil {
		return nil, fmt.Errorf("reddit token refresh: %w", err)
	}

	return parseRedditTokenResponse(respBody)
}

func parseRedditTokenResponse(body []byte) (*TokenResult, error) {
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("decoding reddit token: %w", err)
	}

	return &TokenResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		TokenType:    tokenResp.TokenType,
		Extra: map[string]string{
			"scope": tokenResp.Scope,
		},
	}, nil
}

// GetProfile fetches the authenticated Reddit user.
func (r *RedditAdapter) GetProfile(ctx context.Context, accessToken string) (*UserProfile, error) {
	endpoint := redditAPIBase + "/api/v1/me"
	if strings.HasPrefix(accessToken, "cookie:") {
		endpoint = redditAuthBase + "/api/me.json"
	}

	respBody, err := DoRequest(ctx, http.MethodGet, endpoint, nil, r.redditHeaders(accessToken))
	if err != nil {
		return nil, fmt.Errorf("reddit profile: %w", err)
	}

	if strings.HasPrefix(accessToken, "cookie:") {
		return parseRedditCookieProfile(respBody)
	}

	var profile struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		IconImg    string `json:"icon_img"`
		TotalKarma int    `json:"total_karma"`
		Subreddit  struct {
			Title string `json:"title"`
		} `json:"subreddit"`
	}
	if err := json.Unmarshal(respBody, &profile); err != nil {
		return nil, fmt.Errorf("decoding reddit profile: %w", err)
	}

	displayName := profile.Name
	if profile.Subreddit.Title != "" {
		displayName = profile.Subreddit.Title
	}

	return &UserProfile{
		ID:          profile.ID,
		Username:    profile.Name,
		DisplayName: displayName,
		CapabilityState: map[string]string{
			"reddit_user_id":  profile.ID,
			"reddit_username": profile.Name,
			"connection_type": "oauth",
		},
	}, nil
}

func parseRedditCookieProfile(body []byte) (*UserProfile, error) {
	var resp struct {
		Data struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			IconImg string `json:"icon_img"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decoding reddit cookie profile: %w", err)
	}
	return &UserProfile{
		ID:          resp.Data.ID,
		Username:    resp.Data.Name,
		DisplayName: resp.Data.Name,
		CapabilityState: map[string]string{
			"reddit_user_id":  resp.Data.ID,
			"reddit_username": resp.Data.Name,
			"connection_type": "cookie",
		},
	}, nil
}

// ListAccountSelections returns the user's subscribed subreddits as publishable destinations.
func (r *RedditAdapter) ListAccountSelections(ctx context.Context, token *TokenResult) ([]AccountSelectionOption, error) {
	if token == nil || strings.TrimSpace(token.AccessToken) == "" {
		return nil, fmt.Errorf("reddit access token is required")
	}

	profile, err := r.GetProfile(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}

	// Fetch subscribed subreddits
	subreddits, err := r.listSubscribedSubreddits(ctx, token.AccessToken)
	if err != nil {
		// Subreddit listing can fail; fall back to just the user profile
		return []AccountSelectionOption{
			{
				ID:          "u_" + profile.Username,
				Username:    profile.Username,
				DisplayName: "u/" + profile.Username,
				Kind:        "User profile",
				Description: "Post to your Reddit profile.",
				AvatarURL:   token.Extra["icon_img"],
			},
		}, nil
	}

	options := make([]AccountSelectionOption, 0, 1+len(subreddits))
	options = append(options, AccountSelectionOption{
		ID:          "u_" + profile.Username,
		Username:    profile.Username,
		DisplayName: "u/" + profile.Username,
		Kind:        "User profile",
		Description: "Post to your Reddit profile.",
	})

	for _, sr := range subreddits {
		options = append(options, AccountSelectionOption{
			ID:          sr.ID,
			Username:    sr.DisplayName,
			DisplayName: "r/" + sr.DisplayName,
			Kind:        "Subreddit",
			Description: firstNonEmptyString(sr.Title, sr.PublicDescription, "Subreddit"),
		})
	}

	return options, nil
}

type redditSubreddit struct {
	ID                string `json:"id"`
	DisplayName       string `json:"display_name"`
	Title             string `json:"title"`
	PublicDescription string `json:"public_description"`
	Subscribers       int    `json:"subscribers"`
	Over18            bool   `json:"over_18"`
}

func (r *RedditAdapter) listSubscribedSubreddits(ctx context.Context, accessToken string) ([]redditSubreddit, error) {
	endpoint := r.apiBase(accessToken) + "/subreddits/mine/subscriber?limit=100"

	respBody, err := DoRequest(ctx, http.MethodGet, endpoint, nil, r.redditHeaders(accessToken))
	if err != nil {
		return nil, fmt.Errorf("reddit subscribed subreddits: %w", err)
	}

	var result struct {
		Data struct {
			Children []struct {
				Data redditSubreddit `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decoding reddit subreddits: %w", err)
	}

	subreddits := make([]redditSubreddit, 0, len(result.Data.Children))
	for _, child := range result.Data.Children {
		subreddits = append(subreddits, child.Data)
	}

	return subreddits, nil
}

// SelectAccount resolves a subreddit selection ID to a SelectedAccount.
func (r *RedditAdapter) SelectAccount(ctx context.Context, token *TokenResult, selectionID string) (*SelectedAccount, error) {
	selectionID = strings.TrimSpace(selectionID)
	options, err := r.ListAccountSelections(ctx, token)
	if err != nil {
		return nil, err
	}

	for _, option := range options {
		if option.ID != selectionID {
			continue
		}

		displayName := firstNonEmptyString(option.Username, option.DisplayName)

		return &SelectedAccount{
			AccountID:        selectionID,
			AccountUsername:  displayName,
			AccountAvatarURL: option.AvatarURL,
			Token:            token,
			CapabilityState: map[string]string{
				"reddit_destination_type": option.Kind,
				"reddit_destination_id":   selectionID,
			},
		}, nil
	}

	return nil, fmt.Errorf("the selected Reddit destination is no longer available")
}

// UploadMedia uploads an image or video to Reddit and returns the asset ID (as a JSON-encoded string).
func (r *RedditAdapter) UploadMedia(ctx context.Context, accessToken, accountID, mimeType string, reader io.Reader) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("reading media: %w", err)
	}

	isVideo := isVideoMime(mimeType)
	if isVideo {
		return r.uploadVideo(ctx, accessToken, mimeType, data)
	}
	return r.uploadImage(ctx, accessToken, mimeType, data)
}

func (r *RedditAdapter) uploadImage(ctx context.Context, accessToken, mimeType string, data []byte) (string, error) {
	// Build multipart upload
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	ext := ".jpg"
	switch {
	case strings.Contains(mimeType, "png"):
		ext = ".png"
	case strings.Contains(mimeType, "webp"):
		ext = ".webp"
	case strings.Contains(mimeType, "gif"):
		ext = ".gif"
	}

	part, err := writer.CreateFormFile("file", "image"+ext)
	if err != nil {
		return "", fmt.Errorf("creating form file: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return "", fmt.Errorf("writing image data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("closing multipart: %w", err)
	}

	headers := r.redditHeaders(accessToken)
	headers[headerContentType] = writer.FormDataContentType()

	respBody, err := DoRequest(ctx, http.MethodPost, r.apiBase(accessToken)+"/api/media/asset.json", &buf, headers)
	if err != nil {
		return "", fmt.Errorf("reddit media upload: %w", err)
	}

	var result struct {
		Args struct {
			Action string `json:"action"`
			Fields []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"fields"`
			WebsocketURL string `json:"websocket_url"`
		} `json:"args"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("decoding reddit media upload: %w", err)
	}

	// Extract asset ID from fields
	assetID := ""
	for _, field := range result.Args.Fields {
		if field.Name == "asset_id" || field.Name == "id" {
			assetID = field.Value
			break
		}
	}

	if assetID == "" {
		// Fall back to websocket polling if needed
		if result.Args.WebsocketURL != "" {
			assetID, err = r.pollRedditMediaWebsocket(ctx, accessToken, result.Args.WebsocketURL)
			if err != nil {
				return "", err
			}
		}
	}

	if assetID == "" {
		return "", fmt.Errorf("reddit media upload did not return an asset ID")
	}

	// Return as JSON so we can store both the asset ID and websocket URL
	mediaInfo := map[string]string{
		"asset_id": assetID,
		"kind":     "image",
	}
	encoded, _ := json.Marshal(mediaInfo)
	return string(encoded), nil
}

func (r *RedditAdapter) uploadVideo(ctx context.Context, accessToken, mimeType string, data []byte) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	ext := ".mp4"
	if strings.Contains(mimeType, "quicktime") {
		ext = ".mov"
	}

	part, err := writer.CreateFormFile("file", "video"+ext)
	if err != nil {
		return "", fmt.Errorf("creating form file: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return "", fmt.Errorf("writing video data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("closing multipart: %w", err)
	}

	headers := r.redditHeaders(accessToken)
	headers[headerContentType] = writer.FormDataContentType()

	respBody, err := DoRequest(ctx, http.MethodPost, r.apiBase(accessToken)+"/api/media/asset.json", &buf, headers)
	if err != nil {
		return "", fmt.Errorf("reddit video upload: %w", err)
	}

	var result struct {
		Args struct {
			Fields []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"fields"`
			WebsocketURL string `json:"websocket_url"`
		} `json:"args"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("decoding reddit video upload: %w", err)
	}

	assetID := ""
	for _, field := range result.Args.Fields {
		if field.Name == "asset_id" || field.Name == "id" {
			assetID = field.Value
			break
		}
	}

	if assetID == "" && result.Args.WebsocketURL != "" {
		assetID, err = r.pollRedditMediaWebsocket(ctx, accessToken, result.Args.WebsocketURL)
		if err != nil {
			return "", err
		}
	}

	if assetID == "" {
		return "", fmt.Errorf("reddit video upload did not return an asset ID")
	}

	pollErr := r.pollRedditMediaProcessing(ctx, accessToken, assetID)
	if pollErr != nil {
		return "", pollErr
	}

	mediaInfo := map[string]string{
		"asset_id": assetID,
		"kind":     "video",
	}
	encoded, _ := json.Marshal(mediaInfo)
	return string(encoded), nil
}

func (r *RedditAdapter) pollRedditMediaWebsocket(ctx context.Context, accessToken, websocketURL string) (string, error) {
	// Reddit media upload returns a websocket URL for processing status.
	// We poll the REST equivalent: GET /api/media/asset.json?asset_id=...
	// For now, we return an error if no asset ID was immediately available.
	return "", fmt.Errorf("reddit media processing requires polling; asset ID not immediately available")
}

func (r *RedditAdapter) pollRedditMediaProcessing(ctx context.Context, accessToken, assetID string) error {
	// Poll for video processing to complete. Reddit returns is_mp4_convertable and processing status.
	for i := 0; i < 60; i++ {
		endpoint := r.apiBase(accessToken) + "/api/media/asset.json?asset_id=" + url.QueryEscape(assetID)
		respBody, err := DoRequest(ctx, http.MethodGet, endpoint, nil, r.redditHeaders(accessToken))
		if err != nil {
			return fmt.Errorf("reddit media status: %w", err)
		}

		var status struct {
			Asset struct {
				AssetID          string `json:"asset_id"`
				ProcessingStatus string `json:"processing_status"`
				IsMP4Convertable bool   `json:"is_mp4_convertable"`
			} `json:"asset"`
		}
		if err := json.Unmarshal(respBody, &status); err != nil {
			return fmt.Errorf("decoding reddit media status: %w", err)
		}

		switch status.Asset.ProcessingStatus {
		case "completed", "success":
			return nil
		case "failed":
			return fmt.Errorf("reddit media processing failed for asset %s", assetID)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	return fmt.Errorf("reddit media processing timed out for asset %s", assetID)
}

// Publish creates a post or comment on Reddit.
func (r *RedditAdapter) Publish(ctx context.Context, accessToken, accountID string, req *PublishRequest) (string, error) {
	if req.ReplyToID != "" {
		return r.postComment(ctx, accessToken, req.ReplyToID, req.Content)
	}

	return r.submitPost(ctx, accessToken, accountID, req)
}

func (r *RedditAdapter) apiBase(accessToken string) string {
	if strings.HasPrefix(accessToken, "cookie:") {
		return redditAuthBase
	}
	return redditAPIBase
}

func (r *RedditAdapter) submitPost(ctx context.Context, accessToken, accountID string, req *PublishRequest) (string, error) {
	// Determine the subreddit / destination
	sr := r.resolveSubreddit(accountID, req.Settings)
	if sr == "" {
		return "", fmt.Errorf("reddit requires a subreddit destination")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		// Fall back to first line of content if no title set
		title = req.Content
		if idx := strings.IndexByte(title, '\n'); idx > 0 {
			title = title[:idx]
		}
	}
	if len(title) > 300 {
		title = title[:297] + "..."
	}
	if title == "" {
		return "", fmt.Errorf("reddit posts require a title")
	}

	values := url.Values{}
	values.Set("sr", sr)
	values.Set("title", title)
	values.Set("api_type", "json")

	// Include modhash for cookie-based auth
	if modhash := r.redditModhash(accessToken); modhash != "" {
		values.Set("uh", modhash)
	}

	if val := settingString(req.Settings, "flair_id"); val != "" {
		values.Set("flair_id", val)
	}
	if val := settingString(req.Settings, "flair_text"); val != "" {
		values.Set("flair_text", val)
	}
	if settingBool(req.Settings, "nsfw") {
		values.Set("nsfw", "true")
	}
	if settingBool(req.Settings, "spoiler") {
		values.Set("spoiler", "true")
	}
	if !settingBoolDefault(req.Settings, "sendreplies", true) {
		values.Set("sendreplies", "false")
	}

	// Determine post kind based on content profile and media
	hasMedia := len(req.PlatformMediaIDs) > 0
	linkURL := firstNonEmptyString(settingString(req.Settings, "url"), settingString(req.Settings, "link_url"))

	switch {
	case hasMedia:
		values.Set("kind", "image")
		// Extract asset ID from stored JSON
		assetID := r.extractAssetID(req.PlatformMediaIDs[0])
		if assetID != "" {
			values.Set("asset_id", assetID)
		}
		if req.Content != "" {
			values.Set("text", req.Content)
		}
	case linkURL != "":
		values.Set("kind", "link")
		values.Set("url", linkURL)
		if req.Content != "" && req.Content != linkURL {
			values.Set("text", req.Content)
		}
	default:
		values.Set("kind", "self")
		values.Set("text", req.Content)
	}

	headers := r.redditHeaders(accessToken)
	headers[headerContentType] = contentTypeForm

	respBody, err := DoRequest(ctx, http.MethodPost, r.apiBase(accessToken)+"/api/submit", strings.NewReader(values.Encode()), headers)
	if err != nil {
		return "", fmt.Errorf("posting to reddit: %w", err)
	}

	// Reddit returns JSON with the post info
	var submitResp struct {
		JSON struct {
			Errors [][]string `json:"errors"`
			Data   struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"data"`
		} `json:"json"`
	}
	if err := json.Unmarshal(respBody, &submitResp); err != nil {
		return "", fmt.Errorf("decoding reddit submit response: %w", err)
	}

	if len(submitResp.JSON.Errors) > 0 && len(submitResp.JSON.Errors[0]) > 0 {
		return "", fmt.Errorf("reddit error: %s", submitResp.JSON.Errors[0][1])
	}

	// Return the full name (t3_xxxxx) as external ID
	name := submitResp.JSON.Data.Name
	if name == "" {
		name = submitResp.JSON.Data.ID
	}

	// Encode extra info as JSON for threading/comments
	externalID, _ := json.Marshal(map[string]string{
		"name": name,
		"id":   submitResp.JSON.Data.ID,
		"url":  submitResp.JSON.Data.URL,
		"kind": values.Get("kind"),
	})
	return string(externalID), nil
}

func (r *RedditAdapter) resolveSubreddit(accountID string, settings map[string]interface{}) string {
	// First check for explicit subreddit in settings
	if sr := settingString(settings, "subreddit"); sr != "" {
		return strings.TrimPrefix(sr, "r/")
	}

	// Fall back to account ID (strip u_ prefix for profile posts)
	if strings.HasPrefix(accountID, "u_") {
		return strings.TrimPrefix(accountID, "u_")
	}

	return accountID
}

func (r *RedditAdapter) extractAssetID(mediaID string) string {
	var info map[string]string
	if err := json.Unmarshal([]byte(mediaID), &info); err != nil {
		return mediaID // might be raw asset ID
	}
	return info["asset_id"]
}

func (r *RedditAdapter) postComment(ctx context.Context, accessToken, parentID, content string) (string, error) {
	var parent map[string]string
	if err := json.Unmarshal([]byte(parentID), &parent); err != nil {
		// Fall back: parentID might be a raw Reddit fullname (t1_xxxxx)
		parent = map[string]string{"name": parentID}
	}

	thingID := firstNonEmptyString(parent["name"], parent["id"], parentID)

	values := url.Values{}
	values.Set("thing_id", thingID)
	values.Set("text", content)
	values.Set("api_type", "json")
	if modhash := r.redditModhash(accessToken); modhash != "" {
		values.Set("uh", modhash)
	}

	headers := r.redditHeaders(accessToken)
	headers[headerContentType] = contentTypeForm

	respBody, err := DoRequest(ctx, http.MethodPost, r.apiBase(accessToken)+"/api/comment", strings.NewReader(values.Encode()), headers)
	if err != nil {
		return "", fmt.Errorf("posting reddit comment: %w", err)
	}

	var commentResp struct {
		JSON struct {
			Errors [][]string `json:"errors"`
			Data   struct {
				Things []struct {
					Data struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"data"`
				} `json:"things"`
			} `json:"data"`
		} `json:"json"`
	}
	if err := json.Unmarshal(respBody, &commentResp); err != nil {
		return "", fmt.Errorf("decoding reddit comment response: %w", err)
	}

	if len(commentResp.JSON.Errors) > 0 && len(commentResp.JSON.Errors[0]) > 0 {
		return "", fmt.Errorf("reddit comment error: %s", commentResp.JSON.Errors[0][1])
	}

	commentID := ""
	if len(commentResp.JSON.Data.Things) > 0 {
		commentID = commentResp.JSON.Data.Things[0].Data.Name
	}

	externalID, _ := json.Marshal(map[string]string{
		"name":   commentID,
		"parent": thingID,
	})
	return string(externalID), nil
}

// ListComments fetches comments on a Reddit post.
func (r *RedditAdapter) ListComments(ctx context.Context, accessToken, accountID string, externalID string) ([]Comment, error) {
	var postInfo map[string]string
	json.Unmarshal([]byte(externalID), &postInfo)

	postID := firstNonEmptyString(postInfo["id"], externalID)
	if strings.HasPrefix(postID, "t3_") {
		postID = strings.TrimPrefix(postID, "t3_")
	}

	// Extract subreddit from accountID or post info
	sr := ""
	if strings.HasPrefix(accountID, "u_") {
		sr = strings.TrimPrefix(accountID, "u_")
	} else {
		sr = accountID
	}

	endpoint := r.apiBase(accessToken) + "/r/" + sr + "/comments/" + postID + ".json?limit=100"
	respBody, err := DoRequest(ctx, http.MethodGet, endpoint, nil, r.redditHeaders(accessToken))
	if err != nil {
		return nil, fmt.Errorf("reddit comments: %w", err)
	}

	// Reddit returns an array: [post_data, comments_data]
	var listing []struct {
		Data struct {
			Children []struct {
				Kind string          `json:"kind"`
				Data json.RawMessage `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &listing); err != nil {
		return nil, fmt.Errorf("decoding reddit comments: %w", err)
	}

	comments := make([]Comment, 0)

	// Skip the first element (post data), parse comments from the second
	if len(listing) > 1 {
		for _, child := range listing[1].Data.Children {
			if child.Kind != "t1" {
				continue
			}

			var commentData struct {
				ID       string      `json:"id"`
				Name     string      `json:"name"`
				Author   string      `json:"author"`
				Body     string      `json:"body"`
				Created  float64     `json:"created_utc"`
				Edited   interface{} `json:"edited"`
				LinkID   string      `json:"link_id"`
				ParentID string      `json:"parent_id"`
				Score    int         `json:"score"`
			}
			if err := json.Unmarshal(child.Data, &commentData); err != nil {
				continue
			}

			if commentData.Body == "" {
				continue
			}

			createdAt := ""
			if commentData.Created > 0 {
				createdAt = time.Unix(int64(commentData.Created), 0).UTC().Format(time.RFC3339)
			}

			comments = append(comments, Comment{
				ID:             commentData.Name,
				ParentID:       commentData.ParentID,
				ConversationID: commentData.LinkID,
				AuthorID:       commentData.Author,
				Text:           commentData.Body,
				CreatedAt:      createdAt,
				Deleted:        commentData.Body == "[deleted]" || commentData.Body == "[removed]",
				CanReply:       true,
				CanDelete:      commentData.Author == accountID || strings.TrimPrefix(accountID, "u_") == commentData.Author,
				CanLike:        true,
				CanUnlike:      true,
				Liked:          false,
				LikeStateKnown: false,
			})
		}
	}

	return comments, nil
}

// ReplyToComment posts a reply to an existing comment.
func (r *RedditAdapter) ReplyToComment(ctx context.Context, accessToken, accountID, commentID, message string) (string, error) {
	return r.postComment(ctx, accessToken, commentID, message)
}

// HideComment is not supported by the Reddit API in this form; use DeleteComment instead.
func (r *RedditAdapter) HideComment(context.Context, string, string, string) error {
	return fmt.Errorf("reddit hide comment: %w", ErrUnsupportedCommentAction)
}

// DeleteComment deletes a Reddit comment.
func (r *RedditAdapter) DeleteComment(ctx context.Context, accessToken, accountID, commentID string) error {
	var commentInfo map[string]string
	json.Unmarshal([]byte(commentID), &commentInfo)

	thingID := firstNonEmptyString(commentInfo["name"], commentInfo["id"], commentID)

	values := url.Values{}
	values.Set("id", thingID)
	if modhash := r.redditModhash(accessToken); modhash != "" {
		values.Set("uh", modhash)
	}

	headers := r.redditHeaders(accessToken)
	headers[headerContentType] = contentTypeForm

	_, err := DoRequest(ctx, http.MethodPost, r.apiBase(accessToken)+"/api/del", strings.NewReader(values.Encode()), headers)
	if err != nil {
		return fmt.Errorf("deleting reddit comment: %w", err)
	}

	return nil
}

// AnalyticsSupport reports what analytics Reddit provides.
func (r *RedditAdapter) AnalyticsSupport() AnalyticsSupport {
	return AnalyticsSupport{
		Account: true,
		Content: true,
	}
}

// FetchAccountAnalytics returns account-level Reddit metrics.
func (r *RedditAdapter) FetchAccountAnalytics(ctx context.Context, accessToken string, input AccountAnalyticsRequest) (AnalyticsValues, error) {
	if !strings.HasPrefix(input.AccountID, "u_") && input.AccountID != "" {
		return r.fetchSubredditAnalytics(ctx, accessToken, input.AccountID)
	}

	endpoint := redditAPIBase + "/api/v1/me"
	if strings.HasPrefix(accessToken, "cookie:") {
		endpoint = redditAuthBase + "/api/me.json"
	}
	respBody, err := DoRequest(ctx, http.MethodGet, endpoint, nil, r.redditHeaders(accessToken))
	if err != nil {
		return nil, fmt.Errorf("reddit account analytics: %w", err)
	}

	var profile struct {
		Subreddit struct {
			Subscribers *int64 `json:"subscribers"`
		} `json:"subreddit"`
	}
	if strings.HasPrefix(accessToken, "cookie:") {
		var wrapped struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(respBody, &wrapped); err != nil {
			return nil, fmt.Errorf("decoding reddit account analytics: %w", err)
		}
		respBody = wrapped.Data
	}
	if err := json.Unmarshal(respBody, &profile); err != nil {
		return nil, fmt.Errorf("decoding reddit account analytics: %w", err)
	}

	values := AnalyticsValues{}
	addOptionalMetric(values, MetricFollowers, profile.Subreddit.Subscribers)
	return values, nil
}

func (r *RedditAdapter) fetchSubredditAnalytics(ctx context.Context, accessToken, accountID string) (AnalyticsValues, error) {
	fullname := "t5_" + strings.TrimPrefix(accountID, "t5_")
	endpoint := r.apiBase(accessToken) + "/api/info.json?id=" + url.QueryEscape(fullname)
	respBody, err := DoRequest(ctx, http.MethodGet, endpoint, nil, r.redditHeaders(accessToken))
	if err != nil {
		return nil, fmt.Errorf("reddit account analytics: %w", err)
	}
	var listing struct {
		Data struct {
			Children []struct {
				Data struct {
					Subscribers *int64 `json:"subscribers"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &listing); err != nil {
		return nil, fmt.Errorf("decoding reddit account analytics: %w", err)
	}
	if len(listing.Data.Children) == 0 {
		return nil, NewAnalyticsError(AnalyticsStatusNotFound, "account_not_found")
	}
	values := AnalyticsValues{}
	addOptionalMetric(values, MetricFollowers, listing.Data.Children[0].Data.Subscribers)
	return values, nil
}

// FetchContentAnalytics returns per-post Reddit metrics.
func (r *RedditAdapter) FetchContentAnalytics(ctx context.Context, accessToken string, input ContentAnalyticsRequest) (AnalyticsValues, error) {
	if len(input.ExternalIDs) == 0 {
		return nil, NewAnalyticsError(AnalyticsStatusNotFound, "missing_external_id")
	}

	fullnames := make([]string, 0, len(input.ExternalIDs))
	for _, externalID := range uniqueNonEmpty(input.ExternalIDs) {
		var postInfo map[string]string
		_ = json.Unmarshal([]byte(externalID), &postInfo)
		postID := strings.TrimSpace(firstNonEmptyString(postInfo["id"], externalID))
		if postID == "" {
			continue
		}
		fullnames = append(fullnames, "t3_"+strings.TrimPrefix(postID, "t3_"))
	}
	if len(fullnames) == 0 {
		return nil, NewAnalyticsError(AnalyticsStatusNotFound, "missing_external_id")
	}

	endpoint := r.apiBase(accessToken) + "/api/info.json?id=" + url.QueryEscape(strings.Join(fullnames, ","))
	respBody, err := DoRequest(ctx, http.MethodGet, endpoint, nil, r.redditHeaders(accessToken))
	if err != nil {
		return nil, fmt.Errorf("reddit content analytics: %w", err)
	}

	var listing struct {
		Data struct {
			Children []struct {
				Data struct {
					Score       *int64 `json:"score"`
					NumComments *int64 `json:"num_comments"`
					Views       *int64 `json:"view_count"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &listing); err != nil {
		return nil, fmt.Errorf("decoding reddit post analytics: %w", err)
	}

	if len(listing.Data.Children) == 0 {
		return nil, NewAnalyticsError(AnalyticsStatusNotFound, "")
	}

	values := AnalyticsValues{}
	for _, child := range listing.Data.Children {
		addOptionalMetric(values, MetricLikes, child.Data.Score)
		addOptionalMetric(values, MetricComments, child.Data.NumComments)
		addOptionalMetric(values, MetricViews, child.Data.Views)
	}
	return values, nil
}

// validateRedditMedia validates media constraints for Reddit.
func validateRedditMedia(media []MediaItem) []MediaValidationIssue {
	if len(media) == 0 {
		return nil
	}

	if len(media) > 1 {
		return []MediaValidationIssue{{
			Provider: providerReddit,
			Severity: severityError,
			Message:  "Reddit supports one attachment per post.",
		}}
	}

	item := media[0]
	if isVideoMime(item.MimeType) {
		if item.Size > 1*1024*1024*1024 {
			return []MediaValidationIssue{{
				Provider: providerReddit,
				MediaID:  item.ID,
				Severity: severityError,
				Message:  "Reddit video must be under 1 GB.",
			}}
		}
		if item.DurationMS > 15*60*1000 {
			return []MediaValidationIssue{{
				Provider: providerReddit,
				MediaID:  item.ID,
				Severity: severityError,
				Message:  "Reddit video must be under 15 minutes.",
			}}
		}
		return nil
	}

	// Image constraints
	maxImageBytes := int64(20 * 1024 * 1024) // 20 MB for images
	if item.Size > maxImageBytes {
		return []MediaValidationIssue{{
			Provider: providerReddit,
			MediaID:  item.ID,
			Severity: severityError,
			Message:  "Reddit images must be under 20 MB.",
		}}
	}

	return nil
}
