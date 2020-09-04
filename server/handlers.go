package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/valyala/fastjson"
	"github.com/vishal1132/cafebucks/eventbus"
	hl "github.com/vishal1132/cafebucks/handlers"
)

type handler struct {
	l *zerolog.Logger
}

type ctxKey uint8

const maxBodySize = 2 * 1024 * 1024 // 2MB

const (
	ctxKeyReqID ctxKey = iota
)

func (s *server) registerHandlers() {
	h := handler{l: &s.Logger}
	s.Mux.HandleFunc("/orderservice/_health_/check", h.handleHealth)
	s.Mux.HandleFunc("/order", h.handleCreateOrder).Methods("POST")
	s.Mux.HandleFunc("/orders", h.handleGetOrders).Methods("GET")
}

func (h *handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "healthy")
}

func (h *handler) handleGetOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.l.With().Str("context", "get order event")
	rid, ok := hl.CtxRequestID(ctx)
	if ok {
		l = l.Str("request id", rid)
	}
	lg := l.Logger()

	if len(orderMap) == 0 {
		io.WriteString(w, "We didn't get any orders yet")
		lg.Info().
			Str("Order Count", "0").
			Msg("No orders Available")
		return
	}
	for _, v := range orderMap {
		io.WriteString(w, fmt.Sprintf("Order ID %d Coffee %s Price %v Status %s\n", v.OrderID, v.Cof.Name, v.Cof.Price, v.Status))
		lg.Info().
			Str("OrderId", strconv.Itoa(v.OrderID)).
			Str("Coffee", v.Cof.Name).
			Str("Price", strconv.FormatFloat(v.Cof.Price, 'g', 3, 64)).
			Str("Status", string(v.Status)).
			Msg("")
	}
}

func (h *handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	l := h.l.With().Str("context", "create order event")

	rid, ok := hl.CtxRequestID(ctx)
	if ok {
		l = l.Str("request_id", rid)
	}

	lg := l.Logger()

	mtype, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		lg.Error().
			Err(err).
			Msg("Failed to parse content type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if mtype != "application/json" {
		lg.Error().
			Str("content_type", mtype).
			Msg("content type was not JSON")
		io.WriteString(w, "Not Json Type")
		w.Header().Set("Accept", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Now unmarshaling the body into Coffee struct from cafebucks/eventbus
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		lg.Error().
			Err(err).
			Msg("failed to read request body")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	document, err := fastjson.ParseBytes(body)
	if err != nil {
		lg.Error().
			Err(err).
			Msg("failed to unmarshal JSON document")

		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	coffee, err := hl.GetJSONString(document, "coffee")
	if err != nil {
		lg.Error().
			Err(err).
			Msg("no coffee field in request body")
		return
	}

	price, exist := checkExist(coffee)
	if !exist {
		io.WriteString(w, "Sorry, your coffee not available yet in our café.")
		lg.Error().
			Str("Coffee", coffee).
			Err(fmt.Errorf("Coffee Not Available"))
		return
	}

	io.WriteString(w, "Order Placed")
	// Deduct Payment And Create Order

	var order = eventbus.Order{}

	order.OrderID = len(orderMap) + 1
	order.Cof.Name = coffee
	order.Cof.Price = price
	order.Status = eventbus.OrderReceived

	lg.Info().
		Str("OrderID", strconv.Itoa(order.OrderID)).
		Str("Coffee", order.Cof.Name).
		Str("Price", strconv.FormatFloat(order.Cof.Price, 'g', 3, 64)).
		Str("Status", string(order.Status)).
		Str("Event", string(order.Status)).
		Msg("Order Placed")

	orderMap[order.OrderID] = order

	return
}

func checkExist(coffee string) (float64, bool) {
	for _, val := range coffees {
		if val.Name == coffee {
			return val.Price, true
		}
	}
	return 0, false
}
