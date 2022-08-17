
<!--
order: 2
-->

# AnteHandlers

Section describes the module ante handlers.


## TxGasTrackingDecorator

The [TxGasTrackingDecorator](https://github.com/archway-network/archway/blob/e130d74bd456be037b4e60dea7dada5d7a8760b5/x/tracking/ante/tracking.go#L15) handler kickstarts transaction tracking by creating an empty [TxInfo](01_state.md#TxInfo)

