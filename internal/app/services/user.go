package services

import (
	"context"

	"github.com/superwhys/billiard-helper/config"
	"github.com/superwhys/billiard-helper/internal/app/assembler"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

type UserApp struct {
	userService   user.IUserService
	userAssembler *assembler.UserAssembler
	jwtConfig     *config.JwtConfig
}

func NewUserApp(
	userService user.IUserService,
	userAssembler *assembler.UserAssembler,
	jwtConfig *config.JwtConfig,
) *UserApp {
	return &UserApp{
		userService:   userService,
		userAssembler: userAssembler,
		jwtConfig:     jwtConfig,
	}
}

// SendRegisterCode 发送注册验证码
func (a *UserApp) SendRegisterCode(ctx context.Context, req *dto.SendRegisterCodeReq) error {
	return a.userService.SendRegisterCode(ctx, req.Email)
}

// Register 注册用户
// 流程：调用领域服务注册 -> 返回结果
func (a *UserApp) Register(ctx context.Context, req *dto.RegisterReq) error {
	// 1. 调用领域服务完成核心业务逻辑
	return a.userService.RegisterWithCode(ctx, req.Email, req.Code, req.Password, req.Name)
}

// Login 用户登录
// 流程：调用领域服务验证 -> 生成 Token -> 组装 DTO 返回
func (a *UserApp) Login(ctx context.Context, req *dto.LoginReq) (string, *dto.User, error) {
	// 1. 调用领域服务验证账号密码
	u, err := a.userService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return "", nil, err
	}

	// 2. 生成 JWT Token
	token, err := jwt.GenerateToken(
		[]byte(a.jwtConfig.JwtSecret),
		a.jwtConfig.JwtTimeout,
		u.ID,
	)
	if err != nil {
		return "", nil, err
	}

	// 3. 组装 User DTO
	userDTO := a.userAssembler.ToDTO(u)
	return token, userDTO, nil
}

func (a *UserApp) GetUserTokenClaims(ctx context.Context, tokenStr string) (*jwt.UserTokenClaims, error) {
	return jwt.ParseToken(tokenStr, []byte(a.jwtConfig.JwtSecret))
}
