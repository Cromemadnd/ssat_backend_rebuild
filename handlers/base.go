package handlers

import (
	"encoding/json"
	"fmt"
	"ssat_backend_rebuild/utils"

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

func (h *BaseHandler[T]) UpdateObject(query *gorm.DB, object *T, data map[string]any) error {
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
	updaterFn func(query *gorm.DB, object *T, data map[string]any) error,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		query := h.Select(fields)

		result, err := h.CreateObject(query)
		if err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}

		fieldsIn := make(map[string]any)
		if err := c.ShouldBindJSON(&fieldsIn); err != nil {
			utils.Respond(c, nil, utils.ErrMissingParam)
			return
		}

		if updaterFn == nil {
			updaterFn = h.UpdateObject
		}
		if err := updaterFn(query, &result, fieldsIn); err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}

		query.Save(&result)
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

		results, err := h.GetObjects(query)
		if err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
			return
		}

		resultJson := make([]map[string]any, 0, len(results))
		for _, result := range results {
			resultJson = append(resultJson, StructToJsonMap(result, fields))
		}
		utils.Respond(c, resultJson, utils.ErrOK)
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
	updaterFn func(query *gorm.DB, object *T, data map[string]any) error,
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
		if err := c.ShouldBindJSON(&fieldsIn); err != nil {
			utils.Respond(c, nil, utils.ErrMissingParam)
			return
		}

		if updaterFn == nil {
			updaterFn = h.UpdateObject
		}
		if err := updaterFn(query, &result, fieldsIn); err != nil {
			utils.Respond(c, nil, utils.ErrInternalServer)
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

// func DefaultFilterFn(c *gin.Context) (conditions []any, err error) {
// 	return nil, nil
// }

// func DefaultGetterFn(c *gin.Context) (conditions []any, err error) {
// 	return []any{"uuid = ?", c.Param("uuid")}, nil
// }

// func DefaultUpdaterFn[T any](c *gin.Context, data *any, model *T) error {
// 	deviceValue := reflect.ValueOf(model).Elem() // 获取 Device 的值
// 	dataValue := reflect.ValueOf(data)           // 获取 A 的值
// 	dataType := reflect.TypeOf(data)             // 获取 A 的类型

// 	for i := 0; i < dataType.NumField(); i++ {
// 		field := dataType.Field(i) // 获取 A 的每个字段
// 		if deviceField := deviceValue.FieldByName(field.Name); deviceField.IsValid() && deviceField.CanSet() {
// 			// 如果 Device 中存在同名字段，并且可以设置值
// 			deviceField.Set(dataValue.Field(i))
// 		}
// 	}
// 	return nil
// }

// func (h *BaseHandler[T]) List(
// 	fieldsOut []string,
// 	filterFn func(c *gin.Context) (conditions []any, err error),
// ) func(c *gin.Context) {
// 	return func(c *gin.Context) {
// 		// 处理查询参数
// 		query := h.DB
// 		if fieldsOut != nil {
// 			query = query.Select(fieldsOut)
// 		}

// 		// 调用过滤函数获取条件，查询模型实例
// 		if filterFn == nil {
// 			filterFn = DefaultFilterFn
// 		}
// 		conditions, err := filterFn(c)
// 		if err != nil {
// 			utils.Respond(c, nil, utils.ErrInternalServer)
// 			return
// 		}
// 		var results []map[string]any
// 		if result := query.Model(new(T)).Find(&results, conditions...); result.Error != nil {
// 			utils.Respond(c, nil, utils.ErrInternalServer)
// 			return
// 		}

// 		utils.Respond(c, results, utils.ErrOK)
// 	}
// }

// func (h *BaseHandler[T]) Retrieve(
// 	fieldsOut []string,
// 	getterFn func(c *gin.Context) (conditions []any, err error),
// ) func(c *gin.Context) {
// 	return func(c *gin.Context) {
// 		// 处理查询参数
// 		query := h.DB
// 		if fieldsOut != nil {
// 			query = query.Select(fieldsOut)
// 		}

// 		// 调用过滤函数获取条件，查询模型实例
// 		conditions, err := getterFn(c)
// 		if err != nil {
// 			utils.Respond(c, nil, utils.ErrInternalServer)
// 			return
// 		}
// 		var result map[string]any
// 		if result := query.Model(new(T)).First(&result, conditions...); result.Error != nil {
// 			utils.Respond(c, nil, utils.ErrNotFound)
// 			return
// 		}

// 		utils.Respond(c, result, utils.ErrOK)
// 	}
// }

// func (h *BaseHandler[T]) Create(
// 	fieldsIn any,
// 	updaterFn func(c *gin.Context, data *any, model *T) error,
// ) func(c *gin.Context) {
// 	return func(c *gin.Context) {
// 		// 处理传入参数
// 		query := h.DB
// 		if fieldsIn == nil {
// 			fieldsIn = new(T)
// 		}
// 		if err := c.ShouldBindJSON(&fieldsIn); err != nil {
// 			utils.Respond(c, nil, utils.ErrMissingParam)
// 			return
// 		}

// 		var model T

// 		// 调用更新函数
// 		if updaterFn == nil {
// 			updaterFn = DefaultUpdaterFn[T]
// 		}
// 		if err := updaterFn(c, &fieldsIn, &model); err != nil {
// 			utils.Respond(c, nil, utils.ErrInternalServer)
// 			return
// 		}

// 		// 创建新模型实例
// 		if result := query.Create(&model); result.Error != nil {
// 			utils.Respond(c, nil, utils.ErrInternalServer)
// 			return
// 		}
// 		utils.Respond(c, model, utils.ErrCreated)
// 	}
// }

// func (h *BaseHandler[T]) Update(
// 	fieldsIn map[string]any,
// 	fieldsOut []string,
// 	getterFn func(c *gin.Context) (conditions []any, err error),
// 	updaterFn func(c *gin.Context, data *map[string]any, model *T) error,
// ) func(c *gin.Context) {
// 	return func(c *gin.Context) {
// 		// 处理传入参数
// 		var model T
// 		query := h.DB
// 		if fieldsOut != nil {
// 			query = query.Select(fieldsOut)
// 		}
// 		if fieldsIn == nil {
// 			if err := c.ShouldBindJSON(&model); err != nil {
// 				utils.Respond(c, nil, utils.ErrMissingParam)
// 				return
// 			}
// 		} else {
// 			if err := c.ShouldBindJSON(&fieldsIn); err != nil {
// 				utils.Respond(c, nil, utils.ErrMissingParam)
// 				return
// 			}
// 		}

// 		// 调用过滤函数获取条件，查询模型实例
// 		if getterFn == nil {
// 			getterFn = DefaultGetterFn
// 		}
// 		conditions, err := getterFn(c)
// 		if err != nil {
// 			utils.Respond(c, nil, utils.ErrInternalServer)
// 			return
// 		}
// 		if result := query.First(&model, conditions...); result.Error != nil {
// 			utils.Respond(c, nil, utils.ErrNotFound)
// 			return
// 		}

// 		// 更新模型实例
// 		if updaterFn == nil {
// 			updaterFn = DefaultUpdaterFn[T]
// 		}
// 		if err := updaterFn(c, &fieldsIn, &model); err != nil {
// 			utils.Respond(c, nil, utils.ErrMissingParam)
// 			return
// 		}

// 		// 保存模型实例
// 		query.Save(&model)
// 		utils.Respond(c, model, utils.ErrOK)
// 	}
// }

// func (h *BaseHandler[T]) Destroy(
// 	getterFn func(c *gin.Context) (conditions []any, err error),
// ) func(c *gin.Context) {
// 	return func(c *gin.Context) {
// 		query := h.DB

// 		// 调用过滤函数获取条件，查询模型实例
// 		var model T
// 		if getterFn == nil {
// 			getterFn = DefaultGetterFn
// 		}
// 		conditions, err := getterFn(c)
// 		if err != nil {
// 			utils.Respond(c, nil, utils.ErrInternalServer)
// 			return
// 		}
// 		if result := query.First(&model, conditions...); result.Error != nil {
// 			utils.Respond(c, nil, utils.ErrNotFound)
// 			return
// 		}

// 		// 删除模型实例
// 		if result := query.Delete(&model); result.Error != nil {
// 			utils.Respond(c, nil, utils.ErrInternalServer)
// 			return
// 		}

// 		utils.Respond(c, gin.H{"message": "删除成功"}, utils.ErrOK)
// 	}
// }
