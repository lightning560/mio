golang-lru
==========
This provides the `lru` package which implements a fixed-size
thread safe LRU cache with expire feature. It is based on [golang-lru](https://github.com/hashicorp/golang-lru).

Documentation
=============

Example
=======

Using the LRU is very simple:

```go
l, _ := NewWithExpire(128, 30*time.Second)
for i := 0; i < 256; i++ {
    l.Add(i, nil)
}
if l.Len() != 128 {
    panic(fmt.Sprintf("bad len: %v", l.Len()))
}
```

