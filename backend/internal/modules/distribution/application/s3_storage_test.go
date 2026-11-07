package application

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
)

// mockS3 is a minimal path-style S3 mock: objects, metadata, and tags.
type mockS3 struct {
	mu      sync.Mutex
	objects map[string]*storedMockObject
	signed  int // requests carrying a SigV4 Authorization header
}

type storedMockObject struct {
	content            []byte
	contentType        string
	cacheControl       string
	contentDisposition string
	sha256             string
	tags               map[string]string
}

func newMockS3() *mockS3 {
	return &mockS3{objects: map[string]*storedMockObject{}}
}

type taggingXML struct {
	XMLName xml.Name `xml:"Tagging"`
	TagSet  []tagXML `xml:"TagSet>Tag"`
}

type tagXML struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}

func (m *mockS3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256") {
		m.signed++
	}
	// Path-style: /{bucket}/{key...}
	path := strings.TrimPrefix(r.URL.Path, "/")
	slash := strings.Index(path, "/")
	if slash < 0 {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}
	id := path // bucket/key
	if r.URL.Query().Has("tagging") {
		m.serveTagging(w, r, id)
		return
	}
	switch r.Method {
	case http.MethodPut:
		body, _ := io.ReadAll(r.Body)
		obj := &storedMockObject{
			content:            body,
			contentType:        r.Header.Get("Content-Type"),
			cacheControl:       r.Header.Get("Cache-Control"),
			contentDisposition: r.Header.Get("Content-Disposition"),
			sha256:             r.Header.Get("X-Amz-Meta-Sha256"),
			tags:               map[string]string{},
		}
		if raw := r.Header.Get("X-Amz-Tagging"); raw != "" {
			if values, err := url.ParseQuery(raw); err == nil {
				for k := range values {
					obj.tags[k] = values.Get(k)
				}
			}
		}
		m.objects[id] = obj
		w.WriteHeader(http.StatusOK)
	case http.MethodGet:
		obj, ok := m.objects[id]
		if !ok {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`<?xml version="1.0"?><Error><Code>NoSuchKey</Code><Message>missing</Message></Error>`))
			return
		}
		if obj.contentType != "" {
			w.Header().Set("Content-Type", obj.contentType)
		}
		if obj.cacheControl != "" {
			w.Header().Set("Cache-Control", obj.cacheControl)
		}
		if obj.contentDisposition != "" {
			w.Header().Set("Content-Disposition", obj.contentDisposition)
		}
		if obj.sha256 != "" {
			w.Header().Set("X-Amz-Meta-Sha256", obj.sha256)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(obj.content)
	case http.MethodDelete:
		delete(m.objects, id)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "unsupported", http.StatusBadRequest)
	}
}

