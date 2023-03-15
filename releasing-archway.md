# Archway Protocol Release Versioning

This document captures the release and versioning strategy of archway protocol software.

This document DOES NOT deal with versioning or releases of Various networks using archway protocol.

This Document DOES NOT deal with API versioning or API conventions of archway protocol

Reference: [Semantic Versioning](http://semver.org)

Legend:

- **X.Y.Z** refers to the version (git tag) of Archway Protocol that is released.
- **Network Operator** refers to an entity running a node and/or a validator and/or a relayer.
- **archway-1** refers to the chain id of archway protocl main network.
- **constantine-1** refers to the chain-id of the archway protocol "staging" public test network,
  with long term support, gurantee of state persistance during upgrades and reasonably mirrors mainnet.
  Dapp developers may build
  and test their dapps against this testnet before launching on mainnet.
- **titus-1** refers to the the chain-id of the archway protocol "development" public test network.
This network is not expected to mirror mainnet and may have experimental features. There is no guarantee of stability or state
persistance during upgrades.

Note: Please refer to https://github.com/archway-network/networks for a registry of all public networks running archway protocol
and network specific information.

## Release versioning

### Minor version scheme and timeline
- X.Y.Z-rc.W (Branch: release-X.Y)
  - When main is feature-complete for X.Y, we will cut the release-X.Y
    branch and cherrypick only PRs essential to X.Y.
  - If we're not satisfied with X.Y.0-rc.0, we'll release other rc releases,
    (X.Y.0-rc.W | W > 0) as necessary.
- X.Y.0 (Branch: release-X.Y)
  - Final release, cut from the release-X.Y branch.
  - X.Y.0-rc.0 will be tagged at the same commit on the same branch.
- X.Y.Z, Z > 0 (Branch: release-X.Y) ([Patch releases](#patch-releases))
  - [Patch releases](#patch-releases) are released as we cherrypick commits from main into
    the release-X.Y branch, as needed.
  - X.Y.Z is cut straight from the release-X.Y branch, and X.Y.Z+1-beta.0 is
    tagged on the followup commit.
- X.Y.Z, Z > 0 (Branch: release-X.Y.Z) (Branched [patch releases](#patch-releases) only for hotfix situations)
  - These are rarely used and are special in that the X.Y.Z tag is branched to isolate
    the emergency/critical fix from all other changes that have landed on the
    release branch since the previous tag
  - Cut release-X.Y.Z branch to hold the isolated patch release
  - Tag release-X.Y.Z branch + fixes with X.Y.(Z+1)
  - Branched [patch releases](#patch-releases) are rarely needed but used for
    emergency/critical fixes to the latest release

### Major version timeline

There is currently no mandated timeline for major versions beyond version 1.Y.Z, Y,Z >= 0. We haven't so far applied a rigorous interpretation of semantic
versioning with respect to incompatible changes of any kind.

TODO: Major versioning criteria need to be put up for discussion once Archway protocol reaches 1.Y.Z release.

## Patch releases

Patch releases are intended for critical bug fixes to the latest minor version,
such as addressing security vulnerabilities, fixes to problems affecting a large
number of users and severe problems with no workaround.

They should not contain miscellaneous feature additions or improvements, and
especially no incompatibilities should be introduced between patch versions of
the same minor version (or even major version).

Dependencies, such as cosmos-sdk or tendermint, should also not be changed unless
absolutely necessary, and also just to fix critical bugs (so, at most patch
version changes, not new major nor minor versions).

## Network Upgrades

- Upgrade documentation and a graceful migration path must be provided between all **consective** minor releases.
  e.g. going from v1.3.Z, Z >= 0 to v1.4.Z, Z >= 0 must be accompanied by relevant documentation and automated migration path.
- Network upgrade must also follow sequential minor releases. e.g. if a network is running v1.2.Z, Z >= 0, it must first
  upgrade to v1.3.Z, Z >= 0 before upgrading to v1.4.Z, Z >= 0
- Network upgrades may skip patch release e.g. a network running v1.2.1 may directly upgrade to v1.2.4
- Network upgrades on archway-1 and constantine-1 must happen through software upgrade proposals, with the exception of patch releases
  e.g. for security hotfixes.
- Once a software upgrade proposal passes, all network operator are expected to upgrade to the exact
  same archway protocol release within a reasonable timeframe.
- Each new release must be adopted and tested on constantine-1 before being adopted on archway-1.
- titus-1 is an unstable development network, with no guarantees for stability or graceful upgrades. Network upgrades on titus-1 are
  expected to reset network state and start again from block 1 at any given time.

### Open Questions in Network upgrades:

- How to build artifacts for titus-1?
  - Should there be a build on each new commit on main branch?
  - use X.Y.0-{alpha,beta}.W, W > 0 (Branch: master)
    Alpha and beta releases are created from tags on main branch branch and should be used to run titus-1

### Important upgrade scenarios with examples

The following scenarios have hypothetical versions and do not reflect versions of actual live networks

#### archway-1 needs an urget critical security fix

Current Hypothetical State:

- archway-1 is running v1.2.1 and cosntantine-1 is also running v1.2.1
- A critical security issue has been identified that effects v1.2.1
- v1.2.1 is the latest tag on the branch release-1.2
- release-1.2 also has previous tag v1.2.0 and possibly various release candidate tags e.g. v1.2.0-rc1, v1.2.0-rc2 etc
- release-1.2 branch was cut from main branch at a previous point in time to mark the feature completeness of 1.2.Z
  release cycle and main branch has since moved on with various new features

Upgrade Path:

- Cut release-1.2.1 branch from release-1.2 at tag v1.2.1 to hold the isolated patch release
- Add the specific security patch to release-1.2.1 branch
- Tag release-1.2.1 + security patch as v1.2.2
- Upgrade archway-1 to v1.2.2, at this point constantine-1 will still be at v1.2.1. Validators
  may choose to do the upgrade with an upgrade proposal or with an emergency chain halt, depending
  on the severity of the fix
- Security patch will be cherry pickted on release-1.2 branch and also on main
- Next tag on release on release-1.2 branch would be v1.2.3
- constantine-1 will upgrade from v1.2.1 -> v1.2.2
