# Copy-On-Write Map (CowMap)

| en | cn |
| ---- | ---- |
| There is a scenario where a map with a small amount of data is in use with very many reads and very few writes. In order to optimize for this scenario, I implemented a copy-on-write map.<br/>The principle is that all reads can be concurrently read without locking, and once a write is needed, a copy of the original map is made, modified on the backup, and then the pointer is switched to the new object via an atomic operation. <br/><br/>I compared the performance of CowMap (Copy-On-Write Map) and sync.Map , and ordinary map + read/write locks.  | 有这样一种场景：数据量不多的map，在使用中读极多写极少。为了在这种场景下做极致的优化，我实现了 copy-on-write 的map: <br/>其实现原理为：所有的读都可以不加锁的并发读取，一旦需要写，则 copy 一份原来的map，在备份上修改，然后通过原子操作把指针切换到新的对象上。<br/><br/>我对比了CowMap(Copy-On-Write Map) 和 sync.Map , 以及普通map + 读写锁三种方式的性能。 | 

| Read-write rate | Type | ns/op |
| ---- | ---- | ---- |
| 99.99% : 0.01% | CowMap | 4.649  |
| 99.99% : 0.01% | map + sync.RWMutex | 187.5 |
| 99.99% : 0.01% | sync.Map |  15.06 |
| 99.90% : 0.10% | CowMap | 32.70  |
| 99.90% : 0.10% | map + sync.RWMutex | 159.9 |
| 99.90% : 0.10% | sync.Map |  14.08 |
| 99.00% : 1.00% | CowMap | 303.6  |
| 99.00% : 1.00% | map + sync.RWMutex | 105.7 |
| 99.00% : 1.00% | sync.Map | 14.08 |

因此，当读的比例超过 99.99%时，`CowMap` 是 `sync.Map` 的 3.24 倍。是 `map+sync.RWMutex` 的 40.33 倍。
Thus, when reading more than 99.99%, `CowMap` is 3.24 times as large as `sync.Map`. 40.33 times more than `map+sync.RWMutex`.

## How to use

```bash
go get github.com/ahfuzhang/cowmap@latest
```

```go
import (
    "github.com/ahfuzhang/cowmap"
)

func main(){
    var m cowmap.CowMap[uint64, uint64]
	m.Set(1, 2)
    m.Set(3, 4)
    value, has := m.Get(1)
    fmt.Println(value, has)
    m.Delete(3)
}

```
