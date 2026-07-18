package services

import (
	"time"

	models "apotek-clean/internal/core/entities"

	"gorm.io/gorm"
)

func LogStockTrace(db *gorm.DB, productID, location, mutationType string, qtyChange, balanceAfter int, referenceID, referenceType, note, userID, branchID string) error {
	trace := models.StockTrace{
		ProductID:     productID,
		Location:      location,
		MutationType:  mutationType,
		QtyChange:     qtyChange,
		BalanceAfter:  balanceAfter,
		ReferenceID:   referenceID,
		ReferenceType: referenceType,
		Note:          note,
		UserID:        userID,
		BranchID:      branchID,
	}
	return db.Create(&trace).Error
}

func GetProductTrace(db *gorm.DB, productID, branchID string, page, limit int) ([]models.StockTrace, int64, error) {
	var traces []models.StockTrace
	var total int64

	query := db.Model(&models.StockTrace{}).Where("product_id = ? AND branch_id = ?", productID, branchID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&traces).Error; err != nil {
		return nil, 0, err
	}

	return traces, total, nil
}

func GetProductTraceByDateRange(db *gorm.DB, productID, branchID string, startDate, endDate time.Time, page, limit int) ([]models.StockTrace, int64, error) {
	var traces []models.StockTrace
	var total int64

	query := db.Model(&models.StockTrace{}).
		Where("product_id = ? AND branch_id = ? AND created_at >= ? AND created_at <= ?", productID, branchID, startDate, endDate)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&traces).Error; err != nil {
		return nil, 0, err
	}

	return traces, total, nil
}
