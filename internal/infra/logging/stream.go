package logging

import "sync"

type Entry struct {
	Level   string
	Message string
	Time    string
	Fields  map[string]any
}

type Stream struct {
	mu      sync.RWMutex
	history []Entry
	subs    map[chan Entry]struct{}
	limit   int
}

func NewStream(limit int) *Stream {
	if limit <= 0 {
		limit = 100
	}

	return &Stream{
		history: make([]Entry, 0, limit),
		subs:    make(map[chan Entry]struct{}),
		limit:   limit,
	}
}

func (s *Stream) Publish(entry Entry) {
	if s == nil {
		return
	}

	entry.Fields = cloneFields(entry.Fields)

	s.mu.Lock()
	s.history = append(s.history, entry)
	if len(s.history) > s.limit {
		s.history = append([]Entry(nil), s.history[len(s.history)-s.limit:]...)
	}

	for ch := range s.subs {
		select {
		case ch <- entry:
		default:
		}
	}
	s.mu.Unlock()
}

func (s *Stream) Snapshot() []Entry {
	if s == nil {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	history := make([]Entry, 0, len(s.history))
	for _, entry := range s.history {
		history = append(history, Entry{
			Level:   entry.Level,
			Message: entry.Message,
			Time:    entry.Time,
			Fields:  cloneFields(entry.Fields),
		})
	}

	return history
}

func (s *Stream) Subscribe(buffer int) (<-chan Entry, func()) {
	ch := make(chan Entry, buffer)
	if s == nil {
		close(ch)
		return ch, func() {}
	}

	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()

	cancel := func() {
		s.mu.Lock()
		if _, ok := s.subs[ch]; ok {
			delete(s.subs, ch)
			close(ch)
		}
		s.mu.Unlock()
	}

	return ch, cancel
}

func cloneFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return nil
	}

	cloned := make(map[string]any, len(fields))
	for k, v := range fields {
		cloned[k] = v
	}

	return cloned
}
