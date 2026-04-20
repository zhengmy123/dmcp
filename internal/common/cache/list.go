package cache

type ListItem[T any] struct {
	Value T
	prev  *ListItem[T]
	next  *ListItem[T]
}

type List[T any] struct {
	head *ListItem[T]
	tail *ListItem[T]
	len  int
}

func NewList[T any]() *List[T] {
	return &List[T]{}
}

func (l *List[T]) Len() int {
	return l.len
}

func (l *List[T]) Front() *ListItem[T] {
	return l.head
}

func (l *List[T]) Back() *ListItem[T] {
	return l.tail
}

func (l *List[T]) PushFront(item *ListItem[T]) {
	item.prev = nil
	item.next = l.head
	if l.head != nil {
		l.head.prev = item
	}
	l.head = item
	if l.tail == nil {
		l.tail = item
	}
	l.len++
}

func (l *List[T]) PushBack(item *ListItem[T]) {
	item.next = nil
	item.prev = l.tail
	if l.tail != nil {
		l.tail.next = item
	}
	l.tail = item
	if l.head == nil {
		l.head = item
	}
	l.len++
}

func (l *List[T]) Remove(item *ListItem[T]) {
	if item.prev != nil {
		item.prev.next = item.next
	} else {
		l.head = item.next
	}
	if item.next != nil {
		item.next.prev = item.prev
	} else {
		l.tail = item.prev
	}
	l.len--
}

func (l *List[T]) MoveToFront(item *ListItem[T]) {
	if item == l.head {
		return
	}
	l.Remove(item)
	l.PushFront(item)
}

func (l *List[T]) Clear() {
	l.head = nil
	l.tail = nil
	l.len = 0
}
