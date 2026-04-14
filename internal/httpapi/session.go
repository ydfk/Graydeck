package httpapi

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

type sessionEntry struct {
	username  string
	expiresAt time.Time
}

type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]sessionEntry
	ttl      time.Duration
}

func newSessionStore(ttl time.Duration) *sessionStore {
	return &sessionStore{
		sessions: make(map[string]sessionEntry),
		ttl:      ttl,
	}
}

func (s *sessionStore) Create(username string) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupLocked()
	s.sessions[token] = sessionEntry{
		username:  username,
		expiresAt: time.Now().Add(s.ttl),
	}
	return token, nil
}

func (s *sessionStore) Get(token string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.sessions[token]
	if !ok {
		return "", false
	}

	if time.Now().After(entry.expiresAt) {
		delete(s.sessions, token)
		return "", false
	}

	return entry.username, true
}

func (s *sessionStore) Delete(token string) {
	if token == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

func (s *sessionStore) cleanupLocked() {
	now := time.Now()
	for token, entry := range s.sessions {
		if now.After(entry.expiresAt) {
			delete(s.sessions, token)
		}
	}
}
