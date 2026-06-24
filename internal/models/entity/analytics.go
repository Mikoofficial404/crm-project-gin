package entity

type AnalyticsResponse struct {
	TotalRevenue     float64 `json:"total_revenue" gorm:"column:total_revenue"`
	ProjectedRevenue float64 `json:"projected_revenue" gorm:"column:projected_revenue"`
	ConversionRate   float64 `json:"conversion_rate" gorm:"column:conversion_rate"`
}
