package host

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func streamBusAuthTest(t *testing.T) (*streamBusEndpoint, http.Handler) {
	t.Helper()
	e := newStreamBusEndpoint(NewWSRuntime(WithSSCStreamBusAuth(), WithAuthFunc(func(r *http.Request) bool {
		cookie, err := r.Cookie("session")
		return err == nil && cookie.Value == "test-user"
	})))
	e.auth = &streamBusAuth{tickets: make(map[string]streamBusCredential)}
	return e, e.runtime.Guard(http.HandlerFunc(e.issueCredential))
}

func streamBusAuthRequest() *http.Request {
	r := httptest.NewRequest(http.MethodPost, "https://app.test"+streamBusAuthPath, nil)
	r.Header.Set("Origin", "https://app.test")
	r.Header.Set("Content-Type", "application/json")
	r.AddCookie(&http.Cookie{Name: "session", Value: "test-user", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	return r
}

func issueStreamBusTestCredential(t *testing.T, handler http.Handler) string {
	t.Helper()
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, streamBusAuthRequest())
	var response struct{ Token string }
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &response) != nil || len(response.Token) != 64 {
		t.Fatalf("credential issuance: status=%d token bytes=%d", w.Code, len(response.Token))
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("credential cache policy = %q", w.Header().Get("Cache-Control"))
	}
	return response.Token
}

func streamBusAuthUpgrade() *http.Request {
	r := httptest.NewRequest(http.MethodGet, "https://app.test"+streamBusPath, nil)
	r.Method = http.MethodConnect
	r.Header.Set("Origin", "https://app.test")
	return r
}

func TestStreamBusCredentialRequiresAuthenticatedBootstrap(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	r := streamBusAuthRequest()
	r.Header.Del("Cookie")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized || len(e.auth.tickets) != 0 {
		t.Fatalf("anonymous bootstrap: status=%d tickets=%d", w.Code, len(e.auth.tickets))
	}
}

func TestStreamBusCredentialPreservesIdentityForOneUpgradeOnly(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	token := issueStreamBusTestCredential(t, handler)
	r, err := e.consumeCredential(streamBusAuthUpgrade(), token)
	if err != nil {
		t.Fatal(err)
	}
	if r.Method != http.MethodConnect || r.URL.Path != streamBusPath || r.Header.Get("Cookie") != "session=test-user" {
		t.Fatalf("authenticated upgrade: method=%s path=%s", r.Method, r.URL.Path)
	}
	if _, err := e.consumeCredential(streamBusAuthUpgrade(), token); err == nil {
		t.Fatal("credential reused")
	}
}

func TestStreamBusCredentialRequiresSameOrConfiguredOrigin(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	r := streamBusAuthRequest()
	r.Header.Set("Origin", "https://other.test")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden || len(e.auth.tickets) != 0 {
		t.Fatalf("foreign bootstrap: status=%d tickets=%d", w.Code, len(e.auth.tickets))
	}
}

func TestStreamBusCredentialRejectsMissingOrigin(t *testing.T) {
	_, handler := streamBusAuthTest(t)
	r := streamBusAuthRequest()
	r.Header.Del("Origin")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("missing origin: %d", w.Code)
	}
}

func TestStreamBusCredentialIsBoundToBootstrapOrigin(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	token := issueStreamBusTestCredential(t, handler)
	r := streamBusAuthUpgrade()
	r.Header.Set("Origin", "https://other.test")
	if _, err := e.consumeCredential(r, token); err == nil {
		t.Fatal("foreign upgrade accepted")
	}
}

func TestStreamBusCredentialExpiresBeforeConsumption(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	token := issueStreamBusTestCredential(t, handler)
	credential := e.auth.tickets[token]
	credential.expires = time.Now().Add(-time.Second)
	e.auth.tickets[token] = credential
	if _, err := e.consumeCredential(streamBusAuthUpgrade(), token); err == nil {
		t.Fatal("expired credential accepted")
	}
}

func TestStreamBusCredentialRechecksRevocationBeforeConsumption(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	token := issueStreamBusTestCredential(t, handler)
	e.runtime.authFunc = func(*http.Request) bool { return false }
	if _, err := e.consumeCredential(streamBusAuthUpgrade(), token); err == nil {
		t.Fatal("revoked credential accepted")
	}
}

func TestStreamBusCredentialHasOneConcurrentConsumer(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	token := issueStreamBusTestCredential(t, handler)
	var accepted atomic.Int64
	var group sync.WaitGroup
	for i := 0; i < 16; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := e.consumeCredential(streamBusAuthUpgrade(), token); err == nil {
				accepted.Add(1)
			}
		}()
	}
	group.Wait()
	if accepted.Load() != 1 {
		t.Fatalf("concurrent consumers = %d", accepted.Load())
	}
}

