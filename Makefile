# From https://marcofranssen.nl/manage-go-tools-via-go-modules/
.PHONY: go-tools-bootstrap
go-tools-bootstrap:
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %
	asdf reshim golang

gen-mocks:
	@go generate -run mockgen ./...
