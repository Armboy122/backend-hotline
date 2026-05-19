package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"backend-hotlines3/internal/feature/capability/entity"
)

type fakeCapabilityRepo struct {
	listUserID     uint
	listCodes      []string
	listErr        error
	replaceUserID  uint
	replaceCodes   []string
	replaceGranted *uint
	replaceCalls   int
	replaceErr     error
}

func (f *fakeCapabilityRepo) ListCodesByUserID(_ context.Context, userID uint) ([]string, error) {
	f.listUserID = userID
	return f.listCodes, f.listErr
}

func (f *fakeCapabilityRepo) ReplaceCodes(_ context.Context, userID uint, codes []string, grantedBy *uint) error {
	f.replaceCalls++
	f.replaceUserID = userID
	f.replaceCodes = append([]string{}, codes...)
	if grantedBy != nil {
		value := *grantedBy
		f.replaceGranted = &value
	}
	return f.replaceErr
}

func TestPRDCapabilityCatalogContainsOnlyRoundOneCapability(t *testing.T) {
	svc := NewService(&fakeCapabilityRepo{})
	capabilities := svc.ListAvailable()

	if len(capabilities) != 1 {
		t.Fatalf("capability count = %d, want only round-1 capability", len(capabilities))
	}
	if capabilities[0].Code != entity.CanUploadApprovedMonthlyPlan {
		t.Fatalf("capability code = %q, want %q", capabilities[0].Code, entity.CanUploadApprovedMonthlyPlan)
	}
	if !entity.IsValidCapability(entity.CanUploadApprovedMonthlyPlan) {
		t.Fatalf("round-1 capability should be valid")
	}
	if entity.IsValidCapability("can_export_reports") {
		t.Fatalf("unexpected non-PRD capability accepted")
	}
}

func TestReplaceForUserDedupesAndRejectsUnknownCapabilitiesBeforeWrite(t *testing.T) {
	repo := &fakeCapabilityRepo{}
	svc := NewService(repo)

	out, err := svc.ReplaceForUser(context.Background(), 10, 7, []string{
		entity.CanUploadApprovedMonthlyPlan,
		entity.CanUploadApprovedMonthlyPlan,
	})
	if err != nil {
		t.Fatalf("ReplaceForUser valid capability: %v", err)
	}
	if out.UserID != 7 || !reflect.DeepEqual(out.Capabilities, []string{entity.CanUploadApprovedMonthlyPlan}) {
		t.Fatalf("output = %#v, want deduped round-1 capability", out)
	}
	if repo.replaceUserID != 7 || repo.replaceGranted == nil || *repo.replaceGranted != 10 {
		t.Fatalf("replace metadata = user:%d granted:%#v, want user 7 granted by 10", repo.replaceUserID, repo.replaceGranted)
	}

	repo = &fakeCapabilityRepo{}
	svc = NewService(repo)
	if _, err := svc.ReplaceForUser(context.Background(), 10, 7, []string{"can_delete_everything"}); !errors.Is(err, entity.ErrInvalidCapability) {
		t.Fatalf("invalid capability got %v, want ErrInvalidCapability", err)
	}
	if repo.replaceCalls != 0 {
		t.Fatalf("repository should not be called for invalid capability, got %d calls", repo.replaceCalls)
	}
}

func TestCapabilityUserIDValidation(t *testing.T) {
	svc := NewService(&fakeCapabilityRepo{})
	if _, err := svc.ListForUser(context.Background(), 0); !errors.Is(err, entity.ErrInvalidUserID) {
		t.Fatalf("ListForUser(0) got %v, want ErrInvalidUserID", err)
	}
	if _, err := svc.ReplaceForUser(context.Background(), 10, 0, nil); !errors.Is(err, entity.ErrInvalidUserID) {
		t.Fatalf("ReplaceForUser user 0 got %v, want ErrInvalidUserID", err)
	}
}
