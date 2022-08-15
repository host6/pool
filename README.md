# Pool
This is a wrapper over standard `sync.Pool` which solves few problems:
- provides `Release()` logic which protects against retuning the object back to the pool twice accidentally
- provides cleanup logic
- provides `GetObjectsInUse()` and tracking of code points where an object was borrowed but not returned back to the pool (in debug mode only). That makes control of objects usage easier

# Install
`go get`

# Basic usage
```go

```