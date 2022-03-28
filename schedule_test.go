package cronjob_test

import (
	"testing"
	"time"

	"github.com/Lambels/cronjob"
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
		schedule    cronjob.Schedule
		expectedDur time.Duration
	}{
		// At.
		{
			schedule: cronjob.At(time.Date(
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
			schedule: cronjob.In(nowTesting, 5*time.Minute),

			expectedDur: 5 * time.Minute,
		},

		// Every.
		{
			schedule: cronjob.Every(5 * time.Minute),

			expectedDur: 5 * time.Minute,
		},

		// EveryFixed.
		{
			schedule: cronjob.EveryFixed(10 * time.Minute),

			expectedDur: 5 * time.Minute,
		},
	}

	for _, c := range cases {
		var got time.Duration

		switch c.schedule.(type) {
		case cronjob.CyclicSchedule:
			sched := c.schedule.(cronjob.CyclicSchedule)
			sched.MoveNextAvtivation(nowTesting)
			got = sched.Calculate(nowTesting)

		case cronjob.Schedule:
			got = c.schedule.Calculate(nowTesting)

		}

		if want := c.expectedDur; got != want {
			t.Fatalf("want: %v got: %v\n", want, got)
		}
	}
}
