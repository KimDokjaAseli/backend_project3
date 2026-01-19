package database

import (
	"log"
	"wallet-point/internal/audit"
	"wallet-point/internal/auth"
	"wallet-point/internal/marketplace"
	"wallet-point/internal/mission"
	"wallet-point/internal/transfer"
	"wallet-point/internal/wallet"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	log.Println("🔄 Running database migrations...")

	// Manual Fix: Ensure enum types are updated (GORM AutoMigrate doesn't update existing enums)
	db.Exec("ALTER TABLE users MODIFY COLUMN role ENUM('admin', 'dosen', 'mahasiswa', 'merchant') NOT NULL")
	db.Exec("ALTER TABLE missions MODIFY COLUMN type ENUM('quiz', 'task', 'assignment') NOT NULL")
	db.Exec("ALTER TABLE mission_submissions MODIFY COLUMN status ENUM('pending', 'approved', 'rejected') NOT NULL DEFAULT 'pending'")
	db.Exec("ALTER TABLE wallet_transactions MODIFY COLUMN type ENUM('mission', 'task', 'transfer_in', 'transfer_out', 'marketplace', 'marketplace_sale', 'external', 'adjustment', 'topup') NOT NULL")

	// Cleanup: Remove legacy tables
	db.Exec("DROP TABLE IF EXISTS task_submissions")
	db.Exec("DROP TABLE IF EXISTS tasks")

	err := db.AutoMigrate(
		&auth.User{},
		&wallet.Wallet{},
		&wallet.WalletTransaction{},
		&wallet.PaymentToken{},
		&transfer.Transfer{},
		&marketplace.Product{},
		&marketplace.MarketplaceTransaction{},
		&audit.AuditLog{},
		&mission.Mission{},
		&mission.MissionQuestion{},
		&mission.MissionSubmission{},
	)

	if err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	log.Println("✅ Database migration completed")
}
