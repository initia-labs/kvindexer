# NFT indexer

This indexer make an index of pairs of L1 assets and L2 assets

## Configuration

```toml
# submodule name (mandatory)
name = "pair"

# max value of pagination.limit (optional, 100 by default)
limit = 100

# cron pattern to query L1 assets (mandatory)
l1_query_pattern = "* * * * *" # for every minutes

# L1's LCD URL to query L1 assets (mandatory)
l1_lcd_url = "https://lcd..."

# OPinit bridge ID (mandatory)
bridge_id = number

# L1's chain ID (mandatorty)
l1_chain_id = "mahalo-2"
```
