package keypool

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/atopos31/llmio/models"
	"github.com/atopos31/llmio/providers"
	"gorm.io/gorm"
)

// SyncProviderConfigKeys 同步单个 Provider 的配置 keys 到 key pool
func SyncProviderConfigKeys(ctx context.Context, db *gorm.DB, providerID uint, config string) error {
	// 解析配置获取 keys
	var configData struct {
		Keys []providers.KeyConfig `json:"keys"`
	}
	if err := json.Unmarshal([]byte(config), &configData); err != nil {
		return err
	}

	// 如果配置中没有 keys，跳过
	if len(configData.Keys) == 0 {
		return nil
	}

	// 获取已存在的 keys
	existingKeys, err := gorm.G[models.ProviderKey](db).
		Where("provider_id = ?", providerID).
		Find(ctx)
	if err != nil {
		return err
	}

	// 构建已存在 key 的映射
	existingKeyMap := make(map[string]models.ProviderKey)
	for _, k := range existingKeys {
		existingKeyMap[k.Key] = k
	}

	// 同步 keys
	for _, configKey := range configData.Keys {
		if configKey.Term == "" {
			continue
		}

		if existing, found := existingKeyMap[configKey.Term]; found {
			// Key 已存在，更新 remark 和 status
			if err := db.WithContext(ctx).Model(&models.ProviderKey{}).
				Where("id = ?", existing.ID).
				Updates(map[string]interface{}{
					"remark": configKey.Remark,
					"status": configKey.Status,
				}).Error; err != nil {
				slog.Error("Failed to update provider key", "error", err, "key_id", existing.ID)
			}
		} else {
			// Key 不存在，创建新记录
			newKey := models.ProviderKey{
				ProviderID: providerID,
				Key:        configKey.Term,
				Remark:     configKey.Remark,
				Status:     configKey.Status,
			}
			if err := gorm.G[models.ProviderKey](db).Create(ctx, &newKey); err != nil {
				slog.Error("Failed to create provider key", "error", err, "provider_id", providerID)
			}
		}
	}

	return nil
}

// SyncAllProvidersFromConfig 同步所有 Provider 的配置 keys 到 key pool
func SyncAllProvidersFromConfig(ctx context.Context, db *gorm.DB) error {
	providers, err := gorm.G[models.Provider](db).Find(ctx)
	if err != nil {
		return err
	}

	for _, provider := range providers {
		if err := SyncProviderConfigKeys(ctx, db, provider.ID, provider.Config); err != nil {
			slog.Error("Failed to sync provider keys", "error", err, "provider_id", provider.ID)
		}
	}

	return nil
}
