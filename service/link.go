package service

import (
	"context"
	"log"
	"math/rand/v2"
	"strconv"
	"time"
	"url-shortener/repository"
)

type LinkService struct {
	repo *repository.LinkRepository
}

func NewLinkService(r *repository.LinkRepository) *LinkService {
	return &LinkService{repo: r}
}

func (s *LinkService) GetLink(ctx context.Context, short string) (repository.Link, error) {
	return s.repo.Get(ctx, short)
}

func (s *LinkService) GetLinks(ctx context.Context) ([]repository.Link, error) {
	return s.repo.GetAll()
}

func (s *LinkService) DeleteLink(ctx context.Context, short string) error {
	return s.repo.Delete(ctx, short)
}

func (s *LinkService) CreateLink(ctx context.Context, long string) (repository.Link, error) {
	short := strconv.FormatInt(int64(rand.Uint32()), 36)
	expiresAt := time.Now().Add(time.Hour * 24)
	return s.repo.Create(ctx, repository.Link{Long: long, Short: short, ExpiresAt: &expiresAt})
}

func (s *LinkService) Increase(short string) {
	go s.repo.Increase(context.Background(), short)
}

func (s *LinkService) GetClicks(ctx context.Context, short string) (int, error) {
	return s.repo.GetClicks(ctx, short)
}

func (s *LinkService) StartCleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err := s.repo.DeleteExpired(ctx)
				if err != nil {
					log.Printf("cleanup failed: %v\n", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
