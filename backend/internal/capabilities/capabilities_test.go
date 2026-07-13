package capabilities

import (
	"testing"

	"github.com/openpost/backend/internal/models"
	"github.com/stretchr/testify/require"
)

func TestValidateBlocksMissingMediaAnalysisForVideoProfiles(t *testing.T) {
	issues := Validate(ProviderTikTok, models.ContentProfileShortVideo, "caption", "", "", []MediaItem{{
		ID:             "video-1",
		MimeType:       "video/mp4",
		Size:           1024,
		AnalysisStatus: "pending",
	}}, map[string]any{"content_posting_method": "DIRECT_POST", "privacy_level": "SELF_ONLY"})

	requireIssueCode(t, issues, "media_analysis_pending")
}

func TestValidateBlocksFailedPublicURLVerification(t *testing.T) {
	issues := Validate(ProviderInstagram, models.ContentProfileShortVideo, "caption", "", "", []MediaItem{{
		ID:              "video-1",
		MimeType:        "video/mp4",
		Size:            1024,
		Width:           1080,
		Height:          1920,
		DurationMS:      20_000,
		AnalysisStatus:  "ready",
		PublicURLReady:  false,
		PublicURLError:  "403 forbidden",
		PublicURLStatus: 403,
		URL:             "https://cdn.example/video.mp4",
	}}, map[string]any{})

	requireIssueCode(t, issues, "public_url_unreachable")
}

func TestCapabilitiesExposeValidationCategories(t *testing.T) {
	capability, ok := Find(ProviderYouTube, models.ContentProfileLongVideo)

	require.True(t, ok)
	require.Contains(t, capability.ValidationCategories, "duration")
	require.Contains(t, capability.ValidationCategories, "title")
	require.Contains(t, capability.ValidationCategories, "thumbnail")
}

func TestXCapabilitiesExposePostSettings(t *testing.T) {
	capability, ok := Find(ProviderX, models.ContentProfileShortText)

	require.True(t, ok)
	keys := make([]string, 0, len(capability.Settings))
	for _, setting := range capability.Settings {
		keys = append(keys, setting.Key)
	}
	require.Contains(t, keys, "quote_tweet_id")
	require.Contains(t, keys, "poll_options")
	require.Contains(t, keys, "poll_duration_minutes")
	require.Contains(t, keys, "reply_settings")
	require.Contains(t, keys, "paid_partnership")
	require.Contains(t, keys, "made_with_ai")
}

func TestValidateBlocksXMutuallyExclusiveSettings(t *testing.T) {
	issues := Validate(ProviderX, models.ContentProfileImagePost, "caption", "", "", []MediaItem{{
		ID:       "image-1",
		MimeType: "image/jpeg",
		Size:     1024,
	}}, map[string]any{"poll_options": "One\nTwo"})

	requireIssueCode(t, issues, "x_mutually_exclusive_attachment")

	issues = Validate(ProviderX, models.ContentProfileImagePost, "caption", "", "", []MediaItem{{
		ID:       "image-1",
		MimeType: "image/jpeg",
		Size:     1024,
	}}, map[string]any{"quote_tweet_id": "1346889436626259968"})

	requireIssueCode(t, issues, "x_mutually_exclusive_attachment")
}

func TestMastodonCapabilitiesExposeStatusSettings(t *testing.T) {
	capability, ok := Find(ProviderMastodon, models.ContentProfileShortText)

	require.True(t, ok)
	require.True(t, capability.NativeScheduling)
	keys := make([]string, 0, len(capability.Settings))
	for _, setting := range capability.Settings {
		keys = append(keys, setting.Key)
	}
	require.Contains(t, keys, "visibility")
	require.Contains(t, keys, "spoiler_text")
	require.Contains(t, keys, "sensitive")
	require.Contains(t, keys, "language")
	require.Contains(t, keys, "scheduled_at")
	require.Contains(t, keys, "poll_options")
	require.Contains(t, keys, "poll_expires_in_seconds")
}

func TestBlueskyCapabilitiesExposeVideoAndPostSettings(t *testing.T) {
	capability, ok := Find(ProviderBluesky, models.ContentProfileShortVideo)

	require.True(t, ok)
	require.Equal(t, 1, capability.Media.MinCount)
	require.Equal(t, 1, capability.Media.MaxCount)
	require.Contains(t, capability.Media.AllowedMIMEs, "video/mp4")
	require.Equal(t, int64(100*1024*1024), capability.Media.MaxSizeBytes)

	keys := make([]string, 0, len(capability.Settings))
	for _, setting := range capability.Settings {
		keys = append(keys, setting.Key)
	}
	require.Contains(t, keys, "link_url")
	require.Contains(t, keys, "quote_uri")
	require.Contains(t, keys, "quote_cid")
	require.Contains(t, keys, "self_labels")
	require.Contains(t, keys, "mention_dids")
}

