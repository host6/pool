# Pool
This is a wrapper over standard `sync.Pool` which solves few problems:
- automatically provides `Release()` logic which protects against retuning the object back to the pool twice accidentally
- automatic cleanup before put and init after get if `Cleanup()` and\or `Init()` struct funcs are defined
- automatic lifetime control of owned objects: owned object can not be released manually. Automaticaly will be released on owner release
- debug tools
  - `GetObjectsInUse()`
  - tracking of code points where an object was borrowed but not returned back to the pool (in debug mode only)
  - easy to switch off pooling. Just rename `NewPool()` to `NewPoolStub()` and instances will be created new always on `Get()` and nothing will happen on `Put()`

# Install
`go get`

# Basic usage
```go

```