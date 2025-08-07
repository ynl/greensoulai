package security

import (
	"crypto/sha1"
	"time"

	"github.com/google/uuid"
)

// Fingerprint 组件指纹
type Fingerprint struct {
	UUID      string                 `json:"uuid"`
	CreatedAt time.Time              `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// NewFingerprint 创建新的指纹
func NewFingerprint() *Fingerprint {
	return &Fingerprint{
		UUID:      uuid.New().String(),
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// GenerateDeterministic 生成确定性指纹
func GenerateDeterministic(seed string) *Fingerprint {
	if seed == "" {
		return NewFingerprint()
	}

	// 使用SHA1生成确定性UUID
	hash := sha1.Sum([]byte(seed))
	uuidBytes := make([]byte, 16)
	copy(uuidBytes, hash[:16])

	// 设置UUID版本和变体
	uuidBytes[6] = (uuidBytes[6] & 0x0f) | 0x50 // 版本5
	uuidBytes[8] = (uuidBytes[8] & 0x3f) | 0x80 // 变体

	deterministicUUID := uuid.Must(uuid.FromBytes(uuidBytes))

	return &Fingerprint{
		UUID:      deterministicUUID.String(),
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"seed": seed,
		},
	}
}

// GetUUID 获取UUID字符串
func (f *Fingerprint) GetUUID() string {
	return f.UUID
}

// GetAge 获取指纹年龄
func (f *Fingerprint) GetAge() time.Duration {
	return time.Since(f.CreatedAt)
}

// AddMetadata 添加元数据
func (f *Fingerprint) AddMetadata(key string, value interface{}) {
	f.Metadata[key] = value
}

// GetMetadata 获取元数据
func (f *Fingerprint) GetMetadata(key string) (interface{}, bool) {
	value, exists := f.Metadata[key]
	return value, exists
}

// String 字符串表示
func (f *Fingerprint) String() string {
	return f.UUID
}

// Equals 比较指纹是否相等
func (f *Fingerprint) Equals(other *Fingerprint) bool {
	return f.UUID == other.UUID
}
