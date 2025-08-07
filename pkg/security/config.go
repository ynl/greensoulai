package security

import (
	"time"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	Version     string        `json:"version"`
	Fingerprint *Fingerprint  `json:"fingerprint"`
	EnableAudit bool          `json:"enable_audit"`
	MaxAge      time.Duration `json:"max_age"`
}

// NewSecurityConfig 创建新的安全配置
func NewSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		Version:     "1.0.0",
		Fingerprint: NewFingerprint(),
		EnableAudit: false,
		MaxAge:      24 * time.Hour,
	}
}

// NewSecurityConfigWithFingerprint 使用指定指纹创建安全配置
func NewSecurityConfigWithFingerprint(fingerprint *Fingerprint) *SecurityConfig {
	return &SecurityConfig{
		Version:     "1.0.0",
		Fingerprint: fingerprint,
		EnableAudit: false,
		MaxAge:      24 * time.Hour,
	}
}

// SetAuditEnabled 设置审计启用状态
func (sc *SecurityConfig) SetAuditEnabled(enabled bool) {
	sc.EnableAudit = enabled
}

// IsAuditEnabled 检查审计是否启用
func (sc *SecurityConfig) IsAuditEnabled() bool {
	return sc.EnableAudit
}

// SetMaxAge 设置最大年龄
func (sc *SecurityConfig) SetMaxAge(maxAge time.Duration) {
	sc.MaxAge = maxAge
}

// GetMaxAge 获取最大年龄
func (sc *SecurityConfig) GetMaxAge() time.Duration {
	return sc.MaxAge
}

// IsExpired 检查是否过期
func (sc *SecurityConfig) IsExpired() bool {
	if sc.Fingerprint == nil {
		return true
	}
	return sc.Fingerprint.GetAge() > sc.MaxAge
}
