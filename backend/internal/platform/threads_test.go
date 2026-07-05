package platform

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestThreadsListCommentsMapsReplies(t *testing.T) {
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet || req.URL.String() != "https://graph.threads.net/v1.0/thread-1/replies?fields=id%2Ctext%2Cusername%2Ctimestamp%2Chide_status&access_token=threads-token" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.String())
		}
		return jsonResponse(req, `{"data":[{"id":"reply-1","text":"Nice thread","username":"rita","timestamp":"2026-07-04T10:00:00+0000","hide_status":"HIDDEN"}]}`), nil
	})}

	comments, err := NewThreadsAdapter("", "", "").ListComments(context.Background(), "threads-token", "user-1", "thread-1")
	if err != nil {
		t.Fatalf("ListComments returned error: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("expected one comment, got %#v", comments)
	}
	comment := comments[0]
	if comment.ID != "reply-1" || comment.AuthorName != "rita" || comment.Text != "Nice thread" || !comment.Hidden || !comment.CanReply || !comment.CanHide || comment.CanDelete {
		t.Fatalf("unexpected comment mapping: %#v", comment)
	}
}

func TestThreadsReplyAndHideComment(t *testing.T) {
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodPost && req.URL.String() == "https://graph.threads.net/v1.0/user-1/threads":
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading reply body: %v", err)
			}
			form, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatalf("parsing reply body: %v", err)
			}
			if form.Get(jsonFieldText) != "Thanks" || form.Get("reply_to_id") != "reply-1" || form.Get("media_type") != "TEXT" || form.Get(oauthParamAccessToken) != "threads-token" {
				t.Fatalf("unexpected reply form %#v", form)
			}
			return jsonResponse(req, `{"id":"creation-1"}`), nil
		case req.Method == http.MethodGet && req.URL.String() == "https://graph.threads.net/v1.0/creation-1?fields=status,error_message":
			if req.Header.Get(headerAuthorization) != bearerPrefix+"threads-token" {
				t.Fatalf("unexpected status auth header %q", req.Header.Get(headerAuthorization))
			}
			return jsonResponse(req, `{"status":"FINISHED"}`), nil
		case req.Method == http.MethodPost && req.URL.String() == "https://graph.threads.net/v1.0/user-1/threads_publish":
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading publish body: %v", err)
			}
			form, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatalf("parsing publish body: %v", err)
			}
			if form.Get("creation_id") != "creation-1" || form.Get(oauthParamAccessToken) != "threads-token" {
				t.Fatalf("unexpected publish form %#v", form)
			}
			return jsonResponse(req, `{"id":"reply-post-1"}`), nil
		case req.Method == http.MethodPost && req.URL.String() == "https://graph.threads.net/v1.0/reply-1/manage_reply":
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading hide body: %v", err)
			}
			form, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatalf("parsing hide body: %v", err)
			}
			if form.Get("hide") != "true" || form.Get(oauthParamAccessToken) != "threads-token" {
				t.Fatalf("unexpected hide form %#v", form)
			}
			return jsonResponse(req, `{"success":true}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})}

	adapter := NewThreadsAdapter("", "", "")
	replyID, err := adapter.ReplyToComment(context.Background(), "threads-token", "user-1", "reply-1", " Thanks ")
	if err != nil {
		t.Fatalf("ReplyToComment returned error: %v", err)
	}
	if replyID != "reply-post-1" {
		t.Fatalf("expected reply post ID, got %q", replyID)
	}
	if err := adapter.HideComment(context.Background(), "threads-token", "user-1", "reply-1"); err != nil {
		t.Fatalf("HideComment returned error: %v", err)
	}
}

func TestThreadsDeleteCommentUnsupported(t *testing.T) {
	err := NewThreadsAdapter("", "", "").DeleteComment(context.Background(), "threads-token", "user-1", "reply-1")
	if !errors.Is(err, ErrUnsupportedCommentAction) {
		t.Fatalf("expected unsupported comment action, got %v", err)
	}
}
