package maps

type Map[K comparable, V any] map[K]V

func (m Map[K, V]) Upsert(k K, v V, onConflict func(v, newV V) V) {
	newV := v
	if oldV, exists := m[k]; exists {
		newV = onConflict(oldV, newV)
	}
	m[k] = newV
}

func (m Map[K, V]) Add(other Map[K, V], onConflict func(v, newV V) V) {
	for k, v := range other {
		m.Upsert(k, v, onConflict)
	}
}
