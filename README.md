dockertest
=====

dockertest is golang test tool to launch docker container every test running. 


Quickstart
---

```
go get github.com/abema/dockertest
```

Run docker in your tests
---

```go
import (
  "github.com/abema/dockertest"
  "gopkg.in/redis.v3"
)

func TestFoo(t *testing.T) {

  c := dockertest.Run("redis")
  defer c.Close()

  addr := c.Addr(6379)
  client := redis.NewClient(&redis.Options{Addr: addr})
  ...

}
```

Run docker for package tests
---

```go
import "github.com/abema/dockertest"

func TestMain(m *testing.M) {
  os.Exit(testRun(m))
}

func testRun(m *testing.M) {

  c := dockertest.Run("redis")
  defer c.Close()
  os.Setenv("REDIS_ADDR", c.Addr(6379)

  c = dockertest.Run("mongodb")
  defer c.Close()
  os.Setenv("MONGODB_URL", "mongodb://" + c.Addr(27017))

  return m.Run()
}
```
