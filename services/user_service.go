package services

import (
	models "apotek-clean/internal/core/entities"
	gorm "gorm.io/gorm"
)

func FindUserByID(db *gorm.DB, userID string) (*models.User, error) {
	var user models.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
