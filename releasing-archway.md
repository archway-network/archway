# Archway Protocol Release Process

## Release Versioning

This document outlines the release process for Archway protocol software.

The **Archway** protocol follows a **state versioning system**, it does not utilize traditional [semantic versioning](http://semver.org); SemVer is specifically about APIs, but API breaking changes is not the only way in which libraries can "break". In deterministic systems, such as a blockchain being deterministic state machine, *state breaking changes far outweigh API breaking changes*. As such, Archway follows a State Versioning specification.

Given a version number Major.Minor.Patch, increment the:

1. **Major** version when any state breaking changes are introduced;
2. **Minor** version when any API changes, both API-compatible or API-incompatible, are introduced;
3. **Patch** version when any state-compatible bug fixes are introduced;

Additional labels for release-candidates, pre-release versions and other build metadata are available as extensions to the Major.Minor.Patch format.

**Note:** Any dependency updates that are not state breaking, e.g. updating the Go version, fall under minor.

## State Compatability

**Note:** State breaking changes include changes that impact the amount of gas needed to execute a transaction as well as any changes to error handling. This is because `AppHash` and `LastResultsHash` contains:

1. Tx `GasWanted`;
2. Tx `GasUsed` - which is Merkelized, thus any logic affecting this will result in state changes;
3. Tx response `Data` - protobuf encoding changes result in state changes;
4. Tx response `Code` - any changes to error handling flow, or custom error codes will result in state changes;

## Release Process

The standard release process progresses through the following steps:

1. Tag a new release once a major release, update or patch is deemed ready for deployment (release atrifacts are created via automation);
2. Deploy the new release version to Constantine (testnet);
3. Conduct final formal verification for the release on testnet, including for upgrade handlers;
4. Repeat steps 1-3 until all formal verification passes, e.g. relevant tests, remediations, etc;
5. Deploy the release to mainnet via upgrade proposal;

**Note:** Steps 1-3 may include a number of iterations with various release candidates;

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
