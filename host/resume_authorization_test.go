package host

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

const resumeIdentityHeader = "X-Test-User"

func resumeRequest(t *testing.T, user string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, "/ws", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	request.Header.Set(resumeIdentityHeader, user)
	return request
}

func bindResumeIdentity(request *http.Request, session *Session) error {
	session.ContextSet("user", request.Header.Get(resumeIdentityHeader))
	return nil
}

func authorizeSameUserResume(request *http.Request, session *Session) error {
	user, ok := session.ContextGet("user")
	if !ok || user != request.Header.Get(resumeIdentityHeader) {
		return errors.New("resume denied")
	}
	return nil
}

func resumableRuntime(opts ...MuxOption) *WSRuntime {
	base := []MuxOption{WithSSCLimits(SSCLimits{ReplayMessages: 4, ResumeTTL: time.Minute})}
	return NewWSRuntime(append(base, opts...)...)
}

func detachedSession(t *testing.T, runtime *WSRuntime, user string) *Session {
	t.Helper()
	session, resumed, err := runtime.OpenSession(resumeRequest(t, user), "")
	if err != nil {
		t.Fatalf("open session: %v", err)
	}
	if resumed || session.ResumeToken() == "" {
		t.Fatalf("expected a new resumable session, resumed=%v token=%q", resumed, session.ResumeToken())
	}
	SuspendSession(session, runtime.ResumeTTL())
	return session
}

func assertDetached(t *testing.T, session *Session) {
	t.Helper()
	session.deliveryMu.Lock()
	attached := session.attached
	expires := session.expires
	session.deliveryMu.Unlock()
	if attached {
		t.Fatal("rejected attempt attached the retained session")
	}
	if expires.IsZero() {
		t.Fatal("rejected attempt cleared the retention deadline")
	}
}

// A request authenticated as another user must not reattach a retained
// session, and the refusal must leave that session resumable by its owner.
func TestResumeAuthorizerRejectsAnotherUser(t *testing.T) {
	initializations := 0
	runtime := resumableRuntime(
		WithSSCSessionInitializer(func(request *http.Request, session *Session) error {
			initializations++
			return bindResumeIdentity(request, session)
		}),
		WithSSCResumeAuthorizer(authorizeSameUserResume),
	)
	owner := detachedSession(t, runtime, "alice")
	defer ReleaseSession(owner)
	token := owner.ResumeToken()

	intruder, resumed, err := runtime.OpenSession(resumeRequest(t, "mallory"), token)
	if err != nil {
		t.Fatalf("open session for rejected resume: %v", err)
	}
	defer ReleaseSession(intruder)
	if resumed || intruder == owner {
		t.Fatalf("foreign request resumed the retained session: resumed=%v", resumed)
	}
	if user, _ := owner.ContextGet("user"); user != "alice" {
		t.Fatalf("retained session identity changed to %v", user)
	}
	assertDetached(t, owner)

	restored, resumed, err := runtime.OpenSession(resumeRequest(t, "alice"), token)
	if err != nil {
		t.Fatalf("open session for owner resume: %v", err)
	}
	if !resumed || restored != owner {
		t.Fatalf("owner could not resume after a rejected attempt: resumed=%v", resumed)
	}
	if user, _ := restored.ContextGet("user"); user != "alice" {
		t.Fatalf("resume rebound the session identity to %v", user)
	}
	// The two new sessions were initialized; the resume was not.
	if initializations != 2 {
		t.Fatalf("session initializations = %d, want 2", initializations)
	}
}

// Identity alone is not enough: an authorizer that consults revocation state
// keeps a revoked credential from reattaching its own session.
func TestResumeAuthorizerRejectsRevokedIdentity(t *testing.T) {
	revoked := map[string]bool{"alice": true}
	runtime := resumableRuntime(
		WithSSCSessionInitializer(bindResumeIdentity),
		WithSSCResumeAuthorizer(func(request *http.Request, session *Session) error {
			if err := authorizeSameUserResume(request, session); err != nil {
				return err
			}
			if revoked[request.Header.Get(resumeIdentityHeader)] {
				return errors.New("credential revoked")
			}
			return nil
		}),
	)
	owner := detachedSession(t, runtime, "alice")
	defer ReleaseSession(owner)
	token := owner.ResumeToken()

	replacement, resumed, err := runtime.OpenSession(resumeRequest(t, "alice"), token)
	if err != nil {
		t.Fatalf("open session for revoked resume: %v", err)
	}
	defer ReleaseSession(replacement)
	if resumed || replacement == owner {
		t.Fatalf("revoked identity resumed the session: resumed=%v", resumed)
	}
	assertDetached(t, owner)

	revoked["alice"] = false
	restored, resumed, err := runtime.OpenSession(resumeRequest(t, "alice"), token)
	if err != nil {
		t.Fatalf("open session after revocation cleared: %v", err)
	}
	if !resumed || restored != owner {
		t.Fatalf("cleared credential could not resume: resumed=%v", resumed)
	}
}

