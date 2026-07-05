package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	googleOAuthURL          = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL          = "https://oauth2.googleapis.com/token"
	googleUserInfoURL       = "https://www.googleapis.com/oauth2/v2/userinfo"
	youtubeAPIBaseURL       = "https://www.googleapis.com/youtube/v3"
	youtubeUploadBaseURL    = "https://www.googleapis.com/upload/youtube/v3"
	youtubeDefaultVideoName = "OpenPost video"
	youtubeTitleMaxRunes    = 100
)

type YouTubeAdapter struct {
	clientID     string
	clientSecret string
	redirectURI  string
}

func NewYouTubeAdapter(clientID, clientSecret, redirectURI string) *YouTubeAdapter {
	return &YouTubeAdapter{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
}

func (y *YouTubeAdapter) GenerateAuthURL(state string) (string, map[string]string) {
	params := url.Values{}
	params.Set(oauthParamClientID, y.clientID)
	params.Set(oauthParamRedirectURI, y.redirectURI)
	params.Set("response_type", oauthResponseType)
	params.Set("scope", strings.Join(youtubeScopes(), " "))
	params.Set("state", state)
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")
	params.Set("include_granted_scopes", "true")
	return googleOAuthURL + "?" + params.Encode(), nil
}

func (y *YouTubeAdapter) ExchangeCode(ctx context.Context, code string, _ map[string]string) (*TokenResult, error) {
	values := map[string]string{
		oauthParamClientID:     y.clientID,
		oauthParamClientSecret: y.clientSecret,
		oauthParamCode:         code,
		oauthParamRedirectURI:  y.redirectURI,
		grantType:              oauthGrantAuthCode,
	}
	return y.exchangeToken(ctx, values, "youtube token exchange")
}

func (y *YouTubeAdapter) RefreshCapability() RefreshCapability {
	return RefreshCapability{
		Supported:        true,
		CredentialSource: RefreshCredentialRefreshToken,
	}
}

func (y *YouTubeAdapter) RefreshToken(ctx context.Context, input RefreshTokenInput) (*TokenResult, error) {
	if input.RefreshToken == "" {
		return nil, fmt.Errorf("youtube refresh requires a refresh token")
	}
	values := map[string]string{
		oauthParamClientID:                    y.clientID,
		oauthParamClientSecret:                y.clientSecret,
		grantType:                             oauthGrantRefresh,
		string(RefreshCredentialRefreshToken): input.RefreshToken,
	}
	return y.exchangeToken(ctx, values, "youtube token refresh")
}

func (y *YouTubeAdapter) exchangeToken(ctx context.Context, values map[string]string, label string) (*TokenResult, error) {
	respBody, err := DoFormURLEncoded(ctx, http.MethodPost, googleTokenURL, values, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
		Description  string `json:"error_description"`
	}
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", label, err)
	}
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("%s: %s", label, firstNonEmptyString(tokenResp.Description, tokenResp.Error))
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("%s: missing access token", label)
	}

	extra := map[string]string{}
	if tokenResp.Scope != "" {
		extra["scope"] = tokenResp.Scope
	}
	return &TokenResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		TokenType:    firstNonEmptyString(tokenResp.TokenType, tokenTypeBearer),
		Extra:        extra,
	}, nil
}

func (y *YouTubeAdapter) GetProfile(ctx context.Context, accessToken string) (*UserProfile, error) {
	respBody, err := DoRequest(ctx, http.MethodGet, googleUserInfoURL, nil, bearerHeaders(accessToken))
	if err != nil {
		return nil, fmt.Errorf("youtube google profile: %w", err)
	}

	var profile struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
		Email   string `json:"email"`
		Error   struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &profile); err != nil {
		return nil, fmt.Errorf("decoding youtube google profile: %w", err)
	}
	if profile.Error.Message != "" {
		return nil, fmt.Errorf("youtube google profile: %s", profile.Error.Message)
	}
	return &UserProfile{
		ID:          profile.ID,
		Username:    firstNonEmptyString(profile.Email, profile.Name, profile.ID),
		DisplayName: firstNonEmptyString(profile.Name, profile.Email, profile.ID),
	}, nil
}

