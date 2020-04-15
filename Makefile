all: build

build: build-eshop

build-eshop:
	go build -v -i -o build/eshop .

clean:
	rm -rf build

.PHONY: build build-eshop clean
