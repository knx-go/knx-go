package knx

import (
	"testing"
	"time"

	"github.com/knx-go/knx-go/knx/cemi"
)

type stubMessage struct {
	id int
}

func (m *stubMessage) MessageCode() cemi.MessageCode { return cemi.LDataIndCode }
func (m *stubMessage) Size() uint                    { return 0 }
func (m *stubMessage) Pack(_ []byte)                 {}

func TestCheckRouterConfigAppliesDefaults(t *testing.T) {
	t.Parallel()

	cfg := RouterConfig{}
	got := checkRouterConfig(cfg)

	if got.RetainCount != DefaultRouterConfig.RetainCount {
		t.Fatalf("expected RetainCount %d, got %d", DefaultRouterConfig.RetainCount, got.RetainCount)
	}

	custom := RouterConfig{RetainCount: 5, PostSendPauseDuration: time.Second}
	got = checkRouterConfig(custom)
	if got.RetainCount != custom.RetainCount {
		t.Fatalf("expected to keep RetainCount %d, got %d", custom.RetainCount, got.RetainCount)
	}
	if got.PostSendPauseDuration != custom.PostSendPauseDuration {
		t.Fatalf("unexpected PostSendPauseDuration change: want %v got %v", custom.PostSendPauseDuration, got.PostSendPauseDuration)
	}
}

func TestRouterPushInboundBuffered(t *testing.T) {
	t.Parallel()

	router := &Router{inbound: make(chan cemi.Message, 1)}
	msg := &stubMessage{id: 1}

	router.pushInbound(msg)

	select {
	case got := <-router.inbound:
		if got != msg {
			t.Fatalf("expected %p, got %p", msg, got)
		}
	default:
		t.Fatal("expected message to be delivered without blocking")
	}
}

func TestRouterPushInboundSpawnsWorker(t *testing.T) {
	t.Parallel()

	router := &Router{inbound: make(chan cemi.Message)}
	msg := &stubMessage{id: 2}

	done := make(chan struct{})
	go func() {
		router.pushInbound(msg)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pushInbound should not block when channel is full")
	}

	select {
	case got := <-router.inbound:
		if got != msg {
			t.Fatalf("expected %p, got %p", msg, got)
		}
	case <-time.After(time.Second):
		t.Fatal("expected goroutine to forward message")
	}
}
