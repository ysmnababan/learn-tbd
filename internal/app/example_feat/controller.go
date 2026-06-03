package example_feat

import (
	"fmt"
	"learn-tbd/internal/factory"
	"learn-tbd/internal/utils/response"

	"github.com/labstack/echo/v4"
)

type IUserService interface {
	Get(ctx echo.Context) (out []*UserResponse, err error)
	Create(ctx echo.Context, in *UserCreateRequest) (err error)
	Login(ctx echo.Context, req *UserLoginRequest) (out *UserLoginResponse, err error)
}

type handler struct {
	service IUserService
}

func NewHandler(f *factory.Factory) *handler {
	return &handler{
		service: NewService(f),
	}
}

func (h *handler) GetUsers(c echo.Context) error {
	fmt.Println(c.Get("user_id"))
	res, err := h.service.Get(c)
	if err != nil {
		return err
	}
	return response.WithStatusOKResponse(res, c)
}

func (h *handler) CreateUser(c echo.Context) error {
	req := &UserCreateRequest{}
	err := c.Bind(req)
	if err != nil {
		return response.Wrap(response.ErrUnprocessableEntity, fmt.Errorf("binding error: %w", err))
	}

	err = h.service.Create(c, req)
	if err != nil {
		return err
	}

	return response.WithStatusOKResponse("success", c)
}

func (h *handler) Login(c echo.Context) error {
	req := &UserLoginRequest{}
	err := c.Bind(req)
	if err != nil {
		return response.Wrap(response.ErrUnprocessableEntity, fmt.Errorf("binding error: %w", err))
	}

	res, err := h.service.Login(c, req)
	if err != nil {
		return err
	}

	return response.WithStatusOKResponse(res, c)
}
