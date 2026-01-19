package external

import (
	"time"
)

type ExternalSource struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	SourceName  string    `json:"source_name" gorm:"unique;not null"`
	APIEndpoint string    `json:"api_endpoint" gorm:"not null"`
	APIKeyHash  string    `json:"-"`
	Status      string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ExternalSource) TableName() string {
	return "external_sources"
}

type ExternalPointLog struct {
	ID                    uint      `json:"id" gorm:"primaryKey"`
	WalletID              uint      `json:"wallet_id" gorm:"not null"`
	ExternalSourceID      uint      `json:"external_source_id" gorm:"not null"`
	ExternalTransactionID string    `json:"external_transaction_id" gorm:"unique;not null"`
	Amount                int       `json:"amount" gorm:"not null"`
	Metadata              string    `json:"metadata"`
	Status                string    `json:"status" gorm:"type:enum('success','failed','pending');default:'success'"`
	SyncedAt              time.Time `json:"synced_at"`
	CreatedAt             time.Time `json:"created_at"`
}

func (ExternalPointLog) TableName() string {
	return "external_point_logs"
}

type SyncRequest struct {
	ExternalSourceID uint   `json:"external_source_id" binding:"required"`
	ExternalTxID     string `json:"external_tx_id" binding:"required"`
	Amount           int    `json:"amount" binding:"required,gt=0"`
	Metadata         string `json:"metadata"`
}

type ExternalSourceCreateRequest struct {
	SourceName  string `json:"source_name" binding:"required"`
	APIEndpoint string `json:"api_endpoint" binding:"required"`
	APIKey      string `json:"api_key"`
}
