# Indexer

Indexer listens StreamingManager's stream and indices streamed data

## Submodule

Registered submodules get abci Events(i.e. FinalizeBlock and Commit) and are allowed to CRUD indexer key-value storage.

- block
- tx
- move-nft
- wasm-nft
- evm-nft
- pair: common for move/evm
- wasm-pair: only for wasm
