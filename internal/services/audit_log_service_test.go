package services

import (
	"context"
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/models"
	"testing"
	"time"
)

func TestAuditLogService_List_Table(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		filter     domain.AuditLogFilter
		listErr    error
		listLogs   []models.AuditLog
		listTotal  int64
		expected   error
		assertRepo func(t *testing.T, got domain.AuditLogFilter)
	}{
		{
			name: "invalid time range",
			filter: domain.AuditLogFilter{
				From: ptrTime(now),
				To:   ptrTime(now.Add(-1 * time.Hour)),
			},
			expected: ErrInvalidTimeRange,
		},
		{
			name: "repo error",
			filter: domain.AuditLogFilter{
				Page:     2,
				PageSize: 50,
			},
			listErr:  errors.New("db failed"),
			expected: errors.New("db failed"),
		},
		{
			name: "success and default pagination",
			filter: domain.AuditLogFilter{
				Page:     0,
				PageSize: 0,
			},
			listLogs:  []models.AuditLog{{}},
			listTotal: 1,
			assertRepo: func(t *testing.T, got domain.AuditLogFilter) {
				t.Helper()
				if got.Page != 1 || got.PageSize != 20 {
					t.Fatalf("expected default page=1,pageSize=20 got page=%d,size=%d", got.Page, got.PageSize)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var captured domain.AuditLogFilter
			repo := &auditRepoMock{
				listFn: func(ctx context.Context, filter domain.AuditLogFilter) ([]models.AuditLog, int64, error) {
					captured = filter
					return tc.listLogs, tc.listTotal, tc.listErr
				},
			}
			svc := NewAuditLogService(repo)

			logs, total, err := svc.List(context.Background(), tc.filter)
			if tc.expected == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if len(logs) != len(tc.listLogs) || total != tc.listTotal {
					t.Fatalf("unexpected list result logs=%d total=%d", len(logs), total)
				}
			} else if tc.expected == ErrInvalidTimeRange {
				if !errors.Is(err, tc.expected) {
					t.Fatalf("expected ErrInvalidTimeRange, got %v", err)
				}
				return
			} else if err == nil || err.Error() != tc.expected.Error() {
				t.Fatalf("expected error %v, got %v", tc.expected, err)
			}

			if tc.assertRepo != nil {
				tc.assertRepo(t, captured)
			}
		})
	}
}
