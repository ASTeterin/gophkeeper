package client

import (
	"encoding/json"
	"os"
	"sync"
)

type LocalStore struct {
	mu   sync.RWMutex
	data map[string]*Item
	path string
}

type Item struct {
	Key         string `json:"key"`
	Data        []byte `json:"data"`
	Description string `json:"description"`
	Dirty       bool   `json:"dirty"`
}

func NewLocalStore(path string) *LocalStore {
	s := &LocalStore{
		data: make(map[string]*Item),
		path: path,
	}
	s.load()
	return s
}

func (s *LocalStore) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	json.Unmarshal(data, &s.data)
}

func (s *LocalStore) save() {
	data, _ := json.MarshalIndent(s.data, "", "  ")
	os.WriteFile(s.path, data, 0600)
}

func (s *LocalStore) Get(key string) (*Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.data[key]
	return item, ok
}

func (s *LocalStore) Add(key string, desc string, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = &Item{Key: key, Description: desc, Data: data, Dirty: true}
	s.save()
}

func (s *LocalStore) List() []*Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]*Item, 0, len(s.data))
	for _, item := range s.data {
		items = append(items, item)
	}
	return items
}

func (s *LocalStore) Sync(remoteItems []*Item) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]*Item)
	for _, item := range remoteItems {
		s.data[item.Key] = item
	}
	s.save()
}
