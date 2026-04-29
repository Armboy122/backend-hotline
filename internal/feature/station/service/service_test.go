package service

import (
	"context"
	"errors"
	"testing"

	"backend-hotlines3/internal/feature/station/entity"
)

func TestServiceRejectsInvalidIDs(t *testing.T) {
	svc := NewService(&fakeRepo{})
	ctx := context.Background()

	if _, err := svc.GetByID(ctx, 0); !errors.Is(err, entity.ErrInvalidID) {
		t.Fatalf("GetByID expected ErrInvalidID, got %v", err)
	}
	if _, err := svc.Update(ctx, entity.UpdateInput{ID: -1}); !errors.Is(err, entity.ErrInvalidID) {
		t.Fatalf("Update expected ErrInvalidID, got %v", err)
	}
	if err := svc.Delete(ctx, 0); !errors.Is(err, entity.ErrInvalidID) {
		t.Fatalf("Delete expected ErrInvalidID, got %v", err)
	}
}

func TestServiceDelegatesCRUD(t *testing.T) {
	repo := &fakeRepo{items: []entity.Entity{{ID: 1, Name: "Station A", CodeName: "SA", OperationID: 10}}}
	svc := NewService(repo)
	ctx := context.Background()

	items, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 1 || items[0].CodeName != "SA" {
		t.Fatalf("unexpected list result: %+v", items)
	}

	created, err := svc.Create(ctx, entity.CreateInput{Name: "Station B", CodeName: "SB", OperationID: 11})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.OperationID != 11 {
		t.Fatalf("unexpected created item: %+v", created)
	}

	updated, err := svc.Update(ctx, entity.UpdateInput{ID: 1, Name: "Station C"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "Station C" {
		t.Fatalf("unexpected updated item: %+v", updated)
	}

	if err := svc.Delete(ctx, 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !repo.deleted {
		t.Fatalf("expected repository delete")
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

func (r *fakeRepo) Create(_ context.Context, input entity.CreateInput) (*entity.Entity, error) {
	return &entity.Entity{ID: 2, Name: input.Name, CodeName: input.CodeName, OperationID: input.OperationID}, nil
}

func (r *fakeRepo) Update(_ context.Context, input entity.UpdateInput) (*entity.Entity, error) {
	return &entity.Entity{ID: input.ID, Name: input.Name, CodeName: input.CodeName, OperationID: input.OperationID}, nil
}

func (r *fakeRepo) Delete(context.Context, int64) error {
	r.deleted = true
	return nil
}
