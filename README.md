# Pool
This is a wrapper over standard `sync.Pool` which solves few problems:
- automatically provides `Release()` logic with the folowing features:
  - protect against retuning the object back to the pool twice accidentally
  - automatic cleanup before release and init after borrow if `Cleanup()` and\or `Init()` struct funcs are defined
- automatic lifetime control of owned objects: owned object can not be released manually. Automaticaly will be released on owner release only
- debug tools
  - `GetObjectsInUse()`
  - tracking of code points where an object was borrowed but not returned back to the pool (in debug mode only)
  - easy to switch off pooling. Just rename `NewPool()` to `NewPoolStub()` and instances will be created new always on `Get()` and nothing will happen on `Put()`

# Install
`go get https://github.com/host6/pool`

# Usage
- see [basic usage test](https://github.com/host6/pool/blob/main/impl_test.go#L34)