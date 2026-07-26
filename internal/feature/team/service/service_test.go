package service

import (
	"context"
	"errors"
	"testing"

	"backend-hotlines3/internal/feature/team/entity"
)

func TestServiceRejectsInvalidIDs(t *testing.T) {
	svc := NewService(&fakeRepo{})
	ctx := context.Background()

	if _, err := svc.GetByID(ctx, 0); !errors.Is(err, entity.ErrInvalidID) {
		t.Fatalf("GetByID expected ErrInvalidID, got %v", err)
	}
	if _, err := svc.Update(ctx, -1, entity.UpsertInput{Name: "Team A"}); !errors.Is(err, entity.ErrInvalidID) {
		t.Fatalf("Update expected ErrInvalidID, got %v", err)
	}
	if err := svc.Delete(ctx, 0); !errors.Is(err, entity.ErrInvalidID) {
		t.Fatalf("Delete expected ErrInvalidID, got %v", err)
	}
}

func TestServiceDelegatesCRUD(t *testing.T) {
	repo := &fakeRepo{
		items: []entity.Entity{{ID: 1, Name: "Team A", Tasks: 4}},
	}
	svc := NewService(repo)
	ctx := context.Background()

	items, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 1 || items[0].Tasks != 4 {
		t.Fatalf("unexpected list result: %+v", items)
	}

	created, err := svc.Create(ctx, entity.UpsertInput{Name: "Team B"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Name != "Team B" {
		t.Fatalf("unexpected created item: %+v", created)
	}

	updated, err := svc.Update(ctx, 1, entity.UpsertInput{Name: "Team C"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "Team C" {
		t.Fatalf("unexpected updated item: %+v", updated)
	}

	if err := svc.Delete(ctx, 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !repo.deleted {
		t.Fatalf("expected repository delete")
	}
}

func TestTeamIntegrationCodeIsNormalizedAndImmutable(t *testing.T) {
	code := "  kk-01 "
	repo := &fakeRepo{items: []entity.Entity{{ID: 1, Name: "ชุด 1"}}}
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), entity.UpsertInput{Name: " ชุด 2 ", Code: &code})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Code == nil || *created.Code != "KK-01" {
		t.Fatalf("normalized code = %#v, want KK-01", created.Code)
	}

	currentCode := "T01"
	repo.items[0].Code = &currentCode
	replacement := "T99"
	_, err = svc.Update(context.Background(), 1, entity.UpsertInput{Name: "ชุด 1", Code: &replacement})
	if !errors.Is(err, entity.ErrCodeImmutable) {
		t.Fatalf("Update code error = %v, want ErrCodeImmutable", err)
	}
}

type fakeRepo struct {
	items   []entity.Entity
	deleted bool
}

func (r *fakeRepo) List(context.Context) ([]entity.Entity, error) {
	return r.items, nil
}

func (r *fakeRepo) GetByID(_ context.Context, id int64) (*entity.Entity, error) {
	for _, item := range r.items {
		if item.ID == id {
			out := item
			return &out, nil
		}
	}
	return nil, entity.ErrNotFound
}

func (r *fakeRepo) Create(_ context.Context, input entity.UpsertInput) (*entity.Entity, error) {
	return &entity.Entity{ID: 2, Name: input.Name, Code: input.Code}, nil
}

func (r *fakeRepo) Update(_ context.Context, id int64, input entity.UpsertInput) (*entity.Entity, error) {
	return &entity.Entity{ID: id, Name: input.Name, Code: input.Code}, nil
}

func (r *fakeRepo) Delete(context.Context, int64) error {
	r.deleted = true
	return nil
}
