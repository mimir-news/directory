package service_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/mimir-news/pkg/httputil/auth"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/directory/pkg/repository"
	"github.com/mimir-news/directory/pkg/service"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/id"
	"github.com/mimir-news/pkg/schema/user"
	"github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
	assert := assert.New(t)

	userID := "user-id"
	expectedUser := domain.FullUser{
		User: user.User{
			ID:    userID,
			Email: "mail@mail.com",
		},
	}

	userRepo := &mockUserRepo{
		findUser: expectedUser,
	}
	userSvc := service.NewUserService(nil, nil, nil, userRepo, nil)

	u, err := userSvc.Get(userID)
	assert.NoError(err)
	assert.Equal(userID, u.ID)
	assert.Equal(expectedUser.User.Email, u.Email)
	assert.Equal(userID, userRepo.findArg)

	userRepo = &mockUserRepo{
		findErr: repository.ErrNoSuchUser,
	}
	userSvc = service.NewUserService(nil, nil, nil, userRepo, nil)

	u, err = userSvc.Get(userID)
	assert.Error(err)
	assert.Equal("", u.ID)
	httpErr, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)
	assert.Equal(userID, userRepo.findArg)

	userRepo = &mockUserRepo{
		findErr: testError,
	}
	userSvc = service.NewUserService(nil, nil, nil, userRepo, nil)

	u, err = userSvc.Get(userID)
	assert.Equal(testError, err)
	assert.Equal("", u.ID)
}

func TestDeleteUser(t *testing.T) {
	assert := assert.New(t)

	userID := "user-id"

	userRepo := &mockUserRepo{
		deleteErr: repository.ErrNoSuchUser,
	}
	userSvc := service.NewUserService(nil, nil, nil, userRepo, nil)

	err := userSvc.Delete(userID)
	assert.Error(err)
	httpErr, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpErr.StatusCode)

	userRepo = &mockUserRepo{
		deleteErr: testError,
	}
	userSvc = service.NewUserService(nil, nil, nil, userRepo, nil)

	err = userSvc.Delete(userID)
	assert.Equal(testError, err)
	assert.Equal(userID, userRepo.deleteArg)
}

func TestUserSvcChangePassword(t *testing.T) {
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

	passwordSvc := service.NewPasswordService(userRepo, "my-pepper", "my-encryption-key")
	userSvc := service.NewUserService(passwordSvc, nil, nil, userRepo, nil)

	pwdChange := user.PasswordChange{
		New:      "new-password",
		Repeated: "new-password",
		Old: user.Credentials{
			Email:    storedUser.User.Email,
			Password: "super-secret-password",
		},
	}

	err := userSvc.ChangePassword(pwdChange)
	assert.NoError(err)
	assert.Equal(pwdChange.Old.Email, userRepo.findByEmailArg)
	savedCreds := userRepo.saveArg.Credentials
	assert.NotEqual(storedUser.Credentials.Salt, savedCreds.Salt)
	assert.NotEqual("", savedCreds.Salt)
	assert.NotEqual(storedUser.Credentials.Password, savedCreds.Password)
	assert.NotEqual("", savedCreds.Password)
	assert.Equal(storedUser.Credentials.Email, savedCreds.Email)

	inconsistentPwdChange := user.PasswordChange{
		New:      "new-password",
		Repeated: "other-new-password",
		Old: user.Credentials{
			Email:    storedUser.User.Email,
			Password: "super-secret-password",
		},
	}

	userRepo.findByEmailArg = ""
	userRepo.saveArg = domain.FullUser{}
	err = userSvc.ChangePassword(inconsistentPwdChange)
	assert.Error(err)
	httpError, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusBadRequest, httpError.StatusCode)
	assert.Equal("", userRepo.findByEmailArg)
	savedCreds = userRepo.saveArg.Credentials
	assert.Equal("", savedCreds.Salt)
	assert.Equal("", savedCreds.Password)
	assert.Equal("", savedCreds.Email)

	wrongPwdChange := user.PasswordChange{
		New:      "new-password",
		Repeated: "new-password",
		Old: user.Credentials{
			Email:    storedUser.User.Email,
			Password: "wrong-password",
		},
	}

	userRepo.findByEmailArg = ""
	userRepo.saveArg = domain.FullUser{}
	err = userSvc.ChangePassword(wrongPwdChange)
	assert.Error(err)
	assert.Equal(wrongPwdChange.Old.Email, userRepo.findByEmailArg)
	savedCreds = userRepo.saveArg.Credentials
	assert.Equal("", savedCreds.Salt)
	assert.Equal("", savedCreds.Password)
	assert.Equal("", savedCreds.Email)

}

