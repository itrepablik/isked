package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/itrepablik/itrlog"
)

// ChannelTS is the channel to be used during cancellation of all the tasks
// that are currently running, this is useful when reloading some config variables
// to get the latest values and reload the task scheduler.
var ChannelTS = make(chan bool, 1)

// Name this package as 'gawain' meaning task
const (
	_seconds        = "seconds"
	_minutes        = "minutes"
	_hours          = "hours"
	_everyMonday    = "monday"
	_everyTuesday   = "tuesday"
	_everyWednesday = "wednesday"
	_everyThursday  = "thursday"
	_everyFriday    = "friday"
	_everySaturday  = "saturday"
	_everySunday    = "sunday"
	_onetime        = "onetime"
	_frequently     = "frequently"
	_daily          = "daily"
	_weekly         = "weekly"
	_monthly        = "monthly"
	_timeFormat     = "1504"
	_dateTimeFormat = "Jan 02 2006 03:04:05 PM"
)

// FuncToExec is the function that needs to be executed as parameter
type FuncToExec func()

// TaskScheduler is the task scheduler's format
type TaskScheduler struct {
	TaskList map[string][]Tasks
	mu       sync.Mutex
}

// Tasks is the individual task item to be executed
type Tasks struct {
	Name                   string
	RunType                string       // options: onetime, frequently, daily, weekly, monthly
	FrequencyInterval      string       // use for frequently option only: seconds, minutes, hours
	FrequencyValue         int          // use for frequently option only, minimum value of 1, e.g 1 second
	ExecuteFunc            FuncToExec   // user's defined func to be executed
	runAtHour, runAtMinute string       // 24-hour clock beginning at midnight (0000 hours) and ends at 2359 hours
	isRunAt                bool         // true, if use the '.At("15:04")' method, for frequently it's not applicable
	dayName                time.Weekday // internal usage: dayName such as 'Monday' using time.Weekday format
	monthName              time.Month   // internal usage: monthName such as 'January' using time.Month format
	monthDay               int          // internal usage: monthDay is serve as the specific day of the month
	nextRunTime            int64        // internal usage: next scheduled run
	lastRunTime            int64        // internal usage: last executed task
	created                int64        // internal usage: task created
}

// TS initialize the 'TaskScheduler' struct with an empty values
var TS = TaskScheduler{TaskList: make(map[string][]Tasks)}

// TK initialize the 'Tasks' struct with an empty values
var TK = Tasks{}

// DTFormat is the standard DateTime format to be used for logging information
var DTFormat string = _dateTimeFormat

// LogDTFormat is the DateTime format for each logs
type LogDTFormat struct {
	DTFormat string
	mu       sync.Mutex
}

var logDateTimeFormat string = _dateTimeFormat
var dt *LogDTFormat

func initDT(dtFormat string) *LogDTFormat {
	if len(strings.TrimSpace(dtFormat)) == 0 {
		dtFormat = _dateTimeFormat
	}
	return &LogDTFormat{
		DTFormat: dtFormat,
	}
}

func init() {
	dt = initDT("")
}

// SetLogDT customizes the DateTime logging format to be used for each logs
func SetLogDT(dtFormat string) *LogDTFormat {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	dt = initDT(dtFormat)
	logDateTimeFormat = dt.DTFormat
	return dt
}

// Seconds is the naming convention for the Frequently method as 'seconds' option
func (s *Tasks) Seconds(interval int) *Tasks {
	s.FrequencyInterval = _seconds
	if interval <= 0 {
		s.FrequencyValue = 1 // Default to 1 sec
	} else {
		s.FrequencyValue = interval
	}
	return s
}

// Minutes is the naming convention for the Frequently method as 'minutes' option
func (s *Tasks) Minutes(interval int) *Tasks {
	s.FrequencyInterval = _minutes
	if interval <= 0 {
		s.FrequencyValue = 1 // Default to 1 sec
	} else {
		s.FrequencyValue = interval
	}
	return s
}

// Hours is the naming convention for the Frequently method as 'hours' option
func (s *Tasks) Hours(interval int) *Tasks {
	s.FrequencyInterval = _hours
	if interval <= 0 {
		s.FrequencyValue = 1 // Default to 1 sec
	} else {
		s.FrequencyValue = interval
	}
	return s
}

