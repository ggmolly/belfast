package answer

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

func TestIsVisitBackyardMissingTargetError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "not found",
			err:  db.ErrNotFound,
			want: true,
		},
		{
			name: "wrapped not found",
			err:  fmt.Errorf("lookup failed: %w", db.ErrNotFound),
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("database timeout"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isVisitBackyardMissingTargetError(tt.err); got != tt.want {
				t.Fatalf("isVisitBackyardMissingTargetError() = %v, want %v", got, tt.want)
			}
		})
	}
}
