package route

import (
	"encoding/json"
	"github.com/Dimoonevs/user-service/app/internal/models"
	"github.com/Dimoonevs/user-service/app/internal/service"
	"github.com/Dimoonevs/video-service/app/pkg/respJSON"
	"github.com/valyala/fasthttp"
	"strings"
)

func RequestHandler(ctx *fasthttp.RequestCtx) {
	path := string(ctx.URI().Path())

	if !strings.HasPrefix(path, "/users") {
		respJSON.WriteJSONError(ctx, fasthttp.StatusNotFound, nil, "Endpoint not found")
		return
	}

	remainingPath := path[len("/users"):]

	switch {
	case remainingPath == "/register" && ctx.IsPost():
		handleUserRegister(ctx)
	case remainingPath == "/verify" && ctx.IsGet():
		handleUserVerify(ctx)
	case remainingPath == "/login" && ctx.IsPost():
		handleUserLogin(ctx)
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
