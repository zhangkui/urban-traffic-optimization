package compliance

import ("testing"; "github.com/zhangkui/urban-traffic-optimization/internal/domain")
func TestAuthorizePublishPropagatesFailedReview(t *testing.T) { s := NewService(); plan := domain.TimingPlan{ID:"plan-004", Status:"approved"}; review := Review{PlanID:plan.ID, Passed:false, Findings:[]Finding{{Code:"CYCLE",Severity:"error"}}}; if err := s.CanPublish(plan, review); err == nil { t.Fatal("审核失败不得允许发布") } }
