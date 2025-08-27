package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/MoSed3/otp-server/controller"
	"github.com/MoSed3/otp-server/middleware"
)

var userController = controller.NewUser(controller.OperatorApi)

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

type RequestOTPRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type RequestOTPResponse struct {
	Token string `json:"token"`
}

type VerifyOTPRequest struct {
	Code string `json:"code"`
}

type VerifyOTPResponse struct {
	JWT string `json:"jwt"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type GetCurrentUserResponse struct {
	ID          uint   `json:"id"`
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
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

func (r RequestOTPRequest) validatePhoneNumber() bool {
	return phoneRegex.MatchString(r.PhoneNumber)
}

// requestOTP godoc
// @Summary Request OTP for phone number
// @Description Creates or finds user by phone number and sends OTP for authentication
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RequestOTPRequest true "Phone number in international format"
// @Success 200 {object} RequestOTPResponse "OTP token generated successfully"
// @Failure 400 {object} map[string]string "Invalid request format or phone number"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/request-otp [post]
func requestOTP(w http.ResponseWriter, r *http.Request) {
	var req RequestOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.PhoneNumber == "" {
		http.Error(w, "phone_number is required", http.StatusBadRequest)
		return
	}

	if !req.validatePhoneNumber() {
		http.Error(w, "phone_number must be in international format (e.g., +1234567890)", http.StatusBadRequest)
		return
	}

	token, err := userController.Login(r.Context(), r, req.PhoneNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
// @Failure 400 {object} map[string]string "Invalid request format or OTP code"
// @Failure 401 {object} map[string]string "Invalid bearer token or OTP code"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/verify-otp [post]
func verifyOTP(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Bearer token required", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}

	if len(req.Code) != 6 {
		http.Error(w, "code must be exactly 6 characters", http.StatusBadRequest)
		return
	}

	user, err := userController.VerifyOTP(r.Context(), r, token, req.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	jwtToken, err := middleware.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
		return
	}

	response := VerifyOTPResponse{
		JWT: jwtToken,
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
// @Security BearerAuth
// @Success 200 {object} GetCurrentUserResponse "Current user information"
// @Failure 401 {object} map[string]string "Unauthorized - invalid or missing JWT token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /user/profile [get]
func getCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromRequest(r)

	response := GetCurrentUserResponse{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// updateProfile godoc
// @Summary Update user profile
// @Description Updates the authenticated user's first name and last name
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "User profile information to update"
// @Success 200 {object} GetCurrentUserResponse "Updated user information"
// @Failure 400 {object} map[string]string "Invalid request format or validation error"
// @Failure 401 {object} map[string]string "Unauthorized - invalid or missing JWT token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /user/profile [put]
func updateProfile(w http.ResponseWriter, r *http.Request) {
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
	if err := user.Update(tx, req.FirstName, req.LastName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := GetCurrentUserResponse{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
