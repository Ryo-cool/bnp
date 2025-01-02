package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/my-backend-project/internal/user/auth"
	"github.com/my-backend-project/internal/user/model"
	"github.com/my-backend-project/internal/user/service"
)

type UserHandler struct {
	userService service.UserService
	jwtService  auth.JWTService
}

func NewUserHandler(userService service.UserService, jwtService auth.JWTService) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

func (h *UserHandler) SignUp(c echo.Context) error {
	var req model.SignUpRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.userService.SignUp(c.Request().Context(), &req)
	if err != nil {
		switch err {
		case service.ErrUserAlreadyExists:
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		case service.ErrInvalidEmail:
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		case service.ErrPasswordTooShort:
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *UserHandler) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.userService.Login(c.Request().Context(), &req)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}
	}

	return c.JSON(http.StatusOK, resp)
}

// JWT認証ミドルウェア
func (h *UserHandler) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing authorization token")
		}

		// Bearer トークンの形式を想定
		if len(token) <= 7 || token[:7] != "Bearer " {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token format")
		}

		token = token[7:] // "Bearer "の部分を除去

		// トークンの検証
		claims, err := h.jwtService.ValidateToken(token)
		if err != nil {
			if err == auth.ErrTokenExpired {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token has expired")
			}
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}

		// コンテキストにユーザー情報を設定
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)

		return next(c)
	}
}
