package model

import (
	"time"
)

// NotificationTemplate represents a notification template in the database
type NotificationTemplate struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	TemplateID  string    `gorm:"uniqueIndex;type:varchar(100);not null" json:"template_id"` // Unique template identifier
	Name        string    `gorm:"type:varchar(200);not null" json:"name"`                    // Template name
	Type        string    `gorm:"type:varchar(50);not null;index" json:"type"`               // Template type (build/approval)
	Channel     string    `gorm:"type:varchar(50);not null;index" json:"channel"`            // Target channel
	Title       string    `gorm:"type:varchar(200)" json:"title"`                            // Template title
	Content     string    `gorm:"type:text;not null" json:"content"`                         // Template content
	Variables   string    `gorm:"type:text" json:"variables"`                                // Required variables (JSON array)
	Format      string    `gorm:"type:varchar(50);default:markdown" json:"format"`           // Message format
	Metadata    string    `gorm:"type:text" json:"metadata"`                                 // Additional metadata (JSON)
	Description string    `gorm:"type:text" json:"description"`                              // Template description
	IsActive    bool      `gorm:"default:true;index" json:"is_active"`                       // Active status
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for NotificationTemplate
func (NotificationTemplate) TableName() string {
	return "notification_templates"
}

// NotificationLog represents a notification sending record
type NotificationLog struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	TemplateID string    `gorm:"type:varchar(100);index" json:"template_id"`     // Template ID used
	Channel    string    `gorm:"type:varchar(50);not null;index" json:"channel"` // Channel name
	Recipient  string    `gorm:"type:varchar(500)" json:"recipient"`             // Recipient info
	Content    string    `gorm:"type:text" json:"content"`                       // Rendered content
	Status     string    `gorm:"type:varchar(50);not null;index" json:"status"`  // Status: success/failed
	ErrorMsg   string    `gorm:"type:text" json:"error_msg"`                     // Error message if failed
	Metadata   string    `gorm:"type:text" json:"metadata"`                      // Additional metadata (JSON)
	SentAt     time.Time `gorm:"index" json:"sent_at"`                           // Sent timestamp
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name for NotificationLog
func (NotificationLog) TableName() string {
	return "notification_logs"
}
