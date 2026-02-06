package factory

import (
	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/domain/user"
)

type DomainServiceFactory struct {
	rdb *redisutils.RedisClient
}

func NewDomainServiceFactory(rdb *redisutils.RedisClient) *DomainServiceFactory {
	return &DomainServiceFactory{
		rdb: rdb,
	}
}

func (f *DomainServiceFactory) UserService(repoFactory IRepoFactory) user.IUserService {
	return user.NewUserService(repoFactory.UserRepo())
}

func (f *DomainServiceFactory) MatchService(repoFactory IRepoFactory) match.IMatchService {
	return match.NewMatchService(repoFactory.MatchRepo(), repoFactory.PlayerRepo())
}
