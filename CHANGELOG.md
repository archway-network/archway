# Changelog

All notable changes to this project will be documented in this file.

<!-- 
### Added

Contains the new features.

### Changed

Contains API breaking changes to existing functionality.

### Deprecated

Contains the candidates for removal in a future release.

### Removed

Contains API breaking changes of removed APIs.

### Fixed

Contains bug fixes.

### Improvements

Contains all the PRs that improved the code without changing the behaviours. 
-->

## [v0.3.1]

### Fixed

- [#335](https://github.com/archway-network/archway/pull/335) - fixed `EstimateTxFees` erroring when minConsFee and contract premium are same denom


## [v0.3.0]

### Added

- [#253](https://github.com/archway-network/archway/pull/253) - add wasm bindings for contracts to query the x/gov module.
- [#261](https://github.com/archway-network/archway/pull/261), [#263](https://github.com/archway-network/archway/pull/263), [#264](https://github.com/archway-network/archway/pull/264), [#274](https://github.com/archway-network/archway/pull/274), [#272](https://github.com/archway-network/archway/pull/272), [#280](https://github.com/archway-network/archway/pull/280) - implementing contract premiums
- [#303](https://github.com/archway-network/archway/pull/303) - Add archway protocol versioning and release strategy
- [#326](https://github.com/archway-network/archway/pull/326) - Allow contracts to update another contract's metadata when it is the owner

### Changed

- [#267](https://github.com/archway-network/archway/pull/267) - update `querySrvr.EstimateTxFees` to also consider contract flat fee when returning the estimated fees.
- [#271](https://github.com/archway-network/archway/pull/271) - update the x/rewards/min_cons_fee antehandler to check for contract flat fees
- [#275](https://github.com/archway-network/archway/pull/275) - update the x/rewards/genesis to import/export for contract flat fees

## [v0.1.0]

### Added

- [#196](https://github.com/archway-network/archway/pull/196) - add wasm bindings for contracts to interact with the x/gastracking module.
- [#202](https://github.com/archway-network/archway/pull/202) - added the new x/tracking and x/rewards modules.ù
- [#210](https://github.com/archway-network/archway/pull/210) - wasm bindings API change
- [#217](https://github.com/archway-network/archway/pull/217) - improve the x/rewards withdraw UX by using defaults when params are unset.
- [#227](https://github.com/archway-network/archway/pull/227) - flatten wasmbindings query struct


### Changed

- [#180](https://github.com/archway-network/archway/pull/180) - add x/gastracking params
- [#181](https://github.com/archway-network/archway/pull/181) - simplify params
- [#185](https://github.com/archway-network/archway/pull/185) - remove pointers in proto generated slices.
- [#186](https://github.com/archway-network/archway/pull/186) - improvements on dapp inflationary reward calculation
- [#188](https://github.com/archway-network/archway/pull/188) - improvements on tx tracking
- [#193](https://github.com/archway-network/archway/pull/193) - move tx fees handling to middlewares
- [#231](https://github.com/archway-network/archway/pull/231) - use custom archway wasmd fork

### Deprecated

### Removed 

- [#206](https://github.com/archway-network/archway/pull/206) - remove the legacy x/gastracking module in favour of x/rewards and x/tracking

### Fixed

- [#191](https://github.com/archway-network/archway/pull/191) - make localnet ovveride entrypoint
- [#205](https://github.com/archway-network/archway/pull/205) - fix go.mod
- [#216](https://github.com/archway-network/archway/pull/216) - fix dry-run cmd and bump cosmos-sdk do v0.45.8
- [#218](https://github.com/archway-network/archway/pull/218) - x/rewards unique ID genesis export/import
- [#228](https://github.com/archway-network/archway/pull/228) - testing, fix validator propagation in test chain

### Improvements

- [#182](https://github.com/archway-network/archway/pull/182) - refactor and simplify code.
- [#183](https://github.com/archway-network/archway/pull/183) - update to go1.18
- [#184](https://github.com/archway-network/archway/pull/184) - refactor, move event emission into its own file.
- [#204](https://github.com/archway-network/archway/pull/204) - upgrade IBC to v3 and wasmd to v0.27.0
- [#211](https://github.com/archway-network/archway/pull/211) - update gh action deployment flow
- [#212](https://github.com/archway-network/archway/pull/212) - upgrade to cosmos-sdk v0.45.7
- [#213](https://github.com/archway-network/archway/pull/213) - improve gh action deployment flow
- [#224](https://github.com/archway-network/archway/pull/224) - fix codecov action
- [#225](https://github.com/archway-network/archway/pull/225) - add editorconfig settings
- [#226](https://github.com/archway-network/archway/pull/226) - ci cache go packages to speed up builds
- [#232](https://github.com/archway-network/archway/pull/232) - Makefile to create statically linked binaries
- [#233](https://github.com/archway-network/archway/pull/233) - add the commit version on builds
- [#236](https://github.com/archway-network/archway/pull/236) - add more tests for x/rewards
- [#237](https://github.com/archway-network/archway/pull/237) - add more tests for x/rewards 2
- [#241](https://github.com/archway-network/archway/pull/241) - add golang linter gh action
- [#242](https://github.com/archway-network/archway/pull/242) - add changelog check gh action
- [#243](https://github.com/archway-network/archway/pull/243) - add pr lint gh action
- [#247](https://github.com/archway-network/archway/pull/247) - fix Dockerfile libwasm VM dependencies
- [#249](https://github.com/archway-network/archway/pull/249) - add go releaser, fill changelog history

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
