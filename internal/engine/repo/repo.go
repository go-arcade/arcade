package repo

import (
	"github.com/go-arcade/arcade/pkg/ctx"
	"gorm.io/gorm"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/28 21:58
 * @file: repo.go
 * @description:
 */

func Count(tx *gorm.DB) (int64, error) {
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func Exist(tx *gorm.DB, where interface{}) bool {
	var one interface{}
	if err := tx.Where(where).First(&one).Error; err != nil {
		return false
	}
	return true
}

func Insert(ctx ctx.Context, obj interface{}) error {
	return ctx.DB.Create(obj).Error
}
