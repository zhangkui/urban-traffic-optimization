package optimization

import ("path/filepath"; "testing"; "github.com/zhangkui/urban-traffic-optimization/internal/store")
func TestLatestScoreWithoutInspectionReturnsZero(t *testing.T) { db, err := store.Open(filepath.Join(t.TempDir(), "traffic.db")); if err != nil { t.Fatal(err) }; defer db.Close(); s := NewService(db); if _, err := s.Recommend("task-007", nil); err == nil { t.Fatal("没有候选方案必须返回明确错误") } }
