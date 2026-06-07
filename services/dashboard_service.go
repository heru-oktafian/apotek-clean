package services

import fiber "github.com/gofiber/fiber/v2"

func CalculateProfitPercentages(totalSales, totalProfit int) (int, int, int) {
	totalHPP := totalSales - totalProfit
	hppPercentage := 0
	profitPercentage := 0
	if totalSales > 0 {
		hppPercentage = (totalHPP * 100) / totalSales
		profitPercentage = (totalProfit * 100) / totalSales
	}
	return totalHPP, hppPercentage, profitPercentage
}

func BuildDailyProfitByUserReportData(results []fiber.Map, totalProfit int) []fiber.Map {
	reportData := make([]fiber.Map, 0, len(results))
	for _, row := range results {
		profit, _ := row["profit"].(int)
		percentage := 0
		if totalProfit > 0 {
			percentage = int(float64(profit) / float64(totalProfit) * 100)
		}
		row["profit_percentage"] = percentage
		reportData = append(reportData, row)
	}
	return reportData
}
