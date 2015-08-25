# Secure

A collection of wrapped primitives with thread-safe methods. These structs extends sync.RWMutex and exports Lock, Unlock, RLock and RUnlock methods and Get/Set methods. Please note that just using these structs will not automatically make the values thread safe. It is also necessary to use appropriate locks when using them. More information can be found on [godoc](http://godoc.org/github.com/kadirahq/go-tools/secure).
