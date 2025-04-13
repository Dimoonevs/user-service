package route

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Dimoonevs/go-prometheus-metrics/metrics"
	"github.com/Dimoonevs/user-service/app/internal/models"
	"github.com/Dimoonevs/user-service/app/internal/service"
	"github.com/Dimoonevs/user-service/app/pkg/jwt"
	"github.com/Dimoonevs/video-service/app/pkg/respJSON"
	"github.com/valyala/fasthttp"
	"log"
	"strings"
)

type checkResponse struct {
	ID    float64 `json:"id"`
	Email string  `json:"email"`
}

func RequestHandler(ctx *fasthttp.RequestCtx) {
	if string(ctx.Method()) == fasthttp.MethodOptions {
		ctx.SetStatusCode(fasthttp.StatusOK)
		return
	}

	path := string(ctx.URI().Path())

	if !strings.HasPrefix(path, "/users") {
		respJSON.WriteJSONError(ctx, fasthttp.StatusNotFound, nil, "Endpoint not found")
		return
	}

	remainingPath := path[len("/users"):]

	if strings.HasPrefix(remainingPath, "/check") {
		jwt.JWTMiddleware(func(ctx *fasthttp.RequestCtx) {
			handleCheckRoutes(ctx)
		})(ctx)
		return
	}

	if strings.HasPrefix(remainingPath, "/settings") {
		jwt.JWTMiddleware(func(ctx *fasthttp.RequestCtx) {
			handleSettingsRoutes(ctx)
		})(ctx)
		return
	}

	switch {
	case remainingPath == "/register" && ctx.IsPost():
		handleUserRegister(ctx)
	case remainingPath == "/verify" && ctx.IsPost():
		handleUserVerify(ctx)
	case remainingPath == "/login" && ctx.IsPost():
		handleUserLogin(ctx)
	case remainingPath == "/code" && ctx.IsPost():
		handleSendVerificationEmailAgain(ctx)
	case remainingPath == "/request/reset/password" && ctx.IsPost():
		handleRequestResetPassword(ctx)
	case remainingPath == "/confirm/reset/password" && ctx.IsPost():
		handleConfirmResetPassword(ctx)
	default:
		respJSON.WriteJSONError(ctx, fasthttp.StatusNotFound, nil, "Endpoint not found")
	}

}
func handleSettingsRoutes(ctx *fasthttp.RequestCtx) {
	switch {
	case ctx.IsPost():
		handleSetUserSettings(ctx)
	case ctx.IsGet():
		handleGetUserSettings(ctx)
	case ctx.IsPatch():
		handleUpdateUserSettings(ctx)
	default:
		respJSON.WriteJSONError(ctx, fasthttp.StatusNotFound, nil, "Endpoint not found")
	}
}

func handleCheckRoutes(ctx *fasthttp.RequestCtx) {
	switch {
	case ctx.IsGet():
		handleCheckConnect(ctx)
	default:
		respJSON.WriteJSONError(ctx, fasthttp.StatusNotFound, nil, "Endpoint not found")
	}
}

func handleUserRegister(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()

	var req models.UsersReq

	if err := json.Unmarshal(body, &req); err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Invalid JSON body")
		return
	}
	if req.Email == "" || req.Password == "" {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, nil, "Email or Password is required")
		return
	}

	userID, err := service.RegisterUser(req)
	if err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Failed to register user")
		return
	}

	metrics.UserRegistered.Inc()
	respJSON.WriteJSONResponse(ctx, fasthttp.StatusCreated, "Created user and send code to email", userID)
}

func handleUserVerify(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()
	var req models.UsersReq

	if err := json.Unmarshal(body, &req); err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Invalid JSON body")
		return
	}

	if req.Email == "" || req.Code == "" {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, nil, "Email or Password is required")
		return
	}
	if err := service.VerifyCode(req); err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Failed to verify code")
		return
	}

	metrics.UserVerified.Inc()
	respJSON.WriteJSONResponse(ctx, fasthttp.StatusOK, "Verify user successful", req.Email)
}

func handleUserLogin(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()
	var req models.UsersReq
	if err := json.Unmarshal(body, &req); err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Invalid JSON body")
		return
	}
	if req.Email == "" || req.Password == "" {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, nil, "Email or Password is required")
		return
	}

	token, err := service.LoginUser(req)
	if err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Failed to login")
		return
	}
	respJSON.WriteJSONResponse(ctx, fasthttp.StatusOK, "Login successful", token)
}

