package services

import (
	models "apotek-clean/internal/core/entities"
	gorm "gorm.io/gorm"
)

func GetActiveUserBranchDetails(db *gorm.DB, userID string) ([]models.UserBranchDetail, error) {
	var userBranchDetails []models.UserBranchDetail
	err := db.
		Table("user_branches").
		Select("user_branches.user_id, users.name AS user_name, user_branches.branch_id, branches.branch_name, branches.sia_name, branches.sipa_name, branches.phone").
		Joins("LEFT JOIN users ON users.id = user_branches.user_id").
		Joins("LEFT JOIN branches ON branches.id = user_branches.branch_id").
		Where("branches.branch_status = 'active' AND user_branches.user_id = ?", userID).
		Scan(&userBranchDetails).Error
	return userBranchDetails, err
}
