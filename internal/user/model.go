package user

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UUID     string `json:"uuid"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

func (u *User) CheckPassword(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return fmt.Errorf("password does not match")
	}
	return nil
}

func (u *User) GeneratePasswordHash() error {
	pwd, err := generatePasswordHash(u.Password)
	if err != nil {
		return err
	}
	u.Password = pwd
	return nil
}

type CreateUserDTO struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	RepeatPassword string `json:"repeat_password"`
}

type UpdateUserDTO struct {
	UUID        string `json:"uuid,omitempty"`
	Email       string `json:"email,omitempty"`
	Password    string `json:"password,omitempty"`
	OldPassword string `json:"old_password,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}

type UserFavouriteCityDTO struct {
	UUID     string `json:"uuid,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
	CityID   string `json:"city_id"`
}

func NewUser(dto CreateUserDTO) User {
	return User{
		Email:    dto.Email,
		Password: dto.Password,
	}
}

func UpdatedUser(dto UpdateUserDTO) User {
	return User{
		UUID:     dto.UUID,
		Email:    dto.Email,
		Password: dto.Password,
	}
}

// TODO проверить используется ли
func UserFavouriteCity(dto UserFavouriteCityDTO) User {
	return User{
		UUID:     dto.UUID,
		Email:    dto.Email,
		Password: dto.Password,
	}
}

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password due to error %w", err)
	}
	return string(hash), nil
}
