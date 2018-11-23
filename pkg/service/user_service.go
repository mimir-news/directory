package service

import (
	"net/http"

	"github.com/mimir-news/directory/pkg/domain"
	"github.com/mimir-news/directory/pkg/repository"
	"github.com/mimir-news/pkg/httputil"
	"github.com/mimir-news/pkg/httputil/auth"
	"github.com/mimir-news/pkg/schema/user"
)

var (
	emptyUser  = user.User{}
	emptyToken = user.Token{}
)

// UserService service responsible for handling users.
type UserService interface {
	Get(string) (user.User, error)
	Create(user.Credentials) (user.User, error)
	Delete(string) error
	Authenticate(user.Credentials, string) (user.Token, error)
}

// NewUserService creates a new UserService using the default implementation.
func NewUserService(
	pwdSvc *PasswordService, signer auth.Signer,
	userRepo repository.UserRepo, sessionRepo repository.SessionRepo) UserService {
	return &userSvc{
		passwordSvc: pwdSvc,
		tokenSigner: signer,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

type userSvc struct {
	passwordSvc *PasswordService
	tokenSigner auth.Signer
	userRepo    repository.UserRepo
	sessionRepo repository.SessionRepo
}

// Get gets the user with the provided id.
func (us *userSvc) Get(userID string) (user.User, error) {
	u, err := us.userRepo.Find(userID)
	if err == repository.ErrNoSuchUser {
		return emptyUser, httputil.ErrNotFound()
	} else if err != nil {
		return emptyUser, err
	}

	return u.User, nil
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
func (us *userSvc) Authenticate(credentials user.Credentials, clientID string) (user.Token, error) {
	err := us.passwordSvc.Verify(credentials)
	if err != nil {
		return emptyToken, httputil.ErrUnauthorized()
	}

	u, err := us.userRepo.FindByEmail(credentials.Email)
	if err != nil {
		return emptyToken, err
	}

	encodedToken, err := us.createSessionToken(u.User.ID, clientID)
	if err != nil {
		return emptyToken, err
	}

	return user.NewToken(encodedToken), nil
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

func (us *userSvc) createSessionToken(userID, clientID string) (string, error) {
	token, err := us.tokenSigner.New(userID, clientID)
	if err != nil {
		return "", err
	}

	session := domain.NewSession(userID)
	err = us.sessionRepo.Save(session)
	if err != nil {
		return "", err
	}

	return token, nil
}

func errUserAlreadyExists() error {
	return httputil.NewError("User already exists", http.StatusConflict)
}