# State

Section describes all stored by the module objects and their storage keys.

Refer to the [cwregistry.proto](../../../proto/archway/cwregistry/v1/cwregistry.proto) for objects fields description.


## CodeMetadata

[CodeMetadata](../../../proto/archway/cwregistry/v1/cwregistry.proto#L7) object is used to store the metadata regarding a contract binary which has been uploaded.

Storage keys:
* CodeMetadata: `CodeMetadataKeyPrefix | CodeID -> ProtocolBuffer(CodeMetadata)`