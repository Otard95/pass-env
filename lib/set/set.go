package set

type Set[T comparable] map[T]struct{}

var sentinel = struct{}{}

func (s *Set[T]) Contains(el T) bool {
	_, found := (*s)[el]
	return found
}

func (s *Set[T]) Add(el T) bool {
	if s.Contains(el) {
		return false
	}
	(*s)[el] = sentinel
	return true
}

func (s *Set[T]) Remove(el T) {
	delete(*s, el)
}

func (s *Set[T]) Items() []T {
	items := make([]T, 0, len(*s))
	for el := range *s {
		items = append(items, el)
	}
	return items
}

func (s *Set[T]) Merge(other Set[T]) {
	for el := range other {
		(*s)[el] = sentinel
	}
}