// Two upgrades presenting the same token are authorized against the same
// candidate. Only one may attach it, and the loser must not be handed the
// retained session under the other request's authorization.
func TestConcurrentResumeAttachesOneAuthorizedSession(t *testing.T) {
	authorizing := make(chan struct{}, 2)
	release := make(chan struct{})
	runtime := resumableRuntime(
		WithSSCSessionInitializer(bindResumeIdentity),
		WithSSCResumeAuthorizer(func(request *http.Request, session *Session) error {
			if err := authorizeSameUserResume(request, session); err != nil {
				return err
			}
			authorizing <- struct{}{}
			<-release
			return nil
		}),
	)
	owner := detachedSession(t, runtime, "alice")
	defer ReleaseSession(owner)
	token := owner.ResumeToken()

	type outcome struct {
		session *Session
		resumed bool
		err     error
	}
	requests := []*http.Request{resumeRequest(t, "alice"), resumeRequest(t, "alice")}
	results := make(chan outcome, len(requests))
	for _, request := range requests {
		go func(request *http.Request) {
			session, resumed, err := runtime.OpenSession(request, token)
			results <- outcome{session: session, resumed: resumed, err: err}
		}(request)
	}
	// Both attempts hold an authorized candidate before either commits.
	<-authorizing
	<-authorizing
	close(release)

	attached := 0
	for range requests {
		result := <-results
		if result.err != nil {
			t.Fatalf("concurrent open session: %v", result.err)
		}
		if result.resumed {
			attached++
			if result.session != owner {
				t.Fatal("resume attached a session the request was not authorized for")
			}
			continue
		}
		if result.session == owner {
			t.Fatal("losing attempt received the retained session")
		}
		ReleaseSession(result.session)
	}
	if attached != 1 {
		t.Fatalf("resumed sessions = %d, want 1", attached)
	}
}

// Deployments that configure no resume authorizer keep the historical
// behavior, where the token alone reattaches the session.
func TestResumeWithoutAuthorizerKeepsTokenBehavior(t *testing.T) {
	runtime := resumableRuntime()
	owner := detachedSession(t, runtime, "alice")
	defer ReleaseSession(owner)

	restored, resumed, err := runtime.OpenSession(resumeRequest(t, "mallory"), owner.ResumeToken())
	if err != nil {
		t.Fatalf("open session: %v", err)
	}
	if !resumed || restored != owner {
		t.Fatalf("token resume changed without an authorizer: resumed=%v", resumed)
	}
}

// Releasing a retained session is final. Its token stops naming a candidate,
// so a later upgrade presenting it never reaches the authorizer, the token-only
// path cannot reattach it either, and the caller is served a new session with a
// token of its own.
func TestReleasedSessionTokenStaysUnresumable(t *testing.T) {
	handed := 0
	runtime := resumableRuntime(
		WithSSCSessionInitializer(bindResumeIdentity),
		WithSSCResumeAuthorizer(func(request *http.Request, session *Session) error {
			handed++
			return authorizeSameUserResume(request, session)
		}),
	)
	owner := detachedSession(t, runtime, "alice")
	token := owner.ResumeToken()

	ReleaseSession(owner)
	ReleaseSession(owner)

	for attempt := 0; attempt < 2; attempt++ {
		session, resumed, err := runtime.OpenSession(resumeRequest(t, "alice"), token)
		if err != nil {
			t.Fatalf("attempt %d: open session: %v", attempt, err)
		}
		t.Cleanup(func() { ReleaseSession(session) })
		if resumed || session == owner {
			t.Fatalf("attempt %d: the released session was resumed: resumed=%v", attempt, resumed)
		}
		if session.ResumeToken() == "" || session.ResumeToken() == token {
			t.Fatalf("attempt %d: the fresh session carries token %q", attempt, session.ResumeToken())
		}
	}
	if handed != 0 {
		t.Fatalf("the authorizer was handed a released session %d times", handed)
	}
	if _, registered := SessionByID(owner.ID()); registered {
		t.Fatal("the released session is still registered")
	}
	// Detaching it again must not put its token back into circulation.
	SuspendSession(owner, runtime.ResumeTTL())
	if resumed, ok := ResumeSession(token); ok {
		t.Fatalf("the released token reattached session %q", resumed.ID())
	}
}

// An unknown token is refused before the authorizer runs: there is no
// candidate to authorize, and the caller gets a fresh session.
func TestResumeAuthorizerSkippedForUnknownToken(t *testing.T) {
	calls := 0
	runtime := resumableRuntime(
		WithSSCSessionInitializer(bindResumeIdentity),
		WithSSCResumeAuthorizer(func(*http.Request, *Session) error {
			calls++
			return nil
		}),
	)
	session, resumed, err := runtime.OpenSession(resumeRequest(t, "alice"), "unknown-token")
	if err != nil {
		t.Fatalf("open session: %v", err)
	}
	defer ReleaseSession(session)
	if resumed || calls != 0 {
		t.Fatalf("unknown token reached the authorizer: resumed=%v calls=%d", resumed, calls)
	}
}
