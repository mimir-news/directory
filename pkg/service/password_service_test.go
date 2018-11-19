package service_test

import (
	"errors"
	"testing"

	"github.com/mimir-news/pkg/id"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/directory/pkg/service"
	"github.com/mimir-news/pkg/schema/user"
	"github.com/stretchr/testify/assert"
)

var testError = errors.New("expected test error")

func TestVerify(t *testing.T) {
	assert := assert.New(t)

	// hashedPepper = SHA-3-256(my-pepper) = f280c15e38d4cfaa846e87780bd12e6428631d0ccdbd6857f03a0f5d92e0fe91
	// hashedPwdSalt = SHA-3-256(super-secret-password-my-salt) = 8741924fd10d8f1e7d5c2e42395ea972bebf02c1c8412a6c740f44741c9530e1
	// hashedBody = SHA-3-256(hashedPepper-hashedPwdSalt) = 4bc6172472925ea4243add2dae09adbac71639677aa1b59fa51349386ddd45b0
	// hashedPassword = brcypt-cost-12(hashedBody) = $2y$12$IP3CZFQ.lM8yef1Qrcerb.DJaLkFFmTLQzRX163LvbBMDzjwP7H1W
	// hashedKey : bytes = SHA-3-256(my-encryption-key)
	// encryptedPassword = base64(aes(hashedKey, hashedPassword)) = S5UeZOWCDkIfP/5LUDpyhIY0l6+aow+CmkBEVtHqpebhe04vb6kDbPaD/wo05fs6x1lvJfI/6YZ66zbQ8X2lHaEThp4f1Zl0exk7j/wow740KbWZHf9DSA==
	// encryptedSalt = base64(aes(hashedKey, my-salt)) = 3MQEKd3NVnU+WQFQxo8JpYWrTrqXOiwro4MwLwnsckWXinE=

	storedUser := domain.FullUser{
		User: user.User{
			ID:    id.New(),
			Email: "mail@mail.com",
		},
		Credentials: domain.StoredCredentials{
			Email:    "mail@mail.com",
			Password: "S5UeZOWCDkIfP/5LUDpyhIY0l6+aow+CmkBEVtHqpebhe04vb6kDbPaD/wo05fs6x1lvJfI/6YZ66zbQ8X2lHaEThp4f1Zl0exk7j/wow740KbWZHf9DSA==", // Hashed and encrypted password.
			Salt:     "3MQEKd3NVnU+WQFQxo8JpYWrTrqXOiwro4MwLwnsckWXinE=",                                                                         // Encrypted salt
		},
	}

	userRepo := &mockUserRepo{
		findByEmailUser: storedUser,
	}

	passwordSvc := service.NewPasswordService(userRepo, "my-pepper", "my-encryption-key")

	creds := user.Credentials{
		Email:    storedUser.User.Email,
		Password: "super-secret-password",
	}

	err := passwordSvc.Verify(creds)
	assert.Nil(err)
	assert.Equal(creds.Email, userRepo.findByEmailArg)

	wrongCreds := user.Credentials{
		Email:    storedUser.User.Email,
		Password: "wrong-password",
	}
	userRepo.findByEmailArg = ""

	err = passwordSvc.Verify(wrongCreds)
	assert.Equal(service.ErrInvalidCredentials, err)
	assert.Equal(creds.Email, userRepo.findByEmailArg)

	userRepo.findByEmailArg = ""
	userRepo.findByEmailUser = domain.FullUser{}
	userRepo.findByEmailErr = testError

	err = passwordSvc.Verify(creds)
	assert.NotNil(err)
	assert.NotEqual(service.ErrInvalidCredentials, err)
	assert.Equal(creds.Email, userRepo.findByEmailArg)
}

