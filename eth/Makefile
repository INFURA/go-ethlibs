SRC_FILES = $(filter-out zz_deepcopy_generated.go, $(wildcard *.go))

# This Makefile uses deepcopy-gen to generate new DeepCopy methods if the
# corresponding *.go changed.
#
# To install deepcopy-gen simply run:
# GO111MODULE=off go install k8s.io/gengo/examples/deepcopy-gen
zz_deepcopy_generated.go: ${SRC_FILES}
	deepcopy-gen -o ../../../.. -O zz_deepcopy_generated -i .

