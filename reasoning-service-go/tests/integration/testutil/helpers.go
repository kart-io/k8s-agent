package testutil

import (
	"os"
	"path/filepath"
	"reasoning-service-go/internal/config"
	"testing"
)

// LoadTestConfig 加载测试配置文件
func LoadTestConfig(t *testing.T) *config.Config {
	t.Helper()

	// 获取项目根目录
	rootDir, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	configPath := filepath.Join(rootDir, "configs", "config-test.yaml")

	cfg, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	return cfg
}

// findProjectRoot 查找项目根目录 (包含 go.mod 的目录)
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 向上查找直到找到 go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// 已经到根目录
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

// SetupTestDataDir 创建测试数据目录
func SetupTestDataDir(t *testing.T) string {
	t.Helper()

	rootDir, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	testDataDir := filepath.Join(rootDir, "tests", "integration", "testdata")

	// 创建测试数据目录
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}

	// 创建子目录
	subdirs := []string{"chroma", "learning", "logs"}
	for _, subdir := range subdirs {
		dir := filepath.Join(testDataDir, subdir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create subdirectory %s: %v", subdir, err)
		}
	}

	return testDataDir
}

// CleanupTestDataDir 清理测试数据目录
func CleanupTestDataDir(t *testing.T) {
	t.Helper()

	rootDir, err := findProjectRoot()
	if err != nil {
		t.Logf("Warning: Failed to find project root for cleanup: %v", err)
		return
	}

	testDataDir := filepath.Join(rootDir, "tests", "integration", "testdata")

	// 清空测试数据目录但保留目录本身
	entries, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Logf("Warning: Failed to read test data directory: %v", err)
		return
	}

	for _, entry := range entries {
		path := filepath.Join(testDataDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			t.Logf("Warning: Failed to remove %s: %v", path, err)
		}
	}
}

// SkipIfShort 如果运行短测试则跳过
func SkipIfShort(t *testing.T, reason string) {
	t.Helper()
	if testing.Short() {
		t.Skipf("Skipping integration test in short mode: %s", reason)
	}
}

// RequireEnvVar 要求环境变量存在
func RequireEnvVar(t *testing.T, key string) string {
	t.Helper()
	value := os.Getenv(key)
	if value == "" {
		t.Skipf("Skipping test: required environment variable %s is not set", key)
	}
	return value
}
