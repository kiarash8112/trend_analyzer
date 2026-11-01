package services

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

// Subscribers is a JSON file of the chats that get the daily reminder.
// Each chat maps to the reminder slot it last received, so a restart never sends a slot twice.
type Subscribers struct {
	mu    sync.Mutex
	path  string
	chats map[int64]string
}

// OpenSubscribers loads the file at path; a missing file means no subscribers yet.
func OpenSubscribers(path string) (*Subscribers, error) {
	s := &Subscribers{path: path, chats: map[int64]string{}}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &s.chats); err != nil {
		return nil, err
	}
	return s, nil
}

// Add subscribes a chat, counting slot as already received. Existing chats are left alone.
func (s *Subscribers) Add(chatID int64, slot string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.chats[chatID]; ok {
		return nil
	}
	s.chats[chatID] = slot
	return s.save()
}

// Remove unsubscribes a chat.
func (s *Subscribers) Remove(chatID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.chats[chatID]; !ok {
		return nil
	}
	delete(s.chats, chatID)
	return s.save()
}

// Has reports whether a chat is subscribed.
func (s *Subscribers) Has(chatID int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.chats[chatID]
	return ok
}

// Due returns the chats that have not received slot yet.
func (s *Subscribers) Due(slot string) []int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []int64
	for id, last := range s.chats {
		if last != slot {
			out = append(out, id)
		}
	}
	slices.Sort(out)
	return out
}

// MarkReminded records that a chat received slot. A chat that unsubscribed meanwhile stays out.
func (s *Subscribers) MarkReminded(chatID int64, slot string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.chats[chatID]; !ok {
		return nil
	}
	s.chats[chatID] = slot
	return s.save()
}

// save replaces the file atomically so a crash cannot leave it half written.
func (s *Subscribers) save() error {
	b, err := json.MarshalIndent(s.chats, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
