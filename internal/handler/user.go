package handler

import (
	"bufio"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PoorMercymain/gophermart/internal/domain"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo"
)

type user struct {
	srv domain.UserService
}

func NewUser(srv domain.UserService) *user {
	return &user{srv: srv}
}

func (h *user) Register(c echo.Context) error {
	if !IsJSONContentTypeCorrect(c.Request()) {
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	var user domain.User

	if err := json.NewDecoder(c.Request().Body).Decode(&user); err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		util.LogInfoln(err)
		return err
	}
	defer c.Request().Body.Close()

	if user.Login == "" || user.Password == "" {
		c.Response().WriteHeader(http.StatusBadRequest)
		util.LogInfoln("login or password is empty:", user)
		return nil
	}

	util.LogInfoln("this is a password:", user.Password, "this is a login:", user.Login)

	uniqueLoginErrorChan := make(chan error, 1)

	err := h.srv.Register(c.Request().Context(), &user, uniqueLoginErrorChan)
	if err != nil {
		select {
		case <-uniqueLoginErrorChan:
			c.Response().WriteHeader(http.StatusConflict)
			return err
		default:
			c.Response().WriteHeader(http.StatusInternalServerError)
			return err
		}
	}

	util.LogInfoln("в хэндлере", user)
	jwtStr, err := CreateJWTString(user.Login + " " + user.Password)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return err
	}

	util.LogInfoln("jwt", jwtStr)
	cookie := &http.Cookie{Name: "jwt", Value: jwtStr, Expires: time.Now().Add(time.Hour * 3)}
	http.SetCookie(c.Response(), cookie)
	c.Response().WriteHeader(http.StatusOK)
	return nil
}

func (h *user) Authenticate(c echo.Context) error {
	if !IsJSONContentTypeCorrect(c.Request()) {
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	var user domain.User

	if err := json.NewDecoder(c.Request().Body).Decode(&user); err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		util.LogInfoln(err)
		return err
	}
	defer c.Request().Body.Close()

	if user.Login == "" || user.Password == "" {
		c.Response().WriteHeader(http.StatusBadRequest)
		util.LogInfoln("login or password is empty:", user)
		return nil
	}

	validPair, _ := h.srv.CompareHashAndPassword(c.Request().Context(), &user)
	if validPair {
		jwtStr, err := CreateJWTString(user.Login + " " + user.Password)
		if err != nil {
			c.Response().WriteHeader(http.StatusInternalServerError)
			return err
		}
		util.LogInfoln("this is a password:", user.Password, "this is a login:", user.Login)
		util.LogInfoln("jwt", jwtStr)
		cookie := &http.Cookie{Name: "jwt", Value: jwtStr, Expires: time.Now().Add(time.Hour * 3)}
		http.SetCookie(c.Response(), cookie)
		c.Response().WriteHeader(http.StatusOK)
		return nil
	}

	c.Response().WriteHeader(http.StatusUnauthorized)
	return nil
}

func (h *user) AddOrder(c echo.Context) error {
	if !IsPlaintextContentTypeCorrect(c.Request()) {
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	scanner := bufio.NewScanner(c.Request().Body)
	scanner.Scan()
	defer c.Request().Body.Close()

	orderNumber, err := strconv.Atoi(scanner.Text())
	if err != nil {
		c.Response().WriteHeader(http.StatusUnprocessableEntity)
		return err
	}

	// TODO: add goroutine to send req to accrual
	err = h.srv.AddOrder(c.Request().Context(), orderNumber)
	if errors.Is(err, domain.ErrorAlreadyRegistered) {
		c.Response().WriteHeader(http.StatusOK)
		return err
	} else if errors.Is(err, domain.ErrorAlreadyRegisteredByAnotherUser) {
		c.Response().WriteHeader(http.StatusConflict)
		return err
	} else if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return err
	}
	c.Response().WriteHeader(http.StatusAccepted)
	return nil
}

func IsJSONContentTypeCorrect(r *http.Request) bool {
	if len(r.Header.Values("Content-Type")) == 0 {
		return false
	}

	for contentTypeCurrentIndex, contentType := range r.Header.Values("Content-Type") {
		if contentType == "application/json" {
			break
		}
		if contentTypeCurrentIndex == len(r.Header.Values("Content-Type"))-1 {
			return false
		}
	}

	return true
}

func IsPlaintextContentTypeCorrect(r *http.Request) bool {
	if len(r.Header.Values("Content-Type")) == 0 {
		return false
	}

	for contentTypeCurrentIndex, contentType := range r.Header.Values("Content-Type") {
		if strings.HasPrefix(contentType, "text/plain") {
			break
		}
		if contentTypeCurrentIndex == len(r.Header.Values("Content-Type"))-1 {
			return false
		}
	}

	return true
}

func CreateJWTString(stringToIncludeInJWT string) (string, error) {
	claims := jwt.MapClaims{
		"str": stringToIncludeInJWT,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("ultrasecretkey"))
	if err != nil {
		util.LogInfoln("could not create token", err)
		return "", err
	}
	return tokenString, err
}
