package handlers

import (
	"net/http"

	"encoding/json"

	"github.com/slikk66/newspin/internal/auth"
	"github.com/slikk66/newspin/internal/db"
	"github.com/slikk66/newspin/internal/news"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type Handler struct {
	DB   *db.Client
	News *news.Client
}

type CreatePinRequest struct {
	ArticleId string `json:"articleId"`
	Title     string `json:"title"`
	Url       string `json:"url"`
}

func NewHandler(dbClient *db.Client, newsClient *news.Client) *Handler {
	return &Handler{DB: dbClient, News: newsClient}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {

	var loginRequest LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(loginRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "server error creating hash", http.StatusInternalServerError)
		return
	}

	if err = h.DB.CreateUser(loginRequest.Username, string(hash)); err != nil {
		http.Error(w, "server error creating user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "user created"})

}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {

	var loginRequest LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.DB.GetUser(loginRequest.Username)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginRequest.Password)); err != nil {
		http.Error(w, "incorrect password", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(loginRequest.Username)
	if err != nil {
		http.Error(w, "server error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{Token: token})

}

func (h *Handler) GetPins(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(auth.UsernameKey).(string)
	pins, err := h.DB.GetPins(username)
	if err != nil {
		http.Error(w, "server error GetPins", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pins)
}

func (h *Handler) SearchNews(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	articles, err := h.News.Search(query)
	if err != nil {
		http.Error(w, "server error SearchNews", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(articles)
}

func (h *Handler) CreatePin(w http.ResponseWriter, r *http.Request) {

	var createPinRequest CreatePinRequest

	if err := json.NewDecoder(r.Body).Decode(&createPinRequest); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	username := r.Context().Value(auth.UsernameKey).(string)

	pin := &db.Pin{
		UserId:    username,
		ArticleId: createPinRequest.ArticleId,
		Title:     createPinRequest.Title,
		Url:       createPinRequest.Url,
	}

	if err := h.DB.CreatePin(pin); err != nil {
		http.Error(w, "server error CreatePin", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pin)
}

func (h *Handler) DeletePin(w http.ResponseWriter, r *http.Request) {

	username := r.Context().Value(auth.UsernameKey).(string)
	articleId := r.URL.Query().Get("articleId")

	if err := h.DB.DeletePin(username, articleId); err != nil {
		http.Error(w, "server error DeletePin", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