func TestStreamBusCredentialStoreIsBoundedAndReclaimsExpiredEntries(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	for i := 0; i < streamBusAuthLimit; i++ {
		issueStreamBusTestCredential(t, handler)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, streamBusAuthRequest())
	if w.Code != http.StatusTooManyRequests || len(e.auth.tickets) != streamBusAuthLimit {
		t.Fatalf("capacity: status=%d tickets=%d", w.Code, len(e.auth.tickets))
	}
	for key, credential := range e.auth.tickets {
		credential.expires = time.Now().Add(-time.Second)
		e.auth.tickets[key] = credential
	}
	issueStreamBusTestCredential(t, handler)
	if len(e.auth.tickets) != 1 {
		t.Fatalf("retained credentials = %d", len(e.auth.tickets))
	}
}

func TestStreamBusCredentialRefusesOversizedIdentityHeaders(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	r := streamBusAuthRequest()
	r.Header.Set("X-Large", strings.Repeat("x", streamBusAuthHeaders))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusRequestHeaderFieldsTooLarge || len(e.auth.tickets) != 0 {
		t.Fatalf("header limit: status=%d tickets=%d", w.Code, len(e.auth.tickets))
	}
}

func TestStreamBusCredentialRequiresPost(t *testing.T) {
	_, handler := streamBusAuthTest(t)
	r := streamBusAuthRequest()
	r.Method = http.MethodGet
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("bootstrap method: %d", w.Code)
	}
}

func TestStreamBusCredentialRefusesMissingAuthenticator(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	e.runtime.authFunc = nil
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, streamBusAuthRequest())
	if w.Code != http.StatusServiceUnavailable || len(e.auth.tickets) != 0 {
		t.Fatalf("missing authenticator: status=%d tickets=%d", w.Code, len(e.auth.tickets))
	}
}

func TestStreamBusCredentialFrameRejectsUnknownFields(t *testing.T) {
	if _, err := decodeStreamBusCredential([]byte(`{"token":"redacted","action":"mutate"}`)); err == nil {
		t.Fatal("unknown authentication field accepted")
	}
}

func TestStreamBusCredentialFrameRejectsTrailingJSON(t *testing.T) {
	if _, err := decodeStreamBusCredential([]byte(`{"token":"redacted"}{}`)); err == nil {
		t.Fatal("trailing authentication frame accepted")
	}
}

func TestStreamBusCredentialRefusesFormBootstrap(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	r := streamBusAuthRequest()
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest || len(e.auth.tickets) != 0 {
		t.Fatalf("form bootstrap: status=%d tickets=%d", w.Code, len(e.auth.tickets))
	}
}

func TestStreamBusCredentialRefusesBootstrapBody(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	r := streamBusAuthRequest()
	r.ContentLength = 2
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest || len(e.auth.tickets) != 0 {
		t.Fatalf("bootstrap body: status=%d tickets=%d", w.Code, len(e.auth.tickets))
	}
}

func TestStreamBusCredentialRefusesURLParameters(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	r := streamBusAuthRequest()
	r.URL.RawQuery = "token=redacted"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest || len(e.auth.tickets) != 0 {
		t.Fatalf("URL parameters: status=%d tickets=%d", w.Code, len(e.auth.tickets))
	}
}

func TestStreamBusCredentialOwnsItsIdentityHeaders(t *testing.T) {
	e, handler := streamBusAuthTest(t)
	r := streamBusAuthRequest()
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	var response struct{ Token string }
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &response) != nil {
		t.Fatalf("bootstrap status = %d", w.Code)
	}
	r.Header.Set("Cookie", "session=other-user")
	if _, err := e.consumeCredential(streamBusAuthUpgrade(), response.Token); err != nil {
		t.Fatalf("captured identity changed with caller headers: %v", err)
	}
}

func TestStreamBusAuthenticationDoesNotResumeWithoutAnAuthorizer(t *testing.T) {
	session := AllocateResumableSession(8)
	t.Cleanup(func() { ReleaseSession(session) })
	SuspendSession(session, time.Minute)
	runtime := NewWSRuntime(WithSSCStreamBusAuth())
	if resumed, ok := runtime.resumeSession(streamBusAuthUpgrade(), session.ResumeToken()); ok || resumed != nil {
		t.Fatal("authenticated mode resumed without an identity policy")
	}
	owner := NewWSRuntime(WithSSCStreamBusAuth(), WithSSCResumeAuthorizer(func(*http.Request, *Session) error { return nil }))
	if resumed, ok := owner.resumeSession(streamBusAuthUpgrade(), session.ResumeToken()); !ok || resumed != session {
		t.Fatal("refused resume consumed the owner's detached session")
	}
}
