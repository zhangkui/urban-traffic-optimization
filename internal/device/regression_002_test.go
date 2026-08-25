package device

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/zhangkui/urban-traffic-optimization/internal/store"
)

func TestRefreshStatusesReportsCorruptHeartbeatRecord(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "traffic.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// A truncated telemetry payload can be persisted by an upstream ingest retry.
	if err := db.Save("devices", "det-002", "not-json"); err != nil {
		t.Fatal(err)
	}

	s := NewService(db)
	if err := s.RefreshStatuses(time.Now().UTC()); err == nil {
		t.Fatal("RefreshStatuses silently accepted a corrupt heartbeat payload")
	}
}
