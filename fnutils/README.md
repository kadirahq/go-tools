# Fn Utils

A collection of reusable utilities to extend the behavior of go functions. More information can be found on [godoc](http://godoc.org/github.com/kadirahq/go-tools/fnutils).

## Batch

Batch wraps a function to run only once when called multiple times. While running, all function calls will be collected for the next batch. The Flush method can be used to run the function once and release all waiting go-routines. This can be useful for synchronizing data (ex. running file.Sync).
