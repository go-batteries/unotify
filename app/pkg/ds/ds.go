package ds

// Not concurrent proof
type Set[E comparable] struct {
	data map[E]int
	arr  []E
}

func (set *Set[E]) Has(value E) bool {
	_, ok := set.data[value]
	return ok
}

func (set *Set[E]) Add(value E) {
	ok := set.Has(value)
	if ok {
		return
	}

	set.arr = append(set.arr, value)
	set.data[value] = len(set.arr)
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

func (m Map[E, V]) Add(key E, value V) {
	m[key] = value
}

func (m Map[E, V]) Has(key E) bool {
	_, ok := m[key]
	return ok
}

func (m Map[E, V]) Get(key E) V {
	value, _ := m[key]
	return value
}
