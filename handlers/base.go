package handlers

import (
	"encoding/json"
	"ssat_backend_rebuild/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseHandler[T any] struct {
	DB *gorm.DB
}

// ===== 底层处理器 ======
func (h *BaseHandler[T]) Select(fields []string) *gorm.DB {
	// 筛选整个过程中需要处理的字段
	query := h.DB.Model(new(T))
	if len(fields) > 0 {
		return query.Select(fields)
	}
	return query
}

func (h *BaseHandler[T]) CreateObject(query *gorm.DB) (T, error) {
	// 创建新模型实例
	var object T
	if err := query.Create(&object).Error; err != nil {
		return object, err
	}
	return object, nil
}

func (h *BaseHandler[T]) GetObject(query *gorm.DB) (T, error) {
	// 获取模型实例
	var object T
	if err := query.First(&object).Error; err != nil {
		return object, err
	}
	return object, nil
}

func (h *BaseHandler[T]) GetObjects(query *gorm.DB) ([]T, error) {
	// 获取模型实例列表
	var objects []T
	if err := query.Find(&objects).Error; err != nil {
		return nil, err
	}
	return objects, nil
}

func (h *BaseHandler[T]) UpdateObject(c *gin.Context, query *gorm.DB, object *T, data map[string]any) error {
	// 更新模型实例
	return query.Model(&object).Updates(data).Error
}

func (h *BaseHandler[T]) DeleteObject(query *gorm.DB, object *T) error {
	// 删除模型实例
	return query.Delete(&object).Error
}

// 解析请求体数据的公共方法
func (h *BaseHandler[T]) parseRequestData(c *gin.Context) (map[string]any, error) {
	fieldsIn := make(map[string]any)
	body, err := c.GetRawData()
	if err != nil {
		return nil, err
	}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &fieldsIn); err != nil {
			return nil, err
		}
	}
	return fieldsIn, nil
}

// 获取分页参数的公共方法
func (h *BaseHandler[T]) getPaginationParams(c *gin.Context) (offset, limit int) {
	page := c.Query("page")
	pageSize := c.Query("page_size")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum < 1 {
		pageSizeNum = 10
	}

	offset = (pageNum - 1) * pageSizeNum
	return offset, pageSizeNum
}

// 优化的 StructToJsonMap 函数
func StructToJsonMap(obj any, fields []string) map[string]any {
	b, _ := json.Marshal(obj)
	var m map[string]any
	json.Unmarshal(b, &m)

	if len(fields) == 0 {
		return m
	}

	filtered := make(map[string]any, len(fields))
	for _, field := range fields {
		if value, exists := m[field]; exists {
			filtered[field] = value
		}
	}
	return filtered
}

// ===== CRUD 函数构造器 ======

func (h *BaseHandler[T]) Create(
	fields []string,
	updaterFn func(c *gin.Context, query *gorm.DB, object *T, data map[string]any) error,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		query := h.Select(fields)

		result, err := h.CreateObject(query)
		if err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}

		fieldsIn, err := h.parseRequestData(c)
		if err != nil {
			utils.Respond(c, nil, utils.ErrMissingParam)
			return
		}

		if updaterFn == nil {
			updaterFn = h.UpdateObject
		}
		if err := updaterFn(c, query, &result, fieldsIn); err != nil {
			utils.Respond(c, nil, utils.ErrorCode{
				Code:     4,
				HttpCode: 400,
				Message:  err.Error(),
			})
			return
		}

		if err := h.DB.Save(&result).Error; err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}
		utils.Respond(c, result, utils.ErrCreated)
	}
}

func (h *BaseHandler[T]) List(
	fields []string,
	filterFn func(c *gin.Context, query *gorm.DB) *gorm.DB,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		query := h.Select(fields)
		if filterFn != nil {
			query = filterFn(c, query)
		}

		// 计算总数
		var total int64
		if err := query.Count(&total).Error; err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}

		// 处理分页参数
		offset, limit := h.getPaginationParams(c)
		query = query.Offset(offset).Limit(limit).Order("created_at DESC")

		results, err := h.GetObjects(query)
		if err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}

		resultJson := make([]map[string]any, 0, len(results))
		for _, result := range results {
			resultJson = append(resultJson, StructToJsonMap(result, fields))
		}
		utils.Respond(c, gin.H{"count": total, "items": resultJson}, utils.ErrOK)
	}
}

func (h *BaseHandler[T]) Retrieve(
	fields []string,
	filterFn func(c *gin.Context, query *gorm.DB) *gorm.DB,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		query := h.Select(fields)
		if filterFn != nil {
			query = filterFn(c, query)
		} else {
			query = query.Where("uuid = ?", c.Param("uuid"))
		}

		result, err := h.GetObject(query)
		if err != nil {
			utils.Respond(c, nil, utils.ErrNotFound)
			return
		}

		resultJson := StructToJsonMap(result, fields)
		utils.Respond(c, resultJson, utils.ErrOK)
	}
}

func (h *BaseHandler[T]) Update(
	fields []string,
	filterFn func(c *gin.Context, query *gorm.DB) *gorm.DB,
	updaterFn func(c *gin.Context, query *gorm.DB, object *T, data map[string]any) error,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		query := h.Select(fields)
		if filterFn != nil {
			query = filterFn(c, query)
		} else {
			query = query.Where("uuid = ?", c.Param("uuid"))
		}

		result, err := h.GetObject(query)
		if err != nil {
			utils.Respond(c, nil, utils.ErrNotFound)
			return
		}

		fieldsIn, err := h.parseRequestData(c)
		if err != nil {
			utils.Respond(c, nil, utils.ErrMissingParam)
			return
		}

		if updaterFn == nil {
			updaterFn = h.UpdateObject
		}
		if err := updaterFn(c, query, &result, fieldsIn); err != nil {
			utils.Respond(c, nil, utils.ErrorCode{
				Code:     4,
				HttpCode: 400,
				Message:  err.Error(),
			})
			return
		}

		if err := query.Save(&result).Error; err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}
		utils.Respond(c, result, utils.ErrOK)
	}
}

func (h *BaseHandler[T]) Destroy(
	filterFn func(c *gin.Context, query *gorm.DB) *gorm.DB,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		query := h.DB
		if filterFn != nil {
			query = filterFn(c, query)
		} else {
			query = query.Where("uuid = ?", c.Param("uuid"))
		}

		result, err := h.GetObject(query)
		if err != nil {
			utils.Respond(c, nil, utils.ErrNotFound)
			return
		}

		err = h.DeleteObject(query, &result)
		if err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}

		utils.Respond(c, gin.H{"message": "删除成功"}, utils.ErrOK)
	}
}
