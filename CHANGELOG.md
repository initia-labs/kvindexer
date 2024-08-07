<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry is required to include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The tag should consist of where the change is being made ex. (x/staking), (store)
The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"KVIndexer Breaking" for breaking KVIndexer module.
"Submodule Breaking" for breaking submodules
Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

## [submodules/move-nft/v0.1.4](https://github.com/initia-labs/kvindexer/releases/tag/submodules/move-nft/v0.1.4) - 2024-07-26

* (submodule/move-nft) fix: don't abort on nft index failure

## [submodules/wasm-nft/v0.1.4](https://github.com/initia-labs/kvindexer/releases/tag/submodules/wasm-nft/v0.1.4) - 2024-07-26

* (submodule/wasm-nft) fix: don't abort on nft index failure

## [v0.1.6](https://github.com/initia-labs/kvindexer/releases/tag/v0.1.6) - 2024-07-26

* (submodule) deprecate: pair, wasm-pair
* (cache) Change cache-capacity unit to MiB

## [v0.1.5](https://github.com/initia-labs/kvindexer/releases/tag/v0.1.5) - 2024-07-16

### KVIndexer breaking 

* (cache) [#48](https://github.com/initia-labs/kvindexer/pull/48) Replace record-count-based lru cache with capacity-based one
* (keeper) [#49](https://github.com/initia-labs/kvindexer/pull/49) use db from outside