// Monday is the naming convention for the day called 'Monday' method
func (s *Tasks) Monday() *Tasks {
	s.dayName = time.Monday
	return s
}

// Tuesday is the naming convention for the day called 'Tuesday' method
func (s *Tasks) Tuesday() *Tasks {
	s.dayName = time.Tuesday
	return s
}

// Wednesday is the naming convention for the day called 'Wednesday' method
func (s *Tasks) Wednesday() *Tasks {
	s.dayName = time.Wednesday
	return s
}

// Thursday is the naming convention for the day called 'Thursday' method
func (s *Tasks) Thursday() *Tasks {
	s.dayName = time.Thursday
	return s
}

// Friday is the naming convention for the day called 'Friday' method
func (s *Tasks) Friday() *Tasks {
	s.dayName = time.Friday
	return s
}

// Saturday is the naming convention for the day called 'Saturday' method
func (s *Tasks) Saturday() *Tasks {
	s.dayName = time.Saturday
	return s
}

// Sunday is the naming convention for the day called 'Sunday' method
func (s *Tasks) Sunday() *Tasks {
	s.dayName = time.Sunday
	return s
}

// Every is use mainly for the 'Monthly' method that serve as the specific day of each month
func (s *Tasks) Every(day int) *Tasks {
	today := time.Now()
	lastDayOfMonth := getLastDayOfMonth(day, today.Month())
	switch {
	case day == 0:
		s.monthDay = lastDayOfMonth
	case day > lastDayOfMonth:
		s.monthDay = getLastDayOfMonth(day, today.Month())
	default:
		s.monthDay = day
	}
	return s
}

// Get gets the specific task information using the task name
func (t *TaskScheduler) Get(taskName string) ([]Tasks, bool) {
	taskData, ok := t.TaskList[taskName]
	if !ok {
		return taskData, ok
	}
	return taskData, ok
}

// TaskName method is the run type option of each task that execute once only
func (s *Tasks) TaskName(taskName string) *Tasks {
	newTaskName := strings.TrimSpace(taskName)
	if len(newTaskName) == 0 {
		taskName = uuid.New().String() // Assign with random strings if empty
	}
	// Check if any duplicate task name, add extra timestamp using unix format
	payLoad, _ := TS.Get(taskName)
	for _, e := range payLoad {
		newTaskName = e.Name + "_" + fmt.Sprintf("%v", time.Now().Unix())
	}
	TK = Tasks{
		Name:              newTaskName,
		RunType:           "",
		FrequencyInterval: "",
		FrequencyValue:    0,
		ExecuteFunc:       nil,
		runAtHour:         "",
		runAtMinute:       "",
		monthDay:          0,
		monthName:         time.Now().Local().Month(),
		isRunAt:           false,
		nextRunTime:       0,
		lastRunTime:       0,
		created:           time.Now().Unix(),
	}
	return &TK
}

// Frequently method is the run type option of each task that execute frequently
// Options: seconds, minutes, hours
func (s *Tasks) Frequently() *Tasks {
	s.RunType = _frequently
	return s
}

// OneTime method requires unix DateTime format that executes only once
func (s *Tasks) OneTime(dt int64) *Tasks {
	s.RunType = _onetime
	timeNow := time.Now().Unix()

	if dt < timeNow {
		// Set the default DateTime of +24 hours from the current time if entered time is not a future time.
		s.nextRunTime = time.Now().Add(24 * time.Hour).Unix()
	} else {
		s.nextRunTime = dt
	}
	return s
}

// Daily method is the run type option of each task that execute every day
func (s *Tasks) Daily() *Tasks {
	s.RunType = _daily
	return s
}

// Weekly method is the run type option of each task that execute every week
func (s *Tasks) Weekly() *Tasks {
	s.RunType = _weekly
	return s
}

// Monthly method is the run type option of each task that execute every month
func (s *Tasks) Monthly() *Tasks {
	s.RunType = _monthly
	return s
}

