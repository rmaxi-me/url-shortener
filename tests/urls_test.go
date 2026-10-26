package tests_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/http/httptest"
	"rmaxi-me/url-shortner/src/handlers"
	"rmaxi-me/url-shortner/src/models"
	"rmaxi-me/url-shortner/src/stores"
	"strings"
	"testing"
)

type fakeStore struct {
	urls     map[string]*models.ShortUrl
	idLength int

	createErr error
	// collisions is how many Create calls fail with ErrSourceAlreadyExists
	// before one succeeds.
	collisions int
	resolveErr error
	deleteErr  error

	createCalls  []models.ShortUrl
	resolveCalls int
	deleteCalls  int
}

var _ stores.Store = (*fakeStore)(nil)

func newFakeStore() *fakeStore {
	return &fakeStore{urls: map[string]*models.ShortUrl{}, idLength: 8}
}

func (f *fakeStore) seed(source, target, deleteKey string) {
	f.urls[source] = &models.ShortUrl{Source: source, Target: target, DeleteKey: deleteKey}
}

func (f *fakeStore) Create(_ context.Context, shortUrl *models.ShortUrl) (*models.ShortUrl, error) {
	f.createCalls = append(f.createCalls, *shortUrl)
	if f.createErr != nil {
		return nil, fmt.Errorf("fake store: %w", f.createErr)
	}

	if _, exists := f.urls[shortUrl.Source]; exists || len(f.createCalls) <= f.collisions {
		return nil, fmt.Errorf("fake store: %w", models.ErrSourceAlreadyExists)
	}

	stored := *shortUrl
	stored.ID = uint(len(f.urls) + 1)
	f.urls[stored.Source] = &stored
	return &stored, nil
}

func (f *fakeStore) GetIdLength() int {
	return f.idLength
}

func (f *fakeStore) Resolve(ctx context.Context, source string, incrementServe bool) (*models.ShortUrl, error) {
	f.resolveCalls++
	if f.resolveErr != nil {
		return nil, fmt.Errorf("fake store: %w", f.resolveErr)
	}

	url, ok := f.urls[source]
	if !ok {
		return nil, fmt.Errorf("fake store: %w", models.ErrSourceNotFound)
	}
	url.ServeCount++
	return url, nil
}

func (f *fakeStore) Delete(_ context.Context, source, deleteKey string) error {
	f.deleteCalls++
	if f.deleteErr != nil {
		return fmt.Errorf("fake store: %w", f.deleteErr)
	}

	url, ok := f.urls[source]
	if !ok {
		return fmt.Errorf("fake store: %w", models.ErrSourceNotFound)
	}
	if url.DeleteKey != deleteKey {
		return fmt.Errorf("fake store: %w", models.ErrInvalidDeleteKey)
	}
	delete(f.urls, source)
	return nil
}

func serve(store stores.Store, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handlers.RegisterRoutes(store).ServeHTTP(rec, req)
	return rec
}

func newCreateRequest(contentType, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return req
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d (body: %q)", rec.Code, want, rec.Body.String())
	}
}

func assertJSONContentType(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	mediaType, _, err := mime.ParseMediaType(rec.Header().Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", rec.Header().Get("Content-Type"))
	}
}

func assertJSONError(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	assertJSONContentType(t, rec)

	var body struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("error body is not valid JSON: %v", err)
	}
	if body.Error == "" {
		t.Fatalf(`error body has an empty or missing "error" field`)
	}
}

type createResponse struct {
	ID        string `json:"id"`
	DeleteKey string `json:"delete_key"`
}

func TestCreate_Success(t *testing.T) {
	maxLengthUrl := "https://example.com/" + strings.Repeat("a", 2048-len("https://example.com/"))

	tests := []struct {
		name        string
		contentType string
		url         string
	}{
		{name: "https url", contentType: "application/json", url: "https://example.com/some/long/path?q=1#frag"},
		{name: "http url", contentType: "application/json", url: "http://example.com"},
		{name: "content type with charset", contentType: "application/json; charset=utf-8", url: "https://example.com"},
		{name: "url at max length", contentType: "application/json", url: maxLengthUrl},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStore()
			body, _ := json.Marshal(map[string]string{"url": tt.url})

			rec := serve(store, newCreateRequest(tt.contentType, string(body)))

			assertStatus(t, rec, http.StatusCreated)
			assertJSONContentType(t, rec)

			var resp createResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("response is not valid JSON: %v", err)
			}

			if len(store.createCalls) != 1 {
				t.Fatalf("store.Create called %d times, want 1", len(store.createCalls))
			}
			if got := store.createCalls[0].Target; got != tt.url {
				t.Errorf("store.Create target = %q, want %q", got, tt.url)
			}
			assertStoredMatchesResponse(t, store, rec, resp)
		})
	}
}

