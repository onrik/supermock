PROJECT=supermock
VERSION=1.4.1

.PHONY: openapi vendor

openapi:
	docker run --rm -i -v "$(PWD):/src" -w /src onrik/gaws:1.6.0 sh -c "gaws -t 'Supermock API' -path=/src > /src/openapi.yml"

build:
	docker build -t $(PROJECT):$(VERSION) .

push:
	docker tag $(PROJECT):$(VERSION) onrik/$(PROJECT):$(VERSION)
	docker tag $(PROJECT):$(VERSION) onrik/$(PROJECT):latest
	docker push onrik/$(PROJECT):$(VERSION)
	docker push onrik/$(PROJECT):latest

lint:
	docker run --name $(PROJECT)-lint --rm -i -v "$(PWD):/src" -w /src golangci/golangci-lint:v1.56 golangci-lint run ./pkg/... -E gofmt -E bodyclose -E gosec -E goconst -E unparam -E unconvert -E asciicheck -E exportloopref -E nilerr --timeout=10m
