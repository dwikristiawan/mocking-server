package users_svc

import (
	"context"
	"encoding/json"
	"fmt"
	"mocking-server/internal/model"
	"mocking-server/internal/repository/postgres/users"
	"mocking-server/internal/security"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type service struct {
	repository users.Repository
	jwtService security.JwtService
}

func NewService(
	repository users.Repository,
	jwtService security.JwtService) Service {
	return &service{
		repository: repository,
		jwtService: jwtService}
}

type Service interface {
	UserInquiryService(context.Context, *model.Users) (*[]model.Users, error)
	AddUserService(context.Context, *model.Users) error
	UpdateUsersService(context.Context, *model.Users) error
	RemoveUsersService(context.Context, *model.Users) error
	LoginService(context.Context, *model.Users) (*security.Tokens, error)
}

func (svc service) UserInquiryService(c context.Context, req *model.Users) (*[]model.Users, error) {
	response, err := svc.repository.SelectUser(c, req)
	return response, err
}

func (svc service) AddUserService(c context.Context, req *model.Users) error {
	if req.Password != "" {
		hashedPassword, err := security.StrHashing(req.Password)
		if err != nil {
			log.Errorf("Err InsertService.security.StrHashing Err > %v", err)
			return err
		}
		req.Password = hashedPassword
		strRole := `{"none": "none"}`
		jsonRole, err := json.Marshal(strRole)
		if err != nil {
			log.Errorf("AddUserService")
			return err
		}
		req.Roles = append(req.Roles, jsonRole...)
	}
	err := svc.repository.InsertUser(c, req)
	return err
}
func (svc service) UpdateUsersService(c context.Context, req *model.Users) error {
	return svc.repository.UpdateUser(c, req)
}
func (svc service) RemoveUsersService(c context.Context, req *model.Users) error {
	return svc.repository.DeleteUser(c, req)
}
func (svc service) LoginService(c context.Context, req *model.Users) (*security.Tokens, error) {

	users, err := svc.repository.SelectUser(c, &model.Users{Username: req.Username})
	if err != nil {
		return nil, err
	}
	if len(*users) < 1 {
		return nil, fmt.Errorf("not found")
	}
	user := (*users)[0]
	if err = security.CompareHashingData(user.Password, req.Password); err != nil {
		log.Errorf("LoginService.CompareHashingData Err: %v, username : %s", err, req.Username)
		return nil, err
	}
	token, err := svc.jwtService.CreateTokens(c, &security.JwtCustomClaims{
		Uuid:           uuid.NewString(),
		Id:             user.Id,
		Username:       user.Username,
		Name:           user.Name,
		Roles:          user.Roles,
		StandardClaims: jwt.StandardClaims{},
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}
