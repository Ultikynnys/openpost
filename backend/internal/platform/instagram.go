package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type InstagramAdapter struct {
	clientID     string
	clientSecret string
	redirectURI  string
	graphVersion string
}

func NewInstagramAdapter(clientID, clientSecret, redirectURI string) *InstagramAdapter {
	return &InstagramAdapter{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		graphVersion: metaGraphAPIVersion(),
	}
}

func (i *InstagramAdapter) graphURL(path string) string {
	return facebookGraphBaseURL + "/" + i.graphVersion + "/" + strings.TrimPrefix(path, "/")
}

func (i *InstagramAdapter) GenerateAuthURL(state string) (string, map[string]string) {
	params := url.Values{}
	params.Set(oauthParamClientID, i.clientID)
	params.Set(oauthParamRedirectURI, i.redirectURI)
	params.Set("response_type", oauthResponseType)
	params.Set("scope", strings.Join(instagramScopes(), ","))
	params.Set("state", state)
	return facebookOAuthBaseURL + "/" + i.graphVersion + "/dialog/oauth?" + params.Encode(), nil
}

func (i *InstagramAdapter) ExchangeCode(ctx context.Context, code string, _ map[string]string) (*TokenResult, error) {
	return exchangeMetaAuthCode(ctx, i.graphURL, i.clientID, i.clientSecret, i.redirectURI, "instagram", code)
}

func (i *InstagramAdapter) RefreshCapability() RefreshCapability {
	return RefreshCapability{}
}

func (i *InstagramAdapter) RefreshToken(_ context.Context, _ RefreshTokenInput) (*TokenResult, error) {
	return nil, fmt.Errorf("instagram page tokens do not support OpenPost refresh yet")
}

func (i *InstagramAdapter) GetProfile(ctx context.Context, accessToken string) (*UserProfile, error) {
	respBody, err := DoRequest(ctx, http.MethodGet, i.graphURL("me?fields=id,name&access_token="+url.QueryEscape(accessToken)), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("instagram facebook profile: %w", err)
	}
	var profile struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &profile); err != nil {
		return nil, fmt.Errorf("decoding instagram facebook profile: %w", err)
	}
	if profile.Error.Message != "" {
		return nil, fmt.Errorf("instagram facebook profile: %s", profile.Error.Message)
	}
	return &UserProfile{ID: profile.ID, Username: profile.Name, DisplayName: profile.Name}, nil
}

