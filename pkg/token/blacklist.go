package token

import (
	"sync"
	"time"
)

// BlacklistedToken represents a blacklisted token with its expiration
type BlacklistedToken struct {
	Token     string
	ExpiresAt time.Time
}

// TokenBlacklist manages blacklisted tokens
type TokenBlacklist interface {
	Add(token string, expiresAt time.Time) error
	IsBlacklisted(token string) bool
	Cleanup() // Remove expired tokens
}

// InMemoryBlacklist is an in-memory implementation of TokenBlacklist
type InMemoryBlacklist struct {
	mu         sync.RWMutex
	tokens     map[string]time.Time
	cleanupTTL time.Duration
}

// NewInMemoryBlacklist creates a new in-memory token blacklist
func NewInMemoryBlacklist(cleanupInterval time.Duration) *InMemoryBlacklist {
	blacklist := &InMemoryBlacklist{
		tokens:     make(map[string]time.Time),
		cleanupTTL: cleanupInterval,
	}

	// Start cleanup goroutine
	go blacklist.startCleanup()

	return blacklist
}

// Add adds a token to the blacklist
func (b *InMemoryBlacklist) Add(token string, expiresAt time.Time) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.tokens[token] = expiresAt
	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (b *InMemoryBlacklist) IsBlacklisted(token string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	expiresAt, exists := b.tokens[token]
	if !exists {
		return false
	}

	// If token has expired, it's no longer relevant
	if time.Now().After(expiresAt) {
		return false
	}

	return true
}

// Cleanup removes expired tokens from the blacklist
func (b *InMemoryBlacklist) Cleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	for token, expiresAt := range b.tokens {
		if now.After(expiresAt) {
			delete(b.tokens, token)
		}
	}
}

// startCleanup runs periodic cleanup of expired tokens
func (b *InMemoryBlacklist) startCleanup() {
	ticker := time.NewTicker(b.cleanupTTL)
	defer ticker.Stop()

	for range ticker.C {
		b.Cleanup()
	}
}

// Count returns the number of blacklisted tokens (for debugging)
func (b *InMemoryBlacklist) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.tokens)
}
