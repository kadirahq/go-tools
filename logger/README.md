# Logger

> status: completed

A basic logger package which supports multiple levels of logging. Users can enable or disable log levels using the "log" environment variable. By default, logger logs **info** and **error** level logs. Custom log levels can be added anytime as required. More information can be found on [godoc](http://godoc.org/github.com/kadirahq/go-tools/logger).


## Installing

``` shell
go get -u github.com/kadirahq/go-tools/logger
```

## Examples

``` go
// create a new logger
log := logger.New("webapp")
log.Info("starting application on port:", 8080)

// Stack trace will be printed when the go-errors/errors
// package is used to create (or wrap) the error.
err := errors.New("error message")
log.Error(err, "foo =>", "bar")

// debug level is not enabled by default
// use logger.Enable("debug") or the env variable
// ex. log="info,error,debug" go run main.go
log.Debug("debug options", []int{1, 2, 3, 4, 5})

// any number of custom log levels can be used
// use logger.Enable("lvl9001") or the env variable
// ex. log="info,error,lvl9001" go run main.go
log.Print("lvl9001", "!!!")
```

![Example output](https://raw.githubusercontent.com/kadirahq/go-tools/master/logger/assets/logger-example.png)
