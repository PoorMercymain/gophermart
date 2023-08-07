package handler

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
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

	defer c.Request().Body.Close()

	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}
	util.GetLogger().Infoln(string(b))

	bufCopy := append([]byte(nil), b...)
	reader := bytes.NewReader(bufCopy)

	err = util.CheckDuplicatesInJSON(json.NewDecoder(reader), nil)
	if err != nil {
		util.GetLogger().Infoln("json with duplicate properties")
		err = domain.ErrorRequestFormatIncorrect
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	c.Request().Body = io.NopCloser(bytes.NewBuffer(b))

	d := json.NewDecoder(c.Request().Body)
	d.DisallowUnknownFields()

	if err := d.Decode(&user); err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		util.GetLogger().Infoln(err)
		return err
	}

	if user.Login == "" || user.Password == "" {
		c.Response().WriteHeader(http.StatusBadRequest)
		util.GetLogger().Infoln("login or password is empty:", user)
		return nil
	}

	util.GetLogger().Infoln("this is a password:", user.Password, "this is a login:", user.Login)

	uniqueLoginErrorChan := make(chan error, 1)

	err = h.srv.Register(c.Request().Context(), &user, uniqueLoginErrorChan)
	if err != nil {
		select {
		case <-uniqueLoginErrorChan:
			c.Response().WriteHeader(http.StatusConflict)
			return err
		default:
			c.Response().WriteHeader(http.StatusInternalServerError)
			util.GetLogger().Infoln(err)
			return err
		}
	}

	util.GetLogger().Infoln("в хэндлере", user)
	jwtStr, err := CreateJWTString(user.Login + " " + user.Password)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return err
	}

	util.GetLogger().Infoln("jwt", jwtStr)
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

	defer c.Request().Body.Close()

	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	bufCopy := append([]byte(nil), b...)
	reader := bytes.NewReader(bufCopy)

	err = util.CheckDuplicatesInJSON(json.NewDecoder(reader), nil)
	if err != nil {
		util.GetLogger().Infoln("json with duplicate properties")
		err = domain.ErrorRequestFormatIncorrect
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	c.Request().Body = io.NopCloser(bytes.NewBuffer(b))

	d := json.NewDecoder(c.Request().Body)
	d.DisallowUnknownFields()

	if err := d.Decode(&user); err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		util.GetLogger().Infoln(err)
		return err
	}

	if user.Login == "" || user.Password == "" {
		c.Response().WriteHeader(http.StatusBadRequest)
		util.GetLogger().Infoln("login or password is empty:", user)
		return nil
	}

	validPair, _ := h.srv.CompareHashAndPassword(c.Request().Context(), &user)
	if validPair {
		jwtStr, err := CreateJWTString(user.Login + " " + user.Password)
		if err != nil {
			c.Response().WriteHeader(http.StatusInternalServerError)
			return err
		}
		util.GetLogger().Infoln("this is a password:", user.Password, "this is a login:", user.Login)
		util.GetLogger().Infoln("jwt", jwtStr)
		cookie := &http.Cookie{Name: "jwt", Value: jwtStr, Expires: time.Now().Add(time.Hour * 3)}
		http.SetCookie(c.Response(), cookie)
		c.Response().WriteHeader(http.StatusOK)
		return nil
	}

	c.Response().WriteHeader(http.StatusUnauthorized)
	return nil
}

