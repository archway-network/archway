# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added 

### Changed 

### Deprecated

### Removed 

### Fixed 

### Security


## v0.0.5

### Breaking Changes
- Update wasmd to 0.25.

### Fixed
- Fix logs printing total contract rewards instead of gas rebate reward.
- Replace info logs for debug logs.

## v0.0.4

### Breaking Changes
- Replace wasmd KV store to KV Multistore.
- Split WasmVM gas & SDK Gas.

## v0.0.3
- inflation reward calculation now depend upon block gas limit.
- inflation reward is given even when the gas rebate to the user flag is true.
- gastracker's begin blocker takes into account gas rebate to user governance switch.
- fix gas estimation for `--gas auto` flag.
