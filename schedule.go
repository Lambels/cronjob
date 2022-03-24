package cronjob

import "time"

// At returns a schedule that runs at: at (field).
func At(at time.Time) Schedule {
	return &constantSchedule{
		at: at,
	}
}

// In returns a schedule that runs from now (field) in offset (field).
func In(now time.Time, offset time.Duration) Schedule {
	return &constantSchedule{
		at: now.Add(offset),
	}
}

// Every runs from now: now (field) in constant increments of every (field).
func Every(now time.Time, every time.Duration) Schedule {
	schedule := &cyclicSchedule{
		every: every,
	}
	schedule.Calculate(now)

	return schedule
}

// EveryFixed finds the next time interval: every (field) and runs it at that time.
//
// example:
//	cronjob.EveryFixed(time.Now(), time.Hour * 3)
// the schedule will find the next 3 hour interval to run at.
//
// possible 3 hour intervals: 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, 21:00, 24:00
func EveryFixed(now time.Time, every time.Duration) Schedule {
	schedule := &fixedCyclicSchedule{
		every: every,
	}
	schedule.Calculate(now)

	return schedule
}

// ConstantSchedule ------------------------------------------------------------------

type constantSchedule struct {
	// is the date on which the job is scheduled to run on.
	at time.Time
}

// Calculate calculates the duration of time in which the schedule will be active
// in reference to now parameter.
func (s *constantSchedule) Calculate(now time.Time) time.Duration {
	return s.at.Sub(now)
}

// FixedCyclicSchedule ------------------------------------------------------------------

type fixedCyclicSchedule struct {
	every time.Duration

	nextActivation time.Time
}

func (s *fixedCyclicSchedule) Calculate(now time.Time) time.Duration {
	// next activation isnt calculated, calculate it.
	// or if next activation is outdated, calculate it.
	if s.nextActivation.IsZero() || s.nextActivation.Before(now) {
		s.nextActivation = s.calculateNextInterval(now)
	}

	return s.nextActivation.Sub(now)
}

func (s *fixedCyclicSchedule) calculateNextInterval(now time.Time) time.Time {
	switch h, m, s := breakTime(int(s.every.Seconds())); {
	case h > 0:
		return time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			h*(h/now.Hour()+1),
			m,
			s,
			0,
			now.Location(),
		)

	case m > 0:
		return time.Time{}

	default:
		return time.Time{}
	}
}

func breakTime(secs int) (hour, minute, second int) {
	hour = secs / 3600
	secs = secs - hour*3600

	minute = secs / 60
	secs = secs - minute*60

	second = secs
	return
}

// CyclicSchedule ------------------------------------------------------------------

type cyclicSchedule struct {
	// every gets added to the current time to get the next activation cycle.
	every time.Duration

	// nextActivation will hold the time of the next activation
	nextActivation time.Time
}

func (s *cyclicSchedule) Calculate(now time.Time) time.Duration {
	// next activation isnt calculated, calculate it.
	// or if next activation is outdated, calculate it.
	if s.nextActivation.IsZero() || s.nextActivation.Before(now) {
		s.nextActivation = now.Add(s.every)
	}

	return s.nextActivation.Sub(now)
}
