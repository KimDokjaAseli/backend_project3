package transfer

import (
	"gorm.io/gorm"
)

// Repository handles database operations for transfers
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new transfer repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new transfer record
func (r *Repository) Create(transfer *Transfer) error {
	return r.db.Create(transfer).Error
}

// FindByID retrieves a transfer by ID
func (r *Repository) FindByID(id uint) (*Transfer, error) {
	var transfer Transfer
	err := r.db.First(&transfer, id).Error
	return &transfer, err
}

// FindBySenderWallet retrieves all transfers sent by a wallet
func (r *Repository) FindBySenderWallet(walletID uint, limit, offset int) ([]Transfer, int64, error) {
	var transfers []Transfer
	var total int64

	query := r.db.Model(&Transfer{}).Where("sender_wallet_id = ?", walletID)
	query.Count(&total)

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&transfers).Error
	return transfers, total, err
}

// FindByReceiverWallet retrieves all transfers received by a wallet
func (r *Repository) FindByReceiverWallet(walletID uint, limit, offset int) ([]Transfer, int64, error) {
	var transfers []Transfer
	var total int64

	query := r.db.Model(&Transfer{}).Where("receiver_wallet_id = ?", walletID)
	query.Count(&total)

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&transfers).Error
	return transfers, total, err
}

// FindByWallet retrieves all transfers (sent or received) by a wallet
func (r *Repository) FindByWallet(walletID uint, limit, offset int) ([]Transfer, int64, error) {
	var transfers []Transfer
	var total int64

	query := r.db.Model(&Transfer{}).
		Where("sender_wallet_id = ? OR receiver_wallet_id = ?", walletID, walletID)
	query.Count(&total)

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&transfers).Error
	return transfers, total, err
}

// FindAll retrieves all transfers with pagination
func (r *Repository) FindAll(limit, offset int) ([]Transfer, int64, error) {
	var transfers []Transfer
	var total int64

	r.db.Model(&Transfer{}).Count(&total)
	err := r.db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&transfers).Error
	return transfers, total, err
}

// CreateWithTransaction creates a transfer within a database transaction
func (r *Repository) CreateWithTransaction(tx *gorm.DB, transfer *Transfer) error {
	return tx.Create(transfer).Error
}
