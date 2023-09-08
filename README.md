# Pool
This is a wrapper over standard `sync.Pool` which solves few problems:
- automatically provides `Release()` logic with the folowing features:
  - protect against retuning the object back to the pool twice accidentally
  - automatic cleanup before release if `Cleanup()` struct func is defined
- automatic init after borrow if `Init()` struct funcs is defined
- automatic lifetime control of owned objects: owned object can not be released manually. Automaticaly will be released on owner release only
- debug tools
  - `GetObjectsInUse()`. Very useful to have ```require.Zero(t, pool.GetObjectsInUse())``` at the end of each test
  - tracking of code points where an object was borrowed but not returned back to the pool (in debug mode only). Call `pool.PrintNonReleased(os.Stdout)` at the end of your test to catch leaks
  - easy to switch off pooling for debug purposes: just use `NewPoolStub()` instead of `NewPool()` and the pool will not actually be used

# Install
`go get https://github.com/host6/pool`

# Usage
- see [basic usage test](https://github.com/host6/pool/blob/main/impl_test.go#L41)