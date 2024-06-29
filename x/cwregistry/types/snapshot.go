package types

func NewSnapshot(codeId uint64, compressedSchema []byte) *SnapshotPayload {
	return &SnapshotPayload{
		CodeId: codeId,
		Schema: compressedSchema,
	}
}
