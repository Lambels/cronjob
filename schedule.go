package cronjob

import "time"

// ConstantSchedule ------------------------------------------------------------------

type ConstantSchedule struct {
	// is the date on which the job is scheduled to run on.
	at time.Time
}

// Calculate calculates the duration of time in which the schedule will be active
// in reference to now parameter.
func (s *ConstantSchedule) Calculate(now time.Time) time.Duration {
	return s.at.Sub(now)
}

// FixedCyclicSchedule ------------------------------------------------------------------

type FixedCyclicSchedule struct {
	every time.Duration
}

func (s *FixedCyclicSchedule) Calculate(now time.Time) time.Duration {
	switch h, m, s := breakTime(int(s.every.Minutes())); {
	case h > 0:
		nextTime := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			h*(h/now.Hour()+1),
			m,
			s,
			0,
			now.Location(),
		)

		return nextTime.Sub(now)

	case m > 0:
		return 0

	default:
		return 0

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

type CyclicSchedule struct {
	// every gets added to the current time to get the next activation cycle.
	every time.Duration

	// nextActivation will hold the time of the next activation
	nextActivation time.Time
}

func (s *CyclicSchedule) Calculate(now time.Time) time.Duration {
	// next activation isnt calculated, calculate it.
	if s.nextActivation.IsZero() {
		s.nextActivation = now.Add(s.every)
	}

	// next activation is outdated, calculate it.
	if s.nextActivation.Before(now) {
		s.nextActivation = now.Add(s.every)
	}

	return s.nextActivation.Sub(now)
}