func (y *YouTubeAdapter) ListAccountSelections(ctx context.Context, token *TokenResult) ([]AccountSelectionOption, error) {
	channels, err := y.listChannels(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	options := make([]AccountSelectionOption, 0, len(channels))
	for _, channel := range channels {
		options = append(options, AccountSelectionOption{
			ID:          channel.ID,
			Username:    firstNonEmptyString(channel.Snippet.CustomURL, channel.Snippet.Title, channel.ID),
			DisplayName: channel.Snippet.Title,
			AvatarURL:   channel.Snippet.Thumbnails.Default.URL,
			Description: youtubeSubscriberDescription(channel.Statistics.SubscriberCount),
			Kind:        "channel",
		})
	}
	return options, nil
}

func (y *YouTubeAdapter) SelectAccount(ctx context.Context, token *TokenResult, selectionID string) (*SelectedAccount, error) {
	channels, err := y.listChannels(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	for _, channel := range channels {
		if channel.ID != selectionID {
			continue
		}
		selectedToken := *token
		selectedToken.Extra = map[string]string{}
		for key, value := range token.Extra {
			selectedToken.Extra[key] = value
		}
		selectedToken.Extra["channel_id"] = channel.ID

		return &SelectedAccount{
			AccountID:        channel.ID,
			AccountUsername:  firstNonEmptyString(channel.Snippet.CustomURL, channel.Snippet.Title, channel.ID),
			AccountAvatarURL: channel.Snippet.Thumbnails.Default.URL,
			Token:            &selectedToken,
		}, nil
	}
	return nil, fmt.Errorf("youtube channel selection %s was not found", selectionID)
}

func (y *YouTubeAdapter) listChannels(ctx context.Context, accessToken string) ([]youtubeChannel, error) {
	params := url.Values{}
	params.Set("part", "snippet,statistics")
	params.Set("mine", "true")
	params.Set("maxResults", "50")
	endpoint := youtubeAPIBaseURL + "/channels?" + params.Encode()
	respBody, err := DoRequest(ctx, http.MethodGet, endpoint, nil, bearerHeaders(accessToken))
	if err != nil {
		return nil, fmt.Errorf("youtube channels: %w", err)
	}

	var channelsResp struct {
		Items []youtubeChannel `json:"items"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &channelsResp); err != nil {
		return nil, fmt.Errorf("decoding youtube channels: %w", err)
	}
	if channelsResp.Error.Message != "" {
		return nil, fmt.Errorf("youtube channels: %s", channelsResp.Error.Message)
	}
	if len(channelsResp.Items) == 0 {
		return nil, fmt.Errorf("google account has no YouTube channels")
	}
	return channelsResp.Items, nil
}

func (y *YouTubeAdapter) UploadMedia(_ context.Context, _ string, _ string, _ string, _ io.Reader) (string, error) {
	return "", fmt.Errorf("youtube video upload requires post metadata")
}

func (y *YouTubeAdapter) UploadMediaWithMetadata(ctx context.Context, accessToken, _ string, req UploadMediaRequest) (string, error) {
	if req.Reader == nil {
		return "", fmt.Errorf("youtube upload requires a video reader")
	}
	if !isVideoMime(req.MimeType) {
		return "", fmt.Errorf("youtube upload requires a video attachment")
	}
	mediaBytes, err := io.ReadAll(req.Reader)
	if err != nil {
		return "", fmt.Errorf("reading youtube media: %w", err)
	}
	if len(mediaBytes) == 0 {
		return "", fmt.Errorf("youtube upload requires a non-empty video")
	}
	mediaSize := req.Size
	if mediaSize <= 0 {
		mediaSize = int64(len(mediaBytes))
	}
	if mediaSize != int64(len(mediaBytes)) {
		return "", fmt.Errorf("youtube upload size mismatch: expected %d bytes, read %d", mediaSize, len(mediaBytes))
	}

	metadata := youtubeVideoInsertRequest{
		Snippet: youtubeVideoSnippet{
			Title:       youtubeTitle(req),
			Description: strings.TrimSpace(req.Description),
			Tags:        youtubeTags(req.Settings),
			CategoryID:  settingString(req.Settings, "category_id"),
		},
		Status: youtubeVideoStatus{
			PrivacyStatus:           firstNonEmptyString(settingString(req.Settings, "privacy"), "private"),
			SelfDeclaredMadeForKids: settingBool(req.Settings, "self_declared_made_for_kids"),
		},
	}

	sessionURL, err := y.startYouTubeResumableUpload(ctx, accessToken, req, metadata, mediaSize)
	if err != nil {
		return "", err
	}
	videoID, err := y.uploadYouTubeVideoBytes(ctx, sessionURL, req.MimeType, mediaBytes)
	if err != nil {
		return "", err
	}
	if req.ThumbnailReader != nil {
		if err := y.setYouTubeThumbnail(ctx, accessToken, videoID, req); err != nil {
			return "", err
		}
	}
	if playlistID := settingString(req.Settings, "playlist_id"); playlistID != "" {
		if err := y.insertYouTubePlaylistItem(ctx, accessToken, playlistID, videoID); err != nil {
			return "", err
		}
	}
	if err := y.checkYouTubeProcessingStatus(ctx, accessToken, videoID); err != nil {
		return "", err
	}

	return videoID, nil
}

func (y *YouTubeAdapter) setYouTubeThumbnail(ctx context.Context, accessToken, videoID string, req UploadMediaRequest) error {
	thumbnailBytes, err := io.ReadAll(req.ThumbnailReader)
	if err != nil {
		return fmt.Errorf("reading youtube thumbnail: %w", err)
	}
	if len(thumbnailBytes) == 0 {
		return fmt.Errorf("youtube thumbnail upload requires a non-empty image")
	}
	if req.ThumbnailSize > 0 && req.ThumbnailSize != int64(len(thumbnailBytes)) {
		return fmt.Errorf("youtube thumbnail size mismatch: expected %d bytes, read %d", req.ThumbnailSize, len(thumbnailBytes))
	}
	params := url.Values{}
	params.Set("videoId", videoID)
	endpoint := youtubeUploadBaseURL + "/thumbnails/set?" + params.Encode()
	resp, err := doYouTubeRequest(ctx, http.MethodPost, endpoint, bytes.NewReader(thumbnailBytes), map[string]string{
		headerAuthorization: bearerPrefix + accessToken,
		headerContentType:   firstNonEmptyString(req.ThumbnailMimeType, "image/jpeg"),
	})
	if err != nil {
		return fmt.Errorf("youtube thumbnail upload: %w", err)
	}
	return youtubeAPIError("youtube thumbnail upload", resp.statusCode, resp.body)
}

func (y *YouTubeAdapter) startYouTubeResumableUpload(ctx context.Context, accessToken string, req UploadMediaRequest, metadata youtubeVideoInsertRequest, mediaSize int64) (string, error) {
	metaBytes, err := jsonMarshal(metadata)
	if err != nil {
		return "", fmt.Errorf("marshaling youtube metadata: %w", err)
	}

	params := url.Values{}
	params.Set("part", "snippet,status")
	params.Set("uploadType", "resumable")
	params.Set("notifySubscribers", fmt.Sprint(settingBool(req.Settings, "notify_subscribers")))
	endpoint := youtubeUploadBaseURL + "/videos?" + params.Encode()
	resp, err := doYouTubeRequest(ctx, http.MethodPost, endpoint, bytes.NewReader(metaBytes), map[string]string{
		headerAuthorization:       bearerPrefix + accessToken,
		headerContentType:         contentTypeJSON + "; charset=UTF-8",
		"X-Upload-Content-Length": strconv.FormatInt(mediaSize, 10),
		"X-Upload-Content-Type":   firstNonEmptyString(req.MimeType, videoTypeMP4),
	})
	if err != nil {
		return "", fmt.Errorf("youtube resumable upload session: %w", err)
	}
	if err := youtubeAPIError("youtube resumable upload session", resp.statusCode, resp.body); err != nil {
		return "", err
	}
	sessionURL := resp.header.Get("Location")
	if sessionURL == "" {
		return "", fmt.Errorf("youtube resumable upload session: missing session location")
	}
	return sessionURL, nil
}

func (y *YouTubeAdapter) Publish(_ context.Context, _ string, _ string, req *PublishRequest) (string, error) {
	if req.ReplyToID != "" {
		return "", fmt.Errorf("youtube thread replies are not supported")
	}
	if len(req.PlatformMediaIDs) != 1 || len(req.Media) != 1 {
		return "", fmt.Errorf("youtube publishing requires exactly one video attachment")
	}
	if !isVideoMime(req.Media[0].MimeType) {
		return "", fmt.Errorf("youtube publishing requires a video attachment")
	}
	return req.PlatformMediaIDs[0], nil
}

func (y *YouTubeAdapter) uploadYouTubeVideoBytes(ctx context.Context, sessionURL, mimeType string, mediaBytes []byte) (string, error) {
	offset := int64(0)
	for attempt := 0; attempt < 3; attempt++ {
		end := int64(len(mediaBytes)) - 1
		headers := map[string]string{
			headerContentType: "video/mp4",
			"Content-Range":   fmt.Sprintf("bytes %d-%d/%d", offset, end, len(mediaBytes)),
		}
		if mimeType != "" {
			headers[headerContentType] = mimeType
		}
		resp, err := doYouTubeRequest(ctx, http.MethodPut, sessionURL, bytes.NewReader(mediaBytes[offset:]), headers)
		if err != nil {
			return "", fmt.Errorf("youtube video upload: %w", err)
		}
		switch {
		case resp.statusCode == http.StatusOK || resp.statusCode == http.StatusCreated:
			return youtubeVideoIDFromResponse("youtube video upload", resp.body)
		case resp.statusCode == http.StatusPermanentRedirect:
			nextOffset := youtubeNextUploadOffset(resp.header.Get("Range"))
			if nextOffset <= offset {
				return "", fmt.Errorf("youtube video upload: could not advance resumable upload offset")
			}
			offset = nextOffset
		case resp.statusCode >= 500:
			nextOffset, statusErr := y.queryYouTubeUploadOffset(ctx, sessionURL, int64(len(mediaBytes)))
			if statusErr != nil {
				return "", statusErr
			}
			if nextOffset <= offset {
				return "", fmt.Errorf("youtube video upload: could not recover upload offset after transient failure")
			}
			offset = nextOffset
		default:
			return "", youtubeAPIError("youtube video upload", resp.statusCode, resp.body)
		}
	}
	return "", fmt.Errorf("youtube video upload: retry limit exceeded")
}

func (y *YouTubeAdapter) queryYouTubeUploadOffset(ctx context.Context, sessionURL string, mediaSize int64) (int64, error) {
	resp, err := doYouTubeRequest(ctx, http.MethodPut, sessionURL, http.NoBody, map[string]string{
		"Content-Range": fmt.Sprintf("bytes */%d", mediaSize),
	})
	if err != nil {
		return 0, fmt.Errorf("youtube video upload status: %w", err)
	}
	if resp.statusCode == http.StatusPermanentRedirect {
		return youtubeNextUploadOffset(resp.header.Get("Range")), nil
	}
	if resp.statusCode == http.StatusOK || resp.statusCode == http.StatusCreated {
		return mediaSize, nil
	}
	return 0, youtubeAPIError("youtube video upload status", resp.statusCode, resp.body)
}

func (y *YouTubeAdapter) insertYouTubePlaylistItem(ctx context.Context, accessToken, playlistID, videoID string) error {
	body, err := jsonMarshal(youtubePlaylistItemInsertRequest{
		Snippet: youtubePlaylistItemSnippet{
			PlaylistID: playlistID,
			ResourceID: youtubePlaylistItemResourceID{
				Kind:    "youtube#video",
				VideoID: videoID,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("marshaling youtube playlist item: %w", err)
	}
	params := url.Values{}
	params.Set("part", "snippet")
	endpoint := youtubeAPIBaseURL + "/playlistItems?" + params.Encode()
	resp, err := doYouTubeRequest(ctx, http.MethodPost, endpoint, bytes.NewReader(body), map[string]string{
		headerAuthorization: bearerPrefix + accessToken,
		headerContentType:   contentTypeJSON,
	})
	if err != nil {
		return fmt.Errorf("youtube playlist insert: %w", err)
	}
	return youtubeAPIError("youtube playlist insert", resp.statusCode, resp.body)
}

func (y *YouTubeAdapter) checkYouTubeProcessingStatus(ctx context.Context, accessToken, videoID string) error {
	params := url.Values{}
	params.Set("part", "status,processingDetails")
	params.Set("id", videoID)
	endpoint := youtubeAPIBaseURL + "/videos?" + params.Encode()
	respBody, err := DoRequest(ctx, http.MethodGet, endpoint, nil, bearerHeaders(accessToken))
	if err != nil {
		return fmt.Errorf("youtube processing status: %w", err)
	}
	var statusResp youtubeVideosListResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return fmt.Errorf("decoding youtube processing status: %w", err)
	}
	if statusResp.Error.Message != "" {
		return fmt.Errorf("youtube processing status: %s", statusResp.Error.Message)
	}
	if len(statusResp.Items) == 0 {
		return fmt.Errorf("youtube processing status: uploaded video was not returned")
	}
	item := statusResp.Items[0]
	if item.ProcessingDetails.ProcessingStatus == "failed" || item.Status.UploadStatus == "rejected" || item.Status.UploadStatus == "failed" {
		return fmt.Errorf("youtube processing failed: %s", firstNonEmptyString(item.ProcessingDetails.ProcessingFailureReason, item.Status.FailureReason, item.Status.RejectionReason, "unknown"))
	}
	return nil
}

func youtubeVideoIDFromResponse(label string, respBody []byte) (string, error) {
	var resp struct {
		ID    string `json:"id"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("decoding %s: %w", label, err)
	}
	if resp.Error.Message != "" {
		return "", fmt.Errorf("%s: %s", label, resp.Error.Message)
	}
	if resp.ID == "" {
		return "", fmt.Errorf("%s: missing video id", label)
	}
	return resp.ID, nil
}

type youtubeHTTPResponse struct {
	statusCode int
	header     http.Header
	body       []byte
}

func doYouTubeRequest(ctx context.Context, method, endpoint string, body io.Reader, headers map[string]string) (*youtubeHTTPResponse, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		if value != "" {
			req.Header.Set(key, value)
		}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &youtubeHTTPResponse{statusCode: resp.StatusCode, header: resp.Header, body: respBody}, nil
}

func youtubeAPIError(label string, statusCode int, respBody []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	var resp struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(respBody, &resp)
	if resp.Error.Message != "" {
		return fmt.Errorf("%s: %s", label, resp.Error.Message)
	}
	return fmt.Errorf("%s: unexpected status %d", label, statusCode)
}

func youtubeNextUploadOffset(rangeHeader string) int64 {
	if !strings.HasPrefix(rangeHeader, "bytes=0-") {
		return 0
	}
	uploadedEnd, err := strconv.ParseInt(strings.TrimPrefix(rangeHeader, "bytes=0-"), 10, 64)
	if err != nil {
		return 0
	}
	return uploadedEnd + 1
}

func validateYouTubeMedia(media []MediaItem) []MediaValidationIssue {
	if len(media) != 1 {
		return []MediaValidationIssue{{
			Provider: providerYouTube,
			Severity: severityError,
			Message:  "YouTube publishing currently requires exactly one video attachment.",
		}}
	}
	if !isVideoMime(media[0].MimeType) {
		return []MediaValidationIssue{{
			Provider: providerYouTube,
			MediaID:  media[0].ID,
			Severity: severityError,
			Message:  "YouTube publishing supports video attachments only.",
		}}
	}
	return nil
}

func youtubeTitle(req UploadMediaRequest) string {
	title := firstNonEmptyString(settingString(req.Settings, "title"), strings.TrimSpace(req.Title))
	if title == "" {
		for _, line := range strings.Split(req.Description, "\n") {
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				title = trimmed
				break
			}
		}
	}
	if title == "" {
		title = youtubeDefaultVideoName
	}
	return truncateRunes(title, youtubeTitleMaxRunes)
}

func youtubeTags(settings map[string]interface{}) []string {
	raw := settingString(settings, "tags")
	if raw == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n'
	})
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		if tag := strings.TrimSpace(part); tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func youtubeSubscriberDescription(count string) string {
	if strings.TrimSpace(count) == "" {
		return ""
	}
	return count + " subscribers"
}

