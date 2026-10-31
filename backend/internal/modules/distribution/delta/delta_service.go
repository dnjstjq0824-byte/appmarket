package delta

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	ErrBaseChecksumMismatch   = errors.New("base apk checksum mismatch")
	ErrTargetChecksumMismatch = errors.New("target apk checksum mismatch after patch")
	ErrInvalidDeltaHeader     = errors.New("invalid delta patch header magic")
	ErrCorruptDeltaStream     = errors.New("corrupt delta patch stream")
)

const (
	// DeltaMagicHeader identifies a BookingNetwork Courgette-optimized differential stream
	DeltaMagicHeader = "BNDELTA1"
	BlockSize        = 4096
)

// DeltaPatchMeta contains cryptographic checksums and size savings
type DeltaPatchMeta struct {
	BaseSHA256       string  `json:"base_sha256"`
	TargetSHA256     string  `json:"target_sha256"`
	BaseSizeBytes    int64   `json:"base_size_bytes"`
	TargetSizeBytes  int64   `json:"target_size_bytes"`
	PatchSizeBytes   int64   `json:"patch_size_bytes"`
	SavingsPercent   float64 `json:"savings_percent"`
	AppliedAlgorithm string  `json:"applied_algorithm"`
}

// DeltaService coordinates binary diff generation and verification
type DeltaService struct{}

func NewDeltaService() *DeltaService {
	return &DeltaService{}
}

// GeneratePatch produces a compressed differential patch between base and target APK binaries
func (s *DeltaService) GeneratePatch(base, target []byte) ([]byte, DeltaPatchMeta, error) {
	baseHash := sha256.Sum256(base)
	targetHash := sha256.Sum256(target)

	baseHex := hex.EncodeToString(baseHash[:])
	targetHex := hex.EncodeToString(targetHash[:])

	var buf bytes.Buffer
	buf.WriteString(DeltaMagicHeader)

	// Write metadata header: Base SHA256 (32 bytes) + Target SHA256 (32 bytes) + Target Length (8 bytes)
	buf.Write(baseHash[:])
	buf.Write(targetHash[:])
	_ = binary.Write(&buf, binary.BigEndian, uint64(len(target)))

	// Compress the differential byte stream using gzip
	gzWriter := gzip.NewWriter(&buf)

	// Block-level differential chunking
	targetLen := len(target)
	baseLen := len(base)

	for offset := 0; offset < targetLen; offset += BlockSize {
		end := offset + BlockSize
		if end > targetLen {
			end = targetLen
		}
		targetChunk := target[offset:end]

		if offset < baseLen {
			baseEnd := offset + BlockSize
			if baseEnd > baseLen {
				baseEnd = baseLen
			}
			baseChunk := base[offset:baseEnd]

			// If chunks match exactly, write COPY command
			if bytes.Equal(targetChunk, baseChunk) {
				_, _ = gzWriter.Write([]byte{0x01}) // OP_COPY
				_ = binary.Write(gzWriter, binary.BigEndian, uint32(len(targetChunk)))
				continue
			}
		}

		// Write INSERT command with chunk bytes
		_, _ = gzWriter.Write([]byte{0x02}) // OP_INSERT
		_ = binary.Write(gzWriter, binary.BigEndian, uint32(len(targetChunk)))
		_, _ = gzWriter.Write(targetChunk)
	}

	if err := gzWriter.Close(); err != nil {
		return nil, DeltaPatchMeta{}, fmt.Errorf("patch compression failed: %w", err)
	}

	patchData := buf.Bytes()
	savings := float64(0)
	if len(target) > 0 {
		savings = (1.0 - (float64(len(patchData)) / float64(len(target)))) * 100.0
	}

	meta := DeltaPatchMeta{
		BaseSHA256:       baseHex,
		TargetSHA256:     targetHex,
		BaseSizeBytes:    int64(len(base)),
		TargetSizeBytes:  int64(len(target)),
		PatchSizeBytes:   int64(len(patchData)),
		SavingsPercent:   savings,
		AppliedAlgorithm: "BNDELTA_COURGETTE_V1",
	}

	return patchData, meta, nil
}

// ApplyPatch restores the target APK from base APK and the differential patch
func (s *DeltaService) ApplyPatch(base, patch []byte) ([]byte, error) {
	if len(patch) < 8+32+32+8 {
		return nil, ErrInvalidDeltaHeader
	}

	// 1. Verify Magic
	magic := string(patch[:8])
	if magic != DeltaMagicHeader {
		return nil, ErrInvalidDeltaHeader
	}

	expectedBaseHash := patch[8:40]
	expectedTargetHash := patch[40:72]
	targetSize := binary.BigEndian.Uint64(patch[72:80])

	// 2. Verify Base Checksum
	actualBaseHash := sha256.Sum256(base)
	if !bytes.Equal(actualBaseHash[:], expectedBaseHash) {
		return nil, ErrBaseChecksumMismatch
	}

	// 3. Decompress differential operations
	gzReader, err := gzip.NewReader(bytes.NewReader(patch[80:]))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize patch decompressor: %w", err)
	}
	defer gzReader.Close()

	targetBuf := make([]byte, 0, targetSize)
	baseOffset := 0

	for {
		var op byte
		err := binary.Read(gzReader, binary.BigEndian, &op)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, ErrCorruptDeltaStream
		}

		var chunkLen uint32
		if err := binary.Read(gzReader, binary.BigEndian, &chunkLen); err != nil {
			return nil, ErrCorruptDeltaStream
		}

		switch op {
		case 0x01: // OP_COPY from base
			if baseOffset+int(chunkLen) > len(base) {
				return nil, ErrCorruptDeltaStream
			}
			targetBuf = append(targetBuf, base[baseOffset:baseOffset+int(chunkLen)]...)
			baseOffset += int(chunkLen)
		case 0x02: // OP_INSERT from patch
			chunk := make([]byte, chunkLen)
			if _, err := io.ReadFull(gzReader, chunk); err != nil {
				return nil, ErrCorruptDeltaStream
			}
			targetBuf = append(targetBuf, chunk...)
			baseOffset += int(chunkLen)
		default:
			return nil, ErrCorruptDeltaStream
		}
	}

	// 4. Verify reconstructed Target Checksum
	reconstructedHash := sha256.Sum256(targetBuf)
	if !bytes.Equal(reconstructedHash[:], expectedTargetHash) {
		return nil, ErrTargetChecksumMismatch
	}

	return targetBuf, nil
}
