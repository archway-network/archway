# ADR 006: Introduction of the genmsg module for genesis state configuration

Date: 2023-05-19

## Status

Implemented

## Context

With the recent removal of genesis messages from the wasmd module by the CosmWasm team, the feature to store, deploy,
and execute contracts during genesis has been lost. Genesis messages provided us with a mechanism to safely create some
initial and predictable state without meddling too much with wasmd genesis import/export internals.

The CosmWasm team proposed that this feature could be replaced by adding raw golang code to deploy contracts during
genesis. However, this approach is not considered ideal as initial state is considered configuration, not code. 
If we treat it as code, we essentially insert network-specific details into our code, which is not a good practice.

## Decision

We decided to create the genmsg module to fill in the gap created by this feature removal. This module is designed to 
execute after every other module's genesis (excluding crisis) and execute generic state transition messages.
These messages accepted are the ones defined by the installed modules in their msg servers.

Consequences

- Positive: The genmsg module provides a clean, configuration-based approach to initializing network state during genesis,
avoiding the need to write network-specific code.

- Negative: The decision to implement genmsg will require development time and resources, and there will be additional 
overhead for maintenance and integration with other modules. Although the effort should not be big.

- Neutral: This change doesn't affect the operation of other modules' genesis.
