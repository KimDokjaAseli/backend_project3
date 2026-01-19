package marketplace

import (
	"errors"
	"fmt"
	"math"
	"wallet-point/internal/wallet"

	"gorm.io/gorm"
)

type MarketplaceService struct {
	repo          *MarketplaceRepository
	walletService *wallet.WalletService
	db            *gorm.DB
}

func NewMarketplaceService(repo *MarketplaceRepository, walletService *wallet.WalletService, db *gorm.DB) *MarketplaceService {
	return &MarketplaceService{
		repo:          repo,
		walletService: walletService,
		db:            db,
	}
}

// GetAllProducts gets all products with pagination and filters
func (s *MarketplaceService) GetAllProducts(params ProductListParams) (*ProductListResponse, error) {
	// Default pagination
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 20
	}

	products, total, err := s.repo.GetAll(params)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	return &ProductListResponse{
		Products:   products,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetProductByID gets product by ID
func (s *MarketplaceService) GetProductByID(productID uint) (*Product, error) {
	return s.repo.FindByID(productID)
}

// CreateProduct creates a new product
func (s *MarketplaceService) CreateProduct(req *CreateProductRequest, adminID uint) (*Product, error) {
	product := &Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
		Status:      "active",
		CreatedBy:   adminID,
	}

	if err := s.repo.Create(product); err != nil {
		return nil, errors.New("failed to create product")
	}

	return product, nil
}

// UpdateProduct updates product
func (s *MarketplaceService) UpdateProduct(productID uint, req *UpdateProductRequest) (*Product, error) {
	// Check if product exists
	_, err := s.repo.FindByID(productID)
	if err != nil {
		return nil, err
	}

	// Prepare updates
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Price > 0 {
		updates["price"] = req.Price
	}
	if req.Stock >= 0 {
		updates["stock"] = req.Stock
	}
	if req.ImageURL != "" {
		updates["image_url"] = req.ImageURL
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	// Update product
	if len(updates) > 0 {
		if err := s.repo.Update(productID, updates); err != nil {
			return nil, errors.New("failed to update product")
		}
	}

	// Return updated product
	return s.repo.FindByID(productID)
}

// DeleteProduct deletes product
func (s *MarketplaceService) DeleteProduct(productID uint) error {
	// Check if product exists
	_, err := s.repo.FindByID(productID)
	if err != nil {
		return err
	}

	return s.repo.Delete(productID)
}

// PurchaseProduct handles product purchase
func (s *MarketplaceService) PurchaseProduct(userID uint, req *PurchaseRequest) (*MarketplaceTransaction, error) {
	// Start transaction
	tx := s.db.Begin()
	var err error

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 1. Get Product
	var product *Product
	product, err = s.repo.FindByID(req.ProductID)
	if err != nil {
		return nil, err
	}

	if product.Status == "inactive" {
		err = errors.New("product is not active")
		return nil, err
	}

	if product.Stock < 1 {
		err = errors.New("product out of stock")
		return nil, err
	}

	// 2. Validate QR if needed (Legacy / External QR Token)
	if req.PaymentMethod == "qr" && req.PaymentToken != "" {
		if err = s.walletService.ValidateAndConsumeToken(req.PaymentToken, userID, product.Price); err != nil {
			return nil, err
		}
	}

	// 3. Get Wallets
	var studentWallet *wallet.Wallet
	studentWallet, err = s.walletService.GetWalletByUserID(userID)
	if err != nil {
		return nil, err
	}

	var creatorWallet *wallet.Wallet
	creatorWallet, err = s.walletService.GetWalletByUserID(product.CreatedBy)
	if err != nil {
		// Fallback to a system admin if creator wallet not found
		err = s.db.Table("users").Where("role = ?", "admin").Select("id").First(&product.CreatedBy).Error
		if err == nil {
			creatorWallet, _ = s.walletService.GetWalletByUserID(product.CreatedBy)
		}
	}

	// Default quantity to 1 if not provided
	quantity := req.Quantity
	if quantity <= 0 {
		quantity = 1
	}

	totalPrice := product.Price * quantity

	// 4. Check Balance (Only if not already paid via external QR token)
	if req.PaymentToken == "" {
		if studentWallet.Balance < totalPrice {
			err = fmt.Errorf("insufficient balance. Required: %d", totalPrice)
			return nil, err
		}

		// 5. Debit Student Wallet
		err = s.walletService.DebitWithTransaction(tx, studentWallet.ID, totalPrice, "marketplace", fmt.Sprintf("Buy %dx %s", quantity, product.Name))
		if err != nil {
			return nil, err
		}

		// 6. Credit Creator Wallet (Admin/Merchant)
		if creatorWallet != nil {
			err = s.walletService.CreditWithTransaction(tx, creatorWallet.ID, totalPrice, "marketplace_sale", fmt.Sprintf("Sale %dx %s to %s", quantity, product.Name, req.StudentName))
			if err != nil {
				return nil, err
			}
		}
	}

	// 7. Reduce Stock
	err = s.repo.UpdateStock(tx, product.ID, -quantity)
	if err != nil {
		return nil, err
	}

	// 8. Create Transaction Record with Student Data
	txn := &MarketplaceTransaction{
		WalletID:      studentWallet.ID,
		ProductID:     product.ID,
		Amount:        product.Price,
		TotalAmount:   totalPrice,
		Quantity:      quantity,
		StudentName:   req.StudentName,
		StudentNPM:    req.StudentNPM,
		StudentMajor:  req.StudentMajor,
		StudentBatch:  req.StudentBatch,
		PaymentMethod: req.PaymentMethod,
		Status:        "success",
	}

	err = s.repo.CreateTransaction(tx, txn)
	if err != nil {
		return nil, err
	}

	return txn, nil
}

// GetTransactions retrieves all marketplace transactions (Admin)
func (s *MarketplaceService) GetTransactions(limit, offset int) ([]MarketplaceTransactionWithDetails, int64, error) {
	return s.repo.GetAllTransactions(limit, offset)
}
