# supermock

```bash

$ docker run --name supermock onrik/supermock:latest

```

## Usage in code 
```golang
package main

import (
    "github.com/onrik/supermock/pkg/app"
    _ "github.com/mattn/go-sqlite3"
)

func TestMain(t testing.T) {
    s, err := app.New("127.0.0.1:9000", "sqlite://:memory:")

    s.Start()
    defer s.Stop()
}
```
