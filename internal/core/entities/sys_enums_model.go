package models

// Inisialisasi tipe kustom untuk ENUM JournalMethod
type JournalMethod string

const (
	Manual    JournalMethod = "manual"
	Automatic JournalMethod = "automatic"
)

// Inisialisasi tipe kustom untuk ENUM SubcriptionType
type SubcriptionType string

const (
	Quota    SubcriptionType = "quota"
	Month    SubcriptionType = "month"
	Semester SubcriptionType = "semester"
	Year     SubcriptionType = "year"
)

// Inisialisasi tipe kustom untuk ENUM DataStatus
type DataStatus string

const (
	Active   DataStatus = "active"
	Inactive DataStatus = "inactive"
)

// Inisialisasi tipe kustom untuk ENUM DataEnums
type DataEnums string

const (
	EnumTrue  DataEnums = "true"
	EnumFalse DataEnums = "false"
)

// Inisialisasi tipe kustom untuk ENUM UserRole
type UserRole string

const (
	Operator      UserRole = "operator"
	Cashier       UserRole = "cashier"
	Finance       UserRole = "finance"
	Umum          UserRole = "umum"
	Superadmin    UserRole = "superadmin"
	Administrator UserRole = "administrator"
)

// Inisialisasi tipe kustom untuk ENUM UserRole
// Pendaftaran UserRole = "pendaftaran"
// Rekammedis  UserRole = "rekammedis"
// Ralan       UserRole = "ralan"
// Ranap       UserRole = "ranap"
// Vk          UserRole = "vk"
// Lab         UserRole = "lab"
// Klaim       UserRole = "klaim"
// Simrs       UserRole = "simrs"
// Ipsrs       UserRole = "ipsrs"

// Inisialisasi tipe kustom untuk ENUM PaymentStatus
type PaymentStatus string

const (
	Unpaid       PaymentStatus = "unpaid"
	PaidByCash   PaymentStatus = "paid_by_cash"
	PaidByBank   PaymentStatus = "paid_by_bank"
	PaidByCredit PaymentStatus = "paid_by_credit"
	PaidBySaldo  PaymentStatus = "paid_by_saldo"
	Pending      PaymentStatus = "pending"
	Opname       PaymentStatus = "opname"
	Nocost       PaymentStatus = "nocost"
)

// Inisialisasi tipe kustom untuk ENUM TransactionType
type TransactionType string

const (
	Purchase   TransactionType = "purchase"
	Sale       TransactionType = "sale"
	Expense    TransactionType = "expense"
	Income     TransactionType = "income"
	FirstStock TransactionType = "first_stock"
	Ipname     TransactionType = "opname"
	SaleReturn TransactionType = "sale_return"
	BuyReturn  TransactionType = "buy_return"
)

// Inisialisasi tipe kustom untuk ENUM MovementType
type MovementType string

const (
	PurchaseTrans       MovementType = "purchase"
	PurchaseReturnTrans MovementType = "purchase_return"
	SaleTrans           MovementType = "sale"
	SaleReturnTrans     MovementType = "sale_return"
	OpnameTrans         MovementType = "opname"
	FirstStockTrans     MovementType = "first_stock"
)
