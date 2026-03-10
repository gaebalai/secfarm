package alert

import (
	"context"

	"github.com/gaebalai/secfarm/pkg/model"
)

type SearchOptions struct {
	Query string
	Limit int
}

func (u *UseCase) Search(
	ctx context.Context,
	opts SearchOptions,
) ([]*model.Alert, error) {
	return nil, nil
}
