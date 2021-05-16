package handlers

import (
	"diplom/pkg/auth"
	"diplom/pkg/context"
	"diplom/pkg/cookie"
	"diplom/pkg/models"
	"diplom/pkg/reports"
	"net/http"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
)

type CustomError struct {
	Message    string
	InnerError error
}

func (c CustomError) Error() string {
	return c.Message + ": " + c.InnerError.Error()
}

var (
	ErrUserAlreadyExists = errors.New("Пользователь с таким логином уже существует")
	ErrFailedLogin       = errors.New("Пользователь с такой комбинацией логина и пароля не найден")
	ErrNoAuth            = errors.New("Пользователь не авторизован")
)

func pongoContextFromUser(user *models.User) pongo2.Context {
	if user == nil {
		return pongo2.Context{
			"authorized": false,
		}
	}

	return pongo2.Context{
		"authorized": true,
		"user":       user,
	}
}

type LoginRequest struct {
	Login    string `form:"login"`
	Password string `form:"password"`
}

func (l *LoginRequest) Validate() error {

	if len(l.Login) < 4 {
		return CustomError{Message: "Логин должен быть длинной минимум 4 символа"}
	}

	if len(l.Password) < 4 {
		return CustomError{Message: "Пароль должен быть длинной минимум 4 символа"}
	}

	return nil
}

type SignUpRequest struct {
	LoginRequest
	RepeatPassword string `form:"passwordRepeat"`
}

func (s *SignUpRequest) Validate() error {
	err := s.LoginRequest.Validate()
	if err != nil {
		return err
	}

	if s.RepeatPassword != s.Password {
		return CustomError{Message: "Пароли не совпадают"}
	}

	return nil
}

func (h *Handler) Index(c echo.Context) error {
	user := context.MustGetUser(c.Request().Context())
	data := pongoContextFromUser(user)

	reps, err := h.ReportRepo.List()
	if err != nil {
		return errors.Wrap(err, "failed to get reps")
	}

	data["reports"] = reps
	return c.Render(http.StatusOK, "index.html", data)
}

func (h *Handler) SignUpPage(c echo.Context) error {
	user := context.MustGetUser(c.Request().Context())
	data := pongoContextFromUser(user)

	return c.Render(http.StatusOK, "signup.html", data)
}
func (h *Handler) LogInPage(c echo.Context) error {
	user := context.MustGetUser(c.Request().Context())
	data := pongoContextFromUser(user)

	return c.Render(http.StatusOK, "login.html", data)
}

func (h *Handler) signUp(c echo.Context, req *SignUpRequest) error {
	err := req.Validate()
	if err != nil {
		return errors.Wrap(err, "failed to validate request")
	}

	_, err = h.AuthRepo.AddUser(req.Login, req.Password)
	if errors.Is(err, auth.ErrAlreadyExists) {
		return CustomError{Message: "Пользователь с таким логином уже существует", InnerError: err}
	}
	if err != nil {
		return errors.Wrap(err, "failed to add user")
	}

	sessionID, err := h.AuthRepo.LogIn(req.Login, req.Password)
	if err != nil {
		return errors.Wrap(err, "failed to auth user")
	}

	cookie.SetSessionCookie(c, sessionID)

	return nil
}

func (h *Handler) SignUp(c echo.Context) error {
	user := context.MustGetUser(c.Request().Context())
	data := pongoContextFromUser(user)

	req := &SignUpRequest{}
	err := c.Bind(req)
	if err != nil {
		return errors.Wrap(err, "failed to bind request")
	}
	data["login"] = req.Login

	httpErr := &CustomError{}
	err = h.signUp(c, req)
	if errors.As(err, httpErr) {
		data["error"] = httpErr.Message
		return c.Render(http.StatusOK, "signup.html", data)
	}
	if err != nil {
		h.Logger.Error(err)
		data["error"] = "Неивестная ошибка сервера"
		return c.Render(http.StatusOK, "signup.html", data)
	}

	return c.Redirect(http.StatusFound, "/")
}

func (h *Handler) logIn(c echo.Context, req *LoginRequest) error {
	err := req.Validate()
	if err != nil {
		return errors.Wrap(err, "failed to validate request")
	}

	sessionID, err := h.AuthRepo.LogIn(req.Login, req.Password)
	if errors.Is(err, auth.ErrNotFound) || errors.Is(err, auth.ErrBadPassword) {
		return CustomError{Message: "Нет юзера с такой комбинацией логина и пароля", InnerError: err}
	}
	if err != nil {
		return errors.Wrap(err, "failed to auth user")
	}

	cookie.SetSessionCookie(c, sessionID)

	return nil
}

func (h *Handler) LogIn(c echo.Context) error {
	user := context.MustGetUser(c.Request().Context())
	data := pongoContextFromUser(user)

	req := &LoginRequest{}
	err := c.Bind(req)
	if err != nil {
		return errors.Wrap(err, "failed to bind request")
	}

	httpErr := &CustomError{}
	err = h.logIn(c, req)
	if errors.As(errors.Cause(err), httpErr) {
		data["error"] = httpErr.Message
		return c.Render(http.StatusOK, "login.html", data)
	}
	if err != nil {
		h.Logger.Error(err)
		data["error"] = "Неивестная ошибка сервера"
		return c.Render(http.StatusOK, "login.html", data)
	}

	return c.Redirect(http.StatusFound, "/")
}

func (h *Handler) SignOut(c echo.Context) error {
	cookie.SetSessionCookie(c, "")
	return c.Redirect(http.StatusFound, "/")
}

func (h *Handler) Report(c echo.Context) error {
	user := context.MustGetUser(c.Request().Context())
	if user == nil {
		return c.Redirect(http.StatusFound, "/")
	}

	data := pongoContextFromUser(user)
	reportName := c.QueryParam("report")
	data["report"] = reports.Report{
		Name: reportName[:strings.LastIndex(reportName, ".")],
		Href: reportName,
	}

	return c.Render(http.StatusOK, "report.html", data)
}

func (h *Handler) AddReport(c echo.Context) error {
	user := context.MustGetUser(c.Request().Context())
	if user == nil || !user.IsAdmin {
		h.Logger.Error("user is not authorized admin")
		return c.Redirect(http.StatusFound, "/")
	}

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return errors.Wrap(err, "failed to read form value")
	}

	src, err := file.Open()
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	defer src.Close()
	spec := file.Filename[strings.LastIndex(file.Filename, "."):]

	name := c.FormValue("name") + spec
	err = h.ReportRepo.Add(name, src)
	if err != nil {
		return errors.Wrap(err, "failed to add file to repo")
	}

	return c.Redirect(http.StatusFound, "/")
}

func (h *Handler) DeleteReport(c echo.Context) error {
	user := context.MustGetUser(c.Request().Context())
	if user == nil || !user.IsAdmin {
		h.Logger.Error("user is not authorized admin")
		return c.Redirect(http.StatusFound, "/")
	}

	name := c.QueryParam("report")
	err := h.ReportRepo.Delete(name)
	if err != nil {
		return errors.Wrap(err, "failed to delete file from repo")
	}

	return c.Redirect(http.StatusFound, "/")
}

type Handler struct {
	AuthRepo   *auth.Repo
	ReportRepo *reports.Repo
	Logger     *log.Logger
}

func New(authRepo *auth.Repo, logger *log.Logger, reportRepo *reports.Repo) *Handler {
	return &Handler{
		AuthRepo:   authRepo,
		Logger:     logger,
		ReportRepo: reportRepo,
	}
}
