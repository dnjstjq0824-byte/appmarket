// Standalone synthetic-APK generator for binary/release smoke tests.
// Replicates the backend's own test helper (distribution application tests,
// v2 path): plain-text manifest + fake v2 signing block with a self-signed
// cert. Stdlib only. Usage (no go.mod needed, file-arg mode):
//
//	go run scripts/smoke/genapk/main.go C:\tmp\fixture.apk
package main

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"fmt"
	"math/big"
	"os"
	"time"
)

const manifestXML = `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.quokka.livee2e"
    android:versionCode="7"
    android:versionName="1.0.7">
    <uses-sdk android:minSdkVersion="26" android:targetSdkVersion="34" />
    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.CAMERA" />
</manifest>`

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: genapk <output.apk>")
		os.Exit(2)
	}
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	mfWriter, err := zipWriter.Create("AndroidManifest.xml")
	if err != nil {
		panic(err)
	}
	if _, err := mfWriter.Write([]byte(manifestXML)); err != nil {
		panic(err)
	}
	fw, err := zipWriter.Create("lib/arm64-v8a/libnative.so")
	if err != nil {
		panic(err)
	}
	if _, err := fw.Write([]byte("ELF binary stub")); err != nil {
		panic(err)
	}
	if err := zipWriter.Close(); err != nil {
		panic(err)
	}
	zipBytes := buf.Bytes()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{
		SerialNumber:          big.NewInt(987654321),
		Subject:               pkix.Name{CommonName: "Test Official Release", Organization: []string{"Acme Corp"}, Country: []string{"KR"}},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		panic(err)
	}

	v2ValueBuf := new(bytes.Buffer)
	v2ValueBuf.Write(certDER)
	pairBuf := new(bytes.Buffer)
	pairLen := uint64(4 + v2ValueBuf.Len())
	_ = binary.Write(pairBuf, binary.LittleEndian, pairLen)
	_ = binary.Write(pairBuf, binary.LittleEndian, uint32(0x7109871a))
	pairBuf.Write(v2ValueBuf.Bytes())
	blockSize := uint64(pairBuf.Len() + 8 + 16)
	sigBlockBuf := new(bytes.Buffer)
	_ = binary.Write(sigBlockBuf, binary.LittleEndian, blockSize)
	sigBlockBuf.Write(pairBuf.Bytes())
	_ = binary.Write(sigBlockBuf, binary.LittleEndian, blockSize)
	sigBlockBuf.WriteString("APK Sig Block 42")

	eocdIdx := bytes.LastIndex(zipBytes, []byte{0x50, 0x4b, 0x05, 0x06})
	if eocdIdx == -1 {
		panic("EOCD not found")
	}
	cdOffset := binary.LittleEndian.Uint32(zipBytes[eocdIdx+16 : eocdIdx+20])
	newApk := new(bytes.Buffer)
	newApk.Write(zipBytes[:cdOffset])
	newApk.Write(sigBlockBuf.Bytes())
	newCDOffset := cdOffset + uint32(sigBlockBuf.Len())
	newEOCD := make([]byte, len(zipBytes)-eocdIdx)
	copy(newEOCD, zipBytes[eocdIdx:])
	binary.LittleEndian.PutUint32(newEOCD[16:20], newCDOffset)
	newApk.Write(zipBytes[cdOffset:eocdIdx])
	newApk.Write(newEOCD)

	out := os.Args[1]
	if err := os.WriteFile(out, newApk.Bytes(), 0644); err != nil {
		panic(err)
	}
	sum := sha256.Sum256(newApk.Bytes())
	fmt.Printf("APK=%s\nSIZE=%d\nSHA256=%x\n", out, len(newApk.Bytes()), sum)
}