func (i *InstagramAdapter) ListAccountSelections(ctx context.Context, token *TokenResult) ([]AccountSelectionOption, error) {
	pages, err := i.listInstagramPages(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	options := make([]AccountSelectionOption, 0, len(pages))
	for _, page := range pages {
		ig := page.InstagramBusinessAccount
		options = append(options, AccountSelectionOption{
			ID:          ig.ID,
			Username:    firstNonEmptyString(ig.Username, ig.Name, page.Name),
			DisplayName: firstNonEmptyString(ig.Name, ig.Username, page.Name),
			AvatarURL:   firstNonEmptyString(ig.ProfilePictureURL, page.Picture.Data.URL),
			Kind:        "instagram_business_account",
			Extra: map[string]string{
				"page_id": page.ID,
			},
		})
	}
	return options, nil
}

func (i *InstagramAdapter) SelectAccount(ctx context.Context, token *TokenResult, selectionID string) (*SelectedAccount, error) {
	pages, err := i.listInstagramPages(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	for _, page := range pages {
		ig := page.InstagramBusinessAccount
		if ig.ID != selectionID {
			continue
		}
		if page.AccessToken == "" {
			return nil, fmt.Errorf("instagram page %s did not include a page access token", page.ID)
		}
		pageToken := *token
		pageToken.AccessToken = page.AccessToken
		pageToken.RefreshToken = ""
		pageToken.ExpiresIn = 0
		pageToken.Extra = map[string]string{}
		for key, value := range token.Extra {
			pageToken.Extra[key] = value
		}
		pageToken.Extra["page_id"] = page.ID
		pageToken.Extra["instagram_business_account_id"] = ig.ID

		return &SelectedAccount{
			AccountID:        ig.ID,
			AccountUsername:  firstNonEmptyString(ig.Username, ig.Name, page.Name),
			AccountAvatarURL: firstNonEmptyString(ig.ProfilePictureURL, page.Picture.Data.URL),
			Token:            &pageToken,
		}, nil
	}
	return nil, fmt.Errorf("instagram account selection %s was not found", selectionID)
}

func (i *InstagramAdapter) listInstagramPages(ctx context.Context, accessToken string) ([]instagramPage, error) {
	fields := "id,name,username,access_token,picture.type(square),instagram_business_account{id,username,name,profile_picture_url}"
	endpoint := i.graphURL("me/accounts") + "?fields=" + url.QueryEscape(fields) + "&limit=100&access_token=" + url.QueryEscape(accessToken)
	respBody, err := DoRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("instagram accounts: %w", err)
	}

	var pagesResp struct {
		Data  []instagramPage `json:"data"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &pagesResp); err != nil {
		return nil, fmt.Errorf("decoding instagram accounts: %w", err)
	}
	if pagesResp.Error.Message != "" {
		return nil, fmt.Errorf("instagram accounts: %s", pagesResp.Error.Message)
	}

	pages := make([]instagramPage, 0, len(pagesResp.Data))
	for _, page := range pagesResp.Data {
		if page.InstagramBusinessAccount.ID != "" {
			pages = append(pages, page)
		}
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("facebook account has no connected instagram business accounts")
	}
	return pages, nil
}

func (i *InstagramAdapter) UploadMedia(_ context.Context, _ string, _ string, _ string, _ io.Reader) (string, error) {
	return "", fmt.Errorf("instagram uses publicly accessible HTTPS media URLs for the initial adapter")
}

func (i *InstagramAdapter) Publish(ctx context.Context, accessToken, instagramUserID string, req *PublishRequest) (string, error) {
	if req.ReplyToID != "" {
		return i.publishCommentReply(ctx, accessToken, req.ReplyToID, req.Content)
	}
	if len(req.PlatformMediaIDs) == 0 || len(req.PlatformMediaIDs) != len(req.Media) {
		return "", fmt.Errorf("instagram publishing requires media attachment metadata")
	}
	for _, mediaURL := range req.PlatformMediaIDs {
		if !strings.HasPrefix(mediaURL, "https://") {
			return "", fmt.Errorf("instagram requires a publicly-accessible HTTPS media URL. Set OPENPOST_MEDIA_URL to your public media base URL")
		}
	}
	if settingString(req.Settings, "post_type") == "story" || req.Profile == "story" {
		return i.publishStories(ctx, accessToken, instagramUserID, req)
	}
	if len(req.PlatformMediaIDs) > 1 {
		return i.publishCarousel(ctx, accessToken, instagramUserID, req)
	}

	containerID, err := i.createMediaContainer(ctx, accessToken, instagramUserID, req.Content, req.PlatformMediaIDs[0], isVideoMime(req.Media[0].MimeType), false, req)
	if err != nil {
		return "", err
	}
	if err := i.waitForContainer(ctx, accessToken, containerID); err != nil {
		return "", err
	}
	return i.publishMediaContainer(ctx, accessToken, instagramUserID, containerID)
}

//nolint:gocyclo
func (i *InstagramAdapter) createMediaContainer(ctx context.Context, accessToken, instagramUserID, caption, mediaURL string, video bool, carouselItem bool, req *PublishRequest) (string, error) {
	values := map[string]string{
		"caption":             strings.TrimSpace(caption),
		oauthParamAccessToken: accessToken,
	}
	if video {
		values["video_url"] = mediaURL
		if settingString(req.Settings, "post_type") == "story" || req.Profile == "story" {
			values["media_type"] = "STORIES"
		} else if settingBool(req.Settings, "is_reel") || req.Profile == "short_video" {
			values["media_type"] = "REELS"
		}
	} else {
		values["image_url"] = mediaURL
		if settingString(req.Settings, "post_type") == "story" || req.Profile == "story" {
			values["media_type"] = "STORIES"
		}
	}
	if carouselItem {
		values["is_carousel_item"] = "true"
		delete(values, "caption")
	}
	if collaborators := settingString(req.Settings, "collaborators"); collaborators != "" && !carouselItem {
		values["collaborators"] = collaborators
	}
	if settingBool(req.Settings, "is_trial_reel") && !carouselItem {
		values["is_trial_reel"] = "true"
	}
	if graduation := settingString(req.Settings, "graduation_strategy"); graduation != "" && !carouselItem {
		values["graduation_strategy"] = graduation
	}
	if thumb := settingString(req.Settings, "thumbnail_timestamp_ms"); thumb != "" && video && !carouselItem {
		values["thumb_offset"] = thumb
	}

	respBody, err := DoFormURLEncoded(ctx, http.MethodPost, i.graphURL(instagramUserID+"/media"), values, nil)
	if err != nil {
		return "", fmt.Errorf("instagram media container: %w", err)
	}
	return instagramIDFromResponse("instagram media container", respBody)
}

func (i *InstagramAdapter) publishCarousel(ctx context.Context, accessToken, instagramUserID string, req *PublishRequest) (string, error) {
	if len(req.PlatformMediaIDs) < 2 || len(req.PlatformMediaIDs) > 10 {
		return "", fmt.Errorf("instagram carousel requires 2-10 media items")
	}
	childIDs := make([]string, 0, len(req.PlatformMediaIDs))
	for index, mediaURL := range req.PlatformMediaIDs {
		childID, err := i.createMediaContainer(ctx, accessToken, instagramUserID, "", mediaURL, isVideoMime(req.Media[index].MimeType), true, req)
		if err != nil {
			return "", err
		}
		if err := i.waitForContainer(ctx, accessToken, childID); err != nil {
			return "", err
		}
		childIDs = append(childIDs, childID)
	}
	values := map[string]string{
		"media_type":          "CAROUSEL",
		"children":            strings.Join(childIDs, ","),
		"caption":             strings.TrimSpace(req.Content),
		oauthParamAccessToken: accessToken,
	}
	respBody, err := DoFormURLEncoded(ctx, http.MethodPost, i.graphURL(instagramUserID+"/media"), values, nil)
	if err != nil {
		return "", fmt.Errorf("instagram carousel container: %w", err)
	}
	containerID, err := instagramIDFromResponse("instagram carousel container", respBody)
	if err != nil {
		return "", err
	}
	if err := i.waitForContainer(ctx, accessToken, containerID); err != nil {
		return "", err
	}
	return i.publishMediaContainer(ctx, accessToken, instagramUserID, containerID)
}

func (i *InstagramAdapter) publishStories(ctx context.Context, accessToken, instagramUserID string, req *PublishRequest) (string, error) {
	ids := make([]string, 0, len(req.PlatformMediaIDs))
	for index, mediaURL := range req.PlatformMediaIDs {
		containerID, err := i.createMediaContainer(ctx, accessToken, instagramUserID, "", mediaURL, isVideoMime(req.Media[index].MimeType), false, req)
		if err != nil {
			return "", err
		}
		if err := i.waitForContainer(ctx, accessToken, containerID); err != nil {
			return "", err
		}
		publishedID, err := i.publishMediaContainer(ctx, accessToken, instagramUserID, containerID)
		if err != nil {
			return "", err
		}
		ids = append(ids, publishedID)
	}
	return strings.Join(ids, ","), nil
}

func (i *InstagramAdapter) publishCommentReply(ctx context.Context, accessToken, commentID, message string) (string, error) {
	values := map[string]string{
		"message":             strings.TrimSpace(message),
		oauthParamAccessToken: accessToken,
	}
	respBody, err := DoFormURLEncoded(ctx, http.MethodPost, i.graphURL(commentID+"/replies"), values, nil)
	if err != nil {
		return "", fmt.Errorf("instagram comment reply: %w", err)
	}
	return instagramIDFromResponse("instagram comment reply", respBody)
}

func (i *InstagramAdapter) ListComments(ctx context.Context, accessToken, _ string, externalID string) ([]Comment, error) {
	fields := "id,text,timestamp,username,hidden"
	endpoint := i.graphURL(externalID+"/comments") + "?fields=" + url.QueryEscape(fields) + "&access_token=" + url.QueryEscape(accessToken)
	respBody, err := DoRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("instagram comments: %w", err)
	}

	var result struct {
		Data []struct {
			ID        string `json:"id"`
			Text      string `json:"text"`
			Timestamp string `json:"timestamp"`
			Username  string `json:"username"`
			Hidden    bool   `json:"hidden"`
		} `json:"data"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decoding instagram comments: %w", err)
	}
	if result.Error.Message != "" {
		return nil, fmt.Errorf("instagram comments: %s", result.Error.Message)
	}

	comments := make([]Comment, 0, len(result.Data))
	for _, item := range result.Data {
		comments = append(comments, Comment{
			ID:         item.ID,
			AuthorName: item.Username,
			Text:       item.Text,
			CreatedAt:  item.Timestamp,
			Hidden:     item.Hidden,
			CanReply:   true,
			CanHide:    true,
			CanDelete:  true,
		})
	}
	return comments, nil
}

func (i *InstagramAdapter) ReplyToComment(ctx context.Context, accessToken, _ string, commentID, message string) (string, error) {
	return i.publishCommentReply(ctx, accessToken, commentID, message)
}

func (i *InstagramAdapter) HideComment(ctx context.Context, accessToken, _ string, commentID string) error {
	_, err := DoFormURLEncoded(ctx, http.MethodPost, i.graphURL(commentID), map[string]string{
		"hide":                "true",
		oauthParamAccessToken: accessToken,
	}, nil)
	if err != nil {
		return fmt.Errorf("instagram hide comment: %w", err)
	}
	return nil
}

func (i *InstagramAdapter) DeleteComment(ctx context.Context, accessToken, _ string, commentID string) error {
	endpoint := i.graphURL(commentID) + "?access_token=" + url.QueryEscape(accessToken)
	if _, err := DoRequest(ctx, http.MethodDelete, endpoint, nil, nil); err != nil {
		return fmt.Errorf("instagram delete comment: %w", err)
	}
	return nil
}

func (i *InstagramAdapter) waitForContainer(ctx context.Context, accessToken, containerID string) error {
	const maxAttempts = 6
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		respBody, err := DoRequest(ctx, http.MethodGet, i.graphURL(containerID+"?fields=status_code&access_token="+url.QueryEscape(accessToken)), nil, nil)
		if err != nil {
			return fmt.Errorf("instagram container status: %w", err)
		}
		var statusResp struct {
			StatusCode string `json:"status_code"`
			Error      struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(respBody, &statusResp); err != nil {
			return fmt.Errorf("decoding instagram container status: %w", err)
		}
		if statusResp.Error.Message != "" {
			return fmt.Errorf("instagram container status: %s", statusResp.Error.Message)
		}
		switch statusResp.StatusCode {
		case "", "FINISHED":
			return nil
		case "ERROR", "EXPIRED":
			return fmt.Errorf("instagram container processing failed: %s", statusResp.StatusCode)
		}
		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(10 * time.Second):
			}
		}
	}
	return fmt.Errorf("instagram container processing timed out")
}

