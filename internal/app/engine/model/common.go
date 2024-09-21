package model

import (
	"github.com/go-arcade/arcade/pkg/orm"
	"gorm.io/gorm"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 20:42
 * @file: common.go
 * @description:
 */

func DB() *gorm.DB {
	return orm.GetConn()
}

func Count(tx *gorm.DB) (int64, error) {
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func Exist(tx *gorm.DB, where interface{}) bool {
	num, err := Count(tx)
	if err != nil {
		return false
	}
	return num > 0
}

func Insert(obj interface{}) error {
	return DB().Create(obj).Error
}
