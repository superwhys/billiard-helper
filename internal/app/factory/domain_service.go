package factory

import (
	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/domain/user"
)

type DomainServiceFactory struct {
	rdb            *redisutils.RedisClient
	verifyCodeRepo user.IVerifyCodeRepository
}

func NewDomainServiceFactory(rdb *redisutils.RedisClient, verifyCodeRepo user.IVerifyCodeRepository) *DomainServiceFactory {
	return &DomainServiceFactory{
		rdb:            rdb,
		verifyCodeRepo: verifyCodeRepo,
	}
}

func (f *DomainServiceFactory) UserService(repoFactory IRepoFactory) user.IUserService {
	return user.NewUserService(repoFactory.UserRepo(), f.verifyCodeRepo)
}

func (f *DomainServiceFactory) MatchService(repoFactory IRepoFactory) match.IMatchService {
	return match.NewMatchService(
		repoFactory.MatchRepo(),
		repoFactory.PlayerRepo(),
		repoFactory.MatchGameRepo(),
	)
}
