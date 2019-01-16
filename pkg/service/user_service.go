package service

import (
	"net/http"
	"time"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/directory/pkg/repository"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/httputil/auth"
	"github.com/mimir-news/pkg/id"
	"github.com/mimir-news/pkg/schema/user"
)

const year = 365 * 24 * time.Hour

var (
	emptyUser    = user.User{}
	emptyToken   = user.Token{}
	emptySession = domain.Session{}
)

// UserService service responsible for handling users.
type UserService interface {
	Get(userID string) (user.User, error)
	Create(credentials user.Credentials) (user.User, error)
	Delete(userID string) error
	Authenticate(credentials user.Credentials) (user.Token, error)
	RefreshToken(old user.Token) (user.Token, error)
	ChangePassword(change user.PasswordChange) error
	ChangeEmail(userID, newEmail string) error
	GetAnonymousToken() (user.Token, error)
}

// NewUserService creates a new UserService using the default implementation.
func NewUserService(
	pwdSvc *PasswordService, signer auth.Signer, verifier auth.Verifier,
	userRepo repository.UserRepo, sessionRepo repository.SessionRepo) UserService {
	return &userSvc{
		passwordSvc: pwdSvc,
		tokenSigner: signer,
		verifier:    verifier,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

type userSvc struct {
	passwordSvc *PasswordService
	tokenSigner auth.Signer
	verifier    auth.Verifier
	userRepo    repository.UserRepo
	sessionRepo repository.SessionRepo
}

// Get gets the user with the provided id.
func (us *userSvc) Get(userID string) (user.User, error) {
	fullUser, err := us.userRepo.Find(userID)
	if err == repository.ErrNoSuchUser {
		return emptyUser, httputil.ErrNotFound()
	} else if err != nil {
		return emptyUser, err
	}

	u := fullUser.User
	watchlists, err := us.userRepo.FindWatchlists(u.ID)
	if err != nil {
		return emptyUser, err
	}

	u.Watchlists = watchlists
	return u, nil
}

// Create creates new user based the given credentials.
func (us *userSvc) Create(credentials user.Credentials) (user.User, error) {
	err := us.ensureUserDoesNotExist(credentials.Email)
	if err != nil {
		return emptyUser, err
	}

	newUser, err := us.createNewUser(credentials)
	if err != nil {
		return emptyUser, err
	}

	return newUser, nil
}

// Delete deletes the user with the given id.
func (us *userSvc) Delete(userID string) error {
	err := us.userRepo.Delete(userID)
	if err == repository.ErrNoSuchUser {
		return httputil.ErrNotFound()
	}

	return err
}

// Authenticate validates the credentials provided.
func (us *userSvc) Authenticate(credentials user.Credentials) (user.Token, error) {
	err := us.passwordSvc.Verify(credentials)
	if err != nil {
		return emptyToken, httputil.ErrUnauthorized()
	}

	u, err := us.userRepo.FindByEmail(credentials.Email)
	if err != nil {
		return emptyToken, err
	}

	token, err := us.createSessionToken(u.User)
	if err != nil {
		return emptyToken, err
	}

	return token, nil
}

// RefreshToken refreshes an old token if old is valid.
func (us *userSvc) RefreshToken(old user.Token) (user.Token, error) {
	tokenBody, err := us.verifier.Verify(old.Token)
	if err != nil {
		return emptyToken, httputil.ErrUnauthorized()
	}

	oldSession, err := us.sessionRepo.Find(tokenBody.ID)
	if err != nil {
		return emptyToken, httputil.ErrForbidden()
	}

	storedUser, err := us.userRepo.Find(tokenBody.User.ID)
	if err != nil {
		return emptyToken, err
	}

	err = verifyRefreshToken(old.RefreshToken, tokenBody, oldSession)
	if err != nil {
		return emptyToken, httputil.ErrForbidden()
	}

	err = us.sessionRepo.Delete(tokenBody.ID)
	if err != nil {
		return emptyToken, err
	}

	return us.createSessionToken(storedUser.User)
}

// ChangePassword changes a users password if valid credentials are provided.
func (us *userSvc) ChangePassword(change user.PasswordChange) error {
	if change.New != change.Repeated {
		return errPasswordMissmatch()
	}

	newCreds, err := us.passwordSvc.ChangePassword(change.New, change.Old)
	if err != nil {
		return err
	}

	return us.updateUserCredentials(newCreds)
}

// ChangeEmail changes the email of a given user.
func (us *userSvc) ChangeEmail(userID, newEmail string) error {
	savedUser, err := us.userRepo.Find(userID)
	if err == repository.ErrNoSuchUser {
		return httputil.ErrNotFound()
	} else if err != nil {
		return err
	}

	savedUser.User.Email = newEmail
	savedUser.Credentials.Email = newEmail
	return us.userRepo.Save(savedUser)
}

// GetAnonymousToken creates a new anonymous token.
func (us *userSvc) GetAnonymousToken() (user.Token, error) {
	watchlists := []user.Watchlist{getDefaultWatchlist()}
	u := user.New("", auth.AnonymousRole, watchlists)
	accessToken, err := us.tokenSigner.Sign(id.New(), auth.User{ID: u.ID, Role: u.Role})
	if err != nil {
		return emptyToken, err
	}

	return user.NewToken(accessToken, "", u), nil
}

func (us *userSvc) createNewUser(credentials user.Credentials) (user.User, error) {
	secureCreds, err := us.passwordSvc.Create(credentials)
	if err != nil {
		return user.User{}, err
	}

	newUser := domain.NewUser(secureCreds, nil)

	err = us.userRepo.Save(newUser)
	if err != nil {
		return user.User{}, err
	}

	return newUser.User, nil
}

func (us *userSvc) ensureUserDoesNotExist(email string) error {
	_, err := us.userRepo.FindByEmail(email)
	if err == repository.ErrNoSuchUser {
		return nil
	}
	if err == nil {
		return errUserAlreadyExists()
	}
	return err
}

func (us *userSvc) createSessionToken(u user.User) (user.Token, error) {
	session := domain.NewSession(u.ID)
	accessToken, err := us.tokenSigner.Sign(session.ID, auth.User{ID: u.ID, Role: u.Role})
	if err != nil {
		return emptyToken, err
	}

	err = us.sessionRepo.Save(session)
	if err != nil {
		return emptyToken, err
	}

	return user.NewToken(accessToken, session.RefreshToken, u), nil
}

func (us *userSvc) updateUserCredentials(newCreds domain.StoredCredentials) error {
	user, err := us.userRepo.FindByEmail(newCreds.Email)
	if err != nil {
		return err
	}

	newUser := domain.FullUser{
		User:        user.User,
		Credentials: newCreds,
	}

	return us.userRepo.Save(newUser)
}

func verifyRefreshToken(refreshToken string, token auth.Token, session domain.Session) error {
	if token.User.Role != auth.UserRole {
		return httputil.ErrForbidden()
	}

	if token.User.ID != session.UserID {
		return httputil.ErrForbidden()
	}

	if !session.Active {
		return httputil.ErrForbidden()
	}

	if session.CreatedAt.Before(now().Add(-1 * year)) {
		return httputil.ErrForbidden()
	}

	if session.RefreshToken != refreshToken {
		return httputil.ErrForbidden()
	}

	return nil
}

func errUserAlreadyExists() error {
	return httputil.NewError("User already exists", http.StatusConflict)
}

func errPasswordMissmatch() error {
	return httputil.NewError("Passwords do not match", http.StatusBadRequest)
}

func now() time.Time {
	return time.Now().UTC()
}
