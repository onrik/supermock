Super Mock
==========

Mock any http server for functional tests.

# Usage

## Standalone server

```bash
$ docker run --name supermock onrik/supermock:latest
```

```golang
package tests

import (
	"context"
	"net/http"
	
	"github.com/google/uuid"
	"github.com/onrik/supermock/client"
)

func Test() {
	mockClient := client.New("supermock:8000", nil)
	ctx := context.Background()
	testID := "Test"
	
	// add response to mock 
	_ = mockClient.Put(ctx, client.Response{
		UUID:   uuid.NewString(),
		TestID: testID,
		Method: http.MethodPost,
		Path:   "/example",
		Status: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"data":"test"}`,
	})
	
	// do test stuff ....

	// get requests to mocked endpoint
	reqs, _ = mockClient.Get(ctx, testID)
	
	// clean test data from server
	_ = mockClient.Clean(ctx, testID)
}
```

## Running from code

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