func (h *user) AddOrder(wg *sync.WaitGroup) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !IsPlaintextContentTypeCorrect(c.Request()) {
			c.Response().WriteHeader(http.StatusBadRequest)
			return nil
		}

		scanner := bufio.NewScanner(c.Request().Body)
		scanner.Scan()
		defer c.Request().Body.Close()

		orderN := scanner.Text()

		for _, orderNumSymbol := range orderN {
			_, err := strconv.Atoi(string(orderNumSymbol))
			if err != nil {
				c.Response().WriteHeader(http.StatusBadRequest)
				return err
			}
		}

		err := h.srv.AddOrder(c.Request().Context(), orderN)
		if errors.Is(err, domain.ErrorAlreadyRegistered) {
			c.Response().WriteHeader(http.StatusOK)
			return err
		} else if errors.Is(err, domain.ErrorAlreadyRegisteredByAnotherUser) {
			c.Response().WriteHeader(http.StatusConflict)
			return err
		} else if errors.Is(err, domain.ErrorIncorrectOrderNumber) {
			c.Response().WriteHeader(http.StatusUnprocessableEntity)
			return err
		} else if err != nil {
			c.Response().WriteHeader(http.StatusInternalServerError)
			util.GetLogger().Infoln(err)
			return err
		}

		// simple solution (for tests to ignore that code), but mock of accrual would be better
		if c.Request().Context().Value(domain.Key("testing")) != "t" {
			wg.Add(1)
			go func() {
				accrualWithEndpoint := c.Request().Context().Value(domain.Key("accrual_address")).(string) + "/api/orders/" + orderN
				login := c.Request().Context().Value(domain.Key("login")).(string)
				var previousAccrualOrder domain.AccrualOrder
				var accrualOrder domain.AccrualOrder

				util.GetLogger().Infoln(accrualOrder)

				for {
					util.GetLogger().Infoln("requested", accrualWithEndpoint)
					resp, err := http.Get(accrualWithEndpoint)
					if err != nil {
						util.GetLogger().Infoln(err)
						wg.Done()
						return
					}
					defer resp.Body.Close()

					var body []byte

					if resp.StatusCode != http.StatusNoContent {
						body, err = io.ReadAll(resp.Body)
						if err != nil {
							util.GetLogger().Infoln(err)
							wg.Done()
							return
						}
					}

					if body != nil {
						err = json.Unmarshal(body, &accrualOrder)
						if err != nil {
							util.GetLogger().Infoln(string(body))
							util.GetLogger().Infoln(err)
							wg.Done()
							return
						}
						if accrualOrder.Status != previousAccrualOrder.Status {
							previousAccrualOrder = accrualOrder
							cont := context.WithValue(context.Background(), domain.Key("login"), login)
							err = h.srv.UpdateOrder(cont, accrualOrder)
							if err != nil {
								util.GetLogger().Infoln(err)
								wg.Done()
								return
							}
							if accrualOrder.Status == "PROCESSED" || accrualOrder.Status == "INVALID" {
								util.GetLogger().Infoln("got", accrualOrder.Status)
								wg.Done()
								return
							}
						}
					} else if resp.StatusCode == http.StatusNoContent {
						util.GetLogger().Infoln("order was", orderN, "not registred by accrual service")
						util.GetLogger().Infoln(accrualOrder)
					}
					time.Sleep(time.Second * 2)
				}
			}()
		}

		c.Response().WriteHeader(http.StatusAccepted)
		return nil
	}
}

func (h *user) ReadOrders(c echo.Context) error {
	page, err := strconv.Atoi(c.Request().Header.Get("page"))
	if err != nil && c.Request().Header.Get("page") != "" {
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}
	if c.Request().Header.Get("page") == "" || page == 0 {
		page = 1
	} else if page < 0 {
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	ctx := context.WithValue(c.Request().Context(), domain.Key("page"), page)
	c.SetRequest(c.Request().WithContext(ctx))

	orders, err := h.srv.ReadOrders(c.Request().Context())
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return err
	}

	if len(orders) == 0 {
		c.Response().WriteHeader(http.StatusNoContent)
		return nil
	}

	var ordersBytes []byte
	buf := bytes.NewBuffer(ordersBytes)
	err = json.NewEncoder(buf).Encode(orders)
	if err != nil {
		util.GetLogger().Errorln(err)
		c.Response().WriteHeader(http.StatusInternalServerError)
		return err
	}
	c.Response().Header().Set("Content-Type", "application/json")

	if len(buf.Bytes()) > 1024 {
		acceptsEncoding := c.Request().Header.Values("Accept-Encoding")
		for _, encoding := range acceptsEncoding {
			if strings.Contains(encoding, "gzip") {
				c.Response().Header().Set(echo.HeaderContentEncoding, "gzip")
				gz := gzip.NewWriter(c.Response().Writer)
				defer gz.Close()

				c.Response().Writer = domain.GzipResponseWriter{
					Writer:         gz,
					ResponseWriter: c.Response().Writer,
				}
				util.GetLogger().Infoln("gzip used")
				break
			}
		}
	}

	c.Response().Write(buf.Bytes())
	return nil
}

func (h *user) ReadBalance(c echo.Context) error {
	balance, err := h.srv.ReadBalance(c.Request().Context())
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return err
	}

	c.Response().Header().Set("Content-Type", "application/json")
	c.Response().Write(balance.Marshal())
	return nil
}

