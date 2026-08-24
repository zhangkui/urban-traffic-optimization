package reporting

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/zhangkui/urban-traffic-optimization/internal/analysis"
	"github.com/zhangkui/urban-traffic-optimization/internal/domain"
	"github.com/zhangkui/urban-traffic-optimization/internal/store"
	"io"
	"strconv"
	"time"
)

type Service struct {
	store    *store.Store
	analysis *analysis.Service
}

func NewService(s *store.Store, a *analysis.Service) *Service { return &Service{store: s, analysis: a} }

type Report struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Rows      int       `json:"rows"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s *Service) Create(name, kind string, start, end time.Time) Report {
	return Report{ID: fmt.Sprintf("report-%d", time.Now().UnixNano()), Name: name, Kind: kind, Start: start, End: end, CreatedAt: time.Now().UTC()}
}
func (s *Service) TrafficCSV(w io.Writer, readings []domain.TrafficReading) error {
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"id", "intersection_id", "sensor_id", "lane", "volume", "speed", "queue", "occupancy", "recorded_at"}); err != nil {
		return err
	}
	for _, r := range readings {
		if err := writer.Write([]string{r.ID, r.IntersectionID, r.SensorID, r.Lane, strconv.Itoa(r.Volume), strconv.FormatFloat(r.Speed, 'f', 2, 64), strconv.FormatFloat(r.Queue, 'f', 2, 64), strconv.FormatFloat(r.Occupancy, 'f', 2, 64), r.RecordedAt.Format(time.RFC3339)}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
func (s *Service) IntersectionCSV(w io.Writer, items []analysis.IntersectionMetric) error {
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"rank", "intersection_id", "volume", "speed", "delay", "queue", "congestion"}); err != nil {
		return err
	}
	for _, item := range items {
		if err := writer.Write([]string{strconv.Itoa(item.Rank), item.ID, strconv.Itoa(item.Volume), strconv.FormatFloat(item.Speed, 'f', 2, 64), strconv.FormatFloat(item.Delay, 'f', 2, 64), strconv.FormatFloat(item.Queue, 'f', 2, 64), strconv.FormatFloat(item.Congestion, 'f', 2, 64)}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
func (s *Service) Save(report Report) error { return s.store.Save("reports", report.ID, report) }
func (s *Service) List() []Report {
	items := []Report{}
	_ = s.store.List("reports", func(raw []byte) error {
		var r Report
		if err := json.Unmarshal(raw, &r); err != nil {
			return err
		}
		items = append(items, r)
		return nil
	})
	return items
}
