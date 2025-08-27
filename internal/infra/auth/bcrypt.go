package auth

import "golang.org/x/crypto/bcrypt"

type BCrypt struct{ cost int }

func NewBCrypt() *BCrypt { return &BCrypt{cost: bcrypt.DefaultCost} }

func (b *BCrypt) Hash(pw string) (string, error) {
	bt, err := bcrypt.GenerateFromPassword([]byte(pw), b.cost)
	return string(bt), err
}

func (b *BCrypt) Compare(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
