package security

import (
	"testing"
	"time"
)

func TestNewFingerprint(t *testing.T) {
	fp := NewFingerprint()

	if fp.UUID == "" {
		t.Error("expected UUID to be set")
	}

	if fp.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if fp.Metadata == nil {
		t.Error("expected Metadata to be initialized")
	}
}

func TestGenerateDeterministic(t *testing.T) {
	seed := "test_seed"
	fp1 := GenerateDeterministic(seed)
	fp2 := GenerateDeterministic(seed)

	// 相同种子应该生成相同的UUID
	if fp1.UUID != fp2.UUID {
		t.Errorf("deterministic fingerprints should be equal: %s vs %s", fp1.UUID, fp2.UUID)
	}

	// 检查元数据
	if fp1.Metadata["seed"] != seed {
		t.Errorf("expected seed '%s', got '%v'", seed, fp1.Metadata["seed"])
	}
}

func TestGenerateDeterministic_EmptySeed(t *testing.T) {
	fp1 := GenerateDeterministic("")
	fp2 := GenerateDeterministic("")

	// 空种子应该生成不同的UUID（因为使用随机生成）
	if fp1.UUID == fp2.UUID {
		t.Error("empty seed fingerprints should be different")
	}
}

func TestFingerprint_Methods(t *testing.T) {
	fp := NewFingerprint()

	// 测试GetUUID
	if fp.GetUUID() != fp.UUID {
		t.Errorf("GetUUID returned '%s', expected '%s'", fp.GetUUID(), fp.UUID)
	}

	// 测试GetAge
	age := fp.GetAge()
	if age < 0 {
		t.Errorf("expected positive age, got %v", age)
	}

	// 测试AddMetadata和GetMetadata
	fp.AddMetadata("test_key", "test_value")
	value, exists := fp.GetMetadata("test_key")
	if !exists {
		t.Error("expected metadata to exist")
	}
	if value != "test_value" {
		t.Errorf("expected 'test_value', got '%v'", value)
	}

	// 测试不存在的键
	_, exists = fp.GetMetadata("nonexistent")
	if exists {
		t.Error("expected metadata to not exist")
	}
}

func TestFingerprint_String(t *testing.T) {
	fp := NewFingerprint()

	if fp.String() != fp.UUID {
		t.Errorf("String() returned '%s', expected '%s'", fp.String(), fp.UUID)
	}
}

func TestFingerprint_Equals(t *testing.T) {
	fp1 := NewFingerprint()
	fp2 := NewFingerprint()

	// 不同的指纹应该不相等
	if fp1.Equals(fp2) {
		t.Error("different fingerprints should not be equal")
	}

	// 相同的指纹应该相等
	if !fp1.Equals(fp1) {
		t.Error("same fingerprint should be equal to itself")
	}
}

func TestSecurityConfig_New(t *testing.T) {
	config := NewSecurityConfig()

	if config.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", config.Version)
	}

	if config.Fingerprint == nil {
		t.Error("expected Fingerprint to be set")
	}

	if config.EnableAudit {
		t.Error("expected EnableAudit to be false by default")
	}

	if config.MaxAge != 24*time.Hour {
		t.Errorf("expected MaxAge to be 24h, got %v", config.MaxAge)
	}
}

func TestSecurityConfig_WithFingerprint(t *testing.T) {
	fp := NewFingerprint()
	config := NewSecurityConfigWithFingerprint(fp)

	if config.Fingerprint != fp {
		t.Error("expected Fingerprint to be the provided one")
	}
}

func TestSecurityConfig_Methods(t *testing.T) {
	config := NewSecurityConfig()

	// 测试审计设置
	config.SetAuditEnabled(true)
	if !config.IsAuditEnabled() {
		t.Error("expected audit to be enabled")
	}

	config.SetAuditEnabled(false)
	if config.IsAuditEnabled() {
		t.Error("expected audit to be disabled")
	}

	// 测试最大年龄设置
	newMaxAge := 12 * time.Hour
	config.SetMaxAge(newMaxAge)
	if config.GetMaxAge() != newMaxAge {
		t.Errorf("expected MaxAge to be %v, got %v", newMaxAge, config.GetMaxAge())
	}

	// 测试过期检查
	if config.IsExpired() {
		t.Error("new config should not be expired")
	}
}