func (m *mockS3) serveTagging(w http.ResponseWriter, r *http.Request, id string) {
	obj, ok := m.objects[id]
	if !ok {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><Error><Code>NoSuchKey</Code><Message>missing</Message></Error>`))
		return
	}
	switch r.Method {
	case http.MethodGet:
		tags := taggingXML{}
		for k, v := range obj.tags {
			tags.TagSet = append(tags.TagSet, tagXML{Key: k, Value: v})
		}
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_ = xml.NewEncoder(w).Encode(tags)
	case http.MethodPut:
		body, _ := io.ReadAll(r.Body)
		var tags taggingXML
		if err := xml.Unmarshal(body, &tags); err != nil {
			http.Error(w, "bad tagging xml", http.StatusBadRequest)
			return
		}
		obj.tags = map[string]string{}
		for _, t := range tags.TagSet {
			obj.tags[t.Key] = t.Value
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "unsupported", http.StatusBadRequest)
	}
}

func newTestS3Store(t *testing.T, mock *mockS3) (*S3ObjectStorage, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(mock)
	t.Cleanup(server.Close)
	store, err := NewS3ObjectStorage(context.Background(), S3Config{
		Endpoint:        server.URL,
		Region:          "ap-northeast-2",
		AccessKeyID:     "test",
		SecretAccessKey: "test",
		UsePathStyle:    true,
		HTTPClient:      server.Client(),
	})
	if err != nil {
		t.Fatalf("store construction failed: %v", err)
	}
	return store, server
}

func TestS3Storage_PutGetRoundTrip(t *testing.T) {
	ctx := context.Background()
	mock := newMockS3()
	store, _ := newTestS3Store(t, mock)

	original := &StorageObject{
		Key:                "quarantine/org/app/session-1",
		Content:            []byte("%PDF-fake-apk-bytes"),
		SHA256:             "ab" + strings.Repeat("cd", 31),
		ContentType:        "application/vnd.android.package-archive",
		CacheControl:       "public, max-age=31536000, immutable",
		ContentDisposition: `attachment; filename="app-1.apk"`,
		Tags:               map[string]string{"status": "quarantined", "app_id": "app-1"},
	}
	if err := store.PutObject(ctx, "bn-quarantine-bucket", original); err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}
	got, err := store.GetObject(ctx, "bn-quarantine-bucket", original.Key)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	if string(got.Content) != string(original.Content) {
		t.Fatalf("content mismatch: %q", got.Content)
	}
	if got.SHA256 != original.SHA256 || got.ContentType != original.ContentType ||
		got.CacheControl != original.CacheControl || got.ContentDisposition != original.ContentDisposition {
		t.Fatalf("header/metadata mismatch: %+v", got)
	}
	if got.Tags["status"] != "quarantined" || got.Tags["app_id"] != "app-1" {
		t.Fatalf("tag mismatch: %+v", got.Tags)
	}
	if mock.signed == 0 {
		t.Fatalf("expected SigV4-signed requests")
	}
}

func TestS3Storage_GetMissingMapsToNotFound(t *testing.T) {
	ctx := context.Background()
	store, _ := newTestS3Store(t, newMockS3())
	if _, err := store.GetObject(ctx, "bn-quarantine-bucket", "nope/missing"); !errors.Is(err, ErrObjectNotFound) {
		t.Fatalf("expected ErrObjectNotFound, got %v", err)
	}
}

func TestS3Storage_DeleteIsIdempotent(t *testing.T) {
	ctx := context.Background()
	store, _ := newTestS3Store(t, newMockS3())
	if err := store.DeleteObject(ctx, "bn-quarantine-bucket", "never/existed"); err != nil {
		t.Fatalf("delete of missing key must succeed: %v", err)
	}
	if err := store.PutObject(ctx, "bn-quarantine-bucket", &StorageObject{Key: "tmp/x", Content: []byte("x")}); err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}
	if err := store.DeleteObject(ctx, "bn-quarantine-bucket", "tmp/x"); err != nil {
		t.Fatalf("DeleteObject failed: %v", err)
	}
	if _, err := store.GetObject(ctx, "bn-quarantine-bucket", "tmp/x"); !errors.Is(err, ErrObjectNotFound) {
		t.Fatalf("expected ErrObjectNotFound after delete, got %v", err)
	}
}

func TestS3Storage_TagReplaceAndMissing(t *testing.T) {
	ctx := context.Background()
	store, _ := newTestS3Store(t, newMockS3())
	if err := store.PutObject(ctx, "bn-public-cdn-bucket", &StorageObject{
		Key: "packages/a/r/sha.apk", Content: []byte("bin"), Tags: map[string]string{"status": "published"},
	}); err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}
	if err := store.TagObject(ctx, "bn-public-cdn-bucket", "packages/a/r/sha.apk",
		map[string]string{"status": "quarantined_revoked"}); err != nil {
		t.Fatalf("TagObject failed: %v", err)
	}
	got, err := store.GetObject(ctx, "bn-public-cdn-bucket", "packages/a/r/sha.apk")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	if got.Tags["status"] != "quarantined_revoked" {
		t.Fatalf("expected replaced tags, got %+v", got.Tags)
	}
	if err := store.TagObject(ctx, "bn-public-cdn-bucket", "missing/key", map[string]string{"a": "b"}); !errors.Is(err, ErrObjectNotFound) {
		t.Fatalf("expected ErrObjectNotFound for missing key, got %v", err)
	}
}

func TestS3Storage_PromoteFlowAgainstS3Backend(t *testing.T) {
	ctx := context.Background()
	store, _ := newTestS3Store(t, newMockS3())
	publisher := NewCDNPublisher(store, "bn-quarantine-bucket", "bn-public-cdn-bucket", "cdn.example.local")

	content := []byte("fake-signed-apk-bytes")
	sha := domain.ComputeSHA256(content)
	quarantineKey := "quarantine/org/app/sess-9"
	if err := store.PutObject(ctx, "bn-quarantine-bucket", &StorageObject{
		Key: quarantineKey, Content: content, SHA256: sha,
		ContentType: "application/vnd.android.package-archive",
	}); err != nil {
		t.Fatalf("quarantine PutObject failed: %v", err)
	}
	artifact := &domain.ReleaseArtifact{
		ID: "art-s3-1", ReleaseID: "rel-s3-1", PackageName: "com.quokka.s3probe",
		VersionCode: 9, SHA256Hash: sha, S3QuarantineKey: quarantineKey,
	}
	cdnURL, err := publisher.PromoteToPublicCDN(ctx, artifact)
	if err != nil {
		t.Fatalf("PromoteToPublicCDN failed: %v", err)
	}
	if !strings.HasPrefix(cdnURL, "https://cdn.example.local/packages/") {
		t.Fatalf("unexpected cdn url %q", cdnURL)
	}
	publicKey := strings.TrimPrefix(cdnURL, "https://cdn.example.local/")
	publicObj, err := store.GetObject(ctx, "bn-public-cdn-bucket", publicKey)
	if err != nil {
		t.Fatalf("promoted object missing: %v", err)
	}
	if publicObj.CacheControl != "public, max-age=31536000, immutable" || string(publicObj.Content) != string(content) {
		t.Fatalf("promoted object mismatch: %+v", publicObj)
	}
}
