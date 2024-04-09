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

Contains all the PRs that improved the code without changing the behaviors.
-->

## [v7.0.0](https://github.com/archway-network/archway/releases/tag/v7.0.0)

### Added

- [#481](https://github.com/archway-network/archway/pull/481) - Add Release Checklist issue type 
- [#504](https://github.com/archway-network/archway/pull/504) - Interchain test gh workflow now runs on PRs targetting release branches as well as main 
- [#501](https://github.com/archway-network/archway/pull/501) - Adding x/callback module
- [#532](https://github.com/archway-network/archway/pull/532) - Adding ADR-009 for x/callback module
- [#527](https://github.com/archway-network/archway/pull/527) - Add x/cwfees module.
- [#541](https://github.com/archway-network/archway/pull/541) - Update the interchaintest framework to v7
- [#542](https://github.com/archway-network/archway/pull/542), [#548](https://github.com/archway-network/archway/pull/548) - Adding x/cwica module
- [#543](https://github.com/archway-network/archway/pull/543) - Bumping sdk to v0.47.9
- [#544](https://github.com/archway-network/archway/pull/544) - Bumping sdk to v0.47.10
- [#546](https://github.com/archway-network/archway/pull/546) - Adding x/cwerrors module
- [#549](https://github.com/archway-network/archway/pull/549) - Integrating x/cwerrors into x/cwica
- [#550](https://github.com/archway-network/archway/pull/550) - Integrating x/cwerrors into x/callback module
- [#551](https://github.com/archway-network/archway/pull/550) - Adding ADR-012 for x/cwerrors
- [#554](https://github.com/archway-network/archway/pull/554) - Bumping ibc-go to v7.4.0
- [#553](https://github.com/archway-network/archway/pull/553) - Add explicit module licenses
- [#557](https://github.com/archway-network/archway/pull/557) - Updating the module specs to include module error codes

### Improvements

- [#533](https://github.com/archway-network/archway/pull/533) - Add more x/cwfees tests and adjust the gas limit of the RequestGrant call.

### Changed

- [#505](https://github.com/archway-network/archway/pull/505) - Update release process to account for release candidates on Titus

### Fixed
- [#537](https://github.com/archway-network/archway/pull/537) - Fix issue with callback failing when module param is changed
- [#538](https://github.com/archway-network/archway/pull/538) - Fixing the interchain test gh workflow failing cuz rc tags were not recognized 
- [#539](https://github.com/archway-network/archway/pull/539) - Remediations for x/callback audit
- [#552](https://github.com/archway-network/archway/pull/552) - Fix issue with x/callback callback error code was not identified correctly when setting cwerrors


## [v6.0.0](https://github.com/archway-network/archway/releases/tag/v6.0.0)

### Added

- [#431](https://github.com/archway-network/archway/pull/431) - Added gh workflow to run IBC conformance tests on PRs
- [#439](https://github.com/archway-network/archway/pull/439) - Adding containerized localnet
- [#445](https://github.com/archway-network/archway/pull/445) - Adding Archway logo and version number to upgrade logs
- [#459](https://github.com/archway-network/archway/pull/459) - Add missing ADR references to docs index
- [#442](https://github.com/archway-network/archway/pull/442) - Upgrade Cosmos-sdk from v0.45.16 to v0.47.5 and all the other things it depends on
- [#470](https://github.com/archway-network/archway/pull/470) - Bumping x/wasmd to v0.43.0
- [#502](https://github.com/archway-network/archway/pull/502) - Improve rewards withdrawal experience by allowing a Metadata owner to set that rewards directly go to the reward address.
- [#462](https://github.com/archway-network/archway/pull/462) - adding docs ADR-008 – Improvements on rewards withdrawal experience

### Changed

- [#457](https://github.com/archway-network/archway/pull/457) - Modify the upgrade handlers to pass in all the app keepers instead of just the account keeper
- [#465](https://github.com/archway-network/archway/pull/465) - Change the name of the gh workflow job from `build` to `run-tests` as it runs tests
- [#507](https://github.com/archway-network/archway/pull/507) – Version bump wasmd to v0.45.0 and cosmos-sdk to v0.47.6
- [#529](https://github.com/archway-network/archway/pull/529) – Version bump wasmd to v0.47.6 and cosmos-sdk to v0.47.7
- [#531](https://github.com/archway-network/archway/pull/531) - Bump wasmvm from v1.5.0 to v1.5.1. Ref: [CWA-2023-004](https://github.com/CosmWasm/advisories/blob/main/CWAs/CWA-2023-004.md)
- [#534](https://github.com/archway-network/archway/pull/534) - Bump wasmvm to v1.5.2.

### Fixed

- [#496](https://github.com/archway-network/archway/pull/496) - Fix rest endpoints in App
- [#476](https://github.com/archway-network/archway/pull/476) - Fix amd64 binary compatibility on newer linux OS
- [#514](https://github.com/archway-network/archway/pull/514) - Fix snapshot db being hardcoded from goleveldb to based on config 
- [#522](https://github.com/archway-network/archway/pull/522) - Fix Archway module endpoints not showing up in swagger

### Deprecated

- [#439](https://github.com/archway-network/archway/pull/439) - Renaming `debug` image to `dev`
- [#461](https://github.com/archway-network/archway/pull/461) - Remove titus network deployment

### Improvements

- [#478](https://github.com/archway-network/archway/pull/475) – Moves x/rewards state to use collections for state management.

## ~~[RETRACTED - v5.0.1](https://github.com/archway-network/archway/releases/tag/v5.0.1)~~

## ~~[RETRACTED - v5.0.0](https://github.com/archway-network/archway/releases/tag/v5.0.0)~~

## [v4.0.3](https://github.com/archway-network/archway/releases/tag/v4.0.3)

### Changed
- https://github.com/archway-network/archway/pull/530 - [CWA-2023-004](https://github.com/CosmWasm/advisories/blob/main/CWAs/CWA-2023-004.md) - This release fixes patches a flawed dependency on cosmwasm, the patch can be immediately applied and is not consensus breaking.


## [v4.0.2](https://github.com/archway-network/archway/releases/tag/v4.0.2)

### Changed

- [#440](https://github.com/archway-network/archway/pull/440) - Retagging with v4.0.2 to prevent dual tagging of same commit and same tag name

### Fixed

- [#441](https://github.com/archway-network/archway/pull/441) - go-releaser must order tags with create date, when there are multiple tags on the same commit

## [v4.0.1](https://github.com/archway-network/archway/releases/tag/v4.0.1)

### Fixed

- [#437](https://github.com/archway-network/archway/pull/437) - Adding upgrade handler with missing burn permissions for feecollector account

## [v4.0.0](https://github.com/archway-network/archway/releases/tag/v4.0.0)

### Added

- [#429](https://github.com/archway-network/archway/pull/429) - Adding `cosmwasm_1_3` capabilities by bumping wasmd to v0.33.0
- [#430](https://github.com/archway-network/archway/pull/430) - Added gh workflow to run chain upgrade test on PRs

### Changed

- [#428](https://github.com/archway-network/archway/pull/428) - Update go version to 1.20

## [v3.0.0](https://github.com/archway-network/archway/releases/tag/v3.0.0)

### Fixed

- [#424](https://github.com/archway-network/archway/pull/424) - Update titus to v2.0.0

### Added

- [#419](https://github.com/archway-network/archway/pull/419) - Run localnet via make
- [#421](https://github.com/archway-network/archway/pull/421) - Add archwayd darwin binaries
- [#422](https://github.com/archway-network/archway/pull/422) - Add fee burn feature, fees not distributed to contracts get burned
- [#423](https://github.com/archway-network/archway/pull/423) - Update release docs
- [#425](https://github.com/archway-network/archway/pull/425) - Update ADR 004 - Contract Premiums

## [v2.0.0](https://github.com/archway-network/archway/releases/tag/v2.0.0)

### Added

- [#416](https://github.com/archway-network/archway/pull/416) - Enable ICAHost

### Fixed

- [#414](https://github.com/archway-network/archway/pull/414) - Prevent user from setting contract flat fee if rewards address is not set
- [#418](https://github.com/archway-network/archway/pull/418) - Fix authz msg decoding in x/rewards antehandlers

## [v1.0.1](https://github.com/archway-network/archway/releases/tag/v1.0.1)

- [#411](https://github.com/archway-network/archway/pull/411) - Update repository readme with correct docker containers.
- [#413](https://github.com/archway-network/archway/pull/413) - Fix incorrect gas estimation when running with `--dry-run` flag

## [v1.0.0](https://github.com/archway-network/archway/releases/tag/v1.0.0)

Archway Network - Capture the value you create!

## [v1.0.0-rc.4](https://github.com/archway-network/archway/releases/tag/v1.0.0-rc.4)

### Added

- [#409](https://github.com/archway-network/archway/pull/409) - Add cosmwasm_1_1,cosmwasm_1_2 Cosmwasm capabilities

## [v1.0.0-rc.3](https://github.com/archway-network/archway/releases/tag/v1.0.0-rc.3)

### Removed

- [#408](https://github.com/archway-network/archway/pull/408) - Remove genesis msg logging as it impacts network start up performance.

## [v1.0.0-rc.2](https://github.com/archway-network/archway/releases/tag/v1.0.0-rc.2)

### Fixes

- [#401](https://github.com/archway-network/archway/pull/401) - Update libwasmvm version to correct one in Dockerfile.deprecated
- [#402](https://github.com/archway-network/archway/pull/402) - Bump wasmvm version to 1.2.4
- [#403](https://github.com/archway-network/archway/pull/403) - Update libwasmvm version to correct one for wasmvm 1.2.4
- [#404](https://github.com/archway-network/archway/pull/403) - Fix typo in rewards query cli
- [#406](https://github.com/archway-network/archway/pull/406) - Add upgrade handler for v0.6.0 back to prevent downgrade check from panic / consensus failure;

## [v1.0.0-rc.1](https://github.com/archway-network/archway/releases/tag/v1.0.0-rc.1)

### Removed

- [#399](https://github.com/archway-network/archway/pull/399) - Remove the upgrade handler for v1 release

## [v0.6.0](https://github.com/archway-network/archway/releases/tag/v0.6.0)

### Added

- [#387](https://github.com/archway-network/archway/pull/387) - Add genmsgs module
- [#388](https://github.com/archway-network/archway/pull/388) - Add the ibc-go fee middleware
- [#389](https://github.com/archway-network/archway/pull/389) - Add v0.6 upgrade handler
- [#391](https://github.com/archway-network/archway/pull/391) - Add snapshot manager to enable state-synd for wasm
- [#395](https://github.com/archway-network/archway/pull/395) - Add openapi.yml + generate openapi.yml on proto-swagger-gen
- [#396](https://github.com/archway-network/archway/pull/396) - Add repository licenses

### Changed

- [#373](https://github.com/archway-network/archway/pull/373) - Update codeowners
- [#383](https://github.com/archway-network/archway/pull/383), [#385](https://github.com/archway-network/archway/pull/385), [#386](https://github.com/archway-network/archway/pull/386) - Upgrade wasmd to the v0.32.0-archway fork
- [#388](https://github.com/archway-network/archway/pull/388) - Add the ibc-go fee middleware
- [#390](https://github.com/archway-network/archway/pull/390) - Update cosmos-sdk version from v0.45.15 to v0.15.16

### Fixed

- [#392](https://github.com/archway-network/archway/pull/392) - Update to ibc-go v4.3.1 for huckleberry
- [#393](https://github.com/archway-network/archway/pull/393) - Add audit remediations
- [#397](https://github.com/archway-network/archway/pull/397) - Fix map iteration

## [v0.5.2](https://github.com/archway-network/archway/releases/tag/v0.5.2)

### Fixed

- [#382](https://github.com/archway-network/archway/pull/382) - Adjust default power reduction

## [v0.5.0](https://github.com/archway-network/archway/releases/tag/v0.5.0)

### Breaking Changes

- [#357](https://github.com/archway-network/archway/pull/357) - Bump the proto versions for x/rewards and x/tracking from `v1beta1` to `v1`

### Added

- [#330](https://github.com/archway-network/archway/pull/330) - Proper chain upgrade flow.
- [#351](https://github.com/archway-network/archway/pull/351) - Add minimum price of gas.
- [#339](https://github.com/archway-network/archway/pull/339) - Update & Quality Control
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

- [#342](https://github.com/archway-network/archway/pull/342) - updated the contract premium ADR docs to elaborate on the difference between using Contract Premiums and using x/wasmd funds

## [v0.4.0](https://github.com/archway-network/archway/releases/tag/v0.4.0)

### Fixed

- [#338](https://github.com/archway-network/archway/pull/338) - fixed issue where contract premium was not completely being sent to the rewards address

## [v0.3.1](https://github.com/archway-network/archway/releases/tag/v0.3.1)

### Fixed

- [#335](https://github.com/archway-network/archway/pull/335) - fixed `EstimateTxFees` erroring when minConsFee and contract premium are same denom

## [v0.3.0](https://github.com/archway-network/archway/releases/tag/v0.3.0)

### Added

- [#253](https://github.com/archway-network/archway/pull/253) - add wasm bindings for contracts to query the x/gov module.
- [#261](https://github.com/archway-network/archway/pull/261), [#263](https://github.com/archway-network/archway/pull/263), [#264](https://github.com/archway-network/archway/pull/264), [#274](https://github.com/archway-network/archway/pull/274), [#272](https://github.com/archway-network/archway/pull/272), [#280](https://github.com/archway-network/archway/pull/280) - implementing contract premiums
- [#303](https://github.com/archway-network/archway/pull/303) - Add archway protocol versioning and release strategy
- [#326](https://github.com/archway-network/archway/pull/326) - Allow contracts to update another contract's metadata when it is the owner

### Changed

- [#267](https://github.com/archway-network/archway/pull/267) - update `querySrvr.EstimateTxFees` to also consider contract flat fee when returning the estimated fees.
- [#271](https://github.com/archway-network/archway/pull/271) - update the x/rewards/min_cons_fee antehandler to check for contract flat fees
- [#275](https://github.com/archway-network/archway/pull/275) - update the x/rewards/genesis to import/export for contract flat fees

## [v0.1.0](https://github.com/archway-network/archway/releases/tag/v0.1.0)

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

- [#191](https://github.com/archway-network/archway/pull/191) - make localnet override entrypoint
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

## [v0.0.5](https://github.com/archway-network/archway/releases/tag/v0.0.5)

### Breaking Changes

- Update wasmd to 0.25.

### Fixed

- Fix logs printing total contract rewards instead of gas rebate reward.
- Replace info logs for debug logs.

## [v0.0.4](https://github.com/archway-network/archway/releases/tag/v0.0.4)

### Breaking Changes

- Replace wasmd KV store to KV Multistore.
- Split WasmVM gas & SDK Gas.

## v0.0.3

- inflation reward calculation now depend upon block gas limit.
- inflation reward is given even when the gas rebate to the user flag is true.
- gastracker's begin blocker takes into account gas rebate to user governance switch.
- fix gas estimation for `--gas auto` flag.
