package models

// SupplierCategory model adalah model untuk kategori supplier yang akan disimpan di database
type SupplierCategory struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string `gorm:"type:varchar(100);not null" json:"name" validate:"required"`
	BranchID string `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
}

// SupplierCategoryCombo model adalah model untuk kategori supplier yang akan ditampilkan di data combobox
type SupplierCategoryCombo struct {
	SupplierCategoryID   uint   `gorm:"primaryKey;autoIncrement" json:"supplier_category_id"`
	SupplierCategoryName string `gorm:"type:varchar(100);not null" json:"supplier_category_name" validate:"required"`
}

// ComboboxSupplierCategories adalah model untuk combo box kategori supplier
type ComboboxSupplierCategories struct {
	SCID     uint   `gorm:"primaryKey;autoIncrement" json:"sc_id" validate:"required"`
	Category string `gorm:"type:varchar(255);not null" json:"category" validate:"required"`
}
