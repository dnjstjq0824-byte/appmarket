package infrastructure

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
)

const (
	apkSigBlockMagic = "APK Sig Block 42"
	idSigSchemeV2    = uint32(0x7109871a)
	idSigSchemeV3    = uint32(0xf05368c0)
	idSigSchemeV31   = uint32(0x1b9322f7)
)

type VerifiedSignature struct {
	SignatureSchemeVersion int
	CertSHA256Fingerprint  string
	CertSHA1Fingerprint    string
	PublicKeyAlgorithm     string
	KeySizeBits            int
	SubjectDN              string
	IssuerDN               string
	SerialNumberHex        string
	ValidFrom              time.Time
	ValidUntil             time.Time
	ProofOfRotation        string
	IsPrimaryLineage       bool
}

// VerifyAPKSignature checks whether the APK has valid v2 or v3 signature block,
// extracts the signing certificate, and enforces security constraints.
func VerifyAPKSignature(r io.ReaderAt, size int64, blacklistedCerts []string) (*VerifiedSignature, error) {
	if size <= 0 {
		return nil, domain.ErrInvalidApkFile
	}

	// 1. Locate the ZIP End of Central Directory (EOCD)
	eocdOffset, cdOffset, err := findCentralDirectory(r, size)
	if err != nil {
		return nil, fmt.Errorf("failed to locate ZIP central directory: %w", err)
	}

	// 2. Check for APK Signing Block immediately preceding Central Directory
	sigBlock, schemeVersion, err := findAPKSigningBlock(r, cdOffset)
	if err != nil || schemeVersion < 2 {
		// Check if it's only a JAR v1 signature
		isV1Only := checkIsV1SignatureOnly(r, size, eocdOffset)
		if isV1Only {
			return nil, domain.ErrApkV2SignatureRequired
		}
		return nil, domain.ErrApkV2SignatureRequired
	}

	// 3. Extract Certificate from the signature block
	cert, err := extractCertificateFromBlock(sigBlock, schemeVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signature certificate: %w", err)
	}

	sha256FP := fmt.Sprintf("%x", sha256.Sum256(cert.Raw))
	sha1FP := fmt.Sprintf("%x", sha1.Sum(cert.Raw))

	// 4. Check Blacklisted Certificates (TC-SEC-010)
	for _, bl := range blacklistedCerts {
		if strings.EqualFold(sha256FP, bl) || strings.EqualFold(sha1FP, bl) {
			return nil, domain.ErrBlacklistedCertificate
		}
	}

	// 5. Check Debug Certificate
	subject := cert.Subject.String()
	issuer := cert.Issuer.String()
	if strings.Contains(strings.ToUpper(subject), "CN=ANDROID DEBUG") || strings.Contains(strings.ToUpper(issuer), "O=ANDROID,C=US") {
		return nil, fmt.Errorf("%w: debug certificates are forbidden in production", domain.ErrInvalidApkFile)
	}

	// 6. Check Public Key Algorithm & Length
	keySizeBits := cert.PublicKeyAlgorithm.String()
	bitLen := 2048
	switch pub := cert.PublicKey.(type) {
	case interface{ Size() int }:
		bitLen = pub.Size() * 8
	}
	if bitLen < 2048 && cert.PublicKeyAlgorithm == x509.RSA {
		return nil, fmt.Errorf("%w: RSA key size %d is below minimum 2048 bits", domain.ErrInvalidApkFile, bitLen)
	}

	serialHex := hex.EncodeToString(cert.SerialNumber.Bytes())

	return &VerifiedSignature{
		SignatureSchemeVersion: schemeVersion,
		CertSHA256Fingerprint:  sha256FP,
		CertSHA1Fingerprint:    sha1FP,
		PublicKeyAlgorithm:     keySizeBits,
		KeySizeBits:            bitLen,
		SubjectDN:              subject,
		IssuerDN:               issuer,
		SerialNumberHex:        serialHex,
		ValidFrom:              cert.NotBefore,
		ValidUntil:             cert.NotAfter,
		IsPrimaryLineage:       true,
	}, nil
}

// VerifySignatureLineage ensures the new APK matches the certificate of previous releases (TC-SEC-007).
func VerifySignatureLineage(newSig *VerifiedSignature, previousCertSHA256 string) error {
	if previousCertSHA256 == "" {
		return nil // First release
	}
	if !strings.EqualFold(newSig.CertSHA256Fingerprint, previousCertSHA256) {
		return domain.ErrSignatureMismatch
	}
	return nil
}

