include ./make/config.mk

sanitize:
	@echo "--- Sanitizing code"
	CGO_ENABLED=0 go fmt ./...
	CGO_ENABLED=0 go vet ./...

unit:
	@echo "--- Running unit tests"
	CGO_ENABLED=0 go test -tags=unit ./client -v

install:
	echo "--- Installing Pact CLI dependencies"
	curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash

run-client:
	@go run ./client/cmd/main.go

integration: export PACT_TEST := true
integration: install
	@echo "--- Running Client (consumer) Pact integration tests"
	CGO_ENABLED=0 go test -tags=integration -count=1 ./client -v

publish: install
	@echo "--- Publishing Pacts"
	go run client/pact/publish.go
	@echo
	@echo "Pact contract publishing complete!"
	@echo
	@echo "Head over to $(PACT_BROKER_PROTO)://$(PACT_BROKER_URL) and login with $(PACT_BROKER_USERNAME)/$(PACT_BROKER_PASSWORD)"
	@echo "to see your published contracts.	"

.PHONY: sanitize unit install client publish
