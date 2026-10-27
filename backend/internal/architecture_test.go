package internal_test

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const moduleImportRoot = "github.com/bookingnetwork/appmarket/backend/internal/modules/"

func checkDomainImports(root string) ([]string, error) {
	var violations []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) < 2 {
			return fmt.Errorf("module file outside a domain: %s", path)
		}
		owner := parts[0]
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range file.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if strings.HasPrefix(importPath, moduleImportRoot) {
				other := strings.Split(strings.TrimPrefix(importPath, moduleImportRoot), "/")[0]
				if other != owner {
					violations = append(violations, fmt.Sprintf("%s imports domain %s", path, other))
				}
			}
			if strings.Contains(importPath, "/db/queries/") {
				violations = append(violations, fmt.Sprintf("%s imports DB queries directly", path))
			}
		}
		return nil
	})
	return violations, err
}

// Domain modules may not import another module's internals or its SQL storage.
func TestDomainImportBoundaries(t *testing.T) {
	violations, err := checkDomainImports("modules")
	if errors.Is(err, fs.ErrNotExist) {
		t.Log("domain modules are not present yet; boundary checker is exercised by fixture tests")
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, violation := range violations {
		t.Error(violation)
	}
}

func TestDomainImportBoundaryChecker(t *testing.T) {
	root := t.TempDir()
	write := func(rel, source string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("catalog/application/valid.go", `package application
import _ "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
`)
	write("catalog/application/invalid.go", `package application
import (
    _ "github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
    _ "github.com/bookingnetwork/appmarket/backend/internal/db/queries/private"
)
`)
	violations, err := checkDomainImports(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 2 || !strings.Contains(violations[0], "imports domain iam") || !strings.Contains(violations[1], "imports DB queries directly") {
		t.Fatalf("boundary checker found %q, want cross-domain and direct-query violations", violations)
	}
}
