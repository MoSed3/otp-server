package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/MoSed3/otp-server/internal/middleware"
	"github.com/MoSed3/otp-server/internal/models"
	"github.com/MoSed3/otp-server/internal/service"
	"github.com/MoSed3/otp-server/internal/token"
)

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

type RequestOTPRequest struct {
	PhoneNumber string `json:"phone_number"`
}

func (r RequestOTPRequest) validatePhoneNumber() bool {
	return phoneRegex.MatchString(r.PhoneNumber)
}

type RequestOTPResponse struct {
	Token string `json:"token"`
}

type VerifyOTPRequest struct {
	Code string `json:"code"`
}

func (v VerifyOTPRequest) validate() error {
	if len(v.Code) != 6 {
		return errors.New("code must be exactly 6 characters")
	}
	return nil
}

type VerifyOTPResponse struct {
	Token string `json:"token"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (r UpdateProfileRequest) validate() error {
	if len(r.FirstName) > 100 {
		return errors.New("first_name cannot exceed 100 characters")
	}
	if len(r.LastName) > 100 {
		return errors.New("last_name cannot exceed 100 characters")
	}
	return nil
}

type UserResponse struct {
	ID          uint   `json:"id"`
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Status      int    `json:"status"`
}

func UserToResponse(u *models.User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		PhoneNumber: u.PhoneNumber,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Status:      u.Status.Int(),
	}
}

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService service.UserService
	jwtService  *token.JWTService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService service.UserService, jwtService *token.JWTService) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

// requestOTP godoc
// @Summary Request OTP for phone number
// @Description Creates or finds user by phone number and sends OTP for authentication
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RequestOTPRequest true "Phone number in international format"
// @Success 200 {object} RequestOTPResponse "OTP token generated successfully"
// @Failure 400 {string} string "Invalid request format or phone number"
// @Failure 500 {string} string "Internal server error"
// @Router /auth/request-otp [post]
func (h *UserHandler) requestOTP(w http.ResponseWriter, r *http.Request) {
	var req RequestOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if !req.validatePhoneNumber() {
		http.Error(w, "phone_number must be in international format (e.g., +1234567890)", http.StatusBadRequest)
		return
	}

	token, err := h.userService.Login(r.Context(), r, req.PhoneNumber)
	if err != nil {
		if errors.Is(err, service.ErrUserDisabled) {
			http.Error(w, err.Error(), http.StatusForbidden)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	response := RequestOTPResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// verifyOTP godoc
// @Summary Verify OTP code
// @Description Verifies the OTP code and returns JWT token for authenticated user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token from request-otp endpoint"
// @Param request body VerifyOTPRequest true "6-character OTP code"
// @Success 200 {object} VerifyOTPResponse "JWT token for authenticated user"
// @Failure 400 {string} string "Invalid request format or OTP code"
// @Failure 401 {string} string "Invalid bearer token or OTP code"
// @Failure 500 {string} string "Internal server error"
// @Router /auth/verify-otp [post]
func (h *UserHandler) verifyOTP(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Bearer token required", http.StatusUnauthorized)
		return
	}

	uidToken := strings.TrimPrefix(authHeader, "Bearer ")

	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := req.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userService.VerifyOTP(r.Context(), r, uidToken, req.Code)
	if err != nil {
		if errors.Is(err, service.ErrUserDisabled) {
			http.Error(w, err.Error(), http.StatusForbidden)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	jwtToken, err := h.jwtService.GenerateToken(user.ID, token.AudianceUser)
	if err != nil {
		http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
		return
	}

	response := VerifyOTPResponse{
		Token: jwtToken,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// getCurrentUser godoc
// @Summary Get current authenticated user
// @Description Returns current user information for authenticated requests
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuthUser
// @Success 200 {object} UserResponse "Current user information"
// @Failure 401 {string} string "Unauthorized - invalid or missing JWT token"
// @Failure 500 {string} string "Internal server error"
// @Router /user/profile [get]
func (h *UserHandler) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromRequest(r)

	response := UserToResponse(user)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// updateProfile godoc
// @Summary Update user profile
// @Description Updates the authenticated user's first name and last name
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuthUser
// @Param request body UpdateProfileRequest true "User profile information to update"
// @Success 200 {object} UserResponse "Updated user information"
// @Failure 400 {string} string "Invalid request format or validation error"
// @Failure 401 {string} string "Unauthorized - invalid or missing JWT token"
// @Failure 500 {string} string "Internal server error"
// @Router /user/profile [put]
func (h *UserHandler) updateProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromRequest(r)

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := req.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx := middleware.GetTxFromRequest(r)
	if err := h.userService.UpdateProfile(tx, user, req.FirstName, req.LastName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := UserToResponse(user)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
