# Contributing

The Archway project welcomes all contributors.

This document serves as a guide for contributing to the Archway project. It provides an overview of the processes and policies established for quality collaboration and coordination. Please familiarize yourself with these guidelines and feel free to propose changes to this document for improvement.

## License

__Important:__ _All 3rd party contributions to this repository are made under the Apache License 2.0, unless otherwise agreed to in writing._

## 1. IDE

### 1.1 Software Dependencies

The following software should be installed on the target system:

- The Go Programming Language (https://go.dev)
- Git Distributed Version Control (https://git-scm.com)
- Docker (https://www.docker.com)
- GNU Make (https://www.gnu.org/software/make)

### 1.2 Fork & Clone

1. [Fork the Archway repository](https://github.com/archway-network/archway);
2. [Sync your fork](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/syncing-a-fork) to keep it up to date;
3. Clone your fork to an appropriate directory in your local environment;

```sh
git clone https://github.com/${YOU}/archway.git
```

### 1.3 Install and Configure IDE

Download and install an IDE such as Visual Studio Code on your system.

_Contributors may use whatever IDE they are comfortable with, however this guide will focus on VSCode._

Add the following extensions:

1. Go - Rich Go language support for Visual Studio Code;
2. vscode-proto3 - Protobuf 3 support for Visual Studio Code;
3. GitLens - Git supercharged;

Update the Go extension settings: `File -> Preferences -> Settings : @ext:golang.go`

- Set `Go: Lint Tool` to `golangci-lint`;
- Set `Go: Format Tool` to `gofumpt`;

_The required Go Tools may need to be downloaded by your IDE._

For __CosmWasm__ and _the language that shall not be named_, please refer to the [Archway Developer Portal](https://docs.archway.io/developers).

## 2. Issues

Public collaboration and coordination are driven via issues. For a brief overview of using GitHub Issues, [please read this](https://docs.github.com/en/issues/tracking-your-work-with-issues/quickstart).

1. [Find](https://github.com/archway-network/archway/issues) or [Open](https://github.com/archway-network/archway/issues/new) an issue;
2. __Engage__ in _constructive_ and _thoughful_ __discussion__ on the issue;
3. __Coordinate__ with _other contributors_ on the issue;
4. __Follow__ _GitHub best practices_ and the _guidelines_ in the issue templates;

This repository uses the __good first issue__ label to identify issues suitable for first-time contributors.

Ensure a proper title for the issue that clearly defines what the issue is. This will greatly assist others to easily find related issues and avoid duplication.

Make use of the following labels to provide additional context:

- _bug_: Something isn't working;
- _documentation_: Improvements or additions to documentation;
- _enhancement_: New feature or request;
- _question_: Further information is requested;

Example: `Lack of Community Guidelines` <- label as `documentation`

## 3. Branching

The `main` branch represents the latest development features and fixes. Thus, it will most likely not be compatible with running networks such as `mainnet` or any testnets.

Releases are found under `release/#.#.x` using a semantic versioning scheme {braking}.{feature}.{fix} denoting each minor version.

The `main` branch and `release/*` branches are protected and can only be updated via a Pull Request. Thus, do not make any commits to these branches directly.

For increased manageability please address one issue per branch and use the following branch name format:

`<user>/<issue#>-<short_description>`

## 4. Testing

All _features_ and _fixes_ must be accompanied by appropriate unit/integration tests. Please ensure to follow Go and CosmosSDK best practices regarding tests. Archway utilizes table-driven tests, as explained in [this blogpost](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests) by Dave Cheney.

Simple rules for table-driven tests:

1. Keep tests independent (ordering of tests should not impact testing);
2. Test one construct or aspect per test;
3. Every test should make it obvious what is being tested;

## 5. Pull Requests

* All pull requests must target `main`;
* Updates to __release branches__ are cherry-picked from `main`;

Requirements to be merged:

* The source branch must be __up-to-date__ with `main`;
* All __commits__ must be __signed__;
* All GitHub __action checks must pass__;
* __Documentation and Specs__ must be __updated__ accordingly;
* PR must be __reviewed and approved__;

### PR Title & Description

Title format: `<type>: <issue #> - <description>`

Please follow the Conventional Commits standard as specified [here](https://www.conventionalcommits.org/en/v1.0.0/).

Type:

Please use the following types based on the [Angular convention](https://github.com/angular/angular/blob/22b96b9/CONTRIBUTING.md#-commit-message-guidelines).

- _build_: Changes that affect the build system or external dependencies (example scopes: gulp, broccoli, npm)
- _ci_: Changes to our CI configuration files and scripts (example scopes: Travis, Circle, BrowserStack, SauceLabs)
- _docs_: Documentation only changes
- _feat_: A new feature _(minor version bump)_
- _fix_: A bug fix _(patch version bump)_
- _perf_: A code change that improves performance
- _refactor_: A code change that neither fixes a bug nor adds a feature
- _style_: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)
- _test_: Adding missing tests or correcting existing tests

Breaking Changes:

Indicate breaking changes by appending `!` to the type, and/or including `BREAKING CHANGE: ...` in the footer.

The PR __title__ is more than a mere commit message. It should convey something of substantial significance... A review of the repository pull request titles should provide a very clear progression path, that makes sense to someone visiting the repository for the first time.

Provide a clear and concise __description__ of what has changed and why. Use full sentences with appropriate punctuation. It should be a _summary of changes_, taken from the major commits that make up the PR.

Do add references to any applicable issues at the very end using the notation: `Fixes #123`.

Example:

```
docs: #123 - Updates & Quality Control

Add contribution guidelines to provide an overview of the processes and policies to ensure quality collaboration and coordination within the repository.

Add security policy to detail the processes for reporting security issues and vulnerability disclosure.

Fixes #123, #125, #127
```

### Commits & Commit Messages

Commit Messages follow the same conventions as outlined above for PR Titles.

Make smaller commits more often. This allows the commit log to tell the story of what has changed, and why.

The __first line__ of every commit message must complete the following sentence: _"This commit will..."_ (use imperative mood);

An optional __main content__ (separated from the first line with a blank line) should describe or elaborate on what has changed and why. Use full sentences with appropriate punctuation.

Reference issues at the very end using the notation: `Fixes #123`.

Example:

```
docs: Add contribution guidelines

Contribution guidelines provide an overview of the processes and policies to ensure quality collaboration and coordination within the repository.

Fixes #125
```

See [How to Write a Git Commit Message](https://cbea.ms/git-commit/) for more details.

## 6. Review

Verify PR Requirements:

- PR title is appropriate;
- PR description is sufficient;
- Branch-name is appropriate;
- PR target branch is `main`;
- Source is up-to-date with `main`;
- All commits are signed;
- All GitHub action checks pass;
- Documentation and Specs are updated appropriately;

General review considerations:

- The code structure and architecture should follow Go and CosmosSDK best practices;
- Code conventions and naming should be consistent with the overall codebase;
- Code should be well commented, especially functions, variables and custom types;
- Tests should not contain any Personally Identifiable Information;
- Tests should not contain any production keys or addresses;

`Approval` Checklist:

- I understand the code (what the code does and why);
- I understand the impact (how the architecture, security & overall system is affected);
- I accept responsibility for this code;

Only provide `Approval` if the above conditions are met.

Once `Approval` has been granted the reviewer, or original PR author, may proceed to `squash & merge`.
