package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"

	"github.com/MoSed3/otp-server/internal/middleware"
	"github.com/MoSed3/otp-server/internal/models"
	"github.com/MoSed3/otp-server/internal/service"
	"github.com/MoSed3/otp-server/internal/token"
)

type AdminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SearchUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"`
}

func CreateSearchUsersResponse(users []models.User, total int64) SearchUsersResponse {
	userResponses := make([]UserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, UserToResponse(&user))
	}
	return SearchUsersResponse{
		Users: userResponses,
		Total: total,
	}
}

type UserStatusUpdateRequest struct {
	Status models.UserStatus `json:"status"`
}

func (r *UserStatusUpdateRequest) Validate() error {
	switch r.Status {
	case models.UserStatusActive, models.UserStatusDisabled:
		return nil
	default:
		return errors.New("invalid user status")
	}
}

func (r *AdminLoginRequest) Validate() error {
	if r.Username == "" || r.Password == "" {
		return errors.New("username and password are required")
	}
	return nil
}

type AdminLoginResponse struct {
	Token string `json:"token"`
}

type GetCurrentAdminResponse struct {
	ID       uint             `json:"id"`
	Username string           `json:"username"`
	Role     models.AdminRole `json:"role"`
}

// AdminHandler handles admin-related HTTP requests.
type AdminHandler struct {
	adminService service.AdminService
	jwtService   *token.JWTService
	decoder      *schema.Decoder
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(adminService service.AdminService, jwtService *token.JWTService) *AdminHandler {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true) // Ignore unknown keys to prevent errors from other query params
	return &AdminHandler{
		adminService: adminService,
		jwtService:   jwtService,
		decoder:      decoder,
	}
}

// adminLogin godoc
// @Summary Admin login
// @Description Authenticates an admin user and returns a JWT token
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body AdminLoginRequest true "Admin credentials"
// @Success 200 {object} AdminLoginResponse "JWT token for authenticated admin"
// @Failure 400 {string} string "Invalid request format or missing credentials"
// @Failure 401 {string} string "Invalid username or password"
// @Failure 500 {string} string "Internal server error"
// @Router /auth/admin [post]
func (h *AdminHandler) adminLogin(w http.ResponseWriter, r *http.Request) {
	var req AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	admin, err := h.adminService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	jwtToken, err := h.jwtService.GenerateToken(admin.ID, token.AudianceAdmin)
	if err != nil {
		http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
		return
	}

	response := AdminLoginResponse{
		Token: jwtToken,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// getCurrentAdmin godoc
// @Summary Get current authenticated admin
// @Description Returns current admin information for authenticated requests
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuthAdmin
// @Success 200 {object} GetCurrentAdminResponse "Current admin information"
// @Failure 401 {string} string "Unauthorized - invalid or missing JWT token"
// @Failure 500 {string} string "Internal server error"
// @Router /admin/profile [get]
func (h *AdminHandler) getCurrentAdmin(w http.ResponseWriter, r *http.Request) {
	admin := middleware.GetAdminFromRequest(r)

	response := GetCurrentAdminResponse{
		ID:       admin.ID,
		Username: admin.Username,
		Role:     admin.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// searchUsers godoc
// @Summary Search users
// @Description Search users by phone number, first name, or last name (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param id query int false "User ID"
// @Param phone_number query string false "User phone number"
// @Param first_name query string false "User first name"
// @Param last_name query string false "User last name"
// @Param status query int false "User status (1: Active, 2: Disabled)" Enums(1,2)
// @Param limit query int false "Limit for pagination (max 100)"
// @Param offset query int false "Offset for pagination"
// @Param sort_by query string false "Sort by field" Enums(id,phone_number,first_name,last_name,status)
// @Param sort_order query string false "Sort order" Enums(asc, desc)
// @Security BearerAuthAdmin
// @Success 200 {object} SearchUsersResponse "List of matching users with total count"
// @Failure 400 {string} string "Invalid request format"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden: Insufficient privileges"
// @Failure 500 {string} string "Internal server error"
// @Router /admin/users [get]
func (h *AdminHandler) searchUsers(w http.ResponseWriter, r *http.Request) {
	var params models.UserSearchParams
	err := h.decoder.Decode(&params, r.URL.Query())
	if err != nil {
		http.Error(w, "Invalid query parameters: "+err.Error(), http.StatusBadRequest)
		return
	}

	params.SetDefaults()

	// Validate status if provided
	if params.Status != nil {
		userStatus := *params.Status
		if !userStatus.IsValid() {
			http.Error(w, "Invalid user status value", http.StatusBadRequest)
			return
		}
	}

	tx := middleware.GetTxFromRequest(r)
	users, total, err := h.adminService.SearchUsers(tx, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := CreateSearchUsersResponse(users, total)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// getUserByID godoc
// @Summary Get a single user by ID
// @Description Get a user's details by their ID (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Security BearerAuthAdmin
// @Success 200 {object} UserResponse "User details"
// @Failure 400 {string} string "Invalid user ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden: Insufficient privileges"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Router /admin/user/{id} [get]
func (h *AdminHandler) getUserByID(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	tx := middleware.GetTxFromRequest(r)
	user, err := h.adminService.GetUserByID(tx, uint(userID))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := UserToResponse(user)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// updateUserStatus godoc
// @Summary Update user status
// @Description Disable or activate a user (sudo/super admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body UserStatusUpdateRequest true "New user status"
// @Security BearerAuthAdmin
// @Success 200 {object} UserResponse "User details"
// @Failure 400 {string} string "Invalid request format or user ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden: Insufficient privileges"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Router /admin/user/{id}/status [patch]
func (h *AdminHandler) updateUserStatus(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UserStatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx := middleware.GetTxFromRequest(r)
	user, err := h.adminService.UpdateUserStatus(tx, uint(userID), req.Status)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := UserToResponse(user)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
