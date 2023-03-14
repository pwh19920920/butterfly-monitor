package types

type MonitorTaskHomeCountResponse struct {
	TaskCount      *int64 `json:"taskCount"`
	EventCount     *int64 `json:"eventCount"`
	DashboardCount *int64 `json:"dashboardCount"`
	DatabaseCount  *int64 `json:"databaseCount"`
}
