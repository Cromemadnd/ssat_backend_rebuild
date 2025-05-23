package handlers

import (
	"encoding/json"
	"fmt"
	"log"
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
	if fields != nil {
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

// 只保留结构体中指定 fields 的 json tag 字段
func StructToJsonMap(obj any, fields []string) map[string]any {
	b, _ := json.Marshal(obj)
	var m map[string]any
	json.Unmarshal(b, &m)

	if fields == nil {
		return m
	}

	fieldsSet := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		fieldsSet[f] = struct{}{}
	}
	filtered := make(map[string]any)
	for k, v := range m {
		if _, ok := fieldsSet[k]; ok {
			filtered[k] = v
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

		fieldsIn := make(map[string]any)
		body, err := c.GetRawData()
		if err != nil {
			utils.Respond(c, nil, utils.ErrMissingParam)
			return
		}
		if len(body) > 0 {
			if err := json.Unmarshal(body, &fieldsIn); err != nil {
				utils.Respond(c, nil, utils.ErrMissingParam)
				return
			}
		}

		if updaterFn == nil {
			updaterFn = h.UpdateObject
		}
		if err := updaterFn(c, query, &result, fieldsIn); err != nil {
			utils.Respond(c, nil, utils.ErrBadRequest)
			return
		}

		h.DB.Save(&result)
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
		query.Count(&total)

		// 处理分页参数
		page := c.Query("page")
		pageSize := c.Query("page_size")
		log.Println(page, pageSize)
		pageNum, err := strconv.Atoi(page)
		if err != nil || pageNum < 1 {
			pageNum = 1
		}
		pageSizeNum, err := strconv.Atoi(pageSize)
		if err != nil || pageSizeNum < 1 {
			pageSizeNum = 10
		}
		offset := (pageNum - 1) * pageSizeNum
		log.Println("pageNum:", pageNum, "pageSizeNum:", pageSizeNum, "offset:", offset)
		query = query.Offset(offset).Limit(pageSizeNum)

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

		fmt.Println(c.Param("uuid"))

		// resultJson := make(map[string]any)

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

		fieldsIn := make(map[string]any)
		body, err := c.GetRawData()
		if err != nil {
			utils.Respond(c, nil, utils.ErrMissingParam)
			return
		}
		if len(body) > 0 {
			if err := json.Unmarshal(body, &fieldsIn); err != nil {
				utils.Respond(c, nil, utils.ErrMissingParam)
				return
			}
		}

		if updaterFn == nil {
			updaterFn = h.UpdateObject
		}
		if err := updaterFn(c, query, &result, fieldsIn); err != nil {
			utils.Respond(c, nil, utils.ErrBadRequest)
			return
		}

		query.Save(&result)
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
