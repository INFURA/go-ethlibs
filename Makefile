# From https://marcofranssen.nl/manage-go-tools-via-go-modules/
.PHONY: go-tools-bootstrap
go-tools-bootstrap:
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %
	asdf reshim golang

gen-mocks:
	@go generate -run mockgen ./...

.PHONY: asdf-bootstrap
asdf-bootstrap:
	asdf plugin-add kubebuilder https://github.com/virtualstaticvoid/asdf-kubebuilder.git || true
	asdf plugin-add kind https://github.com/johnlayton/asdf-kind.git || true
	asdf plugin-add golang https://github.com/kennyp/asdf-golang.git || true
	asdf plugin-add golangci-lint https://github.com/hypnoglow/asdf-golangci-lint.git || true
	asdf plugin-add yarn https://github.com/twuni/asdf-yarn.git || true
	asdf plugin-add nodejs https://github.com/asdf-vm/asdf-nodejs.git || true
	asdf plugin-add kubectl https://github.com/Banno/asdf-kubectl.git || true
	asdf plugin-add kustomize https://github.com/Banno/asdf-kustomize.git || true
	asdf plugin-add pulumi https://github.com/canha/asdf-pulumi.git || true
	asdf plugin-add protoc https://github.com/paxosglobal/asdf-protoc.git || true 
	asdf plugin-add python https://github.com/danhper/asdf-python.git || true
	asdf plugin add awscli https://github.com/MetricMike/asdf-awscli.git || true
	NODEJS_CHECK_SIGNATURES=no asdf install

