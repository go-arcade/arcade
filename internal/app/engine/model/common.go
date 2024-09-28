package model

import (
	"github.com/go-arcade/arcade/pkg/ctx"
	"gorm.io/gorm"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 20:42
 * @file: common.go
 * @description:
 */

func Count(ctx ctx.Context, tx *gorm.DB) (int64, error) {
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func Exist(ctx ctx.Context, tx *gorm.DB, where interface{}) bool {
	num, err := Count(ctx, tx)
	if err != nil {
		return false
	}
	return num > 0
}

func Insert(ctx ctx.Context, obj interface{}) error {
	return ctx.MySQLIns.Create(obj).Error
}