// At method is when to start executing the task with DateTime in unix time format
func (s *Tasks) At(rt string) *Tasks {
	// Only allowed 'At' method can use this process
	// OneTime and Frequently is not required
	if s.RunType != _onetime && s.RunType != _frequently {
		s.isRunAt = true
		// Check with the correct 24-hour format
		pTime := strings.TrimSpace(rt)
		if strings.Contains(pTime, ":") {
			if len(pTime) == 5 {
				s.runAtHour = strings.Split(pTime, ":")[0]
				s.runAtMinute = strings.Split(pTime, ":")[1]
			} else {
				s.runAtHour = "00"
				s.runAtMinute = "00" // Default to 12-midnight
			}
		} else {
			s.runAtHour = "00"
			s.runAtMinute = "00" // Default to 12-midnight
		}
	}
	return s
}

// ExecFunc method collect the function as parameter that needs to be executed
func (s *Tasks) ExecFunc(fn FuncToExec) *Tasks {
	s.ExecuteFunc = fn
	return s
}

// AddTask create individual task to be executed
func (s *Tasks) AddTask() {
	var nextSchedToRun int64 = 0

	switch s.RunType {
	case _onetime:
		nextSchedToRun = s.nextRunTime

	case _frequently:
		if s.FrequencyInterval == _seconds {
			nextSchedToRun = time.Now().Add(time.Second * time.Duration(s.FrequencyValue)).Unix()
		}
		if s.FrequencyInterval == _minutes {
			nextSchedToRun = time.Now().Add(time.Minute * time.Duration(s.FrequencyValue)).Unix()
		}
		if s.FrequencyInterval == _hours {
			nextSchedToRun = time.Now().Add(time.Hour * time.Duration(s.FrequencyValue)).Unix()
		}

	case _daily:
		runHour, _ := strconv.Atoi(s.runAtHour)
		runMinute, _ := strconv.Atoi(s.runAtMinute)
		nextSchedToRun = time.Date(
			time.Now().Year(),
			time.Now().Month(),
			time.Now().Day()+1,
			runHour, runMinute, 0, 0, time.Local).Unix()

	case _weekly:
		runHour, _ := strconv.Atoi(s.runAtHour)
		runMinute, _ := strconv.Atoi(s.runAtMinute)
		today := time.Now()

		nextSchedToRun = time.Date(
			today.Year(),
			today.Month(),
			today.Day()-int(today.Weekday()-s.dayName),
			runHour, runMinute, 0, 0,
			time.Local).Add(7 * 24 * time.Hour).Unix()

	case _monthly:
		runHour, _ := strconv.Atoi(s.runAtHour)
		runMinute, _ := strconv.Atoi(s.runAtMinute)
		today := time.Now()

		nextSchedToRun = time.Date(
			today.Year(),
			today.Month()+1,
			s.monthDay,
			runHour, runMinute, 0, 0,
			time.Local).Unix()

	default:
		msg := s.Name + " is not running due to incorrect or missing parameters"
		itrlog.Errorw(msg, "log_time", time.Now().Format(logDateTimeFormat))
		color.Red(msg)
	}

	// Format next scheduled run
	nextSched, _ := formatDT(nextSchedToRun, logDateTimeFormat)
	msg := s.Name + " next schedule to run on: " + nextSched
	itrlog.Infow(msg, "log_time", time.Now().Format(logDateTimeFormat))
	color.Magenta(msg)

	newTask := Tasks{
		Name:              s.Name,
		RunType:           s.RunType,
		FrequencyInterval: s.FrequencyInterval,
		FrequencyValue:    s.FrequencyValue,
		ExecuteFunc:       s.ExecuteFunc,
		runAtHour:         s.runAtHour,
		runAtMinute:       s.runAtMinute,
		monthDay:          s.monthDay,
		monthName:         s.monthName,
		isRunAt:           s.isRunAt,
		nextRunTime:       nextSchedToRun,
		lastRunTime:       0,
		created:           time.Now().Unix(),
	}
	TS.TaskList[s.Name] = []Tasks{newTask}
}

// Run executes the task scheduler's individual task item
func (t *TaskScheduler) Run() {
mainloop:
	for {
		for _, e := range t.TaskList {
			for _, s := range e {
				// Check if due for execution
				unixTimeNow := time.Now().Unix()
				if s.nextRunTime == unixTimeNow {
					t.UpdateNextRunTime(&s)
					go s.ExecuteFunc()
				}
			}
		}
		time.Sleep(300 * time.Millisecond)
		select {
		case msg := <-ChannelTS:
			fmt.Println("channel message: ", msg)
			break mainloop
		default:

		}
	}
	TS.Reset()
}

