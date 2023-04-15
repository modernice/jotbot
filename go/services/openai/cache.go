package openai

import (
	"fmt"
	"io"
	"io/fs"
	"sync"
)

type codeCache struct {
	repo fs.FS

	sync.RWMutex
	cache map[string][]byte
}

func newCodeCache(repo fs.FS) *codeCache {
	return &codeCache{
		repo:  repo,
		cache: make(map[string][]byte),
	}
}

func (cache *codeCache) cached(file string) ([]byte, bool) {
	cache.RLock()
	defer cache.RUnlock()
	code, ok := cache.cache[file]
	return code, ok
}

func (cache *codeCache) get(file string) ([]byte, error) {
	if code, ok := cache.cached(file); ok {
		return code, nil
	}

	cache.Lock()
	defer cache.Unlock()

	f, err := cache.repo.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", file, err)
	}
	defer f.Close()

	code, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", file, err)
	}
	cache.cache[file] = code

	return code, nil
}