func TestCreate_RetriesOnCollision(t *testing.T) {
	store := newFakeStore()
	store.collisions = 2

	rec := serve(store, newCreateRequest("application/json", `{"url":"https://example.com"}`))

	assertStatus(t, rec, http.StatusCreated)

	var resp createResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if len(store.createCalls) != 3 {
		t.Fatalf("store.Create called %d times, want 3", len(store.createCalls))
	}
	sources := map[string]bool{}
	for _, call := range store.createCalls {
		sources[call.Source] = true
	}
	if len(sources) != 3 {
		t.Errorf("retries reused an id: attempted sources %v, want 3 distinct", sources)
	}
	if last := store.createCalls[2].Source; resp.ID != last {
		t.Errorf("response id = %q, want the id of the successful attempt %q", resp.ID, last)
	}
	assertStoredMatchesResponse(t, store, rec, resp)
}

// assertStoredMatchesResponse checks that the id and delete key the client
// received are the ones that were generated and persisted.
func assertStoredMatchesResponse(t *testing.T, store *fakeStore, rec *httptest.ResponseRecorder, resp createResponse) {
	t.Helper()

	if len(resp.ID) != store.idLength || !isLowerHex(resp.ID) {
		t.Errorf("id = %q, want a %d character lowercase hex string", resp.ID, store.idLength)
	}
	stored, ok := store.urls[resp.ID]
	if !ok {
		t.Fatalf("response id %q does not match any stored url", resp.ID)
	}
	if resp.DeleteKey == "" || resp.DeleteKey != stored.DeleteKey {
		t.Errorf("delete_key = %q, want the stored key %q", resp.DeleteKey, stored.DeleteKey)
	}
	if got, want := rec.Header().Get("Location"), "/"+resp.ID; got != want {
		t.Errorf("Location = %q, want %q", got, want)
	}
}

func isLowerHex(s string) bool {
	return strings.Trim(s, "0123456789abcdef") == ""
}

func TestCreate_InvalidRequest(t *testing.T) {
	tooLongUrl := "https://example.com/" + strings.Repeat("a", 2049-len("https://example.com/"))

	tests := []struct {
		name        string
		contentType string
		body        string
		wantStatus  int
	}{
		{name: "empty body", contentType: "application/json", body: "", wantStatus: http.StatusBadRequest},
		{name: "invalid json", contentType: "application/json", body: `{"url":`, wantStatus: http.StatusBadRequest},
		{name: "missing url field", contentType: "application/json", body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "empty url", contentType: "application/json", body: `{"url":""}`, wantStatus: http.StatusBadRequest},
		{name: "unparseable url", contentType: "application/json", body: `{"url":"http://%zz"}`, wantStatus: http.StatusBadRequest},
		{name: "relative url", contentType: "application/json", body: `{"url":"/just/a/path"}`, wantStatus: http.StatusBadRequest},
		{name: "ftp scheme", contentType: "application/json", body: `{"url":"ftp://example.com/file"}`, wantStatus: http.StatusBadRequest},
		{name: "javascript scheme", contentType: "application/json", body: `{"url":"javascript:alert(1)"}`, wantStatus: http.StatusBadRequest},
		{name: "no host", contentType: "application/json", body: `{"url":"https://"}`, wantStatus: http.StatusBadRequest},
		{name: "no host with path", contentType: "application/json", body: `{"url":"https:///path"}`, wantStatus: http.StatusBadRequest},
		{name: "url too long", contentType: "application/json", body: `{"url":"` + tooLongUrl + `"}`, wantStatus: http.StatusBadRequest},
		{name: "text/plain content type", contentType: "text/plain", body: `{"url":"https://example.com"}`, wantStatus: http.StatusUnsupportedMediaType},
		{name: "missing content type", contentType: "", body: `{"url":"https://example.com"}`, wantStatus: http.StatusUnsupportedMediaType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStore()

			rec := serve(store, newCreateRequest(tt.contentType, tt.body))

			assertStatus(t, rec, tt.wantStatus)
			assertJSONError(t, rec)
			if len(store.createCalls) != 0 {
				t.Errorf("store.Create called %d times, want 0", len(store.createCalls))
			}
		})
	}
}

