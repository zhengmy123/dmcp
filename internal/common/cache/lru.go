package cache

type entry[K comparable, V any] struct {
	key  K
	val  V
	item *ListItem[*entry[K, V]]
}

type LRU[K comparable, V any] struct {
	capacity int
	items    map[K]*entry[K, V]
	order    *List[*entry[K, V]]
}

func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	return &LRU[K, V]{
		capacity: capacity,
		items:    make(map[K]*entry[K, V]),
		order:    NewList[*entry[K, V]](),
	}
}

func (l *LRU[K, V]) Get(key K) (V, bool) {
	e, ok := l.items[key]
	if !ok {
		return *new(V), false
	}
	l.order.MoveToFront(e.item)
	return e.val, true
}

func (l *LRU[K, V]) Set(key K, value V) {
	if e, ok := l.items[key]; ok {
		l.order.MoveToFront(e.item)
		e.val = value
		return
	}

	if l.order.Len() >= l.capacity {
		l.evict()
	}

	e := &entry[K, V]{key: key, val: value}
	e.item = &ListItem[*entry[K, V]]{Value: e}
	l.order.PushFront(e.item)
	l.items[key] = e
}

func (l *LRU[K, V]) Delete(key K) {
	e, ok := l.items[key]
	if !ok {
		return
	}
	l.order.Remove(e.item)
	delete(l.items, key)
}

func (l *LRU[K, V]) Contains(key K) bool {
	_, ok := l.items[key]
	return ok
}

func (l *LRU[K, V]) Len() int {
	return l.order.Len()
}

func (l *LRU[K, V]) Clear() {
	l.items = make(map[K]*entry[K, V])
	l.order = NewList[*entry[K, V]]()
}

func (l *LRU[K, V]) evict() {
	back := l.order.Back()
	if back != nil {
		e := back.Value
		delete(l.items, e.key)
		l.order.Remove(back)
	}
}
