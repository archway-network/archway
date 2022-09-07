# v0.0.6

## Bug Fixes
- Update Gas Consumed to Decimal values [#159](https://github.com/archway-network/archway/pull/159)

## Breaking Changes
- Refactor codebase
  - Add reward module [#202](https://github.com/archway-network/archway/pull/202)
  - Add tracking module[#202](https://github.com/archway-network/archway/pull/202)
- Add minimum consensus fee middleware [#202](https://github.com/archway-network/archway/pull/202)
- Add lazy distribution feature [#207](https://github.com/archway-network/archway/pull/207)
- Add Wasm Bindings [#196](https://github.com/archway-network/archway/pull/196) [#210](https://github.com/archway-network/archway/pull/210)

# Changes
- Use archway/wasmd fork [#162](https://github.com/archway-network/archway/pull/159)
- Update cosmos sdk to 0.45.7 [#212](https://github.com/archway-network/archway/pull/212)
- Upgrade to IBC v3 [#204](https://github.com/archway-network/archway/pull/204)


# v0.0.5

## Breaking Changes
- Update wasmd to 0.25.

## Bug Fixes
- Fix logs printing total contract rewards instead of gas rebate reward.
- Replace info logs for debug logs.

# v0.0.4
## Breaking Changes
- Replace wasmd KV store to KV Multistore.
- Split WasmVM gas & SDK Gas.

# v0.0.3
- inflation reward calculation now depend upon block gas limit.
- inflation reward is given even when the gas rebate to the user flag is true.
- gastracker's begin blocker takes into account gas rebate to user governance switch.
- fix gas estimation for `--gas auto` flag.