func TestCreate_StoreError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantRetry bool
	}{
		{name: "collision retries exhausted", err: models.ErrSourceAlreadyExists, wantRetry: true},
		{name: "unexpected error", err: errors.New("connection refused"), wantRetry: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStore()
			store.createErr = tt.err

			rec := serve(store, newCreateRequest("application/json", `{"url":"https://example.com"}`))

			assertStatus(t, rec, http.StatusInternalServerError)
			assertJSONError(t, rec)
			if retried := len(store.createCalls) > 1; retried != tt.wantRetry {
				t.Errorf("store.Create called %d times, want retry = %v", len(store.createCalls), tt.wantRetry)
			}
		})
	}
}

func TestCreate_OnlyMatchesRoot(t *testing.T) {
	store := newFakeStore()
	req := newCreateRequest("application/json", `{"url":"https://example.com"}`)
	req.URL.Path = "/abc"

	rec := serve(store, req)

	assertStatus(t, rec, http.StatusMethodNotAllowed)
	if len(store.createCalls) != 0 {
		t.Errorf("store.Create called %d times, want 0", len(store.createCalls))
	}
}

func TestResolve_Success(t *testing.T) {
	const target = "https://example.com/some/long/path?q=1"
	store := newFakeStore()
	store.seed("abc123", target, "secret")

	for i := 1; i <= 3; i++ {
		rec := serve(store, httptest.NewRequest(http.MethodGet, "/abc123", nil))

		assertStatus(t, rec, http.StatusTemporaryRedirect)
		if got := rec.Header().Get("Location"); got != target {
			t.Fatalf("Location = %q, want %q", got, target)
		}
		if got := store.urls["abc123"].ServeCount; got != int64(i) {
			t.Fatalf("serve count after %d requests = %d, want %d", i, got, i)
		}
	}
}

func TestResolve_Errors(t *testing.T) {
	tests := []struct {
		name       string
		storeErr   error
		wantStatus int
	}{
		{name: "unknown id", storeErr: nil, wantStatus: http.StatusNotFound},
		{name: "unexpected store error", storeErr: errors.New("connection refused"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStore()
			store.resolveErr = tt.storeErr

			rec := serve(store, httptest.NewRequest(http.MethodGet, "/unknown", nil))

			assertStatus(t, rec, tt.wantStatus)
			assertJSONError(t, rec)
			if loc := rec.Header().Get("Location"); loc != "" {
				t.Errorf("Location = %q, want none", loc)
			}
		})
	}
}

func TestDelete_Success(t *testing.T) {
	store := newFakeStore()
	store.seed("abc123", "https://example.com", "secret")
	req := httptest.NewRequest(http.MethodDelete, "/abc123", nil)
	req.Header.Set("X-Delete-Key", "secret")

	rec := serve(store, req)

	assertStatus(t, rec, http.StatusNoContent)
	if rec.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", rec.Body.String())
	}
	if _, ok := store.urls["abc123"]; ok {
		t.Errorf("url is still stored after a successful delete")
	}
}

func TestDelete_Errors(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		deleteKey     *string
		storeErr      error
		wantStatus    int
		wantStoreCall bool
	}{
		{name: "missing key header", id: "abc123", deleteKey: nil, wantStatus: http.StatusBadRequest, wantStoreCall: false},
		{name: "empty key header", id: "abc123", deleteKey: new(""), wantStatus: http.StatusBadRequest, wantStoreCall: false},
		{name: "wrong key", id: "abc123", deleteKey: new("wrong"), wantStatus: http.StatusForbidden, wantStoreCall: true},
		{name: "unknown id", id: "unknown", deleteKey: new("secret"), wantStatus: http.StatusNotFound, wantStoreCall: true},
		{name: "unexpected store error", id: "abc123", deleteKey: new("secret"), storeErr: errors.New("connection refused"), wantStatus: http.StatusInternalServerError, wantStoreCall: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStore()
			store.seed("abc123", "https://example.com", "secret")
			store.deleteErr = tt.storeErr
			req := httptest.NewRequest(http.MethodDelete, "/"+tt.id, nil)
			if tt.deleteKey != nil {
				req.Header.Set("X-Delete-Key", *tt.deleteKey)
			}

			rec := serve(store, req)

			assertStatus(t, rec, tt.wantStatus)
			assertJSONError(t, rec)
			if called := store.deleteCalls > 0; called != tt.wantStoreCall {
				t.Errorf("store.Delete called = %v, want %v", called, tt.wantStoreCall)
			}
			if _, ok := store.urls["abc123"]; !ok {
				t.Errorf("url was removed despite the failed delete")
			}
		})
	}
}
