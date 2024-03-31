# Dashboard indexer

This indexer make an index for dashboard to query by range

## Configuration

```toml
# submodule name (mandatory)
name = "dashboard"

# max value of pagination.limit (optional, 100 by default)
limit = 100

# OPinit bridge ID (mandatory)
op-bridge-id = 1

# L1 denom to watch (mandatory)
l1-denom="uinit"
```