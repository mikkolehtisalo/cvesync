package tracker

import (
	"github.com/mikkolehtisalo/cvesync/nvd"
)

type Tracker interface {
	Init()
	// Returns ticket system's ticket ID when creating new one
	Add(nvd.Entry) (string, error)
	// Refer also to the ticket system's ticket ID
	Update(nvd.Entry, string) error
}
