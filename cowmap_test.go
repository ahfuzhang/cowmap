package cowmap

import (
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CowMap_Set(t *testing.T) {
	var m CowMap[uint64, uint64]
	assert.Equal(t, m.Len(), 0)
	m.Set(1, 2)
	assert.Equal(t, m.Len(), 1)
	var found bool
	m.ForEach(func(key, value uint64) (isStop bool) {
		if key == 1 && value == 2 {
			found = true
			return true
		}
		return
	})
	assert.Equal(t, found, true)
	v1, has1 := m.Get(1)
	assert.Equal(t, has1, true)
	assert.Equal(t, v1, uint64(2))
	//
	m.Set(3, 4)
	assert.Equal(t, m.Len(), 2)
	found = false
	m.ForEach(func(key, value uint64) (isStop bool) {
		if key == 3 && value == 4 {
			found = true
			return true
		}
		return
	})
	assert.Equal(t, found, true)
	v1, has1 = m.Get(3)
	assert.Equal(t, has1, true)
	assert.Equal(t, v1, uint64(4))
	// test delete
	m.Delete(5)
	assert.Equal(t, m.Len(), 2)
	m.Delete(3)
	assert.Equal(t, m.Len(), 1)
}

func benchmarkCowmap(b *testing.B, writeEdge int, maxItem int) {
	var m CowMap[int, int]
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v := rand.Intn(10000)
			v1 := rand.Intn(10000)
			if v >= writeEdge { // 99.xxx% read
				m.Set(v1, v1)
			} else {
				_, _ = m.Get(v1)
			}
			if m.Len() > maxItem {
				m.Clear()
			}
		}
	})
}

func benchmarkRWMap(b *testing.B, writeEdge int, maxItem int) {
	var m = map[int]int{}
	var rw sync.RWMutex
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v := rand.Intn(10000)
			v1 := rand.Intn(10000)
			if v >= writeEdge { // 99.xxx% read
				rw.Lock()
				m[v1] = v1
				rw.Unlock()
			} else {
				rw.RLock()
				_, _ = m[v1]
				rw.RUnlock()
			}
			rw.RLock()
			l := len(m)
			rw.RUnlock()
			if l > maxItem {
				rw.Lock()
				arr := make([]int, 0, l)
				for k := range m {
					arr = append(arr, k)
				}
				for _, i := range arr {
					delete(m, i)
				}
				rw.Unlock()
			}
		}
	})
}

func benchmarkSyncMap(b *testing.B, writeEdge int, maxItem int) {
	var m sync.Map
	var cnt atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v := rand.Intn(10000)
			v1 := rand.Intn(10000)
			if v == writeEdge { // 99.xx% read
				m.Store(v1, v1)
				cnt.Add(1)
			} else {
				_, _ = m.Load(v1)
			}
			l := int(cnt.Load())
			if l > maxItem {
				arr := make([]int, 0, l)
				m.Range(func(key, value any) bool {
					arr = append(arr, key.(int))
					return true
				})
				for _, i := range arr {
					m.Delete(i)
					cnt.Add(-1)
				}
			}
		}
	})
}

// go test -benchmem -run=^$ -bench ^Benchmark_cowmap$ github.com/ahfuzhang/cowmap
// 1 core: Benchmark_cowmap-12     30414904                38.59 ns/op           17 B/op          0 allocs/op
/*
// go test -gcflags="-N -l" -benchmem -run=^$ -bench ^Benchmark_cowmap$ github.com/ahfuzhang/cowmap
2 core: Benchmark_cowmap-12     38047376                31.32 ns/op           28 B/op          0 allocs/op
8 core: Benchmark_cowmap-12     36631107                31.47 ns/op           91 B/op          0 allocs/op

go test -benchmem -run=^$ -bench ^Benchmark_cowmap$ github.com/ahfuzhang/cowmap
Benchmark_cowmap-12     43832013                30.07 ns/op          102 B/op          0 allocs/op
*/
func Benchmark_cowmap(b *testing.B) {
	runtime.GOMAXPROCS(8)
	benchmarkCowmap(b, 9990, 1000)
}

/*
go test -benchmem -run=^$ -bench ^Benchmark_map_rwlock$ github.com/ahfuzhang/cowmap
Benchmark_map_rwlock-12          7418427               148.3 ns/op             0 B/op          0 allocs/op
*/
func Benchmark_map_rwlock(b *testing.B) {
	runtime.GOMAXPROCS(8)
	benchmarkRWMap(b, 9990, 1000)
}

/*
// Benchmark_sync_map-12            1877944               580.6 ns/op            21 B/op          0 allocs/op
Benchmark_sync_map-12           33047380                31.07 ns/op            9 B/op          0 allocs/op
*/
func Benchmark_sync_map(b *testing.B) {
	runtime.GOMAXPROCS(8)
	benchmarkSyncMap(b, 9990, 1000)
}

/*
Benchmark_cowmap_99-12           4268751               288.4 ns/op          1076 B/op          0 allocs/op
*/
func Benchmark_cowmap_99(b *testing.B) {
	runtime.GOMAXPROCS(8)
	benchmarkCowmap(b, 9900, 1000)
}

/*
Benchmark_map_rwlock_99-12       6794378               170.2 ns/op             0 B/op          0 allocs/op
*/
func Benchmark_map_rwlock_99(b *testing.B) {
	runtime.GOMAXPROCS(8)
	benchmarkRWMap(b, 9900, 1000)
}

/*
Benchmark_sync_map_99-12        78336960                13.74 ns/op            2 B/op          0 allocs/op
*/
func Benchmark_sync_map_99(b *testing.B) {
	runtime.GOMAXPROCS(8)
	benchmarkSyncMap(b, 9900, 1000)
}

/*
Benchmark_cowmap_9999-12        245512680                4.415 ns/op           3 B/op          0 allocs/op
*/
func Benchmark_cowmap_9999(b *testing.B) {
	runtime.GOMAXPROCS(8)
	benchmarkCowmap(b, 9999, 1000)
}

/*
Benchmark_map_rwlock_9999-12             6451852               177.4 ns/op             0 B/op          0 allocs/op
*/
func Benchmark_map_rwlock_9999(b *testing.B) {
	runtime.GOMAXPROCS(8)
	benchmarkRWMap(b, 9999, 1000)
}

func Benchmark_all(b *testing.B) {
	runtime.GOMAXPROCS(8)
	b.Run("99.99% read, cow", func(b *testing.B) { benchmarkCowmap(b, 9999, 1000) })
	b.Run("99.90% read cow", func(b *testing.B) { benchmarkCowmap(b, 9990, 1000) })
	b.Run("99.00% read cow", func(b *testing.B) { benchmarkCowmap(b, 9900, 1000) })
	//
	b.Run("99.99% read rw ", func(b *testing.B) { benchmarkRWMap(b, 9999, 1000) })
	b.Run("99.90% read rw ", func(b *testing.B) { benchmarkRWMap(b, 9990, 1000) })
	b.Run("99.00% read rw ", func(b *testing.B) { benchmarkRWMap(b, 9900, 1000) })
	//
	b.Run("99.99% read sync", func(b *testing.B) { benchmarkSyncMap(b, 9999, 1000) })
	b.Run("99.90% read sync", func(b *testing.B) { benchmarkSyncMap(b, 9990, 1000) })
	b.Run("99.00% read sync", func(b *testing.B) { benchmarkSyncMap(b, 9900, 1000) })
}
