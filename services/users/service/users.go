package service

import (
	"context"

	"github.com/hungq1205/services/common/genproto/users"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) LogIn(e context.Context)