func TestLinkedInCapabilitiesExposeDocumentCarousel(t *testing.T) {
	capability, ok := Find(ProviderLinkedIn, models.ContentProfileCarousel)

	require.True(t, ok)
	require.Equal(t, 1, capability.Media.MinCount)
	require.Equal(t, 1, capability.Media.MaxCount)
	require.Contains(t, capability.Media.AllowedMIMEs, "application/pdf")
	require.Contains(t, capability.Media.AllowedMIMEs, "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	require.Equal(t, int64(100*1024*1024), capability.Media.MaxSizeBytes)
	require.Contains(t, capability.ValidationCategories, "document")
}

func TestPublicMediaCountsMatchAdapterPublishingModes(t *testing.T) {
	tests := []struct {
		provider string
		profile  string
		min      int
		max      int
	}{
		{ProviderThreads, models.ContentProfileImagePost, 1, 1},
		{ProviderThreads, models.ContentProfileCarousel, 2, 10},
		{ProviderFacebook, models.ContentProfileImagePost, 1, 1},
		{ProviderFacebook, models.ContentProfileCarousel, 2, 10},
		{ProviderFacebook, models.ContentProfileStory, 1, 1},
		{ProviderInstagram, models.ContentProfileImagePost, 1, 1},
		{ProviderInstagram, models.ContentProfileCarousel, 2, 10},
		{ProviderTikTok, models.ContentProfileCarousel, 1, 35},
	}

	for _, tt := range tests {
		t.Run(tt.provider+"/"+tt.profile, func(t *testing.T) {
			capability, ok := Find(tt.provider, tt.profile)
			require.True(t, ok)
			require.Equal(t, tt.min, capability.Media.MinCount)
			require.Equal(t, tt.max, capability.Media.MaxCount)
		})
	}
}

func TestTikTokPhotoCapabilityMatchesDocumentedMediaLimits(t *testing.T) {
	capability, ok := Find(ProviderTikTok, models.ContentProfileCarousel)

	require.True(t, ok)
	require.Equal(t, 4000, capability.TextLimit)
	require.Equal(t, int64(20*1024*1024), capability.Media.MaxSizeBytes)
	require.ElementsMatch(t, []string{"image/jpeg", "image/webp"}, capability.Media.AllowedMIMEs)
}

func TestThreadsCarouselCapabilityAllowsMixedMedia(t *testing.T) {
	capability, ok := Find(ProviderThreads, models.ContentProfileCarousel)

	require.True(t, ok)
	require.Contains(t, capability.Media.AllowedMIMEs, "image/jpeg")
	require.Contains(t, capability.Media.AllowedMIMEs, "video/mp4")
}

func TestValidateBlocksMastodonPollWithMedia(t *testing.T) {
	issues := Validate(ProviderMastodon, models.ContentProfileImagePost, "caption", "", "", []MediaItem{{
		ID:       "image-1",
		MimeType: "image/jpeg",
		Size:     1024,
	}}, map[string]any{"poll_options": "One\nTwo"})

	requireIssueCode(t, issues, "mastodon_poll_media_conflict")
}

func TestValidateFlagsUnsupportedProviderSettings(t *testing.T) {
	issues := Validate(ProviderYouTube, models.ContentProfileLongVideo, "caption", "Title", "", []MediaItem{{
		ID:             "video-1",
		MimeType:       "video/mp4",
		Size:           1024,
		AnalysisStatus: "ready",
	}}, map[string]any{"privacy": "private", "unsupported_field": "value"})

	requireIssueCode(t, issues, "unsupported_setting")
}

func TestValidateEmitsProviderReviewAuditAndQuotaCodes(t *testing.T) {
	tiktokIssues := Validate(ProviderTikTok, models.ContentProfileShortVideo, "caption", "", "", []MediaItem{{
		ID:              "video-1",
		MimeType:        "video/mp4",
		Size:            1024,
		AnalysisStatus:  "ready",
		PublicURLReady:  true,
		PublicURLStatus: 200,
		URL:             "https://cdn.example/video.mp4",
	}}, map[string]any{"content_posting_method": "DIRECT_POST", "privacy_level": "SELF_ONLY"})

	requireIssueCode(t, tiktokIssues, "app_review_required")
	requireIssueCode(t, tiktokIssues, "provider_audit_required")

	youtubeIssues := Validate(ProviderYouTube, models.ContentProfileLongVideo, "caption", "Title", "", []MediaItem{{
		ID:             "video-1",
		MimeType:       "video/mp4",
		Size:           1024,
		AnalysisStatus: "ready",
	}}, map[string]any{"privacy": "private"})

	requireIssueCode(t, youtubeIssues, "provider_audit_required")
	requireIssueCode(t, youtubeIssues, "quota_warning")
}

func requireIssueCode(t *testing.T, issues []ValidationIssue, code string) {
	t.Helper()
	for _, issue := range issues {
		if issue.Code == code {
			return
		}
	}
	require.Failf(t, "missing validation issue", "code %q not found in %#v", code, issues)
}
