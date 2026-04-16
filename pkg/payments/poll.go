package payments

import "time"

// PollOptions configures the PollUntilFinal helper.
type PollOptions struct {
	// Interval is the delay between each poll attempt. Defaults to 2 seconds.
	Interval time.Duration
}

func (o *PollOptions) applyDefaults() {
	if o.Interval == 0 {
		o.Interval = 2 * time.Second
	}
}
