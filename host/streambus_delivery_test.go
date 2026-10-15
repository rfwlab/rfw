package host

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mirkobrombin/go-warp/v2/streambus"
)

func TestStreamBusInvalidPayloadDoesNotConsumeSequence(t *testing.T) {
	connection := &streamBusConnection{}
	session := newSession("stream-invalid", sessionOptions{replayLimit: 4})
	session.streamConnection = connection

	sendStreamBusSession(connection, session, Outbound{Payload: make(chan int)})

	if session.outboundSeq != 0 {
		t.Fatalf("sequence = %d, want 0", session.outboundSeq)
	}
	if len(session.replay) != 0 {
		t.Fatalf("replay = %#v, want empty", session.replay)
	}
}

func TestStreamBusBackpressureClosesBeforeSequence(t *testing.T) {
	bus := streambus.NewInMemory(streambus.Config{DefaultBuffer: 1, MaxBuffer: 1})
	t.Cleanup(func() { _ = bus.Close() })
	subscription, err := bus.Subscribe(context.Background(), streambus.SubscribeOptions{
		Topic: "blocked", Buffer: 1, Overflow: streambus.Block,
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	connection := &streamBusConnection{bus: bus, topic: "blocked", subscription: subscription, buffer: 1}
	session := newSession("stream-full", sessionOptions{replayLimit: 4})
	session.streamConnection = connection

	if _, err := bus.Publish(context.Background(), streambus.Frame{Topic: "blocked", Payload: []byte("one")}); err != nil {
		t.Fatalf("publish first: %v", err)
	}
	waitForStreamBus(t, func() bool { return subscription.Stats().Queued == 0 })
	if _, err := bus.Publish(context.Background(), streambus.Frame{Topic: "blocked", Payload: []byte("two")}); err != nil {
		t.Fatalf("publish second: %v", err)
	}
	waitForStreamBus(t, func() bool { return subscription.Stats().Queued == 1 })

	sendStreamBusSession(connection, session, Outbound{Payload: "late"})

	if session.outboundSeq != 0 {
		t.Fatalf("sequence = %d, want 0", session.outboundSeq)
	}
	select {
	case <-subscription.Done():
	default:
		t.Fatal("subscription stayed open")
	}
}

func TestStreamBusEndpointDisablesTransportReplay(t *testing.T) {
	endpoint := newStreamBusEndpoint(NewWSRuntime())
	t.Cleanup(func() { _ = endpoint.bus.Close() })
	topic := "rfw/connection/replay"
	sequence, err := endpoint.bus.Publish(context.Background(), streambus.Frame{Topic: topic, Payload: []byte("{}")})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	subscription, err := endpoint.bus.Subscribe(context.Background(), streambus.SubscribeOptions{
		Topic: topic, Buffer: 1, Since: sequence,
	})
	if subscription != nil {
		_ = subscription.Close()
	}
	if !errors.Is(err, streambus.ErrReplayUnavailable) {
		t.Fatalf("subscribe since = %v, want %v", err, streambus.ErrReplayUnavailable)
	}
}

func waitForStreamBus(t *testing.T, f func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if f() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition was not reached")
}
