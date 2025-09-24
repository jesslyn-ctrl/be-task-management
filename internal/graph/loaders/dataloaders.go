package loaders

import (
	_model "bitbucket.org/edts/go-task-management/internal/model"
	_repo "bitbucket.org/edts/go-task-management/internal/repository"
	_logger "bitbucket.org/edts/go-task-management/pkg/logger"
	"context"
	"github.com/vikstrous/dataloadgen"
	"net/http"
	"time"
)

var logs = _logger.GetContextLoggerf(nil)

type DataLoaders struct {
	UserLoader *dataloadgen.Loader[string, *_model.User]
	TeamLoader *dataloadgen.Loader[string, *_model.Team]
}

func NewLoaders(repo *_repo.Repository) *DataLoaders {
	return &DataLoaders{
		UserLoader: dataloadgen.NewLoader(func(ctx context.Context, keys []string) ([]*_model.User, []error) {
			users, err := repo.UserRepo.GetUsersByIDs(ctx, keys)
			if err != nil {
				logs.Errorf("ERROR fetching GetUsersByIDs in loader: %s", err.Error())
				errs := make([]error, len(keys))
				for i := range errs {
					errs[i] = err
				}
				return make([]*_model.User, len(keys)), errs
			}

			userMap := make(map[string]*_model.User)
			for _, u := range users {
				userMap[u.ID] = u
			}

			results := make([]*_model.User, len(keys))
			for i, k := range keys {
				results[i] = userMap[k] // nil if not found
			}
			return results, nil
		},
			// Short wait for batching
			dataloadgen.WithWait(1*time.Millisecond),
		),
		TeamLoader: dataloadgen.NewLoader(func(ctx context.Context, keys []string) ([]*_model.Team, []error) {
			teams, err := repo.TeamRepo.GetTeamsByIDs(ctx, keys)
			if err != nil {
				errs := make([]error, len(keys))
				for i := range errs {
					errs[i] = err
				}
				return make([]*_model.Team, len(keys)), errs
			}

			teamMap := make(map[string]*_model.Team)
			for _, t := range teams {
				teamMap[t.ID] = t
			}

			results := make([]*_model.Team, len(keys))
			for i, k := range keys {
				results[i] = teamMap[k] // nil if not found
			}
			return results, nil
		},
			// Short wait for batching
			dataloadgen.WithWait(1*time.Millisecond),
		),
	}
}

type contextKey string

const loadersKey contextKey = "dataloaders"

func Middleware(repo *_repo.Repository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, loadersKey, NewLoaders(repo))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func For(ctx context.Context) *DataLoaders {
	val := ctx.Value(loadersKey)
	if val == nil {
		panic("no dataloaders in context")
	}
	return val.(*DataLoaders)
}
