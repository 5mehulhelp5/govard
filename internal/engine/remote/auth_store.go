package remote

type AuthStore interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

type memoryStore struct {
	data map[string]string
}

func NewInMemoryStore() AuthStore {
	return &memoryStore{data: map[string]string{}}
}

func (m *memoryStore) Get(key string) (string, error) {
	return m.data[key], nil
}

func (m *memoryStore) Set(key, value string) error {
	m.data[key] = value
	return nil
}
