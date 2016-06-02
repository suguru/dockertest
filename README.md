dockertest
=====

[![Build Status](https://travis-ci.org/suguru/dockertest.svg?branch=master)](https://travis-ci.org/suguru/dockertest)
[![codecov](https://codecov.io/gh/suguru/dockertest/branch/master/graph/badge.svg)](https://codecov.io/gh/suguru/dockertest)

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

Wait until container network ports
---

Waiting until port opened.

```go
c := dockertest.Run("redis")
c.WaitPort(6379, 5 * time.Second)
addr := c.Addr(6379)
```

Waiting until HTTP returns valid status code(200-299).

```go
c := dockertest.Run("redis")
c.WaitHTTP(6379, "/", 5 * time.Second)
addr := c.Addr(6379)
```
