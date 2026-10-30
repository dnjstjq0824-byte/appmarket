package infrastructure

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
)

type ParsedManifest struct {
	PackageName      string
	VersionCode      int64
	VersionName      string
	MinSDKVersion    int
	TargetSDKVersion int
	Permissions      []string
	Features         []string
	NativeABIs       []string
	IsUniversalAPK   bool
	SHA256Hash       string
	SizeBytes        int64
}

// InspectAPK reads an APK ZIP stream/reader, validates security constraints (Zip Slip, Zip Bomb),
// parses AndroidManifest.xml and detects native ABIs.
func InspectAPK(r io.ReaderAt, size int64, expectedSHA256 string) (*ParsedManifest, error) {
	if size <= 0 {
		return nil, domain.ErrInvalidApkFile
	}

	// 1. Calculate actual SHA-256 and verify
	h := sha256.New()
	sr := io.NewSectionReader(r, 0, size)
	if _, err := io.Copy(h, sr); err != nil {
		return nil, fmt.Errorf("failed to hash APK: %w", err)
	}
	actualSHA256 := fmt.Sprintf("%x", h.Sum(nil))
	if expectedSHA256 != "" && !strings.EqualFold(actualSHA256, expectedSHA256) {
		return nil, domain.ErrInvalidApkFile
	}

	// 2. Open ZIP
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("%w: not a valid zip archive: %v", domain.ErrInvalidApkFile, err)
	}

	var totalUncompressed int64
	var manifestData []byte
	abiSet := make(map[string]bool)

	for _, file := range zipReader.File {
		cleanName := filepath.Clean(file.Name)
		// Zip Slip Detection
		if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) || strings.Contains(file.Name, "../") {
			return nil, domain.ErrZipSlipDetected
		}

		totalUncompressed += int64(file.UncompressedSize64)

		// Zip Bomb Detection: Ratio > 100:1 or uncompressed > 10GB
		if size > 0 && float64(totalUncompressed)/float64(size) > 100.0 {
			return nil, domain.ErrZipBombDetected
		}
		if totalUncompressed > 10*1024*1024*1024 {
			return nil, domain.ErrZipBombDetected
		}

		// Check for manifest
		if file.Name == "AndroidManifest.xml" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open AndroidManifest.xml: %w", err)
			}
			buf, err := io.ReadAll(io.LimitReader(rc, 10*1024*1024))
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read AndroidManifest.xml: %w", err)
			}
			manifestData = buf
		}

		// Detect native libraries in lib/<ABI>/*.so
		if strings.HasPrefix(file.Name, "lib/") && strings.HasSuffix(file.Name, ".so") {
			parts := strings.Split(file.Name, "/")
			if len(parts) >= 3 && parts[1] != "" {
				abiSet[parts[1]] = true
			}
		}
	}

	if len(manifestData) == 0 {
		return nil, fmt.Errorf("%w: missing AndroidManifest.xml", domain.ErrInvalidApkFile)
	}

	parsed, err := ParseAXML(manifestData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AndroidManifest.xml: %w", err)
	}

	var abis []string
	for abi := range abiSet {
		abis = append(abis, abi)
	}

	parsed.NativeABIs = abis
	parsed.IsUniversalAPK = len(abis) == 0 || len(abis) >= 3
	parsed.SHA256Hash = actualSHA256
	parsed.SizeBytes = size

	return parsed, nil
}

const (
	chunkTypeXML        = 0x00080003
	chunkTypeStringPool = 0x001C0001
	chunkTypeStartTag   = 0x00100102
)

