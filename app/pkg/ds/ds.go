package ds

type Iterator[E any] interface {
	Next() (item E, more bool)
}

type SeqIter[E any] struct {
	items []E
	idx   int
}

func NewSeqItr[E any](items ...E) *SeqIter[E] {
	return &SeqIter[E]{items: items, idx: 0}
}

func (itr *SeqIter[E]) Next() (E, bool) {
	var value E

	if itr.idx >= len(itr.items) {
		return value, false
	}

	value = itr.items[itr.idx]
	itr.idx += 1
	return value, true
}

// Not concurrent proof
type Set[E comparable] struct {
	data map[E]int
	arr  []E
}

func NewSet[E comparable]() *Set[E] {
	set := &Set[E]{
		data: make(map[E]int),
		arr:  []E{},
	}
	return set
}

func (s *Set[E]) Iter() Iterator[E] {
	return NewSeqItr[E](s.arr...)
}

func (set *Set[E]) Has(value E) bool {
	_, ok := set.data[value]
	return ok
}

func (set *Set[E]) Add(value E) *Set[E] {
	ok := set.Has(value)
	if ok {
		return set
	}

	set.arr = append(set.arr, value)
	set.data[value] = len(set.arr)
	return set
}

func (set *Set[E]) Get(idx int) E {
	if idx >= len(set.arr) {
		return *new(E)
	}

	return set.arr[idx]
}

func (set *Set[E]) Count() int {
	return len(set.data)
}

func ToSet[E comparable](values ...E) *Set[E] {
	set := &Set[E]{data: map[E]int{}, arr: []E{}}

	for _, value := range values {
		set.Add(value)
	}

	return set
}

type Map[E comparable, V any] map[E]V

func NewMap[E comparable, V any]() Map[E, V] {
	return make(Map[E, V])
}

func (m Map[E, V]) Add(key E, value V) Map[E, V] {
	m[key] = value
	return m
}

func (m Map[E, V]) Has(key E) bool {
	_, ok := m[key]
	return ok
}

func (m Map[E, V]) Get(key E) V {
	var defaultValue V

	value, ok := m[key]
	if !ok {
		return defaultValue
	}

	return value
}
