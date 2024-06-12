pebble_wasm:
	@tinygo build -o pebble.wasm -scheduler=none --no-debug -target=wasi main.go
	@mkdir -pv wasms && mv pebble.wasm wasms
	@echo built to wasms/pebble.wasm

build: pebble_wasm

clean:
	@rm -rf wasms/pebble.wasm