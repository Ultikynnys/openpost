package capabilities

import (
	"fmt"
	"math"
	"net/url"
	"strings"

	"github.com/openpost/backend/internal/models"
)

const (
	ProviderBluesky   = "bluesky"
	ProviderFacebook  = "facebook"
	ProviderInstagram = "instagram"
	ProviderLinkedIn  = "linkedin"
	ProviderMastodon  = "mastodon"
	ProviderThreads   = "threads"
	ProviderTikTok    = "tiktok"
	ProviderX         = "x"
	ProviderYouTube   = "youtube"
)

type Profile struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SettingField struct {
	Key      string   `json:"key"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Options  []string `json:"options,omitempty"`
	Help     string   `json:"help,omitempty"`
}

type MediaConstraint struct {
	MinCount               int      `json:"min_count"`
	MaxCount               int      `json:"max_count"`
	AllowedMIMEs           []string `json:"allowed_mimes"`
	AspectRatios           []string `json:"aspect_ratios,omitempty"`
	MaxDurationSeconds     int      `json:"max_duration_seconds,omitempty"`
	MaxSizeBytes           int64    `json:"max_size_bytes,omitempty"`
	RequiresPublicURL      bool     `json:"requires_public_url"`
	RequiresHTTPSFetchable bool     `json:"requires_https_fetchable"`
}

type Capability struct {
	Provider             string            `json:"provider"`
	Profile              string            `json:"profile"`
	Label                string            `json:"label"`
	ValidationCategories []string          `json:"validation_categories,omitempty"`
	TextLimit            int               `json:"text_limit,omitempty"`
	TitleRequired        bool              `json:"title_required,omitempty"`
	DescriptionRequired  bool              `json:"description_required,omitempty"`
	NativeScheduling     bool              `json:"native_scheduling"`
	OpenPostQueued       bool              `json:"openpost_queued"`
	RequiresAppReview    bool              `json:"requires_app_review"`
	RequiresPublicMedia  bool              `json:"requires_public_media"`
	Media                MediaConstraint   `json:"media"`
	Settings             []SettingField    `json:"settings,omitempty"`
	Caveats              []string          `json:"caveats,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty"`
}

type MediaItem struct {
	ID              string
	MimeType        string
	Size            int64
	Width           int
	Height          int
	DurationMS      int64
	AnalysisStatus  string
	AnalysisError   string
	PublicURLReady  bool
	PublicURLStatus int
	PublicURLError  string
	URL             string
}

type ValidationIssue struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Provider string `json:"provider,omitempty"`
	Profile  string `json:"profile,omitempty"`
	MediaID  string `json:"media_id,omitempty"`
	Field    string `json:"field,omitempty"`
}

func Profiles() []Profile {
	return []Profile{
		{Key: models.ContentProfileShortText, Name: "Short text", Description: "Fast text-first posts for timelines and feeds."},
		{Key: models.ContentProfileThread, Name: "Thread", Description: "Ordered multi-segment posts and reply chains."},
		{Key: models.ContentProfileLinkShare, Name: "Link share", Description: "URL-driven posts with platform link metadata."},
		{Key: models.ContentProfileImagePost, Name: "Image post", Description: "Single image or simple media feed posts."},
		{Key: models.ContentProfileCarousel, Name: "Carousel", Description: "Multi-image or mixed media swipes."},
		{Key: models.ContentProfileStory, Name: "Story", Description: "Ephemeral vertical story publishing."},
		{Key: models.ContentProfileShortVideo, Name: "Short video", Description: "Reels, Shorts, TikTok, and short-form video."},
		{Key: models.ContentProfileLongVideo, Name: "Long video", Description: "YouTube and feed video uploads with metadata."},
	}
}

func All() []Capability {
	text := MediaConstraint{MinCount: 0, MaxCount: 0}
	image := MediaConstraint{MinCount: 1, MaxCount: 1, AllowedMIMEs: []string{"image/jpeg", "image/png", "image/webp"}}
	feedImages := MediaConstraint{MinCount: 1, MaxCount: 10, AllowedMIMEs: []string{"image/jpeg", "image/png", "image/webp"}}
	document := MediaConstraint{MinCount: 1, MaxCount: 1, AllowedMIMEs: []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
	}, MaxSizeBytes: 100 * 1024 * 1024}
	publicImages := feedImages
	publicImages.RequiresPublicURL = true
	publicImages.RequiresHTTPSFetchable = true
	video := MediaConstraint{MinCount: 1, MaxCount: 1, AllowedMIMEs: []string{"video/mp4", "video/quicktime"}, MaxSizeBytes: 2 * 1024 * 1024 * 1024}
	shortVideo := video
	shortVideo.MaxDurationSeconds = 180
	shortVideo.AspectRatios = []string{"9:16", "1:1"}
	blueskyVideo := MediaConstraint{MinCount: 1, MaxCount: 1, AllowedMIMEs: []string{"video/mp4"}, MaxSizeBytes: 100 * 1024 * 1024}
	publicShortVideo := shortVideo
	publicShortVideo.RequiresPublicURL = true
	publicShortVideo.RequiresHTTPSFetchable = true
	longVideo := video
	longVideo.MaxDurationSeconds = 43200

	defaultQueued := func(c Capability) Capability {
		c.OpenPostQueued = true
		c.ValidationCategories = validationCategories(c)
		return c
	}

	return []Capability{
		defaultQueued(Capability{Provider: ProviderX, Profile: models.ContentProfileShortText, Label: "X post", TextLimit: 280, Media: text, Settings: xSettings()}),
		defaultQueued(Capability{Provider: ProviderX, Profile: models.ContentProfileThread, Label: "X thread", TextLimit: 280, Media: text, Settings: xSettings()}),
		defaultQueued(Capability{Provider: ProviderX, Profile: models.ContentProfileLinkShare, Label: "X link", TextLimit: 280, Media: text, Settings: append(linkSettings(), xSettings()...)}),
		defaultQueued(Capability{Provider: ProviderX, Profile: models.ContentProfileImagePost, Label: "X image post", TextLimit: 280, Media: MediaConstraint{MinCount: 1, MaxCount: 4, AllowedMIMEs: []string{"image/jpeg", "image/png", "image/webp", "image/gif"}}, Settings: xSettings()}),
		defaultQueued(Capability{Provider: ProviderX, Profile: models.ContentProfileShortVideo, Label: "X video", TextLimit: 280, Media: shortVideo, Settings: xSettings()}),

		defaultQueued(Capability{Provider: ProviderBluesky, Profile: models.ContentProfileShortText, Label: "Bluesky post", TextLimit: 300, Media: text, Settings: blueskySettings()}),
		defaultQueued(Capability{Provider: ProviderBluesky, Profile: models.ContentProfileThread, Label: "Bluesky thread", TextLimit: 300, Media: text, Settings: blueskySettings()}),
		defaultQueued(Capability{Provider: ProviderBluesky, Profile: models.ContentProfileLinkShare, Label: "Bluesky link", TextLimit: 300, Media: text, Settings: blueskySettings()}),
		defaultQueued(Capability{Provider: ProviderBluesky, Profile: models.ContentProfileImagePost, Label: "Bluesky images", TextLimit: 300, Media: MediaConstraint{MinCount: 1, MaxCount: 4, AllowedMIMEs: []string{"image/jpeg", "image/png", "image/webp"}}, Settings: blueskySettings()}),
		defaultQueued(Capability{Provider: ProviderBluesky, Profile: models.ContentProfileShortVideo, Label: "Bluesky video", TextLimit: 300, Media: blueskyVideo, Settings: blueskySettings()}),

		defaultQueued(Capability{Provider: ProviderMastodon, Profile: models.ContentProfileShortText, Label: "Mastodon post", TextLimit: 500, Media: text, NativeScheduling: true, Settings: mastodonSettings()}),
		defaultQueued(Capability{Provider: ProviderMastodon, Profile: models.ContentProfileThread, Label: "Mastodon thread", TextLimit: 500, Media: text, NativeScheduling: true, Settings: mastodonSettings()}),
		defaultQueued(Capability{Provider: ProviderMastodon, Profile: models.ContentProfileLinkShare, Label: "Mastodon link", TextLimit: 500, Media: text, NativeScheduling: true, Settings: append(linkSettings(), mastodonSettings()...)}),
		defaultQueued(Capability{Provider: ProviderMastodon, Profile: models.ContentProfileImagePost, Label: "Mastodon media", TextLimit: 500, Media: MediaConstraint{MinCount: 1, MaxCount: 4, AllowedMIMEs: []string{"image/jpeg", "image/png", "image/webp", "image/gif", "video/mp4"}}, NativeScheduling: true, Settings: mastodonSettings()}),

		defaultQueued(Capability{Provider: ProviderThreads, Profile: models.ContentProfileShortText, Label: "Threads post", TextLimit: 500, Media: text}),
		defaultQueued(Capability{Provider: ProviderThreads, Profile: models.ContentProfileThread, Label: "Threads thread", TextLimit: 500, Media: text}),
		defaultQueued(Capability{Provider: ProviderThreads, Profile: models.ContentProfileImagePost, Label: "Threads media", TextLimit: 500, Media: publicImages, RequiresPublicMedia: true}),
		defaultQueued(Capability{Provider: ProviderThreads, Profile: models.ContentProfileCarousel, Label: "Threads carousel", TextLimit: 500, Media: publicImages, RequiresPublicMedia: true}),
		defaultQueued(Capability{Provider: ProviderThreads, Profile: models.ContentProfileShortVideo, Label: "Threads video", TextLimit: 500, Media: publicShortVideo, RequiresPublicMedia: true}),

		defaultQueued(Capability{Provider: ProviderLinkedIn, Profile: models.ContentProfileShortText, Label: "LinkedIn post", TextLimit: 3000, Media: text}),
		defaultQueued(Capability{Provider: ProviderLinkedIn, Profile: models.ContentProfileThread, Label: "LinkedIn comment thread", TextLimit: 1250, Media: text}),
		defaultQueued(Capability{Provider: ProviderLinkedIn, Profile: models.ContentProfileLinkShare, Label: "LinkedIn link", TextLimit: 3000, Media: text, Settings: linkSettings()}),
		defaultQueued(Capability{Provider: ProviderLinkedIn, Profile: models.ContentProfileImagePost, Label: "LinkedIn image", TextLimit: 3000, Media: image}),
		defaultQueued(Capability{Provider: ProviderLinkedIn, Profile: models.ContentProfileCarousel, Label: "LinkedIn document", TextLimit: 3000, Media: document, Settings: []SettingField{{Key: "title", Label: "Document title", Type: "text"}}}),
		defaultQueued(Capability{Provider: ProviderLinkedIn, Profile: models.ContentProfileShortVideo, Label: "LinkedIn video", TextLimit: 3000, Media: shortVideo}),
		defaultQueued(Capability{Provider: ProviderLinkedIn, Profile: models.ContentProfileLongVideo, Label: "LinkedIn long video", TextLimit: 3000, Media: longVideo}),

		defaultQueued(Capability{Provider: ProviderFacebook, Profile: models.ContentProfileShortText, Label: "Facebook Page post", TextLimit: 63206, Media: text, Settings: facebookSettings()}),
		defaultQueued(Capability{Provider: ProviderFacebook, Profile: models.ContentProfileLinkShare, Label: "Facebook Page link", TextLimit: 63206, Media: text, Settings: facebookSettings()}),
		defaultQueued(Capability{Provider: ProviderFacebook, Profile: models.ContentProfileImagePost, Label: "Facebook Page photo", TextLimit: 63206, Media: publicImages, RequiresPublicMedia: true, Settings: facebookSettings()}),
		defaultQueued(Capability{Provider: ProviderFacebook, Profile: models.ContentProfileCarousel, Label: "Facebook multi-photo", TextLimit: 63206, Media: publicImages, RequiresPublicMedia: true, Settings: facebookSettings()}),
		defaultQueued(Capability{Provider: ProviderFacebook, Profile: models.ContentProfileStory, Label: "Facebook Page Story", Media: publicImages, RequiresPublicMedia: true, RequiresAppReview: true, Settings: facebookSettings()}),
		defaultQueued(Capability{Provider: ProviderFacebook, Profile: models.ContentProfileShortVideo, Label: "Facebook Reel/video", TextLimit: 63206, Media: publicShortVideo, RequiresPublicMedia: true, Settings: facebookSettings()}),
		defaultQueued(Capability{Provider: ProviderFacebook, Profile: models.ContentProfileLongVideo, Label: "Facebook video", TextLimit: 63206, Media: longVideo, RequiresPublicMedia: true, Settings: facebookSettings()}),

		defaultQueued(Capability{Provider: ProviderInstagram, Profile: models.ContentProfileImagePost, Label: "Instagram feed", TextLimit: 2200, Media: publicImages, RequiresPublicMedia: true, Settings: instagramSettings()}),
		defaultQueued(Capability{Provider: ProviderInstagram, Profile: models.ContentProfileCarousel, Label: "Instagram carousel", TextLimit: 2200, Media: publicImages, RequiresPublicMedia: true, Settings: instagramSettings()}),
		defaultQueued(Capability{Provider: ProviderInstagram, Profile: models.ContentProfileStory, Label: "Instagram Story", Media: publicImages, RequiresPublicMedia: true, RequiresAppReview: true, Settings: instagramSettings()}),
		defaultQueued(Capability{Provider: ProviderInstagram, Profile: models.ContentProfileShortVideo, Label: "Instagram Reel", TextLimit: 2200, Media: publicShortVideo, RequiresPublicMedia: true, Settings: instagramSettings()}),

		defaultQueued(Capability{Provider: ProviderYouTube, Profile: models.ContentProfileShortVideo, Label: "YouTube Short", TitleRequired: true, DescriptionRequired: false, Media: shortVideo, Settings: youtubeSettings(), Caveats: []string{"Unaudited Google projects can force uploads private."}}),
		defaultQueued(Capability{Provider: ProviderYouTube, Profile: models.ContentProfileLongVideo, Label: "YouTube video", TitleRequired: true, DescriptionRequired: false, Media: longVideo, Settings: youtubeSettings(), Caveats: []string{"Unaudited Google projects can force uploads private."}}),

		defaultQueued(Capability{Provider: ProviderTikTok, Profile: models.ContentProfileShortVideo, Label: "TikTok video", TextLimit: 2200, Media: publicShortVideo, RequiresPublicMedia: true, RequiresAppReview: true, Settings: tiktokSettings()}),
		defaultQueued(Capability{Provider: ProviderTikTok, Profile: models.ContentProfileCarousel, Label: "TikTok photo post", TextLimit: 2200, Media: publicImages, RequiresPublicMedia: true, RequiresAppReview: true, Settings: tiktokSettings()}),
	}
}

func Find(provider, profile string) (Capability, bool) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	profile = strings.TrimSpace(profile)
	for _, capability := range All() {
		if capability.Provider == provider && capability.Profile == profile {
			return capability, true
		}
	}
	return Capability{}, false
}

func ForProfile(profile string) []Capability {
	out := []Capability{}
	for _, capability := range All() {
		if capability.Profile == profile {
			out = append(out, capability)
		}
	}
	return out
}

func Validate(provider, profile, body, title, description string, media []MediaItem, settings map[string]any) []ValidationIssue {
	capability, ok := Find(provider, profile)
	if !ok {
		return []ValidationIssue{{Severity: "error", Code: "unsupported_profile", Message: fmt.Sprintf("%s does not support %s", provider, profile), Provider: provider, Profile: profile}}
	}

	issues := []ValidationIssue{}
	if capability.TextLimit > 0 && len([]rune(body)) > capability.TextLimit {
		issues = append(issues, ValidationIssue{Severity: "error", Code: "text_too_long", Message: fmt.Sprintf("Text is over the %d character limit", capability.TextLimit), Provider: provider, Profile: profile, Field: "body"})
	}
	if capability.TitleRequired && strings.TrimSpace(title) == "" {
		issues = append(issues, ValidationIssue{Severity: "error", Code: "title_required", Message: "Title is required", Provider: provider, Profile: profile, Field: "title"})
	}
	if capability.DescriptionRequired && strings.TrimSpace(description) == "" {
		issues = append(issues, ValidationIssue{Severity: "error", Code: "description_required", Message: "Description is required", Provider: provider, Profile: profile, Field: "description"})
	}
	if len(media) < capability.Media.MinCount || len(media) > capability.Media.MaxCount {
		issues = append(issues, ValidationIssue{Severity: "error", Code: "media_count", Message: fmt.Sprintf("Requires %d-%d media item(s)", capability.Media.MinCount, capability.Media.MaxCount), Provider: provider, Profile: profile, Field: "media"})
	}
	for _, item := range media {
		issues = append(issues, validateMediaItem(capability, item)...)
	}
	issues = append(issues, validateUnsupportedSettings(capability, settings)...)
	for _, field := range capability.Settings {
		if field.Required && settingsValue(settings, field.Key) == "" {
			issues = append(issues, ValidationIssue{Severity: "error", Code: "setting_required", Message: fmt.Sprintf("%s is required", field.Label), Provider: provider, Profile: profile, Field: field.Key})
		}
	}
	issues = append(issues, validateProviderOperationalReadiness(capability)...)
	issues = append(issues, validateProviderSettings(provider, profile, len(media), settings)...)
	return issues
}

func validateUnsupportedSettings(capability Capability, settings map[string]any) []ValidationIssue {
	if len(settings) == 0 {
		return nil
	}
	known := map[string]struct{}{}
	for _, field := range capability.Settings {
		known[field.Key] = struct{}{}
	}
	issues := []ValidationIssue{}
	for key, value := range settings {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok := known[key]; ok {
			continue
		}
		if strings.TrimSpace(fmt.Sprint(value)) == "" {
			continue
		}
		issues = append(issues, ValidationIssue{
			Severity: "error",
			Code:     "unsupported_setting",
			Message:  fmt.Sprintf("%s is not supported for %s", key, capability.Label),
			Provider: capability.Provider,
			Profile:  capability.Profile,
			Field:    key,
		})
	}
	return issues
}

func validateProviderOperationalReadiness(capability Capability) []ValidationIssue {
	issues := []ValidationIssue{}
	if capability.RequiresAppReview {
		issues = append(issues, ValidationIssue{
			Severity: "error",
			Code:     "app_review_required",
			Message:  "Provider app review is required before this profile can be published.",
			Provider: capability.Provider,
			Profile:  capability.Profile,
		})
	}
	switch capability.Provider {
	case ProviderTikTok, ProviderYouTube:
		issues = append(issues, ValidationIssue{
			Severity: "error",
			Code:     "provider_audit_required",
			Message:  "Provider audit or production access must be confirmed before publishing.",
			Provider: capability.Provider,
			Profile:  capability.Profile,
		})
	}
	switch capability.Provider {
	case ProviderX, ProviderYouTube:
		issues = append(issues, ValidationIssue{
			Severity: "warning",
			Code:     "quota_warning",
			Message:  "Provider quota or account tier limits may block publishing.",
			Provider: capability.Provider,
			Profile:  capability.Profile,
		})
	}
	return issues
}

func validateProviderSettings(provider, profile string, mediaCount int, settings map[string]any) []ValidationIssue {
	switch provider {
	case ProviderX:
		return validateXSettings(profile, mediaCount, settings)
	case ProviderMastodon:
		return validateMastodonSettings(profile, mediaCount, settings)
	default:
		return nil
	}
}

func validateXSettings(profile string, mediaCount int, settings map[string]any) []ValidationIssue {
	attachmentKinds := 0
	if mediaCount > 0 {
		attachmentKinds++
	}
	if settingsValue(settings, "quote_tweet_id") != "" {
		attachmentKinds++
	}
	if settingsValue(settings, "poll_options") != "" {
		attachmentKinds++
	}
	if attachmentKinds <= 1 {
		return nil
	}
	return []ValidationIssue{{
		Severity: "error",
		Code:     "x_mutually_exclusive_attachment",
		Message:  "X posts can include only one of media, poll, or quote post.",
		Provider: ProviderX,
		Profile:  profile,
		Field:    "settings",
	}}
}

func validateMastodonSettings(profile string, mediaCount int, settings map[string]any) []ValidationIssue {
	if mediaCount == 0 || settingsValue(settings, "poll_options") == "" {
		return nil
	}
	return []ValidationIssue{{
		Severity: "error",
		Code:     "mastodon_poll_media_conflict",
		Message:  "Mastodon polls cannot be combined with media attachments.",
		Provider: ProviderMastodon,
		Profile:  profile,
		Field:    "poll_options",
	}}
}

//nolint:gocyclo
func validateMediaItem(capability Capability, item MediaItem) []ValidationIssue {
	issues := []ValidationIssue{}
	if len(capability.Media.AllowedMIMEs) > 0 && !mimeAllowed(item.MimeType, capability.Media.AllowedMIMEs) {
		issues = append(issues, ValidationIssue{Severity: "error", Code: "media_mime", Message: fmt.Sprintf("%s is not accepted for %s", item.MimeType, capability.Label), Provider: capability.Provider, Profile: capability.Profile, MediaID: item.ID})
	}
	if capability.Media.MaxSizeBytes > 0 && item.Size > capability.Media.MaxSizeBytes {
		issues = append(issues, ValidationIssue{Severity: "error", Code: "media_size", Message: "Media file is too large", Provider: capability.Provider, Profile: capability.Profile, MediaID: item.ID})
	}
	if strings.HasPrefix(item.MimeType, "video/") && item.AnalysisStatus != "" && item.AnalysisStatus != "ready" {
		issues = append(issues, ValidationIssue{Severity: "error", Code: "media_analysis_pending", Message: "Video analysis must finish before scheduling or publishing", Provider: capability.Provider, Profile: capability.Profile, MediaID: item.ID})
	}
	if strings.HasPrefix(item.MimeType, "video/") && item.AnalysisStatus == "failed" {
		issues = append(issues, ValidationIssue{Severity: "error", Code: "media_analysis_failed", Message: firstNonEmpty(item.AnalysisError, "Video analysis failed"), Provider: capability.Provider, Profile: capability.Profile, MediaID: item.ID})
	}
	if capability.Media.MaxDurationSeconds > 0 && item.DurationMS > int64(capability.Media.MaxDurationSeconds)*1000 {
		issues = append(issues, ValidationIssue{Severity: "error", Code: "media_duration", Message: fmt.Sprintf("Video must be %d seconds or less", capability.Media.MaxDurationSeconds), Provider: capability.Provider, Profile: capability.Profile, MediaID: item.ID})
	}
	if len(capability.Media.AspectRatios) > 0 && item.Width > 0 && item.Height > 0 && !ratioAllowed(item.Width, item.Height, capability.Media.AspectRatios) {
		issues = append(issues, ValidationIssue{Severity: "warning", Code: "media_aspect", Message: "Media should be vertical or square for this profile", Provider: capability.Provider, Profile: capability.Profile, MediaID: item.ID})
	}
	if capability.Media.RequiresPublicURL && !item.PublicURLReady {
		issues = append(issues, ValidationIssue{Severity: "error", Code: "public_url_unreachable", Message: firstNonEmpty(item.PublicURLError, "This provider needs media that is publicly reachable over HTTPS"), Provider: capability.Provider, Profile: capability.Profile, MediaID: item.ID})
	}
	if capability.Media.RequiresHTTPSFetchable && item.URL != "" {
		parsed, err := url.Parse(item.URL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			issues = append(issues, ValidationIssue{Severity: "error", Code: "https_media_required", Message: "Public media URL must be HTTPS", Provider: capability.Provider, Profile: capability.Profile, MediaID: item.ID})
		}
	}
	return issues
}

func validationCategories(c Capability) []string {
	categories := []string{"media_count", "mime"}
	if c.Media.MaxDurationSeconds > 0 {
		categories = append(categories, "duration")
	}
	if len(c.Media.AspectRatios) > 0 {
		categories = append(categories, "aspect")
	}
	if acceptsDocument(c.Media.AllowedMIMEs) {
		categories = append(categories, "document")
	}
	if c.RequiresPublicMedia || c.Media.RequiresPublicURL {
		categories = append(categories, "public_url")
	}
	if c.TitleRequired {
		categories = append(categories, "title")
	}
	if c.DescriptionRequired {
		categories = append(categories, "description")
	}
	for _, setting := range c.Settings {
		if setting.Type == "media" && strings.Contains(setting.Key, "thumbnail") {
			categories = append(categories, "thumbnail")
			break
		}
	}
	if c.RequiresAppReview {
		categories = append(categories, "app_review")
	}
	return categories
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func mimeAllowed(mimeType string, allowed []string) bool {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	for _, candidate := range allowed {
		candidate = strings.ToLower(candidate)
		if candidate == mimeType {
			return true
		}
		if strings.HasSuffix(candidate, "/*") && strings.HasPrefix(mimeType, strings.TrimSuffix(candidate, "*")) {
			return true
		}
	}
	return false
}

func ratioAllowed(width, height int, ratios []string) bool {
	actual := float64(width) / float64(height)
	for _, ratio := range ratios {
		switch ratio {
		case "9:16":
			if math.Abs(actual-(9.0/16.0)) < 0.08 {
				return true
			}
		case "1:1":
			if math.Abs(actual-1) < 0.08 {
				return true
			}
		case "16:9":
			if math.Abs(actual-(16.0/9.0)) < 0.08 {
				return true
			}
		}
	}
	return false
}

func acceptsDocument(allowed []string) bool {
	for _, mimeType := range allowed {
		switch strings.ToLower(strings.TrimSpace(mimeType)) {
		case "application/pdf",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/vnd.ms-powerpoint",
			"application/vnd.openxmlformats-officedocument.presentationml.presentation":
			return true
		}
	}
	return false
}

func settingsValue(settings map[string]any, key string) string {
	if settings == nil {
		return ""
	}
	if value, ok := settings[key]; ok {
		return strings.TrimSpace(fmt.Sprint(value))
	}
	return ""
}

func linkSettings() []SettingField {
	return []SettingField{{Key: "url", Label: "URL", Type: "url", Required: true}}
}

func xSettings() []SettingField {
	return []SettingField{
		{Key: "quote_tweet_id", Label: "Quote post ID", Type: "text"},
		{Key: "poll_options", Label: "Poll options", Type: "textarea", Help: "One option per line. Polls cannot be combined with media or quote posts."},
		{Key: "poll_duration_minutes", Label: "Poll duration", Type: "number"},
		{Key: "reply_settings", Label: "Who can reply", Type: "select", Options: []string{"following", "mentionedUsers", "subscribers", "verified"}},
		{Key: "paid_partnership", Label: "Paid partnership", Type: "boolean"},
		{Key: "made_with_ai", Label: "Made with AI", Type: "boolean"},
	}
}

func mastodonSettings() []SettingField {
	return []SettingField{
		{Key: "visibility", Label: "Visibility", Type: "select", Options: []string{"public", "unlisted", "private", "direct"}},
		{Key: "spoiler_text", Label: "Content warning", Type: "text"},
		{Key: "sensitive", Label: "Sensitive media", Type: "boolean"},
		{Key: "language", Label: "Language", Type: "text", Help: "ISO 639 language code."},
		{Key: "scheduled_at", Label: "Native schedule time", Type: "text", Help: "RFC3339 time at least 5 minutes in the future."},
		{Key: "poll_options", Label: "Poll options", Type: "textarea", Help: "One option per line. Polls cannot be combined with media."},
		{Key: "poll_expires_in_seconds", Label: "Poll duration", Type: "number"},
		{Key: "poll_multiple", Label: "Multiple choice", Type: "boolean"},
		{Key: "poll_hide_totals", Label: "Hide totals", Type: "boolean"},
	}
}

func blueskySettings() []SettingField {
	return []SettingField{
		{Key: "link_url", Label: "Link card URL", Type: "url"},
		{Key: "link_title", Label: "Link card title", Type: "text"},
		{Key: "link_description", Label: "Link card description", Type: "textarea"},
		{Key: "quote_uri", Label: "Quote AT URI", Type: "text"},
		{Key: "quote_cid", Label: "Quote CID", Type: "text"},
		{Key: "self_labels", Label: "Self-labels", Type: "text", Help: "Comma-separated Bluesky self-labels: porn, sexual, nudity, graphic-media, bot, !no-unauthenticated."},
		{Key: "mention_dids", Label: "Mention DIDs", Type: "textarea", Help: "Optional handle=DID mappings for mention facets, one per line."},
	}
}

func facebookSettings() []SettingField {
	return []SettingField{
		{Key: "post_type", Label: "Post type", Type: "select", Options: []string{"post", "story"}},
		{Key: "url", Label: "URL", Type: "url"},
		{Key: "text_format_preset_id", Label: "Text background preset", Type: "text", Help: "Facebook text background preset ID for short text posts."},
		{Key: "first_comment", Label: "First comment", Type: "textarea"},
	}
}

func instagramSettings() []SettingField {
	return []SettingField{
		{Key: "post_type", Label: "Post type", Type: "select", Options: []string{"post", "story"}},
		{Key: "is_reel", Label: "Publish as Reel", Type: "boolean"},
		{Key: "is_trial_reel", Label: "Trial Reel", Type: "boolean"},
		{Key: "graduation_strategy", Label: "Trial graduation", Type: "select", Options: []string{"manual", "automatic"}},
		{Key: "collaborators", Label: "Collaborators", Type: "text"},
		{Key: "audio", Label: "Audio settings", Type: "json"},
		{Key: "cover_media_id", Label: "Cover media", Type: "media"},
		{Key: "thumbnail_timestamp_ms", Label: "Thumbnail time", Type: "number"},
	}
}

func youtubeSettings() []SettingField {
	return []SettingField{
		{Key: "privacy", Label: "Privacy", Type: "select", Required: true, Options: []string{"public", "unlisted", "private"}},
		{Key: "title", Label: "Title", Type: "text"},
		{Key: "description", Label: "Description", Type: "textarea"},
		{Key: "tags", Label: "Tags", Type: "text"},
		{Key: "category_id", Label: "Category", Type: "text"},
		{Key: "playlist_id", Label: "Playlist", Type: "text"},
		{Key: "thumbnail_media_id", Label: "Thumbnail", Type: "media"},
		{Key: "self_declared_made_for_kids", Label: "Made for kids", Type: "boolean"},
		{Key: "contains_synthetic_media", Label: "Synthetic media", Type: "boolean"},
		{Key: "notify_subscribers", Label: "Notify subscribers", Type: "boolean"},
	}
}

func tiktokSettings() []SettingField {
	return []SettingField{
		{Key: "content_posting_method", Label: "Posting method", Type: "select", Required: true, Options: []string{"DIRECT_POST", "UPLOAD"}},
		{Key: "privacy_level", Label: "Privacy", Type: "select", Required: true, Options: []string{"PUBLIC_TO_EVERYONE", "MUTUAL_FOLLOW_FRIENDS", "FOLLOWER_OF_CREATOR", "SELF_ONLY"}},
		{Key: "duet", Label: "Allow Duet", Type: "boolean"},
		{Key: "stitch", Label: "Allow Stitch", Type: "boolean"},
		{Key: "comment", Label: "Allow comments", Type: "boolean"},
		{Key: "auto_add_music", Label: "Auto-add music", Type: "boolean"},
		{Key: "brand_content_toggle", Label: "Branded content", Type: "boolean"},
		{Key: "brand_organic_toggle", Label: "Brand organic", Type: "boolean"},
		{Key: "is_aigc", Label: "AI-generated content", Type: "boolean"},
		{Key: "cover_timestamp_ms", Label: "Cover time", Type: "number"},
	}
}
