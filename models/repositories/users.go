package repositories

import (
	"context"
	"errors"
	"sync"

	"github.com/kzs0/kokoro/koko"
	"github.com/kzs0/pill_manager/models"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Users struct {
	c     int64
	mut   sync.RWMutex
	cache map[int64]*models.User
}

func NewUsers() *Users {
	return &Users{
		c:     1,
		mut:   sync.RWMutex{},
		cache: make(map[int64]*models.User),
	}
}

func (u *Users) Put(ctx context.Context, user *models.User) (_ *models.User, err error) {
	ctx, done := koko.Operation(ctx, "repo_put_user")
	defer done(&ctx, &err)

	u.mut.Lock()
	defer u.mut.Unlock()

	user.ID = u.c

	u.cache[u.c] = user

	u.c = u.c + 1

	return user, nil
}

func (u *Users) Get(ctx context.Context, id int64) (user *models.User, err error) {
	ctx, done := koko.Operation(ctx, "repo_get_user")
	defer done(&ctx, &err)

	u.mut.RLock()
	defer u.mut.RUnlock()

	user, ok := u.cache[id]
	if !ok {
		return nil, ErrUserNotFound
	}

	return user, nil
}
