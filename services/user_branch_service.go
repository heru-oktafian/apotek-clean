package services

import (
	models "apotek-clean/models"
	gorm "gorm.io/gorm"
)

func BuildUserBranchRows(userBranches []models.AllUserBranches) []models.AllUserBranches {
	rows := make([]models.AllUserBranches, 0, len(userBranches))
	for _, item := range userBranches {
		rows = append(rows, models.AllUserBranches{
			UserId:     item.UserId,
			UserName:   item.UserName,
			BranchId:   item.BranchId,
			BranchName: item.BranchName,
		})
	}
	return rows
}

func UserBranchExists(db *gorm.DB, userID, branchID string) (bool, error) {
	var count int64
	err := db.Model(&models.UserBranches{}).Where("user_id = ? AND branch_id = ?", userID, branchID).Count(&count).Error
	return count > 0, err
}
