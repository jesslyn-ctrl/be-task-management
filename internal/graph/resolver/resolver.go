package resolver

import (
	_dl "bitbucket.org/edts/go-task-management/internal/graph/loaders"
	_usecase "bitbucket.org/edts/go-task-management/internal/usecase"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Usecase    *_usecase.Usecase
	DataLoader *_dl.DataLoaders
}

func NewResolver(uc *_usecase.Usecase, dataloader *_dl.DataLoaders) *Resolver {
	return &Resolver{
		Usecase:    uc,
		DataLoader: dataloader,
	}
}
