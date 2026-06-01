package helpers

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

// Inisialisasi status data dalam tipe kustom DataStatus
type DataStatus string

const (
	Active   DataStatus = "active"
	Inactive DataStatus = "inactive"
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

// Inisialisasi status pembayaran dalam tipe kustom PaymentStatus
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
