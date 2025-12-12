package cooldown

import (
	"context"
	"time"

	"github.com/atopos31/llmio/models"
	"gorm.io/gorm"
)

// Manager 缁存姢閿骇涓庢笭閬撶骇鍐峰嵈绐楀彛
type Manager struct {
	db         *gorm.DB
	now        func() time.Time
	maxBackoff time.Duration
}

func NewManager(db *gorm.DB) *Manager {
	return &Manager{
		db:         db,
		now:        time.Now,
		maxBackoff: 30 * time.Minute,
	}
}

// InCooldown 鍒ゆ柇鏄惁澶勪簬浠讳竴鍐峰嵈绐楀彛
func (m *Manager) InCooldown(mp *models.ModelWithProvider) bool {
	return m.CooldownLeft(mp) > 0
}

// CooldownLeft 杩斿洖鍓╀綑鍐峰嵈鏃堕棿锛屾湭鍐峰嵈鍒欎负0
func (m *Manager) CooldownLeft(mp *models.ModelWithProvider) time.Duration {
	now := m.now()
	var remain time.Duration
	if mp.KeyCooldownUntil != nil && now.Before(*mp.KeyCooldownUntil) {
		remain = mp.KeyCooldownUntil.Sub(now)
	}
	if mp.ProviderCooldownUntil != nil && now.Before(*mp.ProviderCooldownUntil) {
		if d := mp.ProviderCooldownUntil.Sub(now); d > remain {
			remain = d
		}
	}
	return remain
}

// OnSuccess 璇锋眰鎴愬姛鍚庢竻鐞嗗喎鍗寸姸鎬?
func (m *Manager) OnSuccess(ctx context.Context, mp *models.ModelWithProvider) error {
	mp.KeyCooldownUntil = nil
	mp.ProviderCooldownUntil = nil
	mp.KeyCooldownStep = 0
	mp.ProviderCooldownStep = 0
	_, err := gorm.G[models.ModelWithProvider](m.db).
		Where("id = ?", mp.ID).
		Updates(ctx, models.ModelWithProvider{
			KeyCooldownUntil:      nil,
			ProviderCooldownUntil: nil,
			KeyCooldownStep:       0,
			ProviderCooldownStep:  0,
		})
	return err
}

// OnError 鎸夊垎绫诲彔鍔犲喎鍗村苟鍐欏簱
func (m *Manager) OnError(ctx context.Context, mp *models.ModelWithProvider, category Category) error {
	switch category {
	case CategoryKey:
		mp.KeyCooldownStep++
		until := m.nextTime(mp.KeyCooldownStep)
		mp.KeyCooldownUntil = &until
		_, err := gorm.G[models.ModelWithProvider](m.db).Where("id = ?", mp.ID).Updates(ctx, models.ModelWithProvider{
			KeyCooldownStep:  mp.KeyCooldownStep,
			KeyCooldownUntil: mp.KeyCooldownUntil,
		})
		return err
	case CategoryProvider:
		mp.ProviderCooldownStep++
		until := m.nextTime(mp.ProviderCooldownStep)
		mp.ProviderCooldownUntil = &until
		_, err := gorm.G[models.ModelWithProvider](m.db).Where("id = ?", mp.ID).Updates(ctx, models.ModelWithProvider{
			ProviderCooldownStep:  mp.ProviderCooldownStep,
			ProviderCooldownUntil: mp.ProviderCooldownUntil,
		})
		return err
	default:
		return nil
	}
}

func (m *Manager) nextTime(step int) time.Time {
	return m.now().Add(m.backoff(step))
}

// backoff 閲囩敤鎸囨暟閫€閬匡紝1s 璧枫€佺炕鍊嶈嚦30min灏侀《
func (m *Manager) backoff(step int) time.Duration {
	if step <= 0 {
		return time.Second
	}
	// 闄愬埗浣嶇Щ閬垮厤婧㈠嚭
	if step > 20 {
		step = 20
	}
	delay := time.Second << (step - 1)
	if delay > m.maxBackoff {
		return m.maxBackoff
	}
	return delay
}
