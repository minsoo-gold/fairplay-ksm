package main

import (
	"context"
	"testing"
)

// 단위 테스트
// https://console.cloud.google.com/firestore/databases/-default-/data/panel?project=fairplaystreaming 콘솔에서 assetID 확인하고
// realAssetID : 입력해서 실행하면 firestore에서 가져오는지 확인
func TestFirestoreContentKey_RealData(t *testing.T) {
	// assetID
	realAssetID := "GRYCzx5wxnkYXliJ|;12132091-d0ef-4f6f-8cca-34989edd0a2b"

	ctx := context.Background()
	contentKey := NewFirestoreContentKey(ctx)

	kidBytes, keyBytes, ivBytes, err := contentKey.FetchContentKey([]byte(realAssetID))

	if err != nil {
		t.Fatalf("FetchContentKey failed: %v", err)
	}

	if len(kidBytes) != 16 {
		t.Errorf("Expected kid length 16, got %d", len(kidBytes))
	}

	if len(keyBytes) != 16 {
		t.Errorf("Expected key length 16, got %d", len(keyBytes))
	}

	if len(ivBytes) != 16 {
		t.Errorf("Expected IV length 16, got %d", len(ivBytes))
	}

	t.Logf("✅ Successfully fetched key and IV for asset: %s", realAssetID)
	t.Logf("Kid: %x", kidBytes)
	t.Logf("Key: %x", keyBytes)
	t.Logf("IV: %x", ivBytes)

	// FetchContentKeyDuration 테스트
	duration, err := contentKey.FetchContentKeyDuration([]byte(realAssetID))

	if err != nil {
		t.Fatalf("FetchContentKeyDuration failed: %v", err)
	}

	if duration == nil {
		t.Fatal("Expected non-nil duration")
	}

	t.Logf("✅ Successfully fetched duration for asset: %s", realAssetID)
	t.Logf("Lease Duration: %d", duration.LeaseDuration)
	t.Logf("Rental Duration: %d", duration.RentalDuration)
}