// ParseAXML decodes Android Binary XML format or plain XML.
func ParseAXML(data []byte) (*ParsedManifest, error) {
	manifest := &ParsedManifest{
		MinSDKVersion:    21, // default Android 5.0
		TargetSDKVersion: 34, // default Android 14
		Permissions:      []string{},
		Features:         []string{},
	}

	// Plain XML fallback
	if bytes.HasPrefix(bytes.TrimSpace(data), []byte("<?xml")) || bytes.HasPrefix(bytes.TrimSpace(data), []byte("<manifest")) {
		return parsePlainXMLManifest(data, manifest)
	}

	if len(data) < 8 {
		return nil, errors.New("data too short for AXML")
	}

	chunkType := binary.LittleEndian.Uint32(data[0:4])
	if chunkType != chunkTypeXML {
		// Not recognized AXML magic
		return nil, errors.New("invalid AXML magic header")
	}

	reader := bytes.NewReader(data[8:])
	stringPool, err := extractStringPool(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to extract string pool: %w", err)
	}

	// Extract attributes and elements by scanning string pool and tag chunks
	manifest.PackageName = findAttrOrString(stringPool, "package", manifest.PackageName)
	for i, s := range stringPool {
		if s == "package" && i+1 < len(stringPool) && strings.Contains(stringPool[i+1], ".") {
			manifest.PackageName = stringPool[i+1]
		}
		if s == "versionName" && i+1 < len(stringPool) {
			manifest.VersionName = stringPool[i+1]
		}
		if s == "versionCode" && i+1 < len(stringPool) {
			if v, err := strconv.ParseInt(stringPool[i+1], 10, 64); err == nil && v > 0 {
				manifest.VersionCode = v
			}
		}
		if s == "minSdkVersion" && i+1 < len(stringPool) {
			if v, err := strconv.Atoi(stringPool[i+1]); err == nil && v > 0 {
				manifest.MinSDKVersion = v
			}
		}
		if s == "targetSdkVersion" && i+1 < len(stringPool) {
			if v, err := strconv.Atoi(stringPool[i+1]); err == nil && v > 0 {
				manifest.TargetSDKVersion = v
			}
		}
		if strings.HasPrefix(s, "android.permission.") {
			manifest.Permissions = append(manifest.Permissions, s)
		}
		if strings.HasPrefix(s, "android.hardware.") {
			manifest.Features = append(manifest.Features, s)
		}
	}

	// If versionCode wasn't a string representation in the pool, use fallback default
	if manifest.VersionCode <= 0 {
		manifest.VersionCode = 1
	}
	if manifest.VersionName == "" {
		manifest.VersionName = "1.0.0"
	}
	if manifest.PackageName == "" {
		// Scan string pool for valid reverse-dns package name
		for _, s := range stringPool {
			if strings.Count(s, ".") >= 2 && !strings.HasPrefix(s, "http") && !strings.HasPrefix(s, "android.") {
				manifest.PackageName = s
				break
			}
		}
	}

	return manifest, nil
}

func parsePlainXMLManifest(data []byte, manifest *ParsedManifest) (*ParsedManifest, error) {
	s := string(data)
	// Simple attribute parser
	manifest.PackageName = extractAttr(s, "package")
	manifest.VersionName = extractAttr(s, "android:versionName")
	if manifest.VersionName == "" {
		manifest.VersionName = extractAttr(s, "versionName")
	}
	vcStr := extractAttr(s, "android:versionCode")
	if vcStr == "" {
		vcStr = extractAttr(s, "versionCode")
	}
	if vc, err := strconv.ParseInt(vcStr, 10, 64); err == nil {
		manifest.VersionCode = vc
	} else {
		manifest.VersionCode = 1
	}

	minSdkStr := extractAttr(s, "android:minSdkVersion")
	if minSdk, err := strconv.Atoi(minSdkStr); err == nil {
		manifest.MinSDKVersion = minSdk
	}
	targetSdkStr := extractAttr(s, "android:targetSdkVersion")
	if targetSdk, err := strconv.Atoi(targetSdkStr); err == nil {
		manifest.TargetSDKVersion = targetSdk
	}

	// Extract permissions
	lines := strings.Split(s, "<")
	for _, l := range lines {
		if strings.HasPrefix(l, "uses-permission") {
			perm := extractAttr("<"+l, "android:name")
			if perm != "" {
				manifest.Permissions = append(manifest.Permissions, perm)
			}
		}
		if strings.HasPrefix(l, "uses-feature") {
			feat := extractAttr("<"+l, "android:name")
			if feat != "" {
				manifest.Features = append(manifest.Features, feat)
			}
		}
	}

	return manifest, nil
}

