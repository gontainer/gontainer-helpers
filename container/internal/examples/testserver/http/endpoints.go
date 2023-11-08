package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gontainer/gontainer-helpers/v3/container/internal/examples/testserver/repositories"
)

// reposProvider is an interface that our container implements.
type reposProvider interface {
	UserRepo(ctx context.Context) repositories.UserRepository
	ImageRepo(ctx context.Context) repositories.ImageRepository
}

func NewMyEndpoint(repos reposProvider) ErrorAwareHandler {
	return ErrorAwareHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		userRepo := repos.UserRepo(r.Context())
		imageRepo := repos.ImageRepo(r.Context())
		s := "MyEndpoint:\n"
		s += fmt.Sprintf("\tTxID: %p\n", userRepo.Tx)
		s += fmt.Sprintf("\tuserRepo.Tx == imageRepo.Tx: %t\n", userRepo.Tx == imageRepo.Tx)

		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		_, _ = w.Write([]byte(s))

		return nil
	})
}
