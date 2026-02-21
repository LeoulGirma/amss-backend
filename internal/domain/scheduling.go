package domain

import (
	"time"

	"github.com/google/uuid"
)

// TaskPriority represents the urgency level of a task
type TaskPriority string

const (
	PriorityRoutine  TaskPriority = "routine"
	PriorityUrgent   TaskPriority = "urgent"
	PriorityAOG      TaskPriority = "aog"
	PriorityCritical TaskPriority = "critical"
)

// DependencyType represents how two tasks are related
type DependencyType string

const (
	DependencyFinishToStart  DependencyType = "finish_to_start"
	DependencyStartToStart   DependencyType = "start_to_start"
	DependencyFinishToFinish DependencyType = "finish_to_finish"
)

// TaskDependency represents a prerequisite relationship between tasks
type TaskDependency struct {
	ID              uuid.UUID
	OrgID           uuid.UUID
	TaskID          uuid.UUID
	DependsOnTaskID uuid.UUID
	DependencyType  DependencyType
	CreatedAt       time.Time
}

// ScheduleChangeType represents the type of schedule change
type ScheduleChangeType string

const (
	ScheduleChangeRescheduled       ScheduleChangeType = "rescheduled"
	ScheduleChangeCancelled         ScheduleChangeType = "cancelled"
	ScheduleChangePriorityChanged   ScheduleChangeType = "priority_changed"
	ScheduleChangeMechanicReassigned ScheduleChangeType = "mechanic_reassigned"
)

// ScheduleChangeEvent records a change to a task's schedule
type ScheduleChangeEvent struct {
	ID              uuid.UUID
	OrgID           uuid.UUID
	TaskID          uuid.UUID
	ChangeType      ScheduleChangeType
	Reason          string
	OldStartTime    *time.Time
	NewStartTime    *time.Time
	OldEndTime      *time.Time
	NewEndTime      *time.Time
	TriggeredBy     *uuid.UUID
	AffectedTaskIDs []uuid.UUID
	CreatedAt       time.Time
}

// RescheduleOption represents a possible rescheduling action
type RescheduleOption struct {
	OptionType      string
	Description     string
	NewStartTime    *time.Time
	NewEndTime      *time.Time
	AffectedTasks   []uuid.UUID
	SubstitutePart  *uuid.UUID
}
