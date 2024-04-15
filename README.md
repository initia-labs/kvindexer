# Indexer

Indexer listens StreamingManager's stream and indices streamed data

## Submodule

Registered submodules get abci Events(i.e. FinalizeBlock and Commit) and are allowed to CRUD indexer key-value storage.

### Configuration

Default configuration will be set when the indexer is initialized.
But, to run your indexer properly, you have to set 2 configuration properties.

* set indexer.enable to true in app.toml
* set indexer.l1-chain-id to the L1's chain id app.toml

Here's example:
```toml
[indexer]
enable = true
l1-chain-id = "mahalo-2"

#...other properties..
```

Other properties are okay with default!