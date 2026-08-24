package domain

import "time"

type PageRequest struct {
	Page    int
	Size    int
	Keyword string
	Status  string
	SortBy  string
	Desc    bool
}

func (r PageRequest) Normalize() PageRequest {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.Size < 1 {
		r.Size = 20
	}
	if r.Size > 200 {
		r.Size = 200
	}
	return r
}

type Page[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Size  int `json:"size"`
	Pages int `json:"pages"`
}

func MakePage[T any](items []T, total int, request PageRequest) Page[T] {
	request = request.Normalize()
	pages := total / request.Size
	if total%request.Size > 0 {
		pages++
	}
	return Page[T]{Items: items, Total: total, Page: request.Page, Size: request.Size, Pages: pages}
}

type User struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	DisplayName  string     `json:"displayName"`
	DepartmentID string     `json:"departmentId"`
	RoleIDs      []string   `json:"roleIds"`
	Enabled      bool       `json:"enabled"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}
type Department struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ParentID   string `json:"parentId"`
	RegionCode string `json:"regionCode"`
	Enabled    bool   `json:"enabled"`
}
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Code        string   `json:"code"`
	Permissions []string `json:"permissions"`
	Enabled     bool     `json:"enabled"`
}

type Road struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	Level        string    `json:"level"`
	SpeedLimit   int       `json:"speedLimit"`
	Lanes        int       `json:"lanes"`
	LengthMeters float64   `json:"lengthMeters"`
	Status       string    `json:"status"`
	Upstream     []string  `json:"upstream"`
	Downstream   []string  `json:"downstream"`
	CreatedAt    time.Time `json:"createdAt"`
}
type RoadLane struct {
	ID        string `json:"id"`
	RoadID    string `json:"roadId"`
	Number    int    `json:"number"`
	Direction string `json:"direction"`
	Movement  string `json:"movement"`
	Allowed   bool   `json:"allowed"`
}
type RoadConnection struct {
	ID                 string  `json:"id"`
	FromRoadID         string  `json:"fromRoadId"`
	ToRoadID           string  `json:"toRoadId"`
	FromIntersectionID string  `json:"fromIntersectionId"`
	ToIntersectionID   string  `json:"toIntersectionId"`
	DistanceMeters     float64 `json:"distanceMeters"`
	TravelTimeSeconds  int     `json:"travelTimeSeconds"`
}

type Intersection struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Code         string  `json:"code"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Level        string  `json:"level"`
	Status       string  `json:"status"`
	RegionCode   string  `json:"regionCode"`
	Congestion   float64 `json:"congestion"`
	AverageSpeed float64 `json:"averageSpeed"`
	AverageDelay float64 `json:"averageDelay"`
}
type Approach struct {
	ID             string  `json:"id"`
	IntersectionID string  `json:"intersectionId"`
	RoadID         string  `json:"roadId"`
	Direction      string  `json:"direction"`
	LaneCount      int     `json:"laneCount"`
	LengthMeters   float64 `json:"lengthMeters"`
}

type Device struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Code           string    `json:"code"`
	IntersectionID string    `json:"intersectionId"`
	Status         string    `json:"status"`
	Version        string    `json:"version"`
	LastHeartbeat  time.Time `json:"lastHeartbeat"`
	Message        string    `json:"message,omitempty"`
}
type SignalPhase struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Sequence  int      `json:"sequence"`
	Green     int      `json:"green"`
	Yellow    int      `json:"yellow"`
	AllRed    int      `json:"allRed"`
	Movements []string `json:"movements"`
	Enabled   bool     `json:"enabled"`
}
type TimingPlan struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	IntersectionID string        `json:"intersectionId"`
	Cycle          int           `json:"cycle"`
	Offset         int           `json:"offset"`
	Status         string        `json:"status"`
	Version        int           `json:"version"`
	Phases         []SignalPhase `json:"phases"`
	EffectiveFrom  *time.Time    `json:"effectiveFrom,omitempty"`
	EffectiveTo    *time.Time    `json:"effectiveTo,omitempty"`
}

