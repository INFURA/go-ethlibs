fuzz: rlp-fuzz.zip
	go-fuzz -bin=./rlp-fuzz.zip

install: 
	go install github.com/dvyukov/go-fuzz/go-fuzz@latest github.com/dvyukov/go-fuzz/go-fuzz-build@latest

rlp-fuzz.zip: fuzz.go go.mod go.sum
	go-fuzz-build

.PHONY: install fuzz
