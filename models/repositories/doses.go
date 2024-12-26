package repositories

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kzs0/kokoro/koko"
	"github.com/kzs0/pill_manager/models"
)

var (
	ErrDosesNotFound = errors.New("doses not found")
	ErrDoseNotFound  = errors.New("dose not found")
)

type Doses struct {
	c     int64
	mut   sync.RWMutex
	cache map[int64]*models.Regimen
}

func NewDoses() *Doses {
	return &Doses{
		c:     1,
		mut:   sync.RWMutex{},
		cache: make(map[int64]*models.Regimen),
	}
}

func (d *Doses) Put(ctx context.Context, rd *models.Regimen) (_ *models.Regimen, err error) {
	ctx, done := koko.Operation(ctx, "repo_put_doses")
	defer done(&ctx, &err)

	d.mut.Lock()
	defer d.mut.Unlock()

	rd.ID = d.c
	d.c++

	for _, dose := range rd.Doses {
		dose.ID = d.c
		d.c++
	}

	d.cache[rd.ID] = rd

	return rd, nil
}

func (d *Doses) Get(ctx context.Context, id int64) (rd *models.Regimen, err error) {
	ctx, done := koko.Operation(ctx, "repo_get_doses")
	defer done(&ctx, &err)

	d.mut.RLock()
	defer d.mut.RUnlock()

	rd, ok := d.cache[id]
	if !ok {
		return nil, ErrDosesNotFound
	}

	return rd, nil
}

func (d *Doses) GetForUser(ctx context.Context, user *models.User) (_ []*models.Regimen, err error) {
	ctx, done := koko.Operation(ctx, "repo_get_doses_for_user")
	defer done(&ctx, &err)

	d.mut.RLock()
	defer d.mut.RUnlock()

	doses := make([]*models.Regimen, 0, 0)

	for _, rd := range d.cache {
		if rd.Patient.ID == user.ID {
			doses = append(doses, rd)
		}
	}

	return doses, nil
}

func (d *Doses) MarkTaken(ctx context.Context, id int64) (err error) {
	ctx, done := koko.Operation(ctx, "repo_mark_taken")
	defer done(&ctx, &err)

	d.mut.Lock()
	defer d.mut.Unlock()

	// TODO make this settable with a particular time
	// TODO make the storage system allow this to be easier and not search all
	for _, rd := range d.cache {
		for _, doses := range rd.Doses {
			if doses.ID == id {
				var t = true
				var now = time.Now()

				doses.Taken = &t
				doses.TimeTaken = &now

				return nil
			}
		}
	}

	return ErrDoseNotFound
}
