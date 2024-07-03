```mermaid
sequenceDiagram
    autonumber
    blockchain/aws mqtt ->> pebble-sequencer: handle blockchain/mqtt event
    pebble-sequencer ->> DA: storing task to da
    DA ->> sprout-coordinator: consuming pebble task
    sprout-coordinator ->> sprout-wasm-prover: dispatch pebble task
    sprout-wasm-prover -> sprout-wasm-prover: handle pebble task(pack and commit tx)
    sprout-wasm-prover -->> blockchain/aws mqtt: -
```
