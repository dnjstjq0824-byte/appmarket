package infrastructure

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type BundletoolRunner interface {
	BuildUniversalAPK(ctx context.Context, aabPath, outputAPKPath, keystorePath, ksPass, keyAlias, keyPass string) error
}

type LocalBundletoolRunner struct {
	jarPath string
}

func NewLocalBundletoolRunner(jarPath string) *LocalBundletoolRunner {
	if jarPath == "" {
		jarPath = "bundletool.jar"
	}
	return &LocalBundletoolRunner{jarPath: jarPath}
}

// BuildUniversalAPK invokes bundletool build-apks --mode=universal to convert an AAB into a standalone universal APK.
func (r *LocalBundletoolRunner) BuildUniversalAPK(ctx context.Context, aabPath, outputAPKPath, keystorePath, ksPass, keyAlias, keyPass string) error {
	tempDir, err := os.MkdirTemp("", "bundletool-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	apksPath := filepath.Join(tempDir, "output.apks")

	args := []string{
		"-jar", r.jarPath,
		"build-apks",
		fmt.Sprintf("--bundle=%s", aabPath),
		fmt.Sprintf("--output=%s", apksPath),
		"--mode=universal",
	}

	if keystorePath != "" {
		args = append(args,
			fmt.Sprintf("--ks=%s", keystorePath),
			fmt.Sprintf("--ks-pass=pass:%s", ksPass),
			fmt.Sprintf("--ks-key-alias=%s", keyAlias),
			fmt.Sprintf("--key-pass=pass:%s", keyPass),
		)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "java", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bundletool execution failed: %w, output: %s", err, string(output))
	}

	// Unzip universal.apk from output.apks
	cmdExtract := exec.CommandContext(timeoutCtx, "unzip", "-p", apksPath, "universal.apk")
	apkBytes, err := cmdExtract.Output()
	if err != nil {
		return fmt.Errorf("failed to extract universal.apk: %w", err)
	}

	if err := os.WriteFile(outputAPKPath, apkBytes, 0644); err != nil {
		return fmt.Errorf("failed to save universal apk: %w", err)
	}

	return nil
}

