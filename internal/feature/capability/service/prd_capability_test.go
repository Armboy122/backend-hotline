package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"backend-hotlines3/internal/feature/capability/entity"
)

type fakeCapabilityRepo struct {
	listUserID      uint
	listCodes       []string
	listErr         error
	replaceUserID   uint
	replaceCodes    []string
	replaceGranted  *uint
	replaceCalls    int
	replaceErr      error
	targetUser      *entity.TargetUser
	targetErr       error
	targetLookupID  uint
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

func (f *fakeCapabilityRepo) GetTargetUser(_ context.Context, userID uint) (*entity.TargetUser, error) {
	f.targetLookupID = userID
	if f.targetUser == nil && f.targetErr == nil {
		return &entity.TargetUser{ID: userID, Role: "user", IsActive: true}, nil
	}
	return f.targetUser, f.targetErr
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

func TestReplaceForUserValidatesTargetExistsActiveAndNonViewerBeforeWrite(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name       string
		targetUser *entity.TargetUser
		targetErr  error
		wantErr    error
	}{
		{name: "missing user", targetErr: entity.ErrTargetUserNotFound, wantErr: entity.ErrTargetUserNotFound},
		{name: "inactive user", targetUser: &entity.TargetUser{ID: 7, Role: "user", IsActive: false}, wantErr: entity.ErrInactiveTargetUser},
		{name: "viewer user", targetUser: &entity.TargetUser{ID: 7, Role: "viewer", IsActive: true}, wantErr: entity.ErrViewerTargetForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeCapabilityRepo{targetUser: tc.targetUser, targetErr: tc.targetErr}
			svc := NewService(repo)

			_, err := svc.ReplaceForUser(ctx, 10, 7, []string{entity.CanUploadApprovedMonthlyPlan})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got %v, want %v", err, tc.wantErr)
			}
			if repo.targetLookupID != 7 {
				t.Fatalf("target lookup id = %d, want 7", repo.targetLookupID)
			}
			if repo.replaceCalls != 0 {
				t.Fatalf("repository ReplaceCodes should not be called, got %d calls", repo.replaceCalls)
			}
		})
	}

	repo := &fakeCapabilityRepo{targetUser: &entity.TargetUser{ID: 7, Role: "team_lead", IsActive: true}}
	svc := NewService(repo)
	if _, err := svc.ReplaceForUser(ctx, 10, 7, []string{entity.CanUploadApprovedMonthlyPlan}); err != nil {
		t.Fatalf("active non-viewer target should be allowed, got %v", err)
	}
	if repo.replaceCalls != 1 {
		t.Fatalf("ReplaceCodes calls = %d, want 1", repo.replaceCalls)
	}
}
