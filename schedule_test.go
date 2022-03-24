package cronjob_test

import (
	"testing"
	"time"

	"github.com/Lambels/cronjob"
)

func TestSchedule(t *testing.T) {
	nowTesting := time.Date(
		2022,
		10,
		7,
		10,
		5,
		0,
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
			schedule: cronjob.In(nowTesting, time.Hour*60),

			expectedDur: time.Hour * 60,
		},

		// Every.
		{
			schedule: cronjob.Every(nowTesting.Add(-time.Hour*20), time.Duration(time.Hour*50)),

			expectedDur: time.Hour * 30,
		},

		// EveryFixed.
		{
			schedule: cronjob.EveryFixed(nowTesting, time.Hour*3),

			expectedDur: time.Hour * 2,
		},
	}

	for _, c := range cases {
		want := c.expectedDur
		v := c.schedule.Calculate(nowTesting)

		if want != v {
			t.Errorf("Expected: %v Got: %v", want, v)
		}
	}
}
