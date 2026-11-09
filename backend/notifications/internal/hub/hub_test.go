package hub

import "testing"

func TestPublishOnlyToThreadSubscribers(t *testing.T) {
	h := New()
	a := h.Subscribe("thread-a")
	b := h.Subscribe("thread-b")
	defer a.Close()
	defer b.Close()

	h.Publish(Event{ID: "1", ThreadID: "thread-a", Data: []byte(`{"postId":"p1"}`)})

	select {
	case evt := <-a.Events():
		if evt.ID != "1" {
			t.Fatalf("unexpected event: %+v", evt)
		}
	default:
		t.Fatal("thread-a subscriber did not receive event")
	}

	select {
	case evt := <-b.Events():
		t.Fatalf("thread-b must not receive event: %+v", evt)
	default:
	}
}

func TestCloseRemovesSubscriber(t *testing.T) {
	h := New()
	c := h.Subscribe("thread")
	c.Close()

	h.Publish(Event{ID: "1", ThreadID: "thread"})

	select {
	case _, ok := <-c.Events():
		if ok {
			t.Fatal("closed client must not receive events")
		}
	default:
		t.Fatal("channel must be closed")
	}
}
