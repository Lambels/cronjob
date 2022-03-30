package cronjob

import (
	"testing"
	"time"
)

func TestSchedule(t *testing.T) {
	nowTesting := time.Date(
		2022, // year
		10,   // month
		7,    // day
		10,   // hour
		5,    // min
		0,    // sec
		0,
		time.Local,
	)

	cases := []struct {
		schedule    Schedule
		expectedDur time.Duration
	}{
		// At.
		{
			schedule: At(time.Date(
				2022,
				10,
				7,
				10,
				5,
				0,
				0,
				time.Local,
			)),

			expectedDur: nowTesting.Sub(time.Date(
				2022,
				10,
				7,
				10,
				5,
				0,
				0,
				time.Local,
			)),
		},

		// In.
		{
			schedule: In(nowTesting, 5*time.Minute),

			expectedDur: 5 * time.Minute,
		},

		// Every.
		{
			schedule: Every(5 * time.Minute),

			expectedDur: 5 * time.Minute,
		},

		// EveryFixed.
		{
			schedule: EveryFixed(10 * time.Minute),

			expectedDur: 5 * time.Minute,
		},
	}

	for _, c := range cases {
		var got time.Duration

		switch c.schedule.(type) {
		case CyclicSchedule:
			sched := c.schedule.(CyclicSchedule)
			sched.MoveNextAvtivation(nowTesting)
			got = sched.Calculate(nowTesting)

		case Schedule:
			got = c.schedule.Calculate(nowTesting)

		}

		if want := c.expectedDur; got != want {
			t.Fatalf("want: %v got: %v\n", want, got)
		}
	}
}