func TestChangeEmail(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	oldEmail := "mail@mail.com"
	storedUser := domain.FullUser{
		User: user.User{
			ID:    userID,
			Email: oldEmail,
		},
		Credentials: domain.StoredCredentials{
			Email:    oldEmail,
			Password: "old-hashed-and-encrypted-password",
			Salt:     "3old-hashed-and-encrypted-salt",
		},
	}

	userRepo := &mockUserRepo{
		findUser: storedUser,
	}

	userSvc := service.NewUserService(nil, nil, nil, userRepo, nil)

	newEmail := "new.email@mail.com"
	err := userSvc.ChangeEmail(userID, newEmail)
	assert.NoError(err)
	savedUser := userRepo.saveArg
	assert.Equal(userID, savedUser.User.ID)
	assert.Equal(newEmail, savedUser.User.Email)
	assert.Equal(newEmail, savedUser.Credentials.Email)
	assert.Equal(storedUser.Credentials.Password, savedUser.Credentials.Password)
	assert.Equal(storedUser.Credentials.Salt, savedUser.Credentials.Salt)

	userRepo.findErr = repository.ErrNoSuchUser
	userRepo.findUser = domain.FullUser{}
	userRepo.saveArg = domain.FullUser{}

	err = userSvc.ChangeEmail(userID, newEmail)
	assert.Error(err)
	httpError, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, httpError.StatusCode)
	savedUser = userRepo.saveArg
	assert.Equal("", savedUser.User.ID)

}

func TestCreateAnonymousUser(t *testing.T) {
	assert := assert.New(t)

	jwtCreds := auth.JWTCredentials{Issuer: "user_service_test", Secret: id.New()}
	signer := auth.NewSigner(jwtCreds, 24*time.Hour)
	verifier := auth.NewVerifier(jwtCreds, 0)
	userSvc := service.NewUserService(nil, signer, nil, nil, nil)

	token, err := userSvc.GetAnonymousToken()
	assert.NoError(err)
	assert.Equal(auth.AnonymousRole, token.User.Role)
	assert.Equal("", token.RefreshToken)
	assert.Equal(1, len(token.User.Watchlists))
	assert.Equal(5, len(token.User.Watchlists[0].Stocks))
	expectedSymbols := []string{"TSLA", "AAPL", "AMZN", "NFLX", "FB"}
	for i, stock := range token.User.Watchlists[0].Stocks {
		es := expectedSymbols[i]
		assert.Equal(es, stock.Symbol)
	}

	content, err := verifier.Verify(token.Token)
	assert.NoError(err)
	assert.Equal(token.User.ID, content.User.ID)
}

