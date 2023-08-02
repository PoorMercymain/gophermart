package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strings"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"

	"github.com/PoorMercymain/gophermart/internal/accrual/calculator"
	"github.com/PoorMercymain/gophermart/internal/accrual/domain"
	"github.com/PoorMercymain/gophermart/internal/accrual/interfaces"
	"github.com/PoorMercymain/gophermart/pkg/util"
)

type StorageHandler struct {
	storage interfaces.Storage
}

func NewStorageHandler(storage interfaces.Storage) *StorageHandler {
	return &StorageHandler{
		storage: storage,
	}
}

func (h *StorageHandler) ProcessGetOrdersRequest(c echo.Context) (err error) {
	orderNumber := c.Param("number")
	util.GetLogger().Infoln(orderNumber)

	//check order number
	err = goluhn.Validate(orderNumber)
	if err != nil {
		util.GetLogger().Infoln(err)
		err = domain.ErrorOrderNotRegistered
		c.Response().WriteHeader(http.StatusNoContent)
		return
	}

	order, err := h.storage.GetOrder(c.Request().Context(), &orderNumber)
	if err != nil {
		util.GetLogger().Infoln(err)
		err = domain.ErrorOrderNotRegistered
		c.Response().WriteHeader(http.StatusNoContent)
		return
	}
	util.GetLogger().Infoln(*order)

	order.Accrual = math.Round(order.Accrual*100) / 100

	util.GetLogger().Infoln("sent from accrual:", order)
	out, err := json.Marshal(order)
	if err != nil {
		util.GetLogger().Infoln(err)
		c.Response().WriteHeader(http.StatusInternalServerError)
		return
	}

	c.Response().Header().Set("Content-Type", "application/json")
	c.Response().Write(out)

	return
}

func (h *StorageHandler) ProcessPostOrdersRequest(c echo.Context) (err error) {
	if !IsJSONContentTypeCorrect(c.Request()) {
		c.Response().WriteHeader(http.StatusBadRequest)
		return
	}

	var order domain.Order

	var buf bytes.Buffer

	_, err = buf.ReadFrom(c.Request().Body)

	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(buf.Bytes(), &order)
	if err != nil {
		util.GetLogger().Infoln(err)
		err = domain.ErrorRequestFormatIncorrect
		c.Response().WriteHeader(http.StatusBadRequest)
		return
	}

	util.GetLogger().Infoln(order)

	util.GetLogger().Infoln(order.Number)
	util.GetLogger().Infoln(order.Goods)
	//check order number
	if order.Number == "" {
		util.GetLogger().Infoln("empty order number")
		err = domain.ErrorRequestFormatIncorrect
		c.Response().WriteHeader(http.StatusBadRequest)
		return
	}
	err = goluhn.Validate(order.Number)
	if err != nil {
		util.GetLogger().Infoln(err)
		err = domain.ErrorRequestFormatIncorrect
		c.Response().WriteHeader(http.StatusBadRequest)
		return
	}

	//register order or check if it's in db
	var orderRecord = domain.OrderRecord{
		Number: order.Number,
		Status: domain.OrderStatusRegistered,
	}
	err = h.storage.StoreOrder(c.Request().Context(), &orderRecord)
	if err != nil {
		util.GetLogger().Infoln(err)
		if errors.Is(err, domain.ErrorOrderAlreadyProcessing) {
			c.Response().WriteHeader(http.StatusConflict)
		}
		return

	}

	err = h.storage.StoreOrderGoods(c.Request().Context(), &order)
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	//enqueue calculation of bonuses in goroutine
	//TODO: make worker pool

	ctx := context.Background()
	go calculator.CalculateAccrual(ctx, &order, h.storage)

	c.Response().WriteHeader(http.StatusAccepted)
	return
}

func (h *StorageHandler) ProcessPostGoodsRequest(c echo.Context) (err error) {
	if !IsJSONContentTypeCorrect(c.Request()) {
		c.Response().WriteHeader(http.StatusBadRequest)
		return
	}

	var goods domain.Goods

	var buf bytes.Buffer

	_, err = buf.ReadFrom(c.Request().Body)

	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(buf.Bytes(), &goods)
	if err != nil {
		util.GetLogger().Infoln(err)
		err = domain.ErrorRequestFormatIncorrect
		c.Response().WriteHeader(http.StatusBadRequest)
		return
	}

	util.GetLogger().Infoln(goods)

	_, err = govalidator.ValidateStruct(goods)

	if err != nil {
		util.GetLogger().Infoln(err)
		err = domain.ErrorRequestFormatIncorrect
		c.Response().WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.storage.StoreGoodsReward(c.Request().Context(), &goods)
	if err != nil {
		util.GetLogger().Infoln(err)
		if errors.Is(err, domain.ErrorMatchAlreadyRegistered) {
			c.Response().WriteHeader(http.StatusConflict)
		}
	}
	return
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