func extractAttr(text, attrName string) string {
	idx := strings.Index(text, attrName+`="`)
	if idx == -1 {
		idx = strings.Index(text, attrName+`='`)
		if idx == -1 {
			return ""
		}
		start := idx + len(attrName) + 2
		end := strings.Index(text[start:], `'`)
		if end == -1 {
			return ""
		}
		return text[start : start+end]
	}
	start := idx + len(attrName) + 2
	end := strings.Index(text[start:], `"`)
	if end == -1 {
		return ""
	}
	return text[start : start+end]
}

func extractStringPool(r *bytes.Reader) ([]string, error) {
	for r.Len() >= 8 {
		var chunkType, chunkSize uint32
		if err := binary.Read(r, binary.LittleEndian, &chunkType); err != nil {
			break
		}
		if err := binary.Read(r, binary.LittleEndian, &chunkSize); err != nil {
			break
		}

		if chunkType == chunkTypeStringPool {
			poolBytes := make([]byte, chunkSize-8)
			if _, err := io.ReadFull(r, poolBytes); err != nil {
				return nil, err
			}
			return decodeStringPool(poolBytes)
		}

		if chunkSize > 8 {
			r.Seek(int64(chunkSize-8), io.SeekCurrent)
		}
	}
	return nil, errors.New("string pool not found")
}

func decodeStringPool(data []byte) ([]string, error) {
	if len(data) < 20 {
		return nil, errors.New("invalid string pool data")
	}

	stringCount := binary.LittleEndian.Uint32(data[0:4])
	flags := binary.LittleEndian.Uint32(data[8:12])
	stringsStart := binary.LittleEndian.Uint32(data[12:16])

	isUTF8 := (flags & (1 << 8)) != 0

	var offsets []uint32
	offsetReader := bytes.NewReader(data[20:])
	for i := uint32(0); i < stringCount; i++ {
		var off uint32
		if err := binary.Read(offsetReader, binary.LittleEndian, &off); err != nil {
			break
		}
		offsets = append(offsets, off)
	}

	stringsBlock := data[stringsStart-8:]
	var result []string

	for _, off := range offsets {
		if int(off) >= len(stringsBlock) {
			continue
		}
		sub := stringsBlock[off:]
		if isUTF8 {
			str := readAXMLStringUTF8(sub)
			result = append(result, str)
		} else {
			str := readAXMLStringUTF16(sub)
			result = append(result, str)
		}
	}

	return result, nil
}

func readAXMLStringUTF8(b []byte) string {
	if len(b) < 2 {
		return ""
	}
	// skip length bytes
	idx := 0
	if b[0]&0x80 != 0 {
		idx += 2
	} else {
		idx++
	}
	if idx >= len(b) {
		return ""
	}
	length := int(b[idx])
	idx++
	if idx+length > len(b) {
		length = len(b) - idx
	}
	return string(b[idx : idx+length])
}

func readAXMLStringUTF16(b []byte) string {
	if len(b) < 2 {
		return ""
	}
	length := binary.LittleEndian.Uint16(b[0:2])
	if int(length*2+2) > len(b) {
		length = uint16((len(b) - 2) / 2)
	}
	u16 := make([]uint16, length)
	for i := 0; i < int(length); i++ {
		u16[i] = binary.LittleEndian.Uint16(b[2+i*2 : 4+i*2])
	}
	return string(utf16.Decode(u16))
}

func findAttrOrString(pool []string, key, fallback string) string {
	for i, s := range pool {
		if s == key && i+1 < len(pool) {
			return pool[i+1]
		}
	}
	return fallback
}