func TestRefreshToken(t *testing.T) {
	assert := assert.New(t)

	userID := id.New()
	tokenID := id.New()

	oldSession := domain.Session{
		ID:           tokenID,
		UserID:       userID,
		Active:       true,
		RefreshToken: id.New(),
		CreatedAt:    time.Now().UTC().Add(-48 * time.Hour),
	}

	authUser := auth.User{
		ID:   userID,
		Role: auth.UserRole,
	}

	expectedUser := domain.FullUser{
		User: user.User{
			ID:    userID,
			Email: "some@email.com",
			Role:  authUser.Role,
		},
	}

	jwtCreds := auth.JWTCredentials{Issuer: "user_service_test", Secret: id.New()}
	signer := auth.NewSigner(jwtCreds, 24*time.Hour)
	verifier := auth.NewVerifier(jwtCreds, 365*24*time.Hour)
	userRepo := &repository.MockUserRepo{
		FindUser: expectedUser,
	}
	sessionRepo := &repository.MockSessionRepo{
		FindSession: oldSession,
	}
	userSvc := service.NewUserService(nil, signer, verifier, userRepo, sessionRepo)

	oldJwt, err := signer.Sign(tokenID, authUser)
	assert.NoError(err)

	oldToken := user.Token{
		Token:        oldJwt,
		RefreshToken: oldSession.RefreshToken,
	}

	newToken, err := userSvc.RefreshToken(oldToken)
	assert.NoError(err)
	assert.Equal(userID, userRepo.FindArg)
	assert.Equal(1, sessionRepo.FindInvocation)
	assert.Equal(1, sessionRepo.SaveInvocation)

	newTokenBody, err := verifier.Verify(newToken.Token)
	assert.NoError(err)
	assert.Equal(userID, newToken.User.ID)
	assert.Equal(sessionRepo.SaveArg.RefreshToken, newToken.RefreshToken)
	assert.NotEqual(oldToken.RefreshToken, sessionRepo.SaveArg.RefreshToken)
	assert.Equal(sessionRepo.SaveArg.ID, newTokenBody.ID)
	assert.NotEqual(tokenID, sessionRepo.SaveArg.ID)

	assert.Equal(tokenID, sessionRepo.FindArg)

	// Test renewing token for with wrong refresh token.
	sessionRepo.UnsetArgs()
	userRepo.FindArg = ""
	otherSession := domain.Session{
		ID:           tokenID,
		UserID:       userID,
		Active:       true,
		RefreshToken: "well-this-is-clearly-the-wrong-token",
		CreatedAt:    time.Now().UTC().Add(-48 * time.Hour),
	}

	sessionRepo.FindSession = otherSession
	_, err = userSvc.RefreshToken(oldToken)
	assert.Error(err)
	httpErr, ok := err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusForbidden, httpErr.StatusCode)
	assert.Equal(1, sessionRepo.FindInvocation)
	assert.Equal(0, sessionRepo.SaveInvocation)
	assert.Equal(userID, userRepo.FindArg)

	// Test renewing token for inactive session.
	sessionRepo.UnsetArgs()
	userRepo.FindArg = ""
	inactiveSession := domain.Session{
		ID:           tokenID,
		UserID:       userID,
		Active:       false,
		RefreshToken: oldSession.RefreshToken,
		CreatedAt:    time.Now().UTC().Add(-48 * time.Hour),
	}

	sessionRepo.FindSession = inactiveSession
	_, err = userSvc.RefreshToken(oldToken)
	assert.Error(err)
	httpErr, ok = err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusForbidden, httpErr.StatusCode)
	assert.Equal(1, sessionRepo.FindInvocation)
	assert.Equal(0, sessionRepo.SaveInvocation)
	assert.Equal(userID, userRepo.FindArg)

	// Test renewing token for too old session.
	sessionRepo.UnsetArgs()
	userRepo.FindArg = ""
	veryOldSession := domain.Session{
		ID:           tokenID,
		UserID:       userID,
		Active:       true,
		RefreshToken: oldSession.RefreshToken,
		CreatedAt:    time.Now().UTC().Add(-366 * 24 * time.Hour),
	}

	sessionRepo.FindSession = veryOldSession
	_, err = userSvc.RefreshToken(oldToken)
	assert.Error(err)
	httpErr, ok = err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusForbidden, httpErr.StatusCode)
	assert.Equal(1, sessionRepo.FindInvocation)
	assert.Equal(0, sessionRepo.SaveInvocation)
	assert.Equal(userID, userRepo.FindArg)

	// Test renewing token for wrong user.
	sessionRepo.UnsetArgs()
	userRepo.FindArg = ""
	wrongUserSession := domain.Session{
		ID:        tokenID,
		UserID:    userID,
		Active:    false,
		CreatedAt: time.Now().UTC().Add(-48 * time.Hour),
	}

	sessionRepo.FindSession = wrongUserSession
	_, err = userSvc.RefreshToken(oldToken)
	assert.Error(err)
	httpErr, ok = err.(*httputil.Error)
	assert.True(ok)
	assert.Equal(http.StatusForbidden, httpErr.StatusCode)
	assert.Equal(1, sessionRepo.FindInvocation)
	assert.Equal(0, sessionRepo.SaveInvocation)
	assert.Equal(userID, userRepo.FindArg)
}
