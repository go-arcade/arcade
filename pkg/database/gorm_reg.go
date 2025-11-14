package database

import "gorm.io/gorm"

var registerysModels []any

// RegisterModels registers the given models for Gorm.
func RegisterModels(models ...any) {
	registerysModels = append(registerysModels, models...)
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(registerysModels...)
}

// GetRegisteredModels returns the registered models for Gorm.
func GetRegisteredModels() []any {
	return registerysModels
}
