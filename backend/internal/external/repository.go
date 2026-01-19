package external

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateSource(source *ExternalSource) error {
	return r.db.Create(source).Error
}

func (r *Repository) GetAllSources() ([]ExternalSource, error) {
	var sources []ExternalSource
	err := r.db.Find(&sources).Error
	return sources, err
}

func (r *Repository) GetSourceByID(id uint) (*ExternalSource, error) {
	var source ExternalSource
	err := r.db.First(&source, id).Error
	return &source, err
}

func (r *Repository) CreateLog(log *ExternalPointLog) error {
	return r.db.Create(log).Error
}

func (r *Repository) CreateLogWithTx(tx *gorm.DB, log *ExternalPointLog) error {
	return tx.Create(log).Error
}

func (r *Repository) GetLogsByWalletID(walletID uint) ([]ExternalPointLog, error) {
	var logs []ExternalPointLog
	err := r.db.Where("wallet_id = ?", walletID).Find(&logs).Error
	return logs, err
}

func (r *Repository) CheckDuplicateTx(externalTxID string) (bool, error) {
	var count int64
	err := r.db.Model(&ExternalPointLog{}).Where("external_transaction_id = ?", externalTxID).Count(&count).Error
	return count > 0, err
}
