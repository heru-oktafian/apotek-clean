package helpers

import (
	rand "crypto/rand"
	fmt "fmt"
	big "math/big" // Untuk konversi byte ke string Base64
	reflect "reflect"
	strings "strings"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
	gorm "gorm.io/gorm"
)

// GenerateID adalah fungsi untuk menghasilkan ID secara otomatis
func GenerateID(prefix string) string {
	// Pastikan prefix memiliki panjang 3 karakter.
	// Jika prefix lebih panjang, potong menjadi 3 karakter.
	if len(prefix) > 3 {
		prefix = prefix[:3]
	}
	// Jika prefix lebih pendek, isi dengan karakter 'X' hingga 3 karakter.
	// Misalnya "SA" akan menjadi "SAX"
	for len(prefix) < 3 {
		prefix += "X"
	}

	// Definisi set karakter
	digits := "0123456789"
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var sb strings.Builder
	sb.Grow(15) // Alokasikan kapasitas awal untuk 15 karakter
	sb.WriteString(prefix)

	// Hasilkan 6 karakter angka acak
	for i := 0; i < 6; i++ {
		// Hasilkan angka acak yang aman secara kriptografi dalam rentang panjang 'digits'
		numIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			// Jika gagal menghasilkan angka acak, ini adalah error sistem yang serius.
			// Melakukan panic agar error ini tidak diabaikan.
			panic(fmt.Sprintf("helpers.GenerateID: Gagal menghasilkan angka acak untuk digit: %v", err))
		}
		sb.WriteByte(digits[numIdx.Int64()]) // Tambahkan karakter angka ke builder
	}

	// Hasilkan 6 karakter alfanumerik acak
	for i := 0; i < 6; i++ {
		// Hasilkan angka acak yang aman secara kriptografi dalam rentang panjang 'alphanum'
		alphaIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphanum))))
		if err != nil {
			// Jika gagal menghasilkan angka acak, ini adalah error sistem yang serius.
			// Melakukan panic agar error ini tidak diabaikan.
			panic(fmt.Sprintf("helpers.GenerateID: Gagal menghasilkan angka acak untuk alfanumerik: %v", err))
		}
		sb.WriteByte(alphanum[alphaIdx.Int64()]) // Tambahkan karakter alfanumerik ke builder
	}

	return sb.String() // Kembalikan ID yang sudah jadi
}

