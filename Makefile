DEBUG_FLAG = $(if $(DEBUG),-debug)

build:
	go build

fmt:
	go fmt ./...

pkg:
	go get github.com/mitchellh/gox/...
	go get github.com/tcnksm/ghr
	mkdir -p pkg && cd pkg && gox ../...

clean:
	rm -f redis-benchmark

