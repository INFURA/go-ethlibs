SRC_FILES = $(filter-out zz_deepcopy_generated.go, $(wildcard *.go))

# This Makefile uses deepcopy-gen to generate new DeepCopy methods if the
# corresponding *.go changed.
#
# To install deepcopy-gen simply run:
# GO111MODULE=off go install k8s.io/code-generator/cmd/deepcopy-gen@latest
zz_deepcopy_generated.go: ${SRC_FILES}
	pushd .. && \
		deepcopy-gen -O zz_deepcopy_generated -i ./eth --go-header-file /dev/null --trim-path-prefix "github.com/justinwongcn/go-ethlibs" ; \
	popd

