package cowmap

import (
	"sync/atomic"
)

// CowMap Copy-On-Write map
type CowMap[K comparable, V any] struct {
	data atomic.Value
}

func (c *CowMap[K, V]) isNull(v interface{}) bool {
	if v == nil {
		return true
	}
	ptr, ok := v.(*map[K]V)
	if !ok {
		panic("impossible error")
	}
	return ptr == nil
}

// Set key and value
func (c *CowMap[K, V]) Set(key K, value V) {
	for {
		m := c.data.Load()
		if c.isNull(m) {
			m1 := c.data.Load()
			if c.isNull(m1) {
				// create a new map
				tempMap := map[K]V{
					key: value,
				}
				if c.data.CompareAndSwap(nil, &tempMap) {
					return
				}
				m = c.data.Load()
				// todo: those are test code
				if c.isNull(m) {
					panic("impossible error")
				}
			} else {
				m = m1
			}
		}
		//
		tempMap, ok := m.(*map[K]V)
		// todo: those are test code
		if !ok {
			panic("impossible error")
		}
		// check key exists
		// if _, has := (*tempMap)[key]; has {
		// 	(*tempMap)[key] = value
		//  // Warning: Although modifying the value of a key in a map is theoretically thread-safe.
		//  // However, this code seems to tell the compiler that this is not a read-only map,
		//  //  and therefore the race condition is bound to cause a panic when assignments are used.
		// 	return
		// }
		// copy map
		newMap := make(map[K]V, len(*tempMap)+1)
		for k1, v1 := range *tempMap {
			newMap[k1] = v1
		}
		newMap[key] = value
		if c.data.CompareAndSwap(m, &newMap) {
			return
		}
		/*
			TODO: The strategy for spin waiting should be four levels:
				1. spin waiting a finite number of times
				2. concatenation level of yield scheduling
				3. operating system physical thread level cede scheduling
				4. operating system process-level give-and-take scheduling
		*/
	}
}

// Get value by key
func (c *CowMap[K, V]) Get(key K) (value V, has bool) {
	m := c.data.Load()
	if c.isNull(m) {
		return
	}
	tempMap, ok := m.(*map[K]V)
	// todo: those are test code
	if !ok {
		panic("impossible error")
	}
	value, has = (*tempMap)[key]
	return
}

// Delete from map by key
func (c *CowMap[K, V]) Delete(key K) {
	for {
		m := c.data.Load()
		if c.isNull(m) {
			return
		}
		//
		tempMap, ok := m.(*map[K]V)
		// todo: those are test code
		if !ok {
			panic("impossible error")
		}
		if _, has := (*tempMap)[key]; !has {
			return
		}
		// copy map
		newMap := make(map[K]V, len(*tempMap)-1)
		for k1, v1 := range *tempMap {
			if k1 != key {
				newMap[k1] = v1
			}
		}
		if c.data.CompareAndSwap(m, &newMap) {
			return
		}
	}
}

// Clear all
func (c *CowMap[K, V]) Clear() {
	c.data.Store(&map[K]V{})
}

// ForEach iterates all keys and values in the map
func (c *CowMap[K, V]) ForEach(callback func(key K, value V) (isStop bool)) {
	m := c.data.Load()
	if c.isNull(m) {
		return
	}
	tempMap, ok := m.(*map[K]V)
	// todo: those are test code
	if !ok {
		panic("impossible error")
	}
	for k1, v1 := range *tempMap {
		if callback(k1, v1) {
			break
		}
	}
}

// Len returns the number of map
func (c *CowMap[K, V]) Len() int {
	m := c.data.Load()
	if c.isNull(m) {
		return 0
	}
	tempMap, ok := m.(*map[K]V)
	// todo: those are test code
	if !ok {
		panic("impossible error")
	}
	return len(*tempMap)
}

func (c *CowMap[K, V]) SetMap(m map[K]V) {
	c.data.Store(&m)
}
