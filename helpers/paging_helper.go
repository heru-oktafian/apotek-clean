package helpers

import (
	fmt "fmt"
	math "math"
	strconv "strconv"
	strings "strings"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	models "github.com/heru-oktafian/fiber-apotek/models"
	gorm "gorm.io/gorm"
)

// Paginate adalah helper untuk menangani pagination dan search pada query
func Paginate(c *fiber.Ctx, query *gorm.DB, model interface{}, searchFields []string) (interface{}, string, int, int, int, int, error) {
	// Parsing body JSON ke struct
	var body models.RequestBody

	// Cek metode request
	if c.Method() == fiber.MethodGet {
		if err := c.QueryParser(&body); err != nil {
			return nil, "", 0, 0, 0, 0, err
		}
	} else {
		if err := c.BodyParser(&body); err != nil {
			return nil, "", 0, 0, 0, 0, err
		}
	}

	// Validasi dan set default untuk page jika tidak valid
	page := body.Page
	if page < 1 {
		page = 1
	}
	limit := 12                              // Tetapkan limit ke 12 data per halaman
	search := strings.TrimSpace(body.Search) // Ambil search key dari body
	offset := (page - 1) * limit

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		if len(searchFields) > 0 {
			whereClause := ""
			args := make([]interface{}, len(searchFields))
			for i, field := range searchFields {
				if whereClause != "" {
					whereClause += " OR "
				}
				whereClause += field
				args[i] = "%" + search + "%"
			}
			query = query.Where(whereClause, args...)
		}
	}

	var total int64

	// Hitung total unit yang sesuai dengan filter
	if err := query.Count(&total).Error; err != nil {
		return nil, "", 0, 0, 0, 0, err
	}

	// Ambil data dengan pagination
	if err := query.Offset(offset).Limit(limit).Scan(model).Error; err != nil {
		return nil, "", 0, 0, 0, 0, err
	}

	// Hitung total halaman berdasarkan hasil filter
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return model, search, int(total), page, totalPages, limit, nil
}

// PaginateWithSearchAndMonth adalah helper untuk menangani pagination dengan parameter search dan month
// Parameters:
//   - c: fiber context
//   - query: gorm query builder
//   - model: interface untuk menampung hasil data
//   - searchColumn: nama kolom untuk filter search (contoh: "A.purchase_id")
//   - dateColumn: nama kolom tanggal untuk filter bulan (contoh: "A.return_date")
//   - page: halaman default jika tidak disediakan
//   - limit: jumlah data per halaman
//
// Return:
//   - data: hasil query yang sudah dipaginate
//   - search: keyword search yang digunakan
//   - total: total data yang sesuai filter
//   - currentPage: halaman saat ini
//   - totalPages: total halaman
//   - error: error jika ada
func PaginateWithSearchAndMonth(
	c *fiber.Ctx,
	query *gorm.DB,
	model interface{},
	searchColumns []string,
	dateColumn string,
	defaultPage int,
	defaultLimit int,
) (interface{}, string, int, int, int, error) {
	// Ambil parameter dari query URL
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))
	month := strings.TrimSpace(c.Query("month"))

	// Konversi page ke int, default ke 1 jika tidak valid
	page := defaultPage
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}

	limit := defaultLimit        // Tetapkan limit ke default
	offset := (page - 1) * limit // Hitung offset berdasarkan halaman dan limit

	// Jika search kosong, set ke string kosong
	if search != "" {
		searchLower := strings.ToLower(search)
		if len(searchColumns) > 0 {
			var conditions []string
			var args []interface{}
			for _, col := range searchColumns {
				conditions = append(conditions, "LOWER("+col+") ILIKE ?")
				args = append(args, "%"+searchLower+"%")
			}
			query = query.Where(strings.Join(conditions, " OR "), args...)
		}
	}

	// Jika month kosong, isi dengan bulan ini (format YYYY-MM)
	if month == "" {
		nowWIB := time.Now()
		month = nowWIB.Format("2006-01")
	}

	// Filter berdasarkan bulan jika disediakan
	if month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return nil, search, 0, page, 0, fmt.Errorf("format bulan tidak valid, gunakan format YYYY-MM")
		}
		startDate := parsedMonth
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		query = query.Where(dateColumn+" BETWEEN ? AND ?", startDate, endDate)
	}

	var total int64

	// Hitung total data yang sesuai dengan filter
	if err := query.Count(&total).Error; err != nil {
		return nil, search, 0, page, 0, err
	}

	// Ambil data dengan pagination
	if err := query.Offset(offset).Limit(limit).Scan(model).Error; err != nil {
		return nil, search, 0, page, 0, err
	}

	// Hitung total halaman berdasarkan hasil filter
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return model, search, int(total), page, totalPages, nil
}
