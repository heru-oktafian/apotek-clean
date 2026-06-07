package services

import (
	models "apotek-clean/models"
	gorm "gorm.io/gorm"
)

func BuildUserBranchRows(userBranches []models.UserBranchDetail) []models.UserBranchDetail {
	rows := make([]models.UserBranchDetail, 0, len(userBranches))
	for _, item := range userBranches {
		rows = append(rows, models.UserBranchDetail{
			UserID:     item.UserID,
			UserName:   item.UserName,
			BranchID:   item.BranchID,
			BranchName: item.BranchName,
			SiaName:    item.SiaName,
			SipaName:   item.SipaName,
			Phone:      item.Phone,
		})
	}
	return rows
}

func UserBranchExists(db *gorm.DB, userID, branchID string) (bool, error) {
	var count int64
	err := db.Model(&models.UserBranch{}).Where("user_id = ? AND branch_id = ?", userID, branchID).Count(&count).Error
	return count > 0, err
}

type UserBranchDetailBranch struct {
	BranchID   string `json:"branch_id"`
	BranchName string `json:"branch_name"`
	Address    string `json:"address"`
	Phone      string `json:"phone"`
}

func BuildUserDetailBranches(branches []models.Branch) []UserBranchDetailBranch {
	rows := make([]UserBranchDetailBranch, 0, len(branches))
	for _, branch := range branches {
		rows = append(rows, UserBranchDetailBranch{
			BranchID:   branch.ID,
			BranchName: branch.BranchName,
			Address:    branch.Address,
			Phone:      branch.Phone,
		})
	}
	return rows
}
