# State

Section describes all stored by the module objects and their stoarage keys

## Params

[Params](../../../proto/archway/cwica/v1/params.proto) object is used to store the module params

The params value can only be updated by x/gov module via a governance upgrade proposal. [More](./02_messages.md#msgupdateparams)

Storage keys:
* Params: `ParamsKey -> ProtocolBuffer(Params)`