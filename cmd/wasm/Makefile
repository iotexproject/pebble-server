pebble_wasm:
	@tinygo build -o pebble.wasm -scheduler=none --no-debug -target=wasi main.go
	@echo built to pebble.wasm

build: pebble_wasm

clean:
	@rm -rf pebble.wasm