package transfer

import (
	"errors"
	"fmt"
	"wallet-point/internal/wallet"

	"gorm.io/gorm"
)

type Service struct {
	repo          *Repository
	walletRepo    *wallet.WalletRepository
	walletService *wallet.WalletService
	db            *gorm.DB
}

func NewService(repo *Repository, walletRepo *wallet.WalletRepository, walletService *wallet.WalletService, db *gorm.DB) *Service {
	return &Service{
		repo:          repo,
		walletRepo:    walletRepo,
		walletService: walletService,
		db:            db,
	}
}

func (s *Service) CreateTransfer(senderUserID, receiverUserID uint, amount int, description string) (*Transfer, error) {
	if senderUserID == receiverUserID {
		return nil, errors.New("cannot transfer points to yourself")
	}

	senderWallet, err := s.walletService.GetWalletByUserID(senderUserID)
	if err != nil {
		return nil, errors.New("sender wallet not found")
	}

	receiverWallet, err := s.walletService.GetWalletByUserID(receiverUserID)
	if err != nil {
		return nil, errors.New("receiver wallet not found: check if user exists and has a wallet")
	}

	if senderWallet.Balance < amount {
		return nil, errors.New("insufficient balance")
	}

	transfer := &Transfer{
		SenderWalletID:   senderWallet.ID,
		ReceiverWalletID: receiverWallet.ID,
		Amount:           amount,
		Description:      description,
		Status:           "success",
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Deduct from sender
		if err := s.walletService.DebitWithTransaction(tx, senderWallet.ID, amount, "transfer_out", fmt.Sprintf("Transfer to user %d", receiverUserID)); err != nil {
			return err
		}

		// 2. Credit to receiver
		if err := s.walletService.CreditWithTransaction(tx, receiverWallet.ID, amount, "transfer_in", fmt.Sprintf("Transfer from user %d", senderUserID)); err != nil {
			return err
		}

		// 3. Create transfer record
		return s.repo.CreateWithTransaction(tx, transfer)
	})

	if err != nil {
		return nil, err
	}

	return transfer, nil
}

func (s *Service) GetUserTransfers(userID uint, limit, offset int) ([]Transfer, int64, error) {
	wallet, err := s.walletService.GetWalletByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	transfers, total, err := s.repo.FindByWallet(wallet.ID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if err := s.populateTransferDetails(transfers); err != nil {
		return nil, 0, err
	}
	return transfers, total, nil
}

func (s *Service) GetSentTransfers(userID uint, limit, offset int) ([]Transfer, int64, error) {
	wallet, err := s.walletService.GetWalletByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	transfers, total, err := s.repo.FindBySenderWallet(wallet.ID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if err := s.populateTransferDetails(transfers); err != nil {
		return nil, 0, err
	}
	return transfers, total, nil
}

func (s *Service) GetReceivedTransfers(userID uint, limit, offset int) ([]Transfer, int64, error) {
	wallet, err := s.walletService.GetWalletByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	transfers, total, err := s.repo.FindByReceiverWallet(wallet.ID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if err := s.populateTransferDetails(transfers); err != nil {
		return nil, 0, err
	}
	return transfers, total, nil
}

func (s *Service) GetAllTransfers(limit, offset int) ([]Transfer, int64, error) {
	transfers, total, err := s.repo.FindAll(limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if err := s.populateTransferDetails(transfers); err != nil {
		return nil, 0, err
	}
	return transfers, total, nil
}

// populateTransferDetails fills in the sender and receiver names
func (s *Service) populateTransferDetails(transfers []Transfer) error {
	if len(transfers) == 0 {
		return nil
	}

	// Collect wallet IDs
	walletIDs := make(map[uint]bool)
	for _, t := range transfers {
		walletIDs[t.SenderWalletID] = true
		walletIDs[t.ReceiverWalletID] = true
	}

	ids := make([]uint, 0, len(walletIDs))
	for id := range walletIDs {
		ids = append(ids, id)
	}

	// Fetch wallets with user info
	type WalletUser struct {
		WalletID uint
		FullName string
		NimNip   string
	}

	var walletUsers []WalletUser
	err := s.db.Table("wallets").
		Select("wallets.id as wallet_id, users.full_name, users.nim_nip").
		Joins("JOIN users ON users.id = wallets.user_id").
		Where("wallets.id IN ?", ids).
		Scan(&walletUsers).Error

	if err != nil {
		return err
	}

	// Map findings
	infoMap := make(map[uint]WalletUser)
	for _, wu := range walletUsers {
		infoMap[wu.WalletID] = wu
	}

	// Assign back to transfers
	for i := range transfers {
		if sender, ok := infoMap[transfers[i].SenderWalletID]; ok {
			transfers[i].SenderName = sender.FullName
			transfers[i].SenderNIM = sender.NimNip
		}
		if receiver, ok := infoMap[transfers[i].ReceiverWalletID]; ok {
			transfers[i].ReceiverName = receiver.FullName
			transfers[i].ReceiverNIM = receiver.NimNip
		}
	}

	return nil
}

func (s *Service) FindRecipient(userID uint) (*RecipientSummary, error) {
	// Check if user has a wallet
	w, err := s.walletService.GetWalletByUserID(userID)
	if err != nil {
		return nil, errors.New("user not found or has no wallet")
	}

	var recipient RecipientSummary
	err = s.db.Table("users").
		Select("id, full_name, role, nim_nip as nim").
		Where("id = ?", w.UserID).
		Scan(&recipient).Error

	if err != nil {
		return nil, err
	}
	if recipient.ID == 0 {
		return nil, errors.New("user not found")
	}

	return &recipient, nil
}
