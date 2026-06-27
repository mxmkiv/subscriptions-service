package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/mxmkiv/subscriptions-service/internal/domain"
	"github.com/mxmkiv/subscriptions-service/internal/service"
)

type Handler struct {
	srv    service.Service
	logger *slog.Logger
}

func New(srv service.Service, logger *slog.Logger) *Handler {
	return &Handler{
		srv:    srv,
		logger: logger,
	}
}

type SubscriptionRequest struct {
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date"`
}

type SubscriptionResponse struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date"`
}

type TotalResponse struct {
	Total int `json:"total"`
}

// helpers func

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)

}

func toSubscriptionResponse(sub domain.Subscription) SubscriptionResponse {

	var endDate *string
	if sub.EndDate != nil {
		format := sub.EndDate.Format("01-2006")
		endDate = &format
	}

	resp := SubscriptionResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   sub.StartDate.Format("01-2006"),
		EndDate:     endDate,
	}

	return resp
}

// Create godoc
// @Summary      Create subscription
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        body  body      SubscriptionRequest  true  "Subscription payload"
// @Success      201   {object}  SubscriptionResponse
// @Failure      400   {string}  string               "invalid request body or input"
// @Failure      500   {string}  string               "internal server error"
// @Router       /subscriptions [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {

	var req SubscriptionRequest

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	dto := service.CreateDTO{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	sub, err := h.srv.Create(r.Context(), dto)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			http.Error(w, "invalid input", http.StatusBadRequest)
		default:
			h.logger.Error("failed to create subscription", "error", err)
			http.Error(w, "failed to create subscription", http.StatusInternalServerError)
		}
		return
	}

	resp := toSubscriptionResponse(*sub)

	writeJSON(w, http.StatusCreated, resp)
}

// GetByID godoc
// @Summary      Get subscription by ID
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      string               true  "Subscription UUID"
// @Success      200  {object}  SubscriptionResponse
// @Failure      400  {string}  string               "invalid id"
// @Failure      404  {string}  string               "subscription not found"
// @Failure      500  {string}  string               "internal server error"
// @Router       /subscriptions/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	reqID := r.PathValue("id")

	id, err := uuid.Parse(reqID)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	sub, err := h.srv.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			http.Error(w, "subscription not found", http.StatusNotFound)
		default:
			h.logger.Error("failed to get subscription", "error", err)
			http.Error(w, "failed to get subscription", http.StatusInternalServerError)
		}
		return
	}

	resp := toSubscriptionResponse(*sub)

	writeJSON(w, http.StatusOK, resp)
}

// Update godoc
// @Summary      Update subscription
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id    path      string               true  "Subscription UUID"
// @Param        body  body      SubscriptionRequest  true  "Subscription payload"
// @Success      200   {object}  SubscriptionResponse
// @Failure      400   {string}  string               "invalid id or input"
// @Failure      404   {string}  string               "subscription not found"
// @Failure      500   {string}  string               "internal server error"
// @Router       /subscriptions/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	reqID := r.PathValue("id")

	id, err := uuid.Parse(reqID)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req SubscriptionRequest

	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	dto := service.UpdateDTO{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	sub, err := h.srv.Update(r.Context(), id, dto)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			http.Error(w, "subscription not found", http.StatusNotFound)
		case errors.Is(err, domain.ErrInvalidInput):
			http.Error(w, "invalid input", http.StatusBadRequest)
		default:
			h.logger.Error("failed to update subscription", "error", err)
			http.Error(w, "failed to update subscription", http.StatusInternalServerError)
		}
		return
	}

	resp := toSubscriptionResponse(*sub)

	writeJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary      Delete subscription
// @Tags         subscriptions
// @Param        id   path      string  true  "Subscription UUID"
// @Success      204
// @Failure      400  {string}  string  "invalid id"
// @Failure      404  {string}  string  "subscription not found"
// @Failure      500  {string}  string  "internal server error"
// @Router       /subscriptions/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	reqID := r.PathValue("id")

	id, err := uuid.Parse(reqID)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.srv.Delete(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			http.Error(w, "subscription not found", http.StatusNotFound)
		default:
			h.logger.Error("failed to delete subscription", "error", err)
			http.Error(w, "failed to delete subscription", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List godoc
// @Summary      List subscriptions
// @Tags         subscriptions
// @Produce      json
// @Param        user_id       query     string               false  "Filter by user UUID"
// @Param        service_name  query     string               false  "Filter by service name"
// @Success      200           {array}   SubscriptionResponse
// @Failure      400           {string}  string               "invalid user_id"
// @Failure      500           {string}  string               "internal server error"
// @Router       /subscriptions [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {

	var filter service.ListFilter

	if userID := r.URL.Query().Get("user_id"); userID != "" {
		id, err := uuid.Parse(userID)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		filter.UserID = &id
	}

	if serviceName := r.URL.Query().Get("service_name"); serviceName != "" {
		filter.ServiceName = &serviceName
	}

	subs, err := h.srv.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to get subscriptions", "error", err)
		http.Error(w, "failed to get subscriptions", http.StatusInternalServerError)
		return
	}

	// convertation
	subsConv := make([]SubscriptionResponse, 0, len(subs))
	for _, sub := range subs {
		subsConv = append(subsConv, toSubscriptionResponse(sub))
	}

	writeJSON(w, http.StatusOK, subsConv)

}

// SumByPeriod godoc
// @Summary      Get total subscription cost for a period
// @Tags         subscriptions
// @Produce      json
// @Param        user_id       query     string         true   "User UUID"
// @Param        start_date    query     string         true   "Period start (MM-YYYY)"
// @Param        end_date      query     string         true   "Period end (MM-YYYY)"
// @Param        service_name  query     string         false  "Filter by service name"
// @Success      200           {object}  TotalResponse
// @Failure      400           {string}  string         "invalid input"
// @Failure      500           {string}  string         "internal server error"
// @Router       /subscriptions/total [get]
func (h *Handler) SumByPeriod(w http.ResponseWriter, r *http.Request) {

	var sumFilter service.SumFilter

	// required
	userIDParam := r.URL.Query().Get("user_id")
	if userIDParam == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	startDate := r.URL.Query().Get("start_date")
	if startDate == "" {
		http.Error(w, "start_date is required", http.StatusBadRequest)
		return
	}

	endDate := r.URL.Query().Get("end_date")
	if endDate == "" {
		http.Error(w, "end_date is required", http.StatusBadRequest)
		return
	}

	sumFilter.UserID = userID
	sumFilter.StartDate = startDate
	sumFilter.EndDate = endDate

	// optional
	if serviceName := r.URL.Query().Get("service_name"); serviceName != "" {
		sumFilter.ServiceName = &serviceName
	}

	sum, err := h.srv.SumByPeriod(r.Context(), sumFilter)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			http.Error(w, "invalid input", http.StatusBadRequest)
		default:
			h.logger.Error("failed to sum subscriptions", "error", err)
			http.Error(w, "failed to sum subscriptions", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, TotalResponse{Total: sum})

}
