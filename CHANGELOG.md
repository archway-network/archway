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

## [Unreleased]

### Added

- [#387](https://github.com/archway-network/archway/pull/387) - Add genmsgs module
- [#388](https://github.com/archway-network/archway/pull/388) - Add the ibc-go fee middleware
- [#389](https://github.com/archway-network/archway/pull/389) - Add v0.6 upgrade handler
- [#391](https://github.com/archway-network/archway/pull/391) - Added snapshot manager to enable state-synd for wasm
- [#395](https://github.com/archway-network/archway/pull/395) - Added openapi.yml + generate openapi.yml on proto-swagger-gen

### Changed

- [#373](https://github.com/archway-network/archway/pull/373) - Update codeowners
- [#383](https://github.com/archway-network/archway/pull/383), [#385](https://github.com/archway-network/archway/pull/385), [#386](https://github.com/archway-network/archway/pull/386) - upgrade wasmd to the v0.32.0-archway fork
- [#388](https://github.com/archway-network/archway/pull/388) - add the ibc-go fee middleware
- [#390](https://github.com/archway-network/archway/pull/390) - update cosmos-sdk version from v0.45.15 to v0.15.16

### Deprecated

### Removed

### Fixed
- [#392](https://github.com/archway-network/archway/pull/392) - Updating to ibc-go v4.3.1 for huckleberry
- [#393](https://github.com/archway-network/archway/pull/393) - Add audit remediations

### Improvements


## [v0.5.2]

### Fixed

- [#382](https://github.com/archway-network/archway/pull/382) - adjust default power reduction

## [v0.5.0]

### Breaking Changes 

- [#357](https://github.com/archway-network/archway/pull/357) - Bumping the proto versions for x/rewards and x/tracking from `v1beta1` to `v1`

### Added

- [#330](https://github.com/archway-network/archway/pull/330) - Proper chain upgrade flow.
- [#351](https://github.com/archway-network/archway/pull/351) - Add minimum price of gas.
- [#339](https://github.com/archway-network/archway/pull/339) - Updates & Quality Control
    - Community Contribution Guidelines
    - Security Policy
    - ADR Log Index
    - Bug report template
    - Feature request template
    - General issue template
- [#347](https://github.com/archway-network/archway/pull/347) - Unified release for cross compiled binaries and docker images
- [#360](https://github.com/archway-network/archway/pull/360) - Fix github access token for release workflow
- [#361](https://github.com/archway-network/archway/pull/361) - Readd missing deprecated Dockerhub build phase
- [#362](https://github.com/archway-network/archway/pull/362) - wrong reference in the deploy pipeline
- [#363](https://github.com/archway-network/archway/pull/363) - move safe dir up in the pipeline
- [#364](https://github.com/archway-network/archway/pull/364) - add CODEOWNERS
- [#365](https://github.com/archway-network/archway/pull/365) - add release tests
- [#367](https://github.com/archway-network/archway/pull/367) - use snapshot for non release builds
- [#372](https://github.com/archway-network/archway/pull/372) - add docker config to release pipeline
- [#375](https://github.com/archway-network/archway/pull/375) - add missing colon from manifest
- [#376](https://github.com/archway-network/archway/pull/376) - fix checksum naming
- [#377](https://github.com/archway-network/archway/pull/377) - artifact naming
- [#378](https://github.com/archway-network/archway/pull/378) - missing end parameter
- [#380](https://github.com/archway-network/archway/pull/380) - titus deployment

### Fixed

- [#365](https://github.com/archway-network/archway/pull/356) - x/rewards genesis runs before x/genutil to correctly process genesis txs.
- [#366](https://github.com/archway-network/archway/pull/366) - github actions should fetch tags as well
- [#368](https://github.com/archway-network/archway/pull/368) - github actions should fetch tags as well for deploy workflow
- [#369](https://github.com/archway-network/archway/pull/369) - CODEOWNERS: small set to expand, not large set that filters
- [#370](https://github.com/archway-network/archway/pull/370) - login to ghcr

### Changed

- [#320](https://github.com/archway-network/archway/pull/320) - Run the lint and test GH actions on all PRs
- [#339](https://github.com/archway-network/archway/pull/339) - Updates & Quality Control
    - README.md
    - docs/README.md
- [#365](https://github.com/archway-network/archway/pull/356) - Disallow setting module accounts as reward address
- [#355](https://github.com/archway-network/archway/pull/355) - chore: Update titus genesis defaults

### Removed

- [#344](https://github.com/archway-network/archway/pull/344) - removed un used ci files

### Improvements

- [#342](https://github.com/archway-network/archway/pull/342) - updated the contract premium ADR docs to elaborate on difference between using Contract Premiums and using x/wasmd funds

## [v0.4.0]

### Fixed

- [#338](https://github.com/archway-network/archway/pull/338) - fixed issue where contract premium was not completly being sent to the rewards address

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