func handleRequestResetPassword(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()
	var req models.UsersReq
	if err := json.Unmarshal(body, &req); err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Invalid JSON body")
		return
	}
	if req.Email == "" {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, nil, "Email is required")
		return
	}
	if err := service.RequestResetPassword(req.Email); err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Failed to reset password")
		return
	}
	respJSON.WriteJSONResponse(ctx, fasthttp.StatusOK, "Reset password successful", req.Email)
}

func handleConfirmResetPassword(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()
	var req models.UsersReq
	if err := json.Unmarshal(body, &req); err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Invalid JSON body")
		return
	}
	if req.Email == "" || req.Code == "" || req.Password == "" {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, nil, "Email or Code or Password is required")
	}
	if err := service.ConfirmResetPassword(req); err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Failed to reset password")
		return
	}
	respJSON.WriteJSONResponse(ctx, fasthttp.StatusOK, "Reset password successful", req.Email)
}

func handleSendVerificationEmailAgain(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()
	var req models.UsersReq
	if err := json.Unmarshal(body, &req); err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Invalid JSON body")
		return
	}
	if req.Email == "" {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, nil, "Email is required")
		return
	}
	if err := service.SendVerificationEmailAgain(req.Email); err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Failed to send verification email")
		return
	}
	respJSON.WriteJSONResponse(ctx, fasthttp.StatusOK, "Sending verification email successful", req.Email)
}

func handleSetUserSettings(ctx *fasthttp.RequestCtx) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusUnauthorized, err, "Error getting user id: ")
		return
	}
	body := ctx.PostBody()
	var req models.UserSettings
	if err = json.Unmarshal(body, &req); err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Invalid JSON body")
		return
	}
	if req.Name == "" {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, nil, "Name is required")
		return
	}
	err = service.UserSettings(userID, req)
	if err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Failed to set user settings")
		return
	}
	respJSON.WriteJSONResponse(ctx, fasthttp.StatusOK, "Set user settings successful", nil)
}

func handleGetUserSettings(ctx *fasthttp.RequestCtx) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusUnauthorized, err, "Error getting user id: ")
		return
	}
	resp, err := service.GetUserSettings(userID)
	if err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Failed to get user settings")
		return
	}
	respJSON.WriteJSONResponse(ctx, fasthttp.StatusOK, "Get user settings successful", resp)
}

func handleUpdateUserSettings(ctx *fasthttp.RequestCtx) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusUnauthorized, err, "Error getting user id: ")
		return
	}
	body := ctx.PostBody()
	var req models.UserSettings
	if err = json.Unmarshal(body, &req); err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Invalid JSON body")
		return
	}
	if req.ID == 0 {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, nil, "ID is required")
		return
	}
	if err = service.UpdateUserSettings(userID, req); err != nil {
		respJSON.WriteJSONError(ctx, fasthttp.StatusBadRequest, err, "Failed to update user settings")
		return
	}
	respJSON.WriteJSONResponse(ctx, fasthttp.StatusOK, "Update user settings successful", nil)
}

func handleCheckConnect(ctx *fasthttp.RequestCtx) {
	id, ok := ctx.UserValue("userID").(float64)
	if !ok {
		respJSON.WriteJSONError(ctx, fasthttp.StatusUnauthorized, errors.New("userID not found or wrong type"), "Invalid token")
		return
	}

	email, ok := ctx.UserValue("email").(string)
	if !ok {
		log.Println("Email not found or wrong type")
		respJSON.WriteJSONError(ctx, fasthttp.StatusUnauthorized, errors.New("email not found or wrong type"), "Invalid token")
		return
	}

	resp := checkResponse{
		ID:    id,
		Email: email,
	}

	log.Printf("Authenticated user: %v", resp)
	respJSON.WriteJSONResponse(ctx, fasthttp.StatusOK, "User is authenticated", resp)
}

func getUserIDFromContext(ctx *fasthttp.RequestCtx) (int, error) {
	userIDValue := ctx.UserValue("userID")
	userIDFloat, ok := userIDValue.(float64)
	if !ok {
		return 0, fmt.Errorf("invalid userID format: %f", userIDFloat)
	}

	return int(userIDFloat), nil
}
