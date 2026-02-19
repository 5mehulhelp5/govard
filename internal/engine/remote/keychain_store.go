package remote

type keychainStore struct {
	fallback AuthStore
}

func NewKeychainStore() AuthStore {
	return &keychainStore{
		fallback: NewFileStore(),
	}
}

func (k *keychainStore) Get(key string) (string, error) {
	return k.fallback.Get(key)
}

func (k *keychainStore) Set(key, value string) error {
	return k.fallback.Set(key, value)
}
