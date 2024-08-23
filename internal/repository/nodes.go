package repository

import (
	"SolProject/internal/model"
	"context"
)

type NodesRepository interface {
	FirstById(id int64) (*model.Nodes, error)
	Count(ctx context.Context) (int64, error)
	Create(ctx context.Context, node *model.Nodes) error
}
type nodesRepository struct {
	*Repository
}

func NewNodesRepository(repository *Repository) NodesRepository {
	return &nodesRepository{
		Repository: repository,
	}
}

func (r *nodesRepository) FirstById(id int64) (*model.Nodes, error) {
	var nodes model.Nodes
	// TODO: query db
	return &nodes, nil
}

func (r *nodesRepository) Create(ctx context.Context, node *model.Nodes) error {
	if err := r.DB(ctx).Model(model.Nodes{}).Create(node).Error; err != nil {
		return err
	}
	return nil
}

func (r *nodesRepository) Count(ctx context.Context) (int64, error) {
	var numbers int64
	if err := r.DB(ctx).Model(model.Nodes{}).Count(&numbers).Error; err != nil {
		return numbers, err
	}
	return numbers / 2, nil
}
