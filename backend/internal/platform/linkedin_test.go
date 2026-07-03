package platform

import (
	"net/url"
	"strings"
	"testing"
)

func TestLinkedInGenerateAuthURLEncodesScopesWithPercentSpaces(t *testing.T) {
	adapter := NewLinkedInAdapter("client-id", "client-secret", "https://app.example/api/v1/accounts/linkedin/callback", true)

	authURL, _ := adapter.GenerateAuthURL("state-123")
	if strings.Contains(authURL, "scope=openid+profile+w_member_social") {
		t.Fatalf("linkedin auth URL used + for scope spaces: %s", authURL)
	}
	if !strings.Contains(authURL, "scope=openid%20profile%20w_member_social") {
		t.Fatalf("linkedin auth URL did not percent-encode scope spaces: %s", authURL)
	}

	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parsing auth URL: %v", err)
	}
	if parsed.Query().Get("scope") != "openid profile w_member_social" {
		t.Fatalf("unexpected parsed scope %q", parsed.Query().Get("scope"))
	}
}
