package ds

import (
	"Day13/models"
	"errors"
	"sync"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")

// UserRegistry manages a thread-safe map of Users by their IDs.
type UserRegistry struct {
	mu    sync.RWMutex
	users map[string]*models.User
}

func NewUserRegistry() *UserRegistry {
	return &UserRegistry{
		users: make(map[string]*models.User),
	}
}

func (r *UserRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.users)
}

func (r *UserRegistry) AddUser(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return ErrUserAlreadyExists
	}
	r.users[user.ID] = user
	return nil
}

func (r *UserRegistry) GetUser(id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *UserRegistry) UpdateUser(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return ErrUserNotFound
	}
	r.users[user.ID] = user
	return nil
}

func (r *UserRegistry) RemoveUser(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
}
