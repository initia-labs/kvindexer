# Indexer

Indexer listens StreamingManager's stream and indices streamed data

## Submodule

Registered submodules get abci Events(i.e. FinalizeBlock and Commit) and are allowed to CRUD indexer key-value storage.

### Configuration

Submodules can be configured via `[[indexer.submodules]]` section array in `config/app.toml`.
The array defines the configuration for each submodule and it should have a key-value pair named 'name', and the value should exist in enabled_submodules.
If a named submodule is not exist in `indexer.enabled-submodules` in `config/app.toml`, it will be ignored.

see `README.md`s for examples

### Basic Submodules

* [block](https://github.com/initia-labs/kvindexer/tree/main/submodule/block)
* [tx](https://github.com/initia-labs/kvindexer/tree/main/submodule/tx)
* [nft](https://github.com/initia-labs/kvindexer/tree/main/submodule/nft)
* [pair](https://github.com/initia-labs/kvindexer/tree/main/submodule/pair)

## Crontab

Indexer handles a simple crontab.

### Configuration

Submodules can be configured via `[[indexer.submodules]]` section array in `config/app.toml`.
The array defines the configuration for each submodule and it should have 2 key-value pairs named `name` and `pattern`, and the name should exist in enabled_submodules.
If a named submodule is not exist in `indexer.enabled-cronjobs` in `config/app.toml`, it will be ignored.
