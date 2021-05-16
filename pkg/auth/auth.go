package auth

import (
	"diplom/pkg/models"

	"github.com/labstack/gommon/log"
	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("not found")
	ErrBadPassword   = errors.New("wrong password")
)

type Repo struct {
	DB  *gorm.DB
	Log *log.Logger
}

func New(db *gorm.DB, logger *log.Logger) *Repo {
	return &Repo{
		DB:  db,
		Log: logger,
	}
}

func hashPassword(rawPass string) models.Password {
	hash, err := bcrypt.GenerateFromPassword([]byte(rawPass), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}

	return models.Password(hash)
}

func finalize(query *gorm.DB) error {
	return processDBError(query.Error)
}

func processDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Wrap(ErrNotFound, err.Error())
	}
	sqliteErr := &sqlite3.Error{}
	if errors.As(err, sqliteErr) {
		if sqliteErr.Code == sqlite3.ErrConstraint {
			return errors.Wrap(ErrAlreadyExists, err.Error())
		}
	}

	return err
}

func (r *Repo) AddUser(login string, rawPassword string) (*models.User, error) {
	hashedPassword := hashPassword(rawPassword)
	r.Log.Info("hashedPassword for %s, %s", login, hashedPassword)
	user := &models.User{
		Login:    models.Login(login),
		Password: hashedPassword,
	}

	err := finalize(r.DB.Create(user))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create user")
	}

	return user, nil
}

func (r *Repo) LogIn(login string, rawPassword string) (models.SessionID, error) {
	user := &models.User{}

	err := finalize(r.DB.Where("login = ?", login).First(user))
	if err != nil {
		return "", errors.Wrapf(err, "failed to find user with login %s", login)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(rawPassword)); err != nil {
		return "", errors.Wrapf(ErrBadPassword, "failed to check user's password, hash : %s", err.Error())
	}

	session := &models.Session{
		SessionID: models.SessionID(xid.New().String()),
		UserID:    user.ID,
	}

	err = finalize(r.DB.Create(session))
	if err != nil {
		return "", errors.Wrap(err, "failed to create session")
	}

	return session.SessionID, nil
}

func (r *Repo) CheckSession(sessionID models.SessionID) (*models.User, error) {
	session := &models.Session{}

	err := finalize(r.DB.Where("session_id = ?", sessionID).First(session))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find session with id %s", sessionID)
	}

	user := &models.User{}
	err = finalize(r.DB.Where("id = ?", session.UserID).First(user))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find user with id %s", session.UserID)
	}

	return user, nil
}
