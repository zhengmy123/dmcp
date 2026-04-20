package cache

import (
	"testing"
)

func TestLRU_BasicOperations(t *testing.T) {
	lru := NewLRU[string, int](3)

	if lru.Len() != 0 {
		t.Errorf("expected empty LRU, got %d", lru.Len())
	}

	lru.Set("a", 1)
	lru.Set("b", 2)
	lru.Set("c", 3)

	if lru.Len() != 3 {
		t.Errorf("expected length 3, got %d", lru.Len())
	}

	val, ok := lru.Get("a")
	if !ok || val != 1 {
		t.Errorf("expected a=1, got ok=%v val=%d", ok, val)
	}

	val, ok = lru.Get("nonexistent")
	if ok {
		t.Errorf("expected not found for nonexistent key")
	}
}

func TestLRU_Eviction(t *testing.T) {
	lru := NewLRU[string, int](3)

	lru.Set("a", 1)
	lru.Set("b", 2)
	lru.Set("c", 3)
	lru.Set("d", 4)

	_, ok := lru.Get("a")
	if ok {
		t.Errorf("expected 'a' to be evicted")
	}

	_, ok = lru.Get("d")
	if !ok {
		t.Errorf("expected 'd' to exist")
	}
}

func TestLRU_Update(t *testing.T) {
	lru := NewLRU[string, int](3)

	lru.Set("a", 1)
	lru.Set("a", 10)

	val, _ := lru.Get("a")
	if val != 10 {
		t.Errorf("expected a=10, got %d", val)
	}

	if lru.Len() != 1 {
		t.Errorf("expected length 1, got %d", lru.Len())
	}
}

func TestLRU_Delete(t *testing.T) {
	lru := NewLRU[string, int](3)

	lru.Set("a", 1)
	lru.Set("b", 2)

	lru.Delete("a")

	_, ok := lru.Get("a")
	if ok {
		t.Errorf("expected 'a' to be deleted")
	}

	val, _ := lru.Get("b")
	if val != 2 {
		t.Errorf("expected b=2, got %d", val)
	}
}

func TestLRU_Clear(t *testing.T) {
	lru := NewLRU[string, int](3)

	lru.Set("a", 1)
	lru.Set("b", 2)

	lru.Clear()

	if lru.Len() != 0 {
		t.Errorf("expected empty LRU after clear, got %d", lru.Len())
	}
}

func TestLRU_Contains(t *testing.T) {
	lru := NewLRU[string, int](3)

	lru.Set("a", 1)

	if !lru.Contains("a") {
		t.Errorf("expected Contains('a') to be true")
	}

	if lru.Contains("b") {
		t.Errorf("expected Contains('b') to be false")
	}
}
