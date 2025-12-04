package providers

import (
	"fmt"
	"sync"
	"time"
)

// keyCooldownManager key冷却管理，标记失败Key避免短时间重复使用
type keyCooldownManager struct {
	mu           sync.RWMutex
	expireAt     map[string]time.Time
	failureCount map[string]int
	duration     time.Duration
}

func newKeyCooldownManager(defaultDuration time.Duration) *keyCooldownManager {
	return &keyCooldownManager{
		expireAt:     make(map[string]time.Time),
		failureCount: make(map[string]int),
		duration:     defaultDuration,
	}
}

func (m *keyCooldownManager) mark(id string, d time.Duration) {
	if id == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	count := m.failureCount[id] + 1
	m.failureCount[id] = count

	if d <= 0 {
		d = calculateBackoff(m.duration, count)
	}
	m.expireAt[id] = time.Now().Add(d)
}

func calculateBackoff(base time.Duration, failureCount int) time.Duration {
	d := base
	for i := 1; i < failureCount && d < 30*time.Minute; i++ {
		d *= 2
	}
	if d > 30*time.Minute {
		d = 30 * time.Minute
	}
	return d
}

func (m *keyCooldownManager) isCooling(id string, now time.Time) bool {
	if id == "" {
		return false
	}
	m.mu.RLock()
	expire, ok := m.expireAt[id]
	m.mu.RUnlock()
	if !ok {
		return false
	}
	if expire.After(now) {
		return true
	}
	// 过期清理
	m.mu.Lock()
	delete(m.expireAt, id)
	delete(m.failureCount, id)
	m.mu.Unlock()
	return false
}

var globalKeyCooldown = newKeyCooldownManager(60 * time.Second)

func init() {
	// 启动定期清理过期条目的 goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			globalKeyCooldown.cleanup()
		}
	}()
}

func (m *keyCooldownManager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for id, expireAt := range m.expireAt {
		if expireAt.Before(now) {
			delete(m.expireAt, id)
			delete(m.failureCount, id)
		}
	}
}

func makeKeyCooldownID(providerID uint, key string) string {
	return fmt.Sprintf("%d:%s", providerID, key)
}

// MarkKeyFailure 将Key标记为冷却，支持自定义冷却时长
func MarkKeyFailure(providerID uint, key string, duration time.Duration) {
	if key == "" {
		return
	}
	globalKeyCooldown.mark(makeKeyCooldownID(providerID, key), duration)
}

// IsKeyCoolingDown 判断Key是否仍在冷却
func IsKeyCoolingDown(providerID uint, key string) bool {
	if key == "" {
		return false
	}
	return globalKeyCooldown.isCooling(makeKeyCooldownID(providerID, key), time.Now())
}
