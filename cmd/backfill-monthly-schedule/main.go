// Command backfill-monthly-schedule imports a reviewed schedule fixture into a
// Hotline draft. It never publishes; publication remains an explicit UI action.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/feature/monthlyschedule/entity"
	"backend-hotlines3/internal/feature/monthlyschedule/repository"
	"backend-hotlines3/internal/feature/monthlyschedule/service"
	dbpkg "backend-hotlines3/pkg/db"
)

type fixture struct {
	Source string        `json:"source"`
	Year   int           `json:"year"`
	Month  int           `json:"month"`
	Teams  []fixtureTeam `json:"teams"`
}

type fixtureTeam struct {
	TeamCode     string              `json:"teamCode"`
	BaseArea     string              `json:"baseArea"`
	CrewType     string              `json:"crewType"`
	DisplayOrder int                 `json:"displayOrder"`
	Assignments  []fixtureAssignment `json:"assignments"`
}

type fixtureAssignment struct {
	AssignmentType string  `json:"assignmentType"`
	StartDate      string  `json:"startDate"`
	EndDate        string  `json:"endDate"`
	Destination    string  `json:"destination"`
	Note           *string `json:"note"`
}

func main() {
	var fixturePath string
	var apply bool
	var actorUserID int64
	flag.StringVar(&fixturePath, "fixture", "fixtures/monthly-schedule/2026-07-clinic-tool.json", "path to reviewed monthly schedule fixture")
	flag.BoolVar(&apply, "apply", false, "persist a draft after validation")
	flag.Int64Var(&actorUserID, "actor-user-id", 0, "super_admin user ID recorded as draft creator")
	flag.Parse()

	ctx := context.Background()
	input, err := readFixture(fixturePath)
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	db, err := dbpkg.Connect(ctx, cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	repo := repository.NewGORM(db)
	teams, err := repo.ListVisibleTeams(ctx)
	if err != nil {
		log.Fatalf("list visible teams: %v", err)
	}
	assignments, missing, err := mapFixture(input, teams)
	if err != nil {
		log.Fatal(err)
	}
	if len(missing) > 0 {
		log.Fatalf("fixture team codes are not configured in Hotline: %s", strings.Join(missing, ", "))
	}

	fmt.Printf("validated %d teams and %d away assignments for %04d-%02d from %s\n", len(input.Teams), len(assignments), input.Year, input.Month, input.Source)
	if !apply {
		fmt.Println("dry run only; rerun with --apply --actor-user-id <super_admin id> to create or replace the draft")
		return
	}
	if actorUserID <= 0 {
		log.Fatal("--actor-user-id must be greater than zero when --apply is used")
	}
	svc := service.New(repo)
	workspace, err := svc.GetWorkspace(ctx, input.Year, input.Month)
	if err != nil {
		log.Fatalf("load current schedule workspace: %v", err)
	}
	var expectedRevisionNo *int
	if workspace.Draft != nil && workspace.Draft.Revision != nil {
		value := workspace.Draft.Revision.RevisionNo
		expectedRevisionNo = &value
	}
	workspace, err = svc.SaveDraft(ctx, entity.Actor{UserID: actorUserID, Role: "super_admin"}, input.Year, input.Month, expectedRevisionNo, assignments)
	if err != nil {
		log.Fatalf("save reviewed fixture as draft: %v", err)
	}
	fmt.Printf("saved draft revision %d; review and publish from Hotline UI\n", workspace.Draft.Revision.RevisionNo)
}

func readFixture(path string) (fixture, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fixture{}, fmt.Errorf("read fixture: %w", err)
	}
	var value fixture
	if err := json.Unmarshal(raw, &value); err != nil {
		return fixture{}, fmt.Errorf("decode fixture: %w", err)
	}
	if value.Year < 2000 || value.Month < 1 || value.Month > 12 || len(value.Teams) == 0 {
		return fixture{}, errors.New("fixture period or teams are invalid")
	}
	return value, nil
}

func mapFixture(input fixture, teams []entity.Team) ([]entity.Assignment, []string, error) {
	byCode := make(map[string]entity.Team, len(teams))
	for _, team := range teams {
		if team.Code != nil {
			byCode[strings.ToUpper(strings.TrimSpace(*team.Code))] = team
		}
	}
	out := []entity.Assignment{}
	missing := []string{}
	for _, fixtureTeam := range input.Teams {
		code := strings.ToUpper(strings.TrimSpace(fixtureTeam.TeamCode))
		team, ok := byCode[code]
		if !ok {
			missing = append(missing, code)
			continue
		}
		for _, item := range fixtureTeam.Assignments {
			start, err := time.Parse(time.DateOnly, item.StartDate)
			if err != nil {
				return nil, nil, fmt.Errorf("%s startDate: %w", code, err)
			}
			end, err := time.Parse(time.DateOnly, item.EndDate)
			if err != nil {
				return nil, nil, fmt.Errorf("%s endDate: %w", code, err)
			}
			out = append(out, entity.Assignment{
				TeamID:         team.ID,
				AssignmentType: item.AssignmentType,
				StartDate:      start,
				EndDate:        end,
				Destination:    item.Destination,
				Note:           item.Note,
				SourceType:     "manual",
			})
		}
	}
	return out, missing, nil
}
