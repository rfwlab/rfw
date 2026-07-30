package host

import (
	"errors"
	"testing"
	"time"
)

func TestSessionResumeAndReplay(t *testing.T) {
	session := AllocateResumableSession(4)
	token := session.ResumeToken()
	first := session.PrepareOutbound(Outbound{Component: "Counter", Payload: map[string]any{"value": 1}})
	second := session.PrepareOutbound(Outbound{Component: "Counter", Payload: map[string]any{"value": 2}})
	if first.Sequence != 1 || second.Sequence != 2 || token == "" {
		t.Fatalf("missing delivery metadata: first=%#v second=%#v", first, second)
	}

	SuspendSession(session, time.Second)
	resumed, ok := ResumeSession(token)
	if !ok || resumed != session {
		t.Fatal("session did not resume")
	}
	replay, err := resumed.ReplayAfter(first.Sequence)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if len(replay) != 1 || replay[0].Sequence != second.Sequence {
		t.Fatalf("unexpected replay: %#v", replay)
	}
	ReleaseSession(session)
}

func TestSessionRejectsDuplicatesAndGaps(t *testing.T) {
	session := newSession("ordered")
	if err := session.AcceptInbound(1); err != nil {
		t.Fatalf("first message: %v", err)
	}
	if err := session.AcceptInbound(1); !errors.Is(err, ErrDuplicateMessage) {
		t.Fatalf("duplicate result: %v", err)
	}
	if err := session.AcceptInbound(3); !errors.Is(err, ErrSequenceGap) {
		t.Fatalf("gap result: %v", err)
	}
	if err := session.AcceptInbound(2); err != nil {
		t.Fatalf("next message: %v", err)
	}
}

func TestSessionReplayReportsEvictedHistory(t *testing.T) {
	session := AllocateResumableSession(1)
	defer ReleaseSession(session)
	session.PrepareOutbound(Outbound{Payload: "one"})
	session.PrepareOutbound(Outbound{Payload: "two"})
	if _, err := session.ReplayAfter(0); !errors.Is(err, ErrReplayUnavailable) {
		t.Fatalf("expected replay error, got %v", err)
	}
}

func TestSessionAllocationLimit(t *testing.T) {
	sessionMu.RLock()
	current := len(sessions)
	sessionMu.RUnlock()

	session, err := allocateSession(1, current+1)
	if err != nil {
		t.Fatalf("allocate within limit: %v", err)
	}
	defer ReleaseSession(session)
	if _, err := allocateSession(1, current+1); !errors.Is(err, ErrSessionLimit) {
		t.Fatalf("expected session limit, got %v", err)
	}
}

func TestWithoutSSCResumeCreatesEphemeralSession(t *testing.T) {
	runtime := NewWSRuntime(WithoutSSCResume())
	session, err := runtime.NewSession(nil)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if session.ResumeToken() != "" || runtime.ResumeTTL() != 0 {
		t.Fatalf("resume was not disabled: token=%q ttl=%s", session.ResumeToken(), runtime.ResumeTTL())
	}
	ReleaseSession(session)
}
