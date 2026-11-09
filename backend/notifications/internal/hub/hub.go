// Package hub хранит активные SSE-подписки в памяти процесса.
package hub

import "sync"

type Event struct {
	ID       string
	ThreadID string
	Data     []byte
}

type Client struct {
	hub      *Hub
	threadID string
	ch       chan Event

	closeOnce sync.Once
}

func (c *Client) Events() <-chan Event { return c.ch }

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.hub.unsubscribe(c.threadID, c)
		close(c.ch)
	})
}

type Hub struct {
	mu       sync.RWMutex
	byThread map[string]map[*Client]struct{}
}

func New() *Hub {
	return &Hub{byThread: make(map[string]map[*Client]struct{})}
}

func (h *Hub) Subscribe(threadID string) *Client {
	c := &Client{hub: h, threadID: threadID, ch: make(chan Event, 32)}

	h.mu.Lock()
	if h.byThread[threadID] == nil {
		h.byThread[threadID] = make(map[*Client]struct{})
	}
	h.byThread[threadID][c] = struct{}{}
	h.mu.Unlock()

	return c
}

// Publish не блокирует consumer: медленный клиент теряет событие, а не тормозит остальных.
func (h *Hub) Publish(evt Event) {
	h.mu.RLock()
	clients := h.byThread[evt.ThreadID]
	h.mu.RUnlock()

	for c := range clients {
		select {
		case c.ch <- evt:
		default:
		}
	}
}

func (h *Hub) unsubscribe(threadID string, c *Client) {
	h.mu.Lock()
	if clients := h.byThread[threadID]; clients != nil {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.byThread, threadID)
		}
	}
	h.mu.Unlock()
}
