package platform

import (
	"testing"
	"time"
)

type consumeOnceXRequestStore struct {
	meta     XRequestMeta
	consume  int
	consumed bool
}

func (s *consumeOnceXRequestStore) Save(_, _, _, _ string, _ time.Time) error {
	return nil
}

func (s *consumeOnceXRequestStore) Consume(_ string, _ time.Duration) (XRequestMeta, bool, error) {
	s.consume++
	if s.consumed {
		return XRequestMeta{}, false, nil
	}
	s.consumed = true
	return s.meta, true, nil
}

func TestXWorkspaceLookupRetainsRequestMetaForTokenExchange(t *testing.T) {
	adapter := NewXAdapter("client-id", "client-secret", "https://app.example/api/v1/accounts/x/callback")
	close(adapter.cleanupDone)

	store := &consumeOnceXRequestStore{meta: XRequestMeta{
		Secret:      "request-secret",
		WorkspaceID: "workspace-1",
		UserID:      "user-1",
		CreatedAt:   time.Now().UTC(),
	}}
	adapter.SetRequestStore(store)

	workspaceID, ok := adapter.GetWorkspaceIDForRequestToken("request-token")
	if !ok {
		t.Fatal("expected workspace lookup to succeed")
	}
	if workspaceID != "workspace-1" {
		t.Fatalf("expected workspace-1, got %q", workspaceID)
	}

	metaRaw, ok := adapter.requestMeta.Load("request-token")
	if !ok {
		t.Fatal("expected consumed request token metadata to be retained for exchange")
	}
	meta := metaRaw.(XRequestMeta)
	if meta.Secret != "request-secret" || meta.UserID != "user-1" {
		t.Fatalf("unexpected retained metadata: %#v", meta)
	}
}
