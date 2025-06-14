package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/zahartd/social-network/src/services/post-service/internal/domain/models"
	"github.com/zahartd/social-network/src/services/post-service/internal/domain/repository"
)

type repo struct {
	mu       sync.RWMutex
	posts    map[string]*models.Post
	likes    map[string]map[string]struct{} // postID -> set(userID)
	comments map[string]*models.Comment     // commentID -> comment
	replies  map[string][]*models.Reply     // parentCommentID -> replies
}

func New() repository.Post {
	return &repo{
		posts:    make(map[string]*models.Post),
		likes:    make(map[string]map[string]struct{}),
		comments: make(map[string]*models.Comment),
		replies:  make(map[string][]*models.Reply),
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

/* ------------------- posts ------------------- */

func (r *repo) Create(_ context.Context, p *models.Post) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.NewString()
	now := time.Now().UTC()
	p.ID, p.CreatedAt, p.UpdatedAt = id, now, now
	cp := *p
	r.posts[id] = &cp
	return id, nil
}

func (r *repo) GetByID(_ context.Context, id string) (*models.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.posts[id]
	if !ok {
		return nil, repository.ErrPostNotFound
	}
	cp := *p
	return &cp, nil
}

func (r *repo) AuthorID(_ context.Context, id string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.posts[id]
	if !ok {
		return "", repository.ErrPostNotFound
	}
	return p.UserID, nil
}

func (r *repo) Update(_ context.Context, p *models.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ex, ok := r.posts[p.ID]
	if !ok {
		return repository.ErrPostNotFound
	}
	ex.Title = p.Title
	ex.Description = p.Description
	ex.IsPrivate = p.IsPrivate
	ex.Tags = p.Tags
	ex.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *repo) Delete(_ context.Context, id, user string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.posts[id]
	if !ok {
		return repository.ErrPostNotFound
	}
	if p.UserID != user {
		return repository.ErrForbidden
	}
	delete(r.posts, id)
	return nil
}

/* ------------------- feeds ------------------- */

func (r *repo) UserFeed(_ context.Context, user string, page, size int) ([]models.Post, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []models.Post
	for _, p := range r.posts {
		if p.UserID == user {
			out = append(out, *p)
		}
	}
	total := len(out)
	start := (page - 1) * size
	if start >= total {
		return nil, total, nil
	}
	end := min(start+size, total)
	return out[start:end], total, nil
}

func (r *repo) PublicFeed(_ context.Context, filter *string, page, size int) ([]models.Post, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []models.Post
	for _, p := range r.posts {
		if p.IsPrivate {
			continue
		}
		if filter != nil && *filter != "" && p.UserID != *filter {
			continue
		}
		out = append(out, *p)
	}
	total := len(out)
	start := (page - 1) * size
	if start >= total {
		return nil, total, nil
	}
	end := min(start+size, total)
	return out[start:end], total, nil
}

/* ------------------- metrics ------------------- */

func (r *repo) RecordView(context.Context, string, string) error { return nil }
func (r *repo) RecordLike(context.Context, string, string) error { return nil }
func (r *repo) RemoveLike(context.Context, string, string) error { return nil }

/* ------------------- comments & replies ------------------- */

func (r *repo) CreateComment(_ context.Context, c *models.Comment) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.NewString()
	now := time.Now().UTC()
	c.ID, c.CreatedAt = id, now
	cp := *c
	r.comments[id] = &cp
	return id, nil
}

func (r *repo) CreateReply(_ context.Context, rp *models.Reply) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.NewString()
	now := time.Now().UTC()
	rp.ID, rp.CreatedAt = id, now
	cp := *rp
	r.replies[rp.ParentCommentID] = append(r.replies[rp.ParentCommentID], &cp)
	return id, nil
}

func (r *repo) ListComments(_ context.Context, postID string, page, size int) ([]models.Comment, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []models.Comment
	for _, c := range r.comments {
		if c.PostID == postID {
			list = append(list, *c)
		}
	}
	total := len(list)
	start := (page - 1) * size
	if start >= total {
		return nil, total, nil
	}
	end := min(start+size, total)
	return list[start:end], total, nil
}

func (r *repo) ListReplies(_ context.Context, parentID string) ([]models.Reply, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	repPtrs := r.replies[parentID]
	var reps []models.Reply
	for _, rp := range repPtrs {
		reps = append(reps, *rp)
	}
	return reps, nil
}
