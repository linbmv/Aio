package handler

import (
	"fmt"

	"github.com/atopos31/llmio/common"
	"github.com/atopos31/llmio/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListProviderKeys 获取 Provider 的 Key 列表
func ListProviderKeys(c *gin.Context) {
	ctx := c.Request.Context()
	providerID := c.Param("id")

	keys, err := gorm.G[models.ProviderKey](models.DB).
		Where("provider_id = ?", providerID).
		Order("created_at DESC").
		Find(ctx)
	if err != nil {
		common.InternalServerError(c, err.Error())
		return
	}

	common.Success(c, keys)
}

// CreateProviderKey 创建 Provider Key
func CreateProviderKey(c *gin.Context) {
	ctx := c.Request.Context()
	providerID := c.Param("id")

	var key models.ProviderKey
	if err := c.ShouldBindJSON(&key); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	// 强制使用 URL 中的 provider_id
	var pid uint
	if _, err := fmt.Sscanf(providerID, "%d", &pid); err != nil {
		common.BadRequest(c, "invalid provider id")
		return
	}
	key.ProviderID = pid
	// 默认启用新创建的 Key
	if !key.Status {
		key.Status = true
	}

	if err := gorm.G[models.ProviderKey](models.DB).Create(ctx, &key); err != nil {
		common.InternalServerError(c, err.Error())
		return
	}

	common.Success(c, key)
}

// UpdateProviderKey 更新 Provider Key
func UpdateProviderKey(c *gin.Context) {
	ctx := c.Request.Context()
	var pid, kid uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &pid); err != nil {
		common.BadRequest(c, "invalid provider id")
		return
	}
	if _, err := fmt.Sscanf(c.Param("keyId"), "%d", &kid); err != nil {
		common.BadRequest(c, "invalid key id")
		return
	}

	var req struct {
		Remark string `json:"remark"`
		Status *bool  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	update := models.ProviderKey{
		Remark: req.Remark,
	}
	if req.Status != nil {
		update.Status = *req.Status
	}

	if _, err := gorm.G[models.ProviderKey](models.DB).
		Where("id = ? AND provider_id = ?", kid, pid).
		Updates(ctx, update); err != nil {
		common.InternalServerError(c, err.Error())
		return
	}

	common.Success(c, nil)
}

// DeleteProviderKey 删除 Provider Key
func DeleteProviderKey(c *gin.Context) {
	ctx := c.Request.Context()
	var pid, kid uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &pid); err != nil {
		common.BadRequest(c, "invalid provider id")
		return
	}
	if _, err := fmt.Sscanf(c.Param("keyId"), "%d", &kid); err != nil {
		common.BadRequest(c, "invalid key id")
		return
	}

	if _, err := gorm.G[models.ProviderKey](models.DB).
		Where("id = ? AND provider_id = ?", kid, pid).
		Delete(ctx); err != nil {
		common.InternalServerError(c, err.Error())
		return
	}

	common.Success(c, nil)
}
