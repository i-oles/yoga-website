package reminder

import (
	"testing"
	"time"
)

func TestIsTimeToRemind(t *testing.T) {
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		start time.Time
		want  bool
	}{
		{
			name:  "within 24h (10h)",
			start: now.Add(10 * time.Hour),
			want:  true,
		},
		{
			name:  "just below 24h",
			start: now.Add(24*time.Hour - time.Nanosecond),
			want:  true,
		},
		{
			name:  "exactly 24h",
			start: now.Add(24 * time.Hour),
			want:  false,
		},
		{
			name:  "just above 24h",
			start: now.Add(24*time.Hour + time.Nanosecond),
			want:  false,
		},
		{
			name:  "exactly now",
			start: now,
			want:  false,
		},
		{
			name:  "in the past",
			start: now.Add(-1 * time.Hour),
			want:  false,
		},
		{
			name:  "zero time",
			start: time.Time{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTimeToRemind(tt.start, now)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