func (h *user) AddWithdrawal(c echo.Context) error {
	util.GetLogger().Infoln("requested withdrawal")
	if !IsJSONContentTypeCorrect(c.Request()) {
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	defer c.Request().Body.Close()

	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	bufCopy := append([]byte(nil), b...)
	reader := bytes.NewReader(bufCopy)

	err = util.CheckDuplicatesInJSON(json.NewDecoder(reader), nil)
	if err != nil {
		util.GetLogger().Infoln("json with duplicate properties")
		err = domain.ErrorRequestFormatIncorrect
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	c.Request().Body = io.NopCloser(bytes.NewBuffer(b))

	var withdrawal domain.Withdrawal

	d := json.NewDecoder(c.Request().Body)
	d.DisallowUnknownFields()

	if err := d.Decode(&withdrawal); err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		util.GetLogger().Infoln(err)
		return err
	}

	for _, orderNumSymbol := range withdrawal.OrderNumber {
		_, err := strconv.Atoi(string(orderNumSymbol))
		if err != nil {
			c.Response().WriteHeader(http.StatusBadRequest)
			return err
		}
	}

	if withdrawal.WithdrawalAmount.Withdrawal < 0 {
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	err = h.srv.AddWithdrawal(c.Request().Context(), withdrawal)
	if err != nil {
		if errors.Is(err, domain.ErrorNotEnoughPoints) {
			c.Response().WriteHeader(http.StatusPaymentRequired)
			return err
		} else if errors.Is(err, domain.ErrorIncorrectOrderNumber) {
			c.Response().WriteHeader(http.StatusUnprocessableEntity)
			return err
		}
		util.GetLogger().Infoln(err)
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}
	c.Response().WriteHeader(http.StatusOK)
	return nil
}

func (h *user) HandleStartup(serverAddress string, wg *sync.WaitGroup) error {
	for batchNumber := 0; ; batchNumber++ {
		unprocessedOrders, err := h.srv.GetUnprocessedBatch(context.Background(), batchNumber)
		if err != nil {
			return err
		}

		if len(unprocessedOrders) == 0 {
			util.GetLogger().Infoln("no unprocessed orders left")
			break
		}

		for _, order := range unprocessedOrders {
			ord := order
			wg.Add(1)
			go func() {
				accrualWithEndpoint := serverAddress + "/api/orders/" + ord.Accrual.Order
				login := ord.Username
				var previousAccrualOrder domain.AccrualOrder
				var accrualOrder domain.AccrualOrder

				util.GetLogger().Infoln(accrualOrder)

				for {
					util.GetLogger().Infoln("requested", accrualWithEndpoint)
					resp, err := http.Get(accrualWithEndpoint)
					if err != nil {
						util.GetLogger().Infoln(err)
						wg.Done()
						return
					}
					defer resp.Body.Close()

					var body []byte

					if resp.StatusCode != http.StatusNoContent {
						body, err = io.ReadAll(resp.Body)
						if err != nil {
							util.GetLogger().Infoln(err)
							return
						}
					}

					if body != nil {
						err = json.Unmarshal(body, &accrualOrder)
						if err != nil {
							util.GetLogger().Infoln(err)
							wg.Done()
							return
						}

						util.GetLogger().Infoln(accrualOrder)

						if accrualOrder.Status != previousAccrualOrder.Status {
							previousAccrualOrder = accrualOrder
							cont := context.WithValue(context.Background(), domain.Key("login"), login)
							err = h.srv.UpdateOrder(cont, accrualOrder)
							if err != nil {
								util.GetLogger().Infoln(err)
								wg.Done()
								return
							}
							if accrualOrder.Status == "PROCESSED" || accrualOrder.Status == "INVALID" {
								util.GetLogger().Infoln("got", accrualOrder.Status)
								wg.Done()
								return
							}
						}
					} else if resp.StatusCode == http.StatusNoContent {
						util.GetLogger().Infoln("order", ord.Accrual.Order, "was not registred by accrual service")
						util.GetLogger().Infoln(accrualOrder)
					}
					time.Sleep(time.Second * 2)
				}
			}()
		}
	}

	return nil
}

func (h *user) ReadWithdrawals(c echo.Context) error {
	page, err := strconv.Atoi(c.Request().Header.Get("page"))
	if err != nil && c.Request().Header.Get("page") != "" {
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}
	if c.Request().Header.Get("page") == "" || page == 0 {
		page = 1
	} else if page < 0 {
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	ctx := context.WithValue(c.Request().Context(), domain.Key("page"), page)
	c.SetRequest(c.Request().WithContext(ctx))

	withdrawals, err := h.srv.ReadWithdrawals(c.Request().Context())
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return err
	}

	if len(withdrawals) == 0 {
		c.Response().WriteHeader(http.StatusNoContent)
		return nil
	}

	var withdrawalsBytes []byte
	buf := bytes.NewBuffer(withdrawalsBytes)
	err = json.NewEncoder(buf).Encode(withdrawals)
	if err != nil {
		util.GetLogger().Errorln(err)
		c.Response().WriteHeader(http.StatusInternalServerError)
		return err
	}
	c.Response().Header().Set("Content-Type", "application/json")

	if len(buf.Bytes()) > 1024 {
		acceptsEncoding := c.Request().Header.Values("Accept-Encoding")
		for _, encoding := range acceptsEncoding {
			if strings.Contains(encoding, "gzip") {
				c.Response().Header().Set(echo.HeaderContentEncoding, "gzip")
				gz := gzip.NewWriter(c.Response().Writer)
				defer gz.Close()

				c.Response().Writer = domain.GzipResponseWriter{
					Writer:         gz,
					ResponseWriter: c.Response().Writer,
				}
				util.GetLogger().Infoln("gzip used")
				break
			}
		}
	}

	c.Response().Write(buf.Bytes())
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
		util.GetLogger().Infoln("could not create token", err)
		return "", err
	}
	return tokenString, err
}
