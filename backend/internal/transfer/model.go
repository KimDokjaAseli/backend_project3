package transfer

import (
	"time"
)

// Transfer represents a point transfer between two wallets
type Transfer struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	SenderWalletID   uint      `json:"sender_wallet_id" gorm:"not null;index"`
	ReceiverWalletID uint      `json:"receiver_wallet_id" gorm:"not null;index"`
	Amount           int       `json:"amount" gorm:"not null"`
	Description      string    `json:"description" gorm:"type:varchar(255)"`
	Status           string    `json:"status" gorm:"type:enum('success','failed');default:'success'"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Virtual fields for response
	SenderName   string `json:"sender_name,omitempty" gorm:"-"`
	ReceiverName string `json:"receiver_name,omitempty" gorm:"-"`
	SenderNIM    string `json:"sender_nim,omitempty" gorm:"-"`
	ReceiverNIM  string `json:"receiver_nim,omitempty" gorm:"-"`
}

// TableName specifies the table name for Transfer model
func (Transfer) TableName() string {
	return "transfers"
}

// TransferRequest represents the request body for creating a transfer
type TransferRequest struct {
	ReceiverUserID uint   `json:"receiver_user_id" binding:"required"`
	Amount         int    `json:"amount" binding:"required,gt=0"`
	Description    string `json:"description" binding:"max=255"`
}

// TransferResponse represents the response for transfer operations
type TransferResponse struct {
	Transfer *Transfer `json:"transfer"`
	Message  string    `json:"message"`
}

// RecipientSummary represents simplified user info for transfer verification
type RecipientSummary struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	NIM      string `json:"nim,omitempty"`
}
