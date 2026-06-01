package models

// ProductCategory adalah model untuk kategori produk
type ProductCategory struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string `gorm:"type:varchar(100);not null" json:"name" validate:"required"`
	BranchID string `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
}

// ComboboxProductCategory adalah model untuk combo box kategori produk
type ComboProductCategory struct {
	ProductCategoryID   uint   `gorm:"primaryKey;autoIncrement" json:"product_category_id"`
	ProductCategoryName string `gorm:"type:varchar(100);not null" json:"product_category_name" validate:"required"`
}

// ComboboxProductCategories adalah model untuk combo box kategori produk
type ComboboxProductCategories struct {
	PCID     uint   `gorm:"primaryKey;autoIncrement" json:"pc_id" validate:"required"`
	Category string `gorm:"type:varchar(255);not null" json:"category" validate:"required"`
}
