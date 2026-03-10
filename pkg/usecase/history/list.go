package history

import (
	"context"

	"github.com/gaebalai/secfarm/pkg/model"
	"github.com/gaebalai/secfarm/pkg/repository"
)

func List(
	ctx context.Context,
	repo repository.Repository,
	offset, limit int,
) ([]*model.History, error) {
	return nil, nil
}
