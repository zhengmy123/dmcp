package cache

import (
	"testing"
	"time"
)

func TestTwoLevelLRU_BasicOperations(t *testing.T) {
	config := Config{
		L1Size:      3,
		L2Size:      3,
		L2Window:    time.Second,
		L2Threshold: 2,
	}
	lru := NewTwoLevelLRU[string, int](config)

	l1, l2 := lru.Len()
	if l1 != 0 || l2 != 0 {
		t.Errorf("expected empty LRU, got L1=%d L2=%d", l1, l2)
	}

	lru.Set("a", 1)
	l1, _ = lru.Len()
	if l1 != 1 {
		t.Errorf("expected L1 length 1, got %d", l1)
	}
}

func TestTwoLevelLRU_L2Admission(t *testing.T) {
	config := Config{
		L1Size:      10,
		L2Size:      10,
		L2Window:    time.Second,
		L2Threshold: 2,
	}
	lru := NewTwoLevelLRU[string, int](config)

	lru.Set("a", 1)
	time.Sleep(10 * time.Millisecond)
	lru.Set("a", 2)
	time.Sleep(10 * time.Millisecond)
	lru.Set("a", 3)

	_, l2 := lru.Len()
	if l2 != 1 {
		t.Errorf("expected 'a' to be in L2 after 3 accesses, got L2=%d", l2)
	}

	val, ok := lru.Get("a")
	if !ok || val != 3 {
		t.Errorf("expected a=3, got ok=%v val=%d", ok, val)
	}
}

func TestTwoLevelLRU_L1Only(t *testing.T) {
	config := Config{
		L1Size:      10,
		L2Size:      10,
		L2Window:    time.Second,
		L2Threshold: 2,
	}
	lru := NewTwoLevelLRU[string, int](config)

	lru.Set("a", 1)

	_, l2 := lru.Len()
	if l2 != 0 {
		t.Errorf("expected 'a' not in L2 after single access, got L2=%d", l2)
	}
}

func TestTwoLevelLRU_L2Eviction(t *testing.T) {
	config := Config{
		L1Size:      2,
		L2Size:      2,
		L2Window:    time.Second,
		L2Threshold: 2,
	}
	lru := NewTwoLevelLRU[string, int](config)

	for i := 0; i < 3; i++ {
		lru.Set(string(rune('a'+i)), i)
		time.Sleep(5 * time.Millisecond)
	}

	for i := 0; i < 3; i++ {
		lru.Set(string(rune('a'+i)), i+10)
		time.Sleep(5 * time.Millisecond)
	}

	_, l2 := lru.Len()
	if l2 > 2 {
		t.Errorf("expected L2 to respect capacity 2, got L2=%d", l2)
	}
}

func TestTwoLevelLRU_Delete(t *testing.T) {
	config := Config{
		L1Size:      10,
		L2Size:      10,
		L2Window:    time.Second,
		L2Threshold: 2,
	}
	lru := NewTwoLevelLRU[string, int](config)

	lru.Set("a", 1)
	lru.Set("a", 2)
	lru.Set("a", 3)

	lru.Delete("a")

	_, l2 := lru.Len()
	if l2 != 0 {
		t.Errorf("expected 'a' deleted from L2, got L2=%d", l2)
	}
}

func TestTwoLevelLRU_Clear(t *testing.T) {
	config := Config{
		L1Size:      10,
		L2Size:      10,
		L2Window:    time.Second,
		L2Threshold: 2,
	}
	lru := NewTwoLevelLRU[string, int](config)

	lru.Set("a", 1)
	lru.Set("b", 2)
	lru.Set("c", 3)

	lru.Clear()

	l1, l2 := lru.Len()
	if l1 != 0 || l2 != 0 {
		t.Errorf("expected empty LRU after clear, got L1=%d L2=%d", l1, l2)
	}
}

func TestTwoLevelLRU_DefaultConfig(t *testing.T) {
	config := Config{}
	lru := NewTwoLevelLRU[string, int](config)

	if lru.config.L2Window != time.Second {
		t.Errorf("expected default L2Window 1s, got %v", lru.config.L2Window)
	}
	if lru.config.L2Threshold != 2 {
		t.Errorf("expected default L2Threshold 2, got %d", lru.config.L2Threshold)
	}
	if lru.config.L1Size != 2000 {
		t.Errorf("expected default L1Size 2000, got %d", lru.config.L1Size)
	}
	if lru.config.L2Size != 2000 {
		t.Errorf("expected default L2Size 2000, got %d", lru.config.L2Size)
	}
}
