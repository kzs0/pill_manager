package repositories

import (
	"context"
	"errors"
	"sync"

	"github.com/kzs0/kokoro/koko"
	"github.com/kzs0/pill_manager/models"
)

var (
	ErrPerscriptionNotFound = errors.New("perscription not found")
)

type Perscriptions struct {
	c     int64
	mut   sync.RWMutex
	cache map[int64]*models.Perscription
}

func NewPerscriptions() *Perscriptions {
	return &Perscriptions{
		c:     1,
		mut:   sync.RWMutex{},
		cache: make(map[int64]*models.Perscription),
	}
}

func (p *Perscriptions) Put(ctx context.Context, perscription *models.Perscription) (_ *models.Perscription, err error) {
	ctx, done := koko.Operation(ctx, "repo_put_perscription")
	defer done(&ctx, &err)

	p.mut.Lock()
	defer p.mut.Unlock()

	perscription.ID = p.c
	perscription.Medication.ID = p.c + 1

	p.cache[p.c] = perscription

	p.c = p.c + 2

	return perscription, nil
}

func (p *Perscriptions) Get(ctx context.Context, id int64) (rx *models.Perscription, err error) {
	ctx, done := koko.Operation(ctx, "repo_get_perscription")
	defer done(&ctx, &err)

	p.mut.RLock()
	defer p.mut.RUnlock()

	perscription, ok := p.cache[id]
	if !ok {
		return nil, ErrPerscriptionNotFound
	}

	return perscription, nil
}

func (p *Perscriptions) GetAll(ctx context.Context) (rx []*models.Perscription, err error) {
	ctx, done := koko.Operation(ctx, "repo_get_all_perscriptions")
	defer done(&ctx, &err)

	p.mut.RLock()
	defer p.mut.RUnlock()

	perscriptions := make([]*models.Perscription, 0, len(p.cache))

	for _, perscription := range p.cache {
		perscriptions = append(perscriptions, perscription)
	}

	return perscriptions, nil
}
