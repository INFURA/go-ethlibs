fuzz: rlp-fuzz.zip
	go-fuzz -bin=./rlp-fuzz.zip

install: 
	go get -u github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-build

rlp-fuzz.zip: fuzz.go go.mod go.sum
	go-fuzz-build

.PHONY: install fuzz
