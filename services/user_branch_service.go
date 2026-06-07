package services

import models "apotek-clean/models"

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