// UpdateNextRunTime modify the next run time
func (t *TaskScheduler) UpdateNextRunTime(s *Tasks) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var nextSchedToRun int64 = 0

	// For OneTime method, no need to auto-create new schedule to run since it's a onetime run only.
	switch s.RunType {
	case _frequently:
		if s.FrequencyInterval == _seconds {
			nextSchedToRun = time.Now().Add(time.Second * time.Duration(s.FrequencyValue)).Unix()
		}
		if s.FrequencyInterval == _minutes {
			nextSchedToRun = time.Now().Add(time.Minute * time.Duration(s.FrequencyValue)).Unix()
		}
		if s.FrequencyInterval == _hours {
			nextSchedToRun = time.Now().Add(time.Hour * time.Duration(s.FrequencyValue)).Unix()
		}

	case _daily:
		runHour, _ := strconv.Atoi(s.runAtHour)
		runMinute, _ := strconv.Atoi(s.runAtMinute)
		nextSchedToRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+1, runHour, runMinute, 0, 0, time.Local).Unix()

	case _weekly:
		runHour, _ := strconv.Atoi(s.runAtHour)
		runMinute, _ := strconv.Atoi(s.runAtMinute)
		today := time.Now()

		nextSchedToRun = time.Date(
			today.Year(),
			today.Month(),
			today.Day()-int(today.Weekday()-s.dayName),
			runHour, runMinute, 0, 0,
			time.Local).Add(7 * 24 * time.Hour).Unix()

	case _monthly:
		runHour, _ := strconv.Atoi(s.runAtHour)
		runMinute, _ := strconv.Atoi(s.runAtMinute)
		today := time.Now()

		nextSchedToRun = time.Date(
			today.Year(),
			today.Month()+1,
			s.monthDay,
			runHour, runMinute, 0, 0,
			time.Local).Unix()

	default:
		msg := s.Name + " is not running due to incorrect or missing parameters"
		itrlog.Errorw(msg, "log_time", time.Now().Format(logDateTimeFormat))
		color.Red(msg)
	}

	// Format next scheduled run
	if s.RunType != _onetime {
		nextSched, _ := formatDT(nextSchedToRun, logDateTimeFormat)
		msg := s.Name + " next schedule to run on: " + nextSched
		itrlog.Infow(msg, "log_time", time.Now().Format(logDateTimeFormat))
		color.Magenta(msg)
	}

	modTask := &Tasks{
		Name:              s.Name,
		RunType:           s.RunType,
		FrequencyInterval: s.FrequencyInterval,
		FrequencyValue:    s.FrequencyValue,
		ExecuteFunc:       s.ExecuteFunc,
		runAtHour:         s.runAtHour,
		runAtMinute:       s.runAtMinute,
		monthDay:          s.monthDay,
		monthName:         s.monthName,
		isRunAt:           s.isRunAt,
		nextRunTime:       nextSchedToRun,
		lastRunTime:       time.Now().Unix(),
	}
	t.TaskList[s.Name] = []Tasks{*modTask}
}

// Reset clear all scheduled tasks
func (t *TaskScheduler) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	TS.TaskList = make(map[string][]Tasks)
	msg := `reloading task schedulers...`
	itrlog.Warnw(msg, "log_time", time.Now().Format(logDateTimeFormat))
	color.Yellow(msg)
}

// Format the DateTime value
func formatDT(dt int64, dtFormat string) (string, error) {
	i, err := strconv.ParseInt(fmt.Sprintf("%v", dt), 10, 64)
	if err != nil {
		panic(err)
	}
	if len(strings.TrimSpace(dtFormat)) == 0 {
		dtFormat = logDateTimeFormat
	}
	dtf := time.Unix(i, 0).Format(dtFormat)
	return dtf, nil
}

// Get the last day of each current month
func getLastDayOfMonth(day int, month time.Month) int {
	// Get the current DateTime and get the last day of this month
	now := time.Now()
	currentYear, _, _ := now.Date()
	firstDayOfMonth := time.Date(currentYear, month, 1, 0, 0, 0, 0, time.Local)
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1).Day()

	switch {
	case day == 0:
		day = lastDayOfMonth // Set default as the last day of this month
	case day > lastDayOfMonth:
		day = lastDayOfMonth // Set the last day of this month
	}
	return day
}
