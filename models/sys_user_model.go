package models

import bcrypt "golang.org/x/crypto/bcrypt"

// Inisialisasi model pengguna
type User struct {
	ID         string     `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Username   string     `gorm:"type:varchar(255);not null;unique" json:"username" validate:"required"`
	Password   string     `gorm:"type:text;not null" json:"password" validate:"required"`
	Name       string     `gorm:"type:varchar(255);not null" json:"name" validate:"required"`
	UserRole   UserRole   `gorm:"type:user_role;not null;default:'operator'" json:"user_role" validate:"required"`
	UserStatus DataStatus `gorm:"type:data_status;not null;default:'inactive'" json:"user_status" validate:"required"`
}

// SetID adalah fungsi untuk menetapkan ID ke User
func (b *User) SetID(id string) {
	b.ID = id
}

// HashPassword adalah fungsi untuk melakukan hash password
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}
