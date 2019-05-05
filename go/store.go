package swagger

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type store interface {
	addSecret(text string, count int, expiredAfterMinutes int) (*Secret, error)
	readSecret(hash string) (*Secret, error)
}

type inMemoryStore struct {
	secrets map[string]*Secret
	mutex   sync.Mutex
}

func newInMemoryStore() *inMemoryStore {
	return &inMemoryStore{
		secrets: make(map[string]*Secret),
	}
}

func (s *inMemoryStore) addSecret(text string, count int, expiredAfterMinutes int) (*Secret, error) {
	if expiredAfterMinutes < 0 {
		return nil, errors.New("minutes can't be < 0")
	}

	if count < 0 {
		return nil, errors.New("view count can't be less < 0")
	}

	uuid, err := uuid.NewUUID()

	if err != nil {
		return nil, err
	}

	hash := uuid.String()
	now := time.Now()

	secret := &Secret{
		Hash:           hash,
		SecretText:     text,
		CreatedAt:      now,
		ExpiresAt:      now.Add(time.Minute * time.Duration(expiredAfterMinutes)),
		RemainingViews: int32(count),
	}

	if expiredAfterMinutes == 0 {
		secret.ExpiresAt = time.Now().AddDate(100, 0, 0) // hack: fix me
	}

	s.mutex.Lock()
	s.secrets[hash] = secret
	s.mutex.Unlock()

	return secret, nil
}

func (s *inMemoryStore) readSecret(hash string) (*Secret, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	secret, ok := s.secrets[hash]

	if !ok {
		return nil, errors.New("Not found")
	}

	if secret.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("Expired secret.")
	}

	if secret.RemainingViews <= 0 {
		return nil, errors.New("No remaining view count.")
	}

	secret.RemainingViews--

	return secret, nil
}
