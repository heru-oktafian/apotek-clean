package services

import (
	"errors"

	
	models "apotek-clean/internal/core/entities"

	"gorm.io/gorm"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock in source location")
	ErrInvalidLocation   = errors.New("invalid location: must be 'warehouse' or 'showcase'")
	ErrSameLocation      = errors.New("source and destination location cannot be the same")
	ErrInvalidQty        = errors.New("quantity must be greater than 0")
)

func CreateStockMutation(db *gorm.DB, productID, fromLoc, toLoc string, qty int, note, userID, branchID string) error {
	if qty <= 0 {
		return ErrInvalidQty
	}
	if fromLoc == toLoc {
		return ErrSameLocation
	}
	validLoc := map[string]bool{"warehouse": true, "showcase": true}
	if !validLoc[fromLoc] || !validLoc[toLoc] {
		return ErrInvalidLocation
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var product models.Product
		if err := tx.First(&product, "id = ? AND branch_id = ?", productID, branchID).Error; err != nil {
			return err
		}

		var sourceStock int
		if fromLoc == "warehouse" {
			sourceStock = product.WarehouseStock
		} else {
			sourceStock = product.ShowcaseStock
		}
		if sourceStock < qty {
			return ErrInsufficientStock
		}

		updates := map[string]interface{}{}
		if fromLoc == "warehouse" {
			updates["warehouse_stock"] = gorm.Expr("warehouse_stock - ?", qty)
		} else {
			updates["showcase_stock"] = gorm.Expr("showcase_stock - ?", qty)
		}
		if toLoc == "warehouse" {
			updates["warehouse_stock"] = gorm.Expr("COALESCE(warehouse_stock,0) + ?", qty)
		} else {
			updates["showcase_stock"] = gorm.Expr("COALESCE(showcase_stock,0) + ?", qty)
		}
		if err := tx.Model(&models.Product{}).Where("id = ?", productID).Updates(updates).Error; err != nil {
			return err
		}

		mutation := models.StockMutation{
			ProductID: productID,
			FromLoc:   fromLoc,
			ToLoc:     toLoc,
			Qty:       qty,
			Note:      note,
			UserID:    userID,
			BranchID:  branchID,
		}
		return tx.Create(&mutation).Error
	})
}

func GetStockMutations(db *gorm.DB, branchID string, page, limit int, productID string) ([]models.StockMutation, int64, error) {
	var mutations []models.StockMutation
	var total int64

	query := db.Model(&models.StockMutation{}).Where("branch_id = ?", branchID)
	if productID != "" {
		query = query.Where("product_id = ?", productID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&mutations).Error; err != nil {
		return nil, 0, err
	}

	return mutations, total, nil
}
