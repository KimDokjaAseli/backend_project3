package marketplace

import (
	"errors"

	"gorm.io/gorm"
)

type MarketplaceRepository struct {
	db *gorm.DB
}

func NewMarketplaceRepository(db *gorm.DB) *MarketplaceRepository {
	return &MarketplaceRepository{db: db}
}

// GetAll gets all products with filters and pagination
func (r *MarketplaceRepository) GetAll(params ProductListParams) ([]Product, int64, error) {
	var products []Product
	var total int64

	query := r.db.Model(&Product{})

	// Apply filters
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (params.Page - 1) * params.Limit
	query = query.Limit(params.Limit).Offset(offset).Order("created_at DESC")

	if err := query.Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// FindByID finds product by ID
func (r *MarketplaceRepository) FindByID(productID uint) (*Product, error) {
	var product Product
	err := r.db.First(&product, productID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}
	return &product, nil
}

// Create creates a new product
func (r *MarketplaceRepository) Create(product *Product) error {
	return r.db.Create(product).Error
}

// Update updates product
func (r *MarketplaceRepository) Update(productID uint, updates map[string]interface{}) error {
	return r.db.Model(&Product{}).Where("id = ?", productID).Updates(updates).Error
}

// Delete deletes product (soft delete by setting status to inactive)
func (r *MarketplaceRepository) Delete(productID uint) error {
	return r.db.Model(&Product{}).Where("id = ?", productID).Update("status", "inactive").Error
}

// UpdateStock updates product stock
func (r *MarketplaceRepository) UpdateStock(tx *gorm.DB, productID uint, delta int) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Model(&Product{}).
		Where("id = ?", productID).
		Update("stock", gorm.Expr("stock + ?", delta)).
		Error
}

func (r *MarketplaceRepository) AddToCart(item *CartItem) error {
	var existing CartItem
	err := r.db.Where("user_id = ? AND product_id = ?", item.UserID, item.ProductID).First(&existing).Error
	if err == nil {
		return r.db.Model(&existing).Update("quantity", existing.Quantity+item.Quantity).Error
	}
	return r.db.Create(item).Error
}

func (r *MarketplaceRepository) GetCart(userID uint) ([]CartItem, error) {
	var items []CartItem
	err := r.db.Preload("Product").Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

func (r *MarketplaceRepository) UpdateCartItem(userID, itemID uint, quantity int) error {
	return r.db.Model(&CartItem{}).Where("user_id = ? AND id = ?", userID, itemID).Update("quantity", quantity).Error
}

func (r *MarketplaceRepository) RemoveFromCart(userID, itemID uint) error {
	return r.db.Where("user_id = ? AND id = ?", userID, itemID).Delete(&CartItem{}).Error
}

func (r *MarketplaceRepository) ClearCart(tx *gorm.DB, userID uint) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Where("user_id = ?", userID).Delete(&CartItem{}).Error
}