func TestChangePassword(t *testing.T) {
	assert := assert.New(t)

	storedUser := domain.FullUser{
		User: user.User{
			ID:    id.New(),
			Email: "mail@mail.com",
		},
		Credentials: domain.StoredCredentials{
			Email:    "mail@mail.com",
			Password: "S5UeZOWCDkIfP/5LUDpyhIY0l6+aow+CmkBEVtHqpebhe04vb6kDbPaD/wo05fs6x1lvJfI/6YZ66zbQ8X2lHaEThp4f1Zl0exk7j/wow740KbWZHf9DSA==", // Hashed and encrypted password.
			Salt:     "3MQEKd3NVnU+WQFQxo8JpYWrTrqXOiwro4MwLwnsckWXinE=",                                                                         // Encrypted salt
		},
	}

	userRepo := &mockUserRepo{
		findByEmailUser: storedUser,
	}

	oldCreds := user.Credentials{
		Email:    storedUser.User.Email,
		Password: "super-secret-password",
	}

	newCreds := user.Credentials{
		Email:    storedUser.User.Email,
		Password: "super-new-password",
	}

	passwordSvc := service.NewPasswordService(userRepo, "my-pepper", "my-encryption-key")
	res, err := passwordSvc.ChangePassword(newCreds.Password, oldCreds)
	assert.Nil(err)
	assert.Equal(oldCreds.Email, userRepo.findByEmailArg)
	assert.Equal(newCreds.Email, res.Email)
	assert.NotEqual(storedUser.Credentials.Password, res.Password)
	assert.NotEqual(storedUser.Credentials.Salt, res.Salt)
	assert.NotEqual("", res.Password)
	assert.NotEqual("", res.Salt)

	newStoredUser := domain.FullUser{
		User: user.User{
			Email: "mail@mail.com",
		},
		Credentials: res,
	}

	userRepo.findByEmailUser = newStoredUser

	verificationErr := passwordSvc.Verify(newCreds)
	assert.Nil(verificationErr)
	assert.Equal(newCreds.Email, userRepo.findByEmailArg)

	userRepo.findByEmailArg = ""

	verificationErr = passwordSvc.Verify(oldCreds)
	assert.Equal(service.ErrInvalidCredentials, verificationErr)
	assert.Equal(newCreds.Email, userRepo.findByEmailArg)

	wrongCreds := user.Credentials{
		Email:    "mail@mail.com",
		Password: "wrong-password",
	}
	userRepo.findByEmailArg = ""

	verificationErr = passwordSvc.Verify(wrongCreds)
	assert.Equal(service.ErrInvalidCredentials, verificationErr)
	assert.Equal(newCreds.Email, userRepo.findByEmailArg)
}

func TestCreate(t *testing.T) {
	assert := assert.New(t)

	oldCredentials := domain.StoredCredentials{
		Password: "S5UeZOWCDkIfP/5LUDpyhIY0l6+aow+CmkBEVtHqpebhe04vb6kDbPaD/wo05fs6x1lvJfI/6YZ66zbQ8X2lHaEThp4f1Zl0exk7j/wow740KbWZHf9DSA==", // Hashed and encrypted password.
		Salt:     "3MQEKd3NVnU+WQFQxo8JpYWrTrqXOiwro4MwLwnsckWXinE=",                                                                         // Encrypted salt
	}

	userRepo := &mockUserRepo{}
	passwordSvc := service.NewPasswordService(userRepo, "my-pepper", "my-encryption-key")

	credentials := user.Credentials{
		Email:    "mail@mail.com",
		Password: "super-secret-password",
	}

	res, err := passwordSvc.Create(credentials)
	assert.Nil(err)
	assert.Equal(credentials.Email, res.Email)
	assert.NotEqual(oldCredentials.Password, res.Password)
	assert.NotEqual(oldCredentials.Salt, res.Salt)
	assert.NotEqual("", res.Password)
	assert.NotEqual("", res.Salt)

	storedUser := domain.FullUser{
		User: user.User{
			Email: "mail@mail.com",
		},
		Credentials: res,
	}

	userRepo.findByEmailUser = storedUser

	verificationErr := passwordSvc.Verify(credentials)
	assert.Nil(verificationErr)
	assert.Equal(credentials.Email, userRepo.findByEmailArg)

	wrongCreds := user.Credentials{
		Email:    "mail@mail.com",
		Password: "wrong-password",
	}
	userRepo.findByEmailArg = ""

	verificationErr = passwordSvc.Verify(wrongCreds)
	assert.Equal(service.ErrInvalidCredentials, verificationErr)
	assert.Equal(credentials.Email, userRepo.findByEmailArg)
}

type mockUserRepo struct {
	findUser domain.FullUser
	findErr  error
	findArg  string

	findByEmailUser domain.FullUser
	findByEmailErr  error
	findByEmailArg  string

	saveErr error
	saveArg domain.FullUser
}

func (r *mockUserRepo) Find(id string) (domain.FullUser, error) {
	r.findArg = id
	return r.findUser, r.findErr
}

func (r *mockUserRepo) FindByEmail(email string) (domain.FullUser, error) {
	r.findByEmailArg = email
	return r.findByEmailUser, r.findByEmailErr
}

func (r *mockUserRepo) Save(user domain.FullUser) error {
	r.saveArg = user
	return r.saveErr
}
