package ksm

// ContentKey is a interface that fetch asset content key and duration.
type ContentKey interface {
	FetchContentKey(assetID []byte) ([]byte, []byte, []byte, error)
	FetchContentKeyDuration(assetID []byte) (*CkcContentKeyDurationBlock, error)
}

// CKCPayload is a object that store ckc payload.
type CKCPayload struct {
	SK             []byte //Session key
	HU             []byte
	R1             []byte
	IntegrityBytes []byte
}