func youtubeScopes() []string {
	return []string{
		"https://www.googleapis.com/auth/userinfo.profile",
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/youtube.readonly",
		"https://www.googleapis.com/auth/youtube.upload",
		"https://www.googleapis.com/auth/youtube",
	}
}

func bearerHeaders(accessToken string) map[string]string {
	return map[string]string{headerAuthorization: bearerPrefix + accessToken}
}

type youtubeVideoInsertRequest struct {
	Snippet youtubeVideoSnippet `json:"snippet"`
	Status  youtubeVideoStatus  `json:"status"`
}

type youtubeVideoSnippet struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	CategoryID  string   `json:"categoryId,omitempty"`
}

type youtubeVideoStatus struct {
	PrivacyStatus           string `json:"privacyStatus"`
	SelfDeclaredMadeForKids bool   `json:"selfDeclaredMadeForKids,omitempty"`
}

type youtubePlaylistItemInsertRequest struct {
	Snippet youtubePlaylistItemSnippet `json:"snippet"`
}

type youtubePlaylistItemSnippet struct {
	PlaylistID string                        `json:"playlistId"`
	ResourceID youtubePlaylistItemResourceID `json:"resourceId"`
}

type youtubePlaylistItemResourceID struct {
	Kind    string `json:"kind"`
	VideoID string `json:"videoId"`
}

type youtubeVideosListResponse struct {
	Items []struct {
		ID                string `json:"id"`
		ProcessingDetails struct {
			ProcessingStatus        string `json:"processingStatus"`
			ProcessingFailureReason string `json:"processingFailureReason"`
		} `json:"processingDetails"`
		Status struct {
			UploadStatus    string `json:"uploadStatus"`
			FailureReason   string `json:"failureReason"`
			RejectionReason string `json:"rejectionReason"`
		} `json:"status"`
	} `json:"items"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type youtubeChannel struct {
	ID      string `json:"id"`
	Snippet struct {
		Title      string `json:"title"`
		CustomURL  string `json:"customUrl"`
		Thumbnails struct {
			Default struct {
				URL string `json:"url"`
			} `json:"default"`
		} `json:"thumbnails"`
	} `json:"snippet"`
	Statistics struct {
		SubscriberCount string `json:"subscriberCount"`
	} `json:"statistics"`
}
