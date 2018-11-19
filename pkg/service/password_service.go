package service

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/directory/pkg/repository"
	"github.com/mimir-news/pkg/httputil/auth"
	"github.com/mimir-news/pkg/schema/user"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("Invalid credentials")
)

var (
	emptyCredentials = domain.StoredCredentials{}
)

// PasswordService service responsible for storing and
// validating user credentials.
type PasswordService struct {
	userRepo  repository.UserRepo
	hasher    auth.Hasher
	encryptor encryptionScheme
}

// NewPasswordService sets up a password service.
func NewPasswordService(userRepo repository.UserRepo, pepper, encryptionKey string) *PasswordService {
	return &PasswordService{
		userRepo: userRepo,
		hasher:   auth.NewHasher(pepper),
		encryptor: encryptionScheme{
			hashedKey: auth.HashKey(encryptionKey),
			encryptor: auth.NewAESEncryptor(),
			decryptor: auth.NewAESDecryptor(),
		},
	}
}

// Verify checks that a set of provided credenitals are valid.
func (p *PasswordService) Verify(credentials user.Credentials) error {
	storedCredentials, err := p.getCredentialsByEmail(credentials.Email)
	if err != nil {

		return err
	}

	hashedPwd, err := p.encryptor.decrypt(storedCredentials.Password)
	if err != nil {
		return err
	}

	salt, err := p.encryptor.decrypt(storedCredentials.Salt)
	if err != nil {
		return err
	}

	saltedPassword := p.saltPassword(credentials.Password, salt)
	err = p.hasher.Verify(saltedPassword, hashedPwd)
	if err == auth.ErrHashDoesNotMatch {
		return ErrInvalidCredentials
	}
	return err
}

// ChangePassword verifies old credentials and updates password and salt if successful.
func (p *PasswordService) ChangePassword(newPassword string, oldCredentials user.Credentials) (domain.StoredCredentials, error) {
	err := p.Verify(oldCredentials)
	if err != nil {
		return emptyCredentials, err
	}

	newCredentials := user.Credentials{
		Email:    oldCredentials.Email,
		Password: newPassword,
	}
	return p.Create(newCredentials)
}

// Create creates and returns a new set of credentials for a user.
func (p *PasswordService) Create(credentials user.Credentials) (domain.StoredCredentials, error) {
	salt, err := auth.GenerateSalt()
	if err != nil {
		return emptyCredentials, err
	}

	saltedPassword := p.saltPassword(credentials.Password, salt)
	hashedPassword, err := p.hasher.Hash(saltedPassword)
	if err != nil {
		return emptyCredentials, err
	}

	encryptedHash, err := p.encryptor.encrypt(hashedPassword)
	if err != nil {
		return emptyCredentials, err
	}

	encryptedSalt, err := p.encryptor.encrypt(salt)
	if err != nil {
		return emptyCredentials, err
	}

	encryptedCredentials := domain.StoredCredentials{
		Email:    credentials.Email,
		Password: encryptedHash,
		Salt:     encryptedSalt,
	}
	return encryptedCredentials, nil
}

// getCredentialsByEmail retrieves a users credentials from their email.
func (p *PasswordService) getCredentialsByEmail(email string) (domain.StoredCredentials, error) {
	user, err := p.userRepo.FindByEmail(email)
	if err != nil {
		return emptyCredentials, err
	}

	return user.Credentials, nil
}

// saltPassword concatenates password and salt and returns its checksum.
func (p *PasswordService) saltPassword(password, salt string) string {
	return fmt.Sprintf("%x", auth.HashKey(password, salt))
}

// encryptionScheme symetric encryption scheme used for encryption and decryption of secrets.
type encryptionScheme struct {
	hashedKey []byte
	encryptor auth.Encryptor
	decryptor auth.Decryptor
}

// encrypt encrypts a string and returns its base64 encoded ciphertext.
func (e encryptionScheme) encrypt(plaintext string) (string, error) {
	ciphertext, err := e.encryptor.Encrypt(e.hashedKey, []byte(plaintext))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decodes and decrypts ciphertext and returns the plaintext.
func (e encryptionScheme) decrypt(ciphertext string) (string, error) {
	decodedCiphertext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	plaintext, err := e.decryptor.Decrypt(e.hashedKey, decodedCiphertext)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