type TrafficReading struct {
	ID             string         `json:"id"`
	IntersectionID string         `json:"intersectionId"`
	SensorID       string         `json:"sensorId"`
	Lane           string         `json:"lane"`
	Volume         int            `json:"volume"`
	Speed          float64        `json:"speed"`
	Queue          float64        `json:"queue"`
	Occupancy      float64        `json:"occupancy"`
	VehicleTypes   map[string]int `json:"vehicleTypes"`
	RecordedAt     time.Time      `json:"recordedAt"`
}
type TrafficAggregate struct {
	IntersectionID   string    `json:"intersectionId"`
	Start            time.Time `json:"start"`
	End              time.Time `json:"end"`
	Volume           int       `json:"volume"`
	AverageSpeed     float64   `json:"averageSpeed"`
	AverageQueue     float64   `json:"averageQueue"`
	AverageOccupancy float64   `json:"averageOccupancy"`
	AverageDelay     float64   `json:"averageDelay"`
	Samples          int       `json:"samples"`
}
type TrafficEvent struct {
	ID             string     `json:"id"`
	Type           string     `json:"type"`
	Level          string     `json:"level"`
	Title          string     `json:"title"`
	Status         string     `json:"status"`
	RoadID         string     `json:"roadId"`
	IntersectionID string     `json:"intersectionId"`
	StartedAt      time.Time  `json:"startedAt"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
	Description    string     `json:"description"`
	Source         string     `json:"source"`
}

type SimulationTask struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Status          string            `json:"status"`
	IntersectionIDs []string          `json:"intersectionIds"`
	TimingPlanID    string            `json:"timingPlanId"`
	StartAt         time.Time         `json:"startAt"`
	EndAt           time.Time         `json:"endAt"`
	FlowMultiplier  float64           `json:"flowMultiplier"`
	Progress        int               `json:"progress"`
	Result          *SimulationResult `json:"result,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
}
type SimulationResult struct {
	TaskID       string    `json:"taskId"`
	AverageDelay float64   `json:"averageDelay"`
	AverageQueue float64   `json:"averageQueue"`
	Throughput   float64   `json:"throughput"`
	Stops        float64   `json:"stops"`
	TravelTime   float64   `json:"travelTime"`
	CreatedAt    time.Time `json:"createdAt"`
}
type OptimizationTask struct {
	ID             string             `json:"id"`
	IntersectionID string             `json:"intersectionId"`
	Objective      string             `json:"objective"`
	Status         string             `json:"status"`
	BaselinePlanID string             `json:"baselinePlanId"`
	Constraints    map[string]float64 `json:"constraints"`
	CreatedAt      time.Time          `json:"createdAt"`
}
type OptimizationCandidate struct {
	ID         string  `json:"id"`
	TaskID     string  `json:"taskId"`
	Cycle      int     `json:"cycle"`
	Offset     int     `json:"offset"`
	GreenRatio float64 `json:"greenRatio"`
	Delay      float64 `json:"delay"`
	Queue      float64 `json:"queue"`
	Stops      float64 `json:"stops"`
	Throughput float64 `json:"throughput"`
	Score      float64 `json:"score"`
}

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"createdAt"`
}
type AuditLog struct {
	ID         string    `json:"id"`
	ActorID    string    `json:"actorId"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resourceId"`
	Detail     string    `json:"detail"`
	IP         string    `json:"ip"`
	CreatedAt  time.Time `json:"createdAt"`
}
type DashboardSummary struct {
	AverageSpeed        float64 `json:"averageSpeed"`
	AverageDelay        float64 `json:"averageDelay"`
	CongestionIndex     float64 `json:"congestionIndex"`
	OnlineDevices       int     `json:"onlineDevices"`
	DeviceTotal         int     `json:"deviceTotal"`
	ActiveEvents        int     `json:"activeEvents"`
	ActiveIntersections int     `json:"activeIntersections"`
	TodayVolume         int     `json:"todayVolume"`
}
