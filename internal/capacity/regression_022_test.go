package capacity

import "testing"
func TestAssessConfiguredReportsMissingApproach(t *testing.T) { s:=NewService(); got:=s.Assess("intersection-22",nil); if got.TotalCapacity!=0||got.TotalDemand!=0{t.Fatalf("缺少进口道却产生容量=%+v",got)} }