func (i *InstagramAdapter) publishMediaContainer(ctx context.Context, accessToken, instagramUserID, containerID string) (string, error) {
	values := map[string]string{
		"creation_id":         containerID,
		oauthParamAccessToken: accessToken,
	}
	respBody, err := DoFormURLEncoded(ctx, http.MethodPost, i.graphURL(instagramUserID+"/media_publish"), values, nil)
	if err != nil {
		return "", fmt.Errorf("instagram media publish: %w", err)
	}
	return instagramIDFromResponse("instagram media publish", respBody)
}

func instagramIDFromResponse(label string, respBody []byte) (string, error) {
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
		return "", fmt.Errorf("%s: missing id", label)
	}
	return resp.ID, nil
}

func validateInstagramMedia(media []MediaItem) []MediaValidationIssue {
	if len(media) < 1 || len(media) > 10 {
		return []MediaValidationIssue{{
			Provider: providerInstagram,
			Severity: severityError,
			Message:  "Instagram publishing requires 1-10 image or video attachments.",
		}}
	}
	for _, item := range media {
		if isInstagramImageMime(item.MimeType) || isInstagramVideoMime(item.MimeType) {
			continue
		}
		return []MediaValidationIssue{{
			Provider: providerInstagram,
			MediaID:  item.ID,
			Severity: severityError,
			Message:  "Instagram supports JPEG, PNG, WebP, MP4, or MOV media.",
		}}
	}
	return nil
}

func isInstagramImageMime(mimeType string) bool {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

func isInstagramVideoMime(mimeType string) bool {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case videoTypeMP4, "video/quicktime":
		return true
	default:
		return false
	}
}

func instagramScopes() []string {
	return []string{
		"instagram_basic",
		"instagram_content_publish",
		"pages_show_list",
		"pages_read_engagement",
		"business_management",
	}
}

type instagramPage struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	Username                 string `json:"username"`
	AccessToken              string `json:"access_token"`
	InstagramBusinessAccount struct {
		ID                string `json:"id"`
		Username          string `json:"username"`
		Name              string `json:"name"`
		ProfilePictureURL string `json:"profile_picture_url"`
	} `json:"instagram_business_account"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}
