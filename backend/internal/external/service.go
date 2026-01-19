package external

import (
	"errors"
	"fmt"
	"time"
	"wallet-point/internal/audit"
	"wallet-point/internal/marketplace" // Add this
	"wallet-point/internal/mission"     // Add this
	"wallet-point/internal/wallet"

	"gorm.io/gorm"
)

type Service struct {
	repo               *Repository
	walletRepo         *wallet.WalletRepository
	walletService      *wallet.WalletService
	marketplaceService *marketplace.MarketplaceService // Add this
	missionService     *mission.MissionService         // Add this
	auditService       *audit.AuditService
	db                 *gorm.DB
}

func NewService(
	repo *Repository,
	walletRepo *wallet.WalletRepository,
	walletService *wallet.WalletService,
	marketplaceService *marketplace.MarketplaceService, // Add this
	missionService *mission.MissionService, // Add this
	auditService *audit.AuditService,
	db *gorm.DB,
) *Service {
	return &Service{
		repo:               repo,
		walletRepo:         walletRepo,
		walletService:      walletService,
		marketplaceService: marketplaceService, // Add this
		missionService:     missionService,     // Add this
		auditService:       auditService,
		db:                 db,
	}
}

func (s *Service) RegisterSource(req *ExternalSourceCreateRequest) (*ExternalSource, error) {
	source := &ExternalSource{
		SourceName:  req.SourceName,
		APIEndpoint: req.APIEndpoint,
		Status:      "active",
	}
	// Note: In a real system, you'd hash the API key if provided
	if err := s.repo.CreateSource(source); err != nil {
		return nil, err
	}
	return source, nil
}

func (s *Service) ListSources() ([]ExternalSource, error) {
	return s.repo.GetAllSources()
}

func (s *Service) SyncPoints(userID uint, req *SyncRequest) (*ExternalPointLog, error) {
	// 1. Validate Source
	source, err := s.repo.GetSourceByID(req.ExternalSourceID)
	if err != nil {
		return nil, errors.New("external source not found")
	}
	if source.Status != "active" {
		return nil, errors.New("external source is inactive")
	}

	// 2. Check Duplicate
	isDup, err := s.repo.CheckDuplicateTx(req.ExternalTxID)
	if err != nil {
		return nil, err
	}
	if isDup {
		return nil, errors.New("transaction already synced")
	}

	// 3. Get Wallet
	w, err := s.walletRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("wallet not found")
	}

	// 4. Create Log & Update Wallet in Transaction
	log := &ExternalPointLog{
		WalletID:              w.ID,
		ExternalSourceID:      source.ID,
		ExternalTransactionID: req.ExternalTxID,
		Amount:                req.Amount,
		Metadata:              req.Metadata,
		Status:                "success",
		SyncedAt:              time.Now(),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create Log
		if err := s.repo.CreateLogWithTx(tx, log); err != nil {
			return err
		}

		// Update Wallet Points
		desc := fmt.Sprintf("External Sync from %s: %s", source.SourceName, req.ExternalTxID)
		if err := s.walletService.CreditWithTransaction(tx, w.ID, req.Amount, "external", desc); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Status = "failed"
		return log, err
	}

	return log, nil
}

func (s *Service) CreateProductFromExternal(sourceID uint, req *marketplace.CreateProductRequest) (*marketplace.Product, error) {
	_, err := s.repo.GetSourceByID(sourceID)
	if err != nil {
		return nil, errors.New("external source not found")
	}

	// For bookkeeping, we use the source ID or a system admin ID
	// Since products table needs a CreatorID (Uint), we'll use a placeholder or the first admin
	product, err := s.marketplaceService.CreateProduct(req, 1) // Using 1 as default admin/system ID
	if err != nil {
		return nil, err
	}

	// Optionally log that this was created externally
	return product, nil
}

func (s *Service) CreateMissionFromExternal(sourceID uint, req *mission.CreateMissionRequest) (*mission.Mission, error) {
	_, err := s.repo.GetSourceByID(sourceID)
	if err != nil {
		return nil, errors.New("external source not found")
	}

	mission, err := s.missionService.CreateMission(req, 1) // Using 1 as default admin/system ID
	if err != nil {
		return nil, err
	}

	return mission, nil
}
