package configs

import (
	context "context"
	log "log"
	os "os"
	strconv "strconv"
	time "time"

	models "github.com/heru-oktafian/fiber-apotek/models"
	redis "github.com/redis/go-redis/v9"
	postgres "gorm.io/driver/postgres"
	gorm "gorm.io/gorm"
	logger "gorm.io/gorm/logger"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
	Ctx = context.Background()
)

// SetupDB menginisialisasi koneksi ke database PostgreSQL dan Redis
func SetupDB() (err error) {

	db_host := os.Getenv("DB_HOST")
	db_port := os.Getenv("DB_PORT")
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_name := os.Getenv("DB_NAME")
	redis_host := os.Getenv("REDIS_HOST")
	redis_port := os.Getenv("REDIS_PORT")
	redis_pass := os.Getenv("REDIS_PASS")

	redis_short := os.Getenv("REDIS_SHORT")

	redis_db, err := strconv.Atoi(redis_short)

	dsn := "user=" + db_user + " password=" + db_pass + " host=" + db_host + " port=" + db_port + " dbname=" + db_name + "  sslmode=disable TimeZone=Asia/Jakarta"
	// DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: time.Second,
				// LogLevel:      logger.Info, // sementara Info biar kelihatan
				LogLevel: logger.Silent,
				Colorful: true,
			},
		),
	})

	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL database: %v", err)
	}

	// Migrasi tabel-tabel dari model
	err = DB.AutoMigrate(
		&models.User{},
		&models.Branch{},
		&models.UserBranch{},
		&models.Unit{},
		&models.UnitConversion{},
		&models.ProductCategory{},
		&models.Product{},
		&models.MemberCategory{},
		&models.Member{},
		&models.SupplierCategory{},
		&models.Supplier{},
		&models.DuplicateReceipts{},
		&models.DuplicateReceiptItems{},
		&models.Sales{},
		&models.SaleItems{},
		&models.SaleReturns{},
		&models.SaleReturnItems{},
		&models.Purchases{},
		&models.PurchaseItems{},
		&models.BuyReturns{},
		&models.BuyReturnItems{},
		&models.FirstStocks{},
		&models.FirstStockItems{},
		&models.Opnames{},
		&models.OpnameItems{},
		&models.Expenses{},
		&models.AnotherIncomes{},
		&models.TransactionReports{},
		&models.DailyProfitReport{},
		&models.DailyAsset{},
		&models.Defectas{},
		&models.DefectaItems{},
	)

	if err != nil {
		log.Fatalf("failed to automigrate: %v", err)
	}

	// Panggil fungsi migrasi jika ada perubahan atau penambahan kolom
	// err = runMigrations(DB)
	// if err != nil {
	// 	panic("failed to migrate: " + err.Error())
	// }

	// Koneksi ke database Redis
	if redis_host == "" || redis_port == "" {
		log.Fatalf("❌ REDIS_HOST or REDIS_PORT is not set in .env")
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     redis_host + ":" + redis_port,
		Password: redis_pass,
		DB:       redis_db,
	})

	// Cek koneksi Redis
	_, err = RDB.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("failed to connect to Redis database: %v", err)
	}

	return nil
}

// runMigrations bertanggung jawab untuk menambahkan kolom baru
func runMigrations(db *gorm.DB) error {
	// m := db.Migrator()

	// Tambahkan kolom 'subscription_type' jika belum ada
	// if !m.HasColumn(&models.Branch{}, "SubscriptionType") {
	// 	err := m.AddColumn(&models.Branch{}, "SubscriptionType")
	// 	if err != nil {
	// 		return fmt.Errorf("failed to add column 'subscription_type': %w", err)
	// 	}
	// 	fmt.Println("Column 'subscription_type' successfully added.")
	// } else {
	// 	fmt.Println("Column 'subscription_type' already exists. Skipping.")
	// }

	return nil
}
