package delta

import (
	"bytes"
	"testing"
)

func TestDeltaService_GenerateAndApplyPatch(t *testing.T) {
	service := NewDeltaService()

	// Base simulated APK (100KB with repeating pattern)
	baseAPK := bytes.Repeat([]byte("BASE_APK_BINARY_DATA_SECTION_12345678_"), 2500)

	// Target simulated APK (100KB with 10% modifications)
	targetAPK := make([]byte, len(baseAPK))
	copy(targetAPK, baseAPK)
	copy(targetAPK[5000:15000], bytes.Repeat([]byte("MODIFIED_BINARY_PATCH_DATA_V2_"), 333))

	// 1. Generate Differential Patch
	patch, meta, err := service.GeneratePatch(baseAPK, targetAPK)
	if err != nil {
		t.Fatalf("GeneratePatch failed: %v", err)
	}

	if len(patch) == 0 {
		t.Fatal("generated patch is empty")
	}

	t.Logf("Base size: %d, Target size: %d, Patch size: %d (Savings: %.2f%%)",
		meta.BaseSizeBytes, meta.TargetSizeBytes, meta.PatchSizeBytes, meta.SavingsPercent)

	if meta.SavingsPercent <= 0 {
		t.Errorf("expected positive savings percent, got: %.2f", meta.SavingsPercent)
	}

	// 2. Apply Differential Patch
	reconstructed, err := service.ApplyPatch(baseAPK, patch)
	if err != nil {
		t.Fatalf("ApplyPatch failed: %v", err)
	}

	if !bytes.Equal(reconstructed, targetAPK) {
		t.Fatal("reconstructed target APK does not match original target APK")
	}
}

func TestDeltaService_CorruptBaseRejection(t *testing.T) {
	service := NewDeltaService()

	baseAPK := []byte("VALID_BASE_BINARY_DATA_CONTENT_0001")
	targetAPK := []byte("VALID_TARGET_BINARY_DATA_CONTENT_0002")

	patch, _, err := service.GeneratePatch(baseAPK, targetAPK)
	if err != nil {
		t.Fatalf("GeneratePatch failed: %v", err)
	}

	// Apply with mismatched corrupted base
	corruptBase := []byte("CORRUPTED_BASE_BINARY_DATA_CONTENT")
	_, err = service.ApplyPatch(corruptBase, patch)
	if err == nil {
		t.Fatal("expected ErrBaseChecksumMismatch, got nil")
	}
}