// CreateResourceUser adalah fungsi untuk membuat resource user baru
func CreateResourceUser(c *fiber.Ctx, db *gorm.DB, model interface{}, generateID func() string) error {
	// Parsing body request ke model
	if err := c.BodyParser(model); err != nil {
		return JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	// Cek apakah model adalah User dan hash passwordnya
	if user, ok := model.(*models.User); ok {
		if err := user.HashPassword(); err != nil {
			return JSONResponse(c, fiber.StatusInternalServerError, "Failed to hash password", err)
		}
	}

	// Set ID
	if generateID != nil {
		if idSetter, ok := model.(interface{ SetID(string) }); ok {
			idSetter.SetID(generateID())
		}
	}

	// Simpan ke database
	if err := db.Create(model).Error; err != nil {
		return JSONResponse(c, fiber.StatusInternalServerError, "Failed to create resource", err)
	}

	return JSONResponse(c, fiber.StatusOK, "Resource created successfully", model)
}

// CreateResource adalah fungsi untuk membuat resource baru
func CreateResource(c *fiber.Ctx, db *gorm.DB, model interface{}, IDCode string) error {
	if err := c.BodyParser(model); err != nil {
		return JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	branchID, _ := services.GetBranchID(c)
	generatedID := GenerateID(IDCode)

	// Gunakan reflection untuk set field
	v := reflect.ValueOf(model).Elem()
	if v.Kind() == reflect.Struct {
		if idField := v.FieldByName("ID"); idField.IsValid() && idField.CanSet() && idField.Kind() == reflect.String {
			idField.SetString(generatedID)
		}
		if branchField := v.FieldByName("BranchID"); branchField.IsValid() && branchField.CanSet() && branchField.Kind() == reflect.String {
			branchField.SetString(branchID)
		}

		// Set Stock ke 0 jika ada field "Stock"
		if stockField := v.FieldByName("Stock"); stockField.IsValid() && stockField.CanSet() {
			switch stockField.Kind() {
			case reflect.Int, reflect.Int32, reflect.Int64:
				stockField.SetInt(0)
			case reflect.Uint, reflect.Uint32, reflect.Uint64:
				stockField.SetUint(0)
			}
		}
	}

	if err := db.Create(model).Error; err != nil {
		return JSONResponse(c, fiber.StatusInternalServerError, "Failed to create resource", err)
	}

	return JSONResponse(c, fiber.StatusOK, "Resource created successfully", model)
}

// CreateResourceInc adalah fungsi untuk membuat resource baru (incremental)
func CreateResourceInc(c *fiber.Ctx, db *gorm.DB, model interface{}) error {
	// Parsing body request ke model
	if err := c.BodyParser(model); err != nil {
		return JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	// Set branch_id
	branchID, _ := services.GetBranchID(c)

	// Gunakan reflection untuk set field jika ada
	v := reflect.ValueOf(model).Elem()
	if v.Kind() == reflect.Struct {
		if branchField := v.FieldByName("BranchID"); branchField.IsValid() && branchField.CanSet() && branchField.Kind() == reflect.String {
			branchField.SetString(branchID) // Set BranchID sebagai string
		}
	}

	// Simpan ke database
	if err := db.Create(model).Error; err != nil {
		return JSONResponse(c, fiber.StatusInternalServerError, "Failed to create resource", err)
	}

	return JSONResponse(c, fiber.StatusOK, "Resource created successfully", model)
}

// GetResource adalah fungsi untuk mendapatkan resource
func GetResource(c *fiber.Ctx, db *gorm.DB, model interface{}, id string) error {
	// Cari resource berdasarkan ID
	if err := db.Where("id = ?", id).First(model).Error; err != nil {
		return JSONResponse(c, fiber.StatusNotFound, "Resource not found", err)
	}

	return JSONResponse(c, fiber.StatusOK, "Resource found", model)
}

// UpdateResource adalah fungsi untuk memperbarui resource
func UpdateResource(c *fiber.Ctx, db *gorm.DB, model interface{}, id string) error {
	// Cari resource berdasarkan ID
	if err := db.Where("id = ?", id).First(model).Error; err != nil {
		return JSONResponse(c, fiber.StatusNotFound, "Resource not found", err)
	}

	// Parsing body ke struct yang sama
	if err := c.BodyParser(model); err != nil {
		return JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	// Gunakan reflection untuk menghapus field Stock dari model sebelum update
	v := reflect.ValueOf(model).Elem()
	if v.Kind() == reflect.Struct {
		if stockField := v.FieldByName("Stock"); stockField.IsValid() && stockField.CanSet() {
			switch stockField.Kind() {
			case reflect.Int, reflect.Int32, reflect.Int64:
				stockField.SetInt(0)
			case reflect.Uint, reflect.Uint32, reflect.Uint64:
				stockField.SetUint(0)
			}
			// Tandai bahwa field Stock tidak akan diupdate (reset ke nilai saat diambil dari DB)
			db.Model(model).Omit("stock").Updates(model)
			return JSONResponse(c, fiber.StatusOK, "Resource updated successfully", model)
		}
	}

	// Jika tidak ada field Stock, lanjut update biasa
	if err := db.Model(model).Updates(model).Error; err != nil {
		return JSONResponse(c, fiber.StatusInternalServerError, "Failed to update resource", err)
	}

	return JSONResponse(c, fiber.StatusOK, "Resource updated successfully", model)
}

// DeleteResource adalah fungsi untuk menghapus resource
func DeleteResource(c *fiber.Ctx, db *gorm.DB, model interface{}, id string) error {
	// Cari resource berdasarkan ID
	if err := db.Where("id = ?", id).First(model).Error; err != nil {
		return JSONResponse(c, fiber.StatusNotFound, "Resource not found", err)
	}

	// Hapus resource
	if err := db.Delete(model).Error; err != nil {
		return JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete resource", err)
	}

	return JSONResponse(c, fiber.StatusOK, "Resource deleted successfully", model)
}

// GetAllResources adalah fungsi untuk mendapatkan semua resource
func GetAllResources(c *fiber.Ctx, db *gorm.DB, models interface{}) error {
	// Dapatkan branch id
	branchID, _ := services.GetBranchID(c)

	// Cari semua resource
	if err := db.Where("branch_id = ?", branchID).Find(models).Error; err != nil {
		return JSONResponse(c, fiber.StatusInternalServerError, "Failed to retrieve resources", err)
	}

	return JSONResponse(c, fiber.StatusOK, "Resources retrieved successfully", models)
}

// GetAllBranches adalah fungsi untuk mendapatkan semua resource cabang
func GetAllBranches(c *fiber.Ctx, db *gorm.DB, models interface{}) error {
	userRole, err := services.GetUserRole(c)
	if err != nil {
		return JSONResponse(c, fiber.StatusUnauthorized, "User role tidak ditemukan di token", nil)
	}

	userID, err := services.GetUserID(c)
	if err != nil {
		return JSONResponse(c, fiber.StatusUnauthorized, "User ID tidak ditemukan di token", nil)
	}

	query := db.Model(&models)

	switch userRole {
	case "administrator":
		// administrator: tampilkan semua branch (tanpa filter)

	case "superadmin":
		// superadmin: hanya branch yang punya relasi dengan user_id di user_branches
		query = query.
			Joins("JOIN user_branches ON user_branches.branch_id = branches.id").
			Where("user_branches.user_id = ?", userID)

	default:
		return JSONResponse(c, fiber.StatusForbidden, "Anda tidak memiliki hak untuk mengakses data ini", nil)
	}

	if err := query.Find(models).Error; err != nil {
		return JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data cabang", err)
	}

	return JSONResponse(c, fiber.StatusOK, "Data cabang berhasil diambil", models)
}

// GetAllUsers adalah fungsi untuk mendapatkan semua resource pengguna
func GetAllUsers(c *fiber.Ctx, db *gorm.DB, models interface{}) error {
	userRole, err := services.GetUserRole(c)
	if err != nil {
		return JSONResponse(c, fiber.StatusUnauthorized, "User role tidak ditemukan di token", nil)
	}

	branchID, err := services.GetBranchID(c)
	if err != nil {
		return JSONResponse(c, fiber.StatusUnauthorized, "Branch ID tidak ditemukan di token", nil)
	}

	query := db.Model(&models)

	switch userRole {
	case "administrator":
		// Tampilkan semua user tanpa filter
	case "superadmin":
		// Join dengan tabel user_branches, lalu filter berdasarkan branch_id
		query = query.
			Joins("JOIN user_branches ON user_branches.user_id = users.id").
			Where("user_branches.branch_id = ?", branchID)
	default:
		return JSONResponse(c, fiber.StatusForbidden, "Anda tidak memiliki hak untuk mengakses menu ini", nil)
	}

	if err := query.Find(models).Error; err != nil {
		return JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data pengguna", err)
	}

	return JSONResponse(c, fiber.StatusOK, "Data pengguna berhasil diambil", models)
}

// UpdateResourceUser adalah fungsi untuk memperbarui resource pengguna
func UpdateResourceUser(c *fiber.Ctx, db *gorm.DB, model interface{}, id string) error {
	// Pastikan model yang diberikan adalah pointer ke struct
	if reflect.ValueOf(model).Kind() != reflect.Ptr || reflect.ValueOf(model).Elem().Kind() != reflect.Struct {
		return JSONResponse(c, fiber.StatusInternalServerError, "Model tidak valid, harus berupa pointer ke struct", nil)
	}

	// Buat instance baru dari tipe model untuk mengambil resource yang sudah ada
	existingResource := reflect.New(reflect.TypeOf(model).Elem()).Interface()

	// Cari resource berdasarkan ID
	if err := db.Where("id = ?", id).First(existingResource).Error; err != nil {
		return JSONResponse(c, fiber.StatusNotFound, "Resource tidak ditemukan", err)
	}

	// Parsing body ke instance baru
	// Ini membantu membedakan antara field yang dikirim dalam request dan field yang sudah ada
	incomingData := reflect.New(reflect.TypeOf(model).Elem()).Interface()
	if err := c.BodyParser(incomingData); err != nil {
		return JSONResponse(c, fiber.StatusBadRequest, "Input tidak valid", err)
	}

	// Gunakan reflection untuk menentukan apakah model memiliki field "Password"
	// Jika ya, dan password yang masuk kosong, kita tidak ingin memperbarui password.
	// Jika password yang masuk tidak kosong, kita hash sebelum memperbarui.
	if userModel, ok := model.(*models.User); ok { // Periksa apakah model adalah tipe *models.User
		incomingUser := incomingData.(*models.User) // Cast incomingData ke *models.User

		// Pertahankan password yang sudah ada jika password yang masuk kosong
		if incomingUser.Password == "" {
			userModel.Password = existingResource.(*models.User).Password
		} else {
			// Hash password baru jika disediakan
			if err := incomingUser.HashPassword(); err != nil {
				return JSONResponse(c, fiber.StatusInternalServerError, "Gagal melakukan hash password", err)
			}
			userModel.Password = incomingUser.Password
		}
		// Salin field lain dari incomingUser ke userModel, kecuali ID yang tidak boleh diubah
		userModel.Username = incomingUser.Username
		userModel.Name = incomingUser.Name
		userModel.UserRole = incomingUser.UserRole
		userModel.UserStatus = incomingUser.UserStatus

		// Gunakan gorm.Updates untuk memperbarui model pengguna.
		// Dengan meneruskan userModel, GORM akan memperbarui semua field yang tidak nol.
		if err := db.Model(existingResource).Updates(userModel).Error; err != nil {
			return JSONResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui pengguna", err)
		}

		userModel.ID = id
		return JSONResponse(c, fiber.StatusOK, "Pengguna berhasil diperbarui", userModel)

	} else {
		// Logika pembaruan umum untuk model lain (jika Anda ingin tetap umum)
		// Untuk model lain, Anda mungkin ingin menghilangkan field tertentu seperti "Stock" dari kode asli Anda.
		if err := db.Model(existingResource).Updates(incomingData).Error; err != nil {
			return JSONResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui resource", err)
		}
		return JSONResponse(c, fiber.StatusOK, "Resource berhasil diperbarui", incomingData)
	}
}

// DeleteResourceUser adalah fungsi untuk menghapus resource user beserta data terkait di user_branches
func DeleteResourceUser(c *fiber.Ctx, db *gorm.DB, model interface{}, id string) error {
	// Pastikan model yang diberikan adalah pointer ke struct User
	userModel, ok := model.(*models.User)
	if !ok {
		return JSONResponse(c, fiber.StatusInternalServerError, "Model tidak valid, harus berupa pointer ke struct User", nil)
	}

	// Cari resource User berdasarkan ID
	if err := db.Where("id = ?", id).First(userModel).Error; err != nil {
		return JSONResponse(c, fiber.StatusNotFound, "User tidak ditemukan", err)
	}

	// Mulai transaksi database
	tx := db.Begin()
	if tx.Error != nil {
		return JSONResponse(c, fiber.StatusInternalServerError, "Gagal memulai transaksi database", tx.Error)
	}

	// Hapus data terkait di tabel user_branches
	// Pastikan kolom UserID di UserBranch sesuai dengan ID User
	if err := tx.Where("user_id = ?", id).Delete(&models.UserBranch{}).Error; err != nil {
		tx.Rollback() // Rollback jika ada kesalahan
		return JSONResponse(c, fiber.StatusInternalServerError, "Gagal menghapus data di user_branches", err)
	}

	// Hapus resource User
	if err := tx.Delete(userModel).Error; err != nil {
		tx.Rollback() // Rollback jika ada kesalahan
		return JSONResponse(c, fiber.StatusInternalServerError, "Gagal menghapus user", err)
	}

	// Commit transaksi jika semua operasi berhasil
	tx.Commit()
	if tx.Error != nil {
		return JSONResponse(c, fiber.StatusInternalServerError, "Gagal meng-commit transaksi", tx.Error)
	}

	return JSONResponse(c, fiber.StatusOK, "User dan data terkait berhasil dihapus", userModel)
}

// FormatIndonesianDate memformat objek time.Time menjadi string tanggal dalam Bahasa Indonesia.
// Contoh: "22 Juni 2025"
func FormatIndonesianDate(t time.Time) string {
	months := []string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	day := t.Day()
	month := months[t.Month()-1]
	year := t.Year()
	return fmt.Sprintf("%d %s %d", day, month, year)
}

// FormatIndonesianDateTime memformat objek time.Time menjadi string tanggal dan waktu dalam Bahasa Indonesia.
// Waktu akan diubah ke timezone WIB (Asia/Jakarta).
// Contoh: "22 Juni 2025 14:05:08"
func FormatIndonesianDateTime(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err == nil {
		t = t.In(loc)
	}

	months := []string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	day := t.Day()
	month := months[t.Month()-1]
	year := t.Year()
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()
	return fmt.Sprintf("%02d %s %d %02d:%02d:%02d", day, month, year, hour, minute, second)
}
