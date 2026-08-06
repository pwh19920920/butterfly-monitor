package persistence

import (
	"gorm.io/gorm"
)

// paginate 基于同一条查询链完成 count + list 分页查询。
// model 指定目标实体，whereSql/whereArgs/notCase 为公共过滤条件：
// 先 Count 取总数，再带 Order/Limit/Offset 取当前页数据，
// 消除分页查询中两段重复构建条件的样板代码（且不丢失 Count 的错误）。
func paginate[T any](db *gorm.DB, model any, whereSql string, whereArgs []interface{}, notCase any, pageSize, offset int, orderBy string) (int64, []T, error) {
	var count int64
	if err := db.Model(model).
		Where(whereSql, whereArgs...).
		Not(notCase).
		Count(&count).Error; err != nil {
		return 0, nil, err
	}

	var data []T
	err := db.Model(model).
		Where(whereSql, whereArgs...).
		Not(notCase).
		Order(orderBy).
		Limit(pageSize).Offset(offset).
		Find(&data).Error
	return count, data, err
}
