# Function

> status: incomplete

A collection of reusable utilities to extend the behavior of go functions. More information about this package and example code can be found on [godoc](http://godoc.org/github.com/kadirahq/go-tools/function).

## Group

Group wraps a function to run only once even when called multiple times. While running, all function calls will be collected for the next batch. The Flush method can be used to run the function once and release all waiting go-routines. This can be useful for synchronizing data (ex. running file.Sync, mmap.Sync).

### Installing

``` shell
go get -u github.com/kadirahq/go-tools/logger
```

### Examples

``` go
// wrap a function with group
fn := function.NewGroup(func() {
  fmt.Println("---")
})

wg := sync.WaitGroup{}
wg.Add(3)

for i := 0; i < 3; i++ {
  go func(i int) {
    fmt.Println("b:", i)
    wg.Done()

    // Wait until the group.Flush method is called.
    // It's called after starting all goroutines.
    fn.Run()

    fmt.Println("a:", i)
  }(i)
}

// wait until all goroutines are started
// this will trigger the payload fn once
// and release all 3 waiting goroutines
wg.Wait()
fn.Flush()
```
