package crons

import (
	log "log"
	time "time"

	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	gorm "gorm.io/gorm"
)

func AssetCounter(db *gorm.DB) error {
	// SQL query untuk menghitung nilai aset per cabang
	query := `
		SELECT 
			branch_id,
			SUM(stock * purchase_price) as total_asset
		FROM 
			products
		GROUP BY 
			branch_id
	`

	type BranchAsset struct {
		BranchID   string
		TotalAsset int
	}

	var branchAssets []BranchAsset
	if err := db.Raw(query).Scan(&branchAssets).Error; err != nil {
		log.Printf("[ASSET COUNTER] Error querying branch assets: %v", err)
		return err
	}

	// Persiapkan peta kredit dan real_asset per cabang
	creditMap := make(map[string]int)
	realAssetMap := make(map[string]string)

	type BranchInfo struct {
		ID        string
		RealAsset string
	}
	var branchesInfo []BranchInfo
	if err := db.Table("branches").Select("id, real_asset").Scan(&branchesInfo).Error; err != nil {
		log.Printf("[ASSET COUNTER] Error querying branches info: %v", err)
		return err
	}
	for _, b := range branchesInfo {
		realAssetMap[b.ID] = b.RealAsset
	}

	// Query untuk total pembelian kredit per cabang
	creditQuery := `
		SELECT 
			branch_id,
			COALESCE(SUM(total_purchase), 0) as total_credit
		FROM 
			purchases
		WHERE 
			payment = 'paid_by_credit'
		GROUP BY 
			branch_id
	`

	type BranchCredit struct {
		BranchID    string
		TotalCredit int
	}

	var branchCredits []BranchCredit
	if err := db.Raw(creditQuery).Scan(&branchCredits).Error; err != nil {
		log.Printf("[ASSET COUNTER] Error querying branch credits: %v", err)
		return err
	}

	// Buat map untuk lookup total kredit per branch
	for _, credit := range branchCredits {
		creditMap[credit.BranchID] = credit.TotalCredit
	}

	// Menyimpan aset harian untuk setiap cabang
	for _, asset := range branchAssets {
		credit := 0
		if realAssetMap[asset.BranchID] == "true" {
			credit = creditMap[asset.BranchID]
		}
		finalAsset := asset.TotalAsset - credit

		// Hitung statistik bulan ini untuk branch ini: jumlah hari tersimpan dan jumlah asset_value
		now := time.Now()
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		monthEnd := monthStart.AddDate(0, 1, 0)

		type BranchMonthStats struct {
			QtyDays      int
			SumAssetDays int
		}

		var stats BranchMonthStats
		statsQuery := `
			SELECT
				COALESCE(COUNT(*), 0) as qty_days,
				COALESCE(SUM(asset_value), 0) as sum_asset_days
			FROM
				daily_assets
			WHERE
				branch_id = ? AND asset_date >= ? AND asset_date < ?
		`

		if err := db.Raw(statsQuery, asset.BranchID, monthStart, monthEnd).Scan(&stats).Error; err != nil {
			log.Printf("[ASSET COUNTER] Error querying monthly stats for branch %s: %v", asset.BranchID, err)
			return err
		}

		vQtyDays := stats.QtyDays
		vSumAssetDays := stats.SumAssetDays

		// Hitung asset_average: (total final_asset bulan ini + finalAsset sekarang) / (qty records bulan ini + 1)
		var assetAverage int
		totalAssetWithCurrent := vSumAssetDays + finalAsset
		totalRecordsWithCurrent := vQtyDays + 1
		assetAverage = totalAssetWithCurrent / totalRecordsWithCurrent

		dailyAsset := models.DailyAsset{
			ID:           helpers.GenerateID("AST"),
			AssetDate:    time.Now(),
			AssetValue:   finalAsset,
			AssetAverage: assetAverage,
			BranchId:     asset.BranchID,
		}

		if err := db.Create(&dailyAsset).Error; err != nil {
			log.Printf("[ASSET COUNTER] Error creating daily asset for branch %s: %v", asset.BranchID, err)
			return err
		}
	}

	log.Println("[ASSET COUNTER] Successfully updated daily assets for all branches")
	return nil
}
