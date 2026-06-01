package helpers

import (
	fmt "fmt"
	strings "strings"

	validator "github.com/go-playground/validator/v10"
	pq "github.com/lib/pq"
)

// validate adalah instance global dari validator.
// Diinisialisasi sekali saja untuk menghindari membuat instance baru di setiap request,
// hal ini menghemat memori dan CPU.
var validate = validator.New()

func IsDuplicateKeyError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		// Error code '23505' adalah untuk unique_violation di PostgreSQL
		return pqErr.Code == "23505"
	}
	return false
}

func ValidateStruct(s interface{}) error {
	// Jalankan validasi pada struct
	if err := validate.Struct(s); err != nil {
		// Jika ada error validasi, kami akan memproses mereka untuk membuat pesan yang lebih informatif.
		var errorMessages []string
		// Konversikan error ke tipe validator.ValidationErrors untuk iterasi.
		for _, err := range err.(validator.ValidationErrors) {
			// Buat pesan error dasar. Contoh: "Field 'MemberId' is required"
			errorMessage := fmt.Sprintf("Field '%s' is %s", err.Field(), err.Tag())

			// Tambahkan parameter validasi jika tersedia (misal: `min`, `max`, `len`, dll)
			if err.Param() != "" {
				errorMessage = fmt.Sprintf("Field '%s' %s %s", err.Field(), err.Tag(), err.Param())
			}

			errorMessages = append(errorMessages, errorMessage)
		}
		// Gabungkan semua pesan error menjadi satu string yang dipisahkan dengan koma.
		return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, ", "))
	}
	// Jika tidak ada error, validasi berhasil dilakukan.
	return nil
}
