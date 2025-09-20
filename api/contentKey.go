package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/minsoo-gold/fairplay-ksm/ksm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FirestoreContentKey struct {
	client   *firestore.Client
	tenantID string
}

// Firestore 초기화 (main에서 호출)
// tenantID를 고정해서 이 인스턴스는 특정 테넌트만 다루도록 합니다.
func NewFirestoreContentKey(ctx context.Context, projectID, tenantID string) (*FirestoreContentKey, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &FirestoreContentKey{client: client, tenantID: tenantID}, nil
}

// (선택) 종료 시 호출
func (f *FirestoreContentKey) Close() error {
	return f.client.Close()
}

// FetchContentKey: Firestore에서 contentKey, IV 가져오기
// 경로: tenants/{tenantID}/asset_keys/{assetID}
func (f *FirestoreContentKey) FetchContentKey(assetID []byte) ([]byte, []byte, error) {
	ctx := context.Background()

	docRef := f.client.
		Collection("tenants").Doc(f.tenantID).
		Collection("asset_keys").Doc(string(assetID))

	doc, err := docRef.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	data := doc.Data()

	keyBytes, err := toBytes(data["contentKey"])
	if err != nil {
		return nil, nil, fmt.Errorf("contentKey decode error: %w", err)
	}
	if len(keyBytes) != 16 {
		return nil, nil, fmt.Errorf("contentKey length must be 16 bytes, got %d", len(keyBytes))
	}

	ivBytes, err := toBytes(data["iv"])
	if err != nil {
		return nil, nil, fmt.Errorf("iv decode error: %w", err)
	}
	if len(ivBytes) != 16 {
		return nil, nil, fmt.Errorf("iv length must be 16 bytes, got %d", len(ivBytes))
	}

	return keyBytes, ivBytes, nil
}

// FetchContentKeyDuration: Firestore에서 Lease/RentalDuration 가져오기
// 경로: tenants/{tenantID}/license_policy/{assetID}
func (f *FirestoreContentKey) FetchContentKeyDuration(assetID []byte) (*ksm.CkcContentKeyDurationBlock, error) {
	ctx := context.Background()

	docRef := f.client.
		Collection("tenants").Doc(f.tenantID).
		Collection("license_policy").Doc(string(assetID))
	doc, err := docRef.Get(ctx)
	if err != nil {
		// 정책 문서가 없다면 0으로 처리 (선택 로직)
		// firestore.IsNotFound is not available, use codes.NotFound
		// import "google.golang.org/grpc/status" and "google.golang.org/grpc/codes"
		if grpcErr, ok := err.(interface{ GRPCStatus() *status.Status }); ok {
			if grpcErr.GRPCStatus().Code() == codes.NotFound {
				return ksm.NewCkcContentKeyDurationBlock(0, 0), nil
			}
		}
		return nil, err
	}
	data := doc.Data()

	lease, err := toUint32(data["leaseDuration"])
	if err != nil {
		lease = 0
	}
	rental, err := toUint32(data["rentalDuration"])
	if err != nil {
		rental = 0
	}

	return ksm.NewCkcContentKeyDurationBlock(lease, rental), nil
}

// ---------- 유틸리티 ----------

// Firestore에 저장된 값을 []byte로 변환
// - []byte (native)
// - string (hex 또는 base64) 를 지원
func toBytes(v interface{}) ([]byte, error) {
	switch t := v.(type) {
	case []byte: // Firestore Binary
		return t, nil
	case string:
		// 우선 hex 시도
		if b, err := hex.DecodeString(t); err == nil {
			return b, nil
		}
		// base64 시도
		if b, err := base64.StdEncoding.DecodeString(t); err == nil {
			return b, nil
		}
		return nil, errors.New("unsupported string encoding (expect hex or base64)")
	default:
		return nil, errors.New("unsupported type (expect []byte or string)")
	}
}

// Firestore 숫자 필드 → uint32
// Firestore는 숫자를 int64(float64) 로 돌려줄 수 있으므로 호환 처리
func toUint32(v interface{}) (uint32, error) {
	switch n := v.(type) {
	case int64:
		if n < 0 {
			return 0, errors.New("negative value")
		}
		return uint32(n), nil
	case int: // 드물지만 방어
		if n < 0 {
			return 0, errors.New("negative value")
		}
		return uint32(n), nil
	case float64:
		if n < 0 {
			return 0, errors.New("negative value")
		}
		return uint32(n), nil
	case uint32:
		return n, nil
	case uint64:
		return uint32(n), nil
	default:
		return 0, errors.New("unsupported numeric type")
	}
}

/*
// NOTE: In production, please implement your content key logic instead of Random Content Key
package main

import (
	"crypto/md5"
	"math/rand"

	"github.com/minsoo-gold/fairplay-ksm/ksm"
)

type RandomContentKey struct {
}

// Implement FetchContentKey func
func (RandomContentKey) FetchContentKey(assetID []byte) ([]byte, []byte, error) {
	return generateDummyKeyIVPair(assetID)
}

// Implement FetchContentKeyDuration func
func (RandomContentKey) FetchContentKeyDuration(assetID []byte) (*ksm.CkcContentKeyDurationBlock, error) {

	LeaseDuration := rand.Uint32()  // The duration of the lease, if any, in seconds.
	RentalDuration := rand.Uint32() // The duration of the rental, if any, in seconds.

	return ksm.NewCkcContentKeyDurationBlock(LeaseDuration, RentalDuration), nil
}

func generateDummyKeyIVPair(assetID []byte) ([]byte, []byte, error) {
	dummyKey := make([]byte, 16)
	dummyIV := make([]byte, 16)
	rand.Read(dummyIV)

	if len(assetID) == 0 {
		rand.Read(dummyKey)
		return dummyKey, dummyIV, nil
	}
	// NOTE: Here is to implement your key generator.
	generator := md5.New()
	generator.Write(assetID)
	dummyKey = generator.Sum(nil)
	return dummyKey, dummyIV, nil
}
*/
