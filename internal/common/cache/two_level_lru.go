package cache

import (
	"sync"
	"time"
)

type Config struct {
	L1Size      int
	L2Size      int
	L2Window    time.Duration
	L2Threshold int
}

type AccessTracker[K comparable] struct {
	mu         sync.RWMutex
	timestamps map[K][]time.Time
	threshold  int
	window     time.Duration
}

func NewAccessTracker[K comparable](threshold int, window time.Duration) *AccessTracker[K] {
	return &AccessTracker[K]{
		timestamps: make(map[K][]time.Time),
		threshold:  threshold,
		window:     window,
	}
}

func (t *AccessTracker[K]) RecordAccess(key K) bool {
	now := time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoff := now.Add(-t.window)
	var recent []time.Time
	for _, ts := range t.timestamps[key] {
		if ts.After(cutoff) {
			recent = append(recent, ts)
		}
	}
	t.timestamps[key] = recent
	t.timestamps[key] = append(t.timestamps[key], now)

	return len(t.timestamps[key]) >= t.threshold
}

func (t *AccessTracker[K]) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.timestamps = make(map[K][]time.Time)
}

type TwoLevelLRU[K comparable, V any] struct {
	l1     *LRU[K, V]
	l2     *LRU[K, V]
	access *AccessTracker[K]
	config Config
}

func NewTwoLevelLRU[K comparable, V any](config Config) *TwoLevelLRU[K, V] {
	if config.L2Window == 0 {
		config.L2Window = time.Second
	}
	if config.L2Threshold == 0 {
		config.L2Threshold = 2
	}
	if config.L1Size == 0 {
		config.L1Size = 2000
	}
	if config.L2Size == 0 {
		config.L2Size = 2000
	}

	return &TwoLevelLRU[K, V]{
		l1:     NewLRU[K, V](config.L1Size),
		l2:     NewLRU[K, V](config.L2Size),
		access: NewAccessTracker[K](config.L2Threshold, config.L2Window),
		config: config,
	}
}

func (t *TwoLevelLRU[K, V]) Get(key K) (V, bool) {
	if val, ok := t.l2.Get(key); ok {
		return val, true
	}
	if val, ok := t.l1.Get(key); ok {
		t.access.RecordAccess(key)
		return val, true
	}
	return *new(V), false
}

func (t *TwoLevelLRU[K, V]) Set(key K, value V) {
	t.l1.Set(key, value)
	if t.access.RecordAccess(key) {
		t.l2.Set(key, value)
	}
}

func (t *TwoLevelLRU[K, V]) Delete(key K) {
	t.l1.Delete(key)
	t.l2.Delete(key)
}

func (t *TwoLevelLRU[K, V]) Len() (l1Len, l2Len int) {
	return t.l1.Len(), t.l2.Len()
}

func (t *TwoLevelLRU[K, V]) Clear() {
	t.l1.Clear()
	t.l2.Clear()
	t.access.Clear()
}
