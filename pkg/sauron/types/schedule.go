package types

// The reconcile operations a Schedule may run (Schedule.metadata.name enum).
const (
	ScheduleSync    = "sync"
	ScheduleUpgrade = "upgrade"
)

// Schedule is an OS-crontab schedule for a reconcile operation, recorded in
// settings.yaml.
type Schedule struct {
	TypeMeta `json:",inline" yaml:",inline"`
	Metadata Metadata     `json:"metadata" yaml:"metadata"`
	Spec     ScheduleSpec `json:"spec" yaml:"spec"`
}

// ScheduleSpec is the spec block of a Schedule document.
type ScheduleSpec struct {
	// Cron is the cron expression registered in the OS crontab.
	Cron string `json:"cron" yaml:"cron"`
}
