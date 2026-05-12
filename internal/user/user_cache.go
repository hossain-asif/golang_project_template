package user

import (
	"go_project_structure/internal/database/models"
	"sync"
	"time"
)

type UserListCache struct {
	mu      sync.RWMutex
	users   []*models.User
	builtAt time.Time
	ttl     time.Duration
}

func NewUserListCache(ttl time.Duration) *UserListCache {
	return &UserListCache{ttl: ttl}
}


// get returns the cached list of users if it is not stale.
// A stale result is one where the time elapsed since the list was built is greater than the configured TTL.
// The method returns a boolean indicating whether the result is from the cache or not.
func (c *UserListCache) Get() ([]*models.User, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.users != nil && time.Since(c.builtAt) < c.ttl {
		return c.users, true
	}
	return nil, false
}


// Set the cached list of users, updating the builtAt timestamp to the current time.
// This method is thread-safe.
func (c *UserListCache) Set(users []*models.User) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.users = users
	c.builtAt = time.Now()
}

// purge invalidates the cache.
// Call this after any create / update / delete operation.
func (c *UserListCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.users = nil
}
