package services

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
