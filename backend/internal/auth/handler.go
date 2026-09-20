package auth

import (
	"errors"
	"net/http"

	"github.com/rohitjoshi6/livebid/backend/internal/httpx"
	"github.com/rohitjoshi6/livebid/backend/internal/users"
)

type Handler struct {
	service *Service
	users   *users.Repository
}

func NewHandler(service *Service, userRepo *users.Repository) *Handler {
	return &Handler{service: service, users: userRepo}
}

type registerRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := h.service.Register(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRegistration):
			httpx.WriteError(w, http.StatusBadRequest, "invalid_registration", "Use a valid email, username, and password of at least 8 characters.")
		case errors.Is(err, users.ErrUserExists):
			httpx.WriteError(w, http.StatusConflict, "user_exists", "A user with that email or username already exists.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "registration_failed", "Could not register user.")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Email or password is incorrect.")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "login_failed", "Could not log in.")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	user, err := h.users.FindByID(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, user.Public())
}