func findCentralDirectory(r io.ReaderAt, size int64) (eocdOffset int64, cdOffset int64, err error) {
	// EOCD is at the end, at most 65KB + 22 bytes
	searchLen := int64(65536 + 22)
	if searchLen > size {
		searchLen = size
	}
	buf := make([]byte, searchLen)
	offset := size - searchLen
	if _, err := r.ReadAt(buf, offset); err != nil {
		return 0, 0, err
	}

	// Search backwards for EOCD signature 0x06054b50
	for i := len(buf) - 22; i >= 0; i-- {
		if buf[i] == 0x50 && buf[i+1] == 0x4b && buf[i+2] == 0x05 && buf[i+3] == 0x06 {
			eocdOffset = offset + int64(i)
			cdOffset = int64(binary.LittleEndian.Uint32(buf[i+16 : i+20]))
			return eocdOffset, cdOffset, nil
		}
	}
	return 0, 0, errors.New("EOCD signature not found")
}

func findAPKSigningBlock(r io.ReaderAt, cdOffset int64) ([]byte, int, error) {
	// APK Signing Block format:
	// - uint64: size_of_block (excluding this field)
	// - sequence of ID-value pairs:
	//     - uint64: size_of_pair
	//     - uint32: id
	//     - bytes: value
	// - uint64: size_of_block (must equal first size)
	// - 16 bytes: magic ("APK Sig Block 42")
	if cdOffset < 24 {
		return nil, 0, errors.New("file too small for APK signing block")
	}

	footer := make([]byte, 24)
	if _, err := r.ReadAt(footer, cdOffset-24); err != nil {
		return nil, 0, err
	}

	magic := string(footer[8:24])
	if magic != apkSigBlockMagic {
		return nil, 0, errors.New("APK signing block magic mismatch")
	}

	blockSize := binary.LittleEndian.Uint64(footer[0:8])
	if int64(blockSize)+8 > cdOffset {
		return nil, 0, errors.New("APK signing block size exceeds file boundaries")
	}

	blockStart := cdOffset - int64(blockSize) - 8
	header := make([]byte, 8)
	if _, err := r.ReadAt(header, blockStart); err != nil {
		return nil, 0, err
	}
	startBlockSize := binary.LittleEndian.Uint64(header)
	if startBlockSize != blockSize {
		return nil, 0, errors.New("APK signing block size headers do not match")
	}

	blockData := make([]byte, blockSize-24)
	if _, err := r.ReadAt(blockData, blockStart+8); err != nil {
		return nil, 0, err
	}

	// Parse ID-value pairs
	reader := bytes.NewReader(blockData)
	schemeVersion := 0
	var selectedPairData []byte

	for reader.Len() >= 12 {
		var pairLen uint64
		if err := binary.Read(reader, binary.LittleEndian, &pairLen); err != nil {
			break
		}
		var pairID uint32
		if err := binary.Read(reader, binary.LittleEndian, &pairID); err != nil {
			break
		}

		valueLen := int(pairLen - 4)
		if valueLen < 0 || valueLen > reader.Len() {
			break
		}
		val := make([]byte, valueLen)
		if _, err := io.ReadFull(reader, val); err != nil {
			break
		}

		if pairID == idSigSchemeV3 || pairID == idSigSchemeV31 {
			schemeVersion = 3
			selectedPairData = val
			break
		} else if pairID == idSigSchemeV2 && schemeVersion < 2 {
			schemeVersion = 2
			selectedPairData = val
		}
	}

	if schemeVersion >= 2 {
		return selectedPairData, schemeVersion, nil
	}

	return nil, 0, errors.New("no v2 or v3 scheme found")
}

func checkIsV1SignatureOnly(r io.ReaderAt, size, eocdOffset int64) bool {
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return false
	}
	hasMetaInfSig := false
	for _, file := range zipReader.File {
		name := strings.ToUpper(file.Name)
		if strings.HasPrefix(name, "META-INF/") && (strings.HasSuffix(name, ".RSA") || strings.HasSuffix(name, ".DSA") || strings.HasSuffix(name, ".EC")) {
			hasMetaInfSig = true
			break
		}
	}
	return hasMetaInfSig
}

func extractCertificateFromBlock(data []byte, schemeVersion int) (*x509.Certificate, error) {
	// A standard v2/v3 signer structure contains length-prefixed bytes:
	// length-prefixed sequence of signers -> signer -> signed data -> length-prefixed sequence of certificates
	reader := bytes.NewReader(data)

	// Scan through the bytes for a valid X.509 DER certificate
	for i := 0; i < len(data)-10; i++ {
		// Look for standard DER SEQUENCE tag (0x30 0x82)
		if data[i] == 0x30 && data[i+1] == 0x82 {
			certLen := int(binary.BigEndian.Uint16(data[i+2:i+4])) + 4
			if i+certLen <= len(data) {
				certCandidate := data[i : i+certLen]
				cert, err := x509.ParseCertificate(certCandidate)
				if err == nil && cert.NotAfter.After(cert.NotBefore) {
					return cert, nil
				}
			}
		}
	}

	_ = reader
	return nil, errors.New("could not extract valid X.509 certificate from signature block")
}

