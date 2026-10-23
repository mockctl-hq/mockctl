package types

import (
	"context"

	"github.com/mockctl-hq/mockctl/internal/app"
)

// TUIContext holds the global dependencies injected into the UI layer.
// This ensures that child components can safely trigger backend actions without global variables.
type TUIContext struct {
	Ctx        context.Context // Context tied to the BubbleTea lifecycle
	Core       *app.App        // The Mock:ctl Application Core
	ActivePort int             // The bound port of the local daemon for API testing
	Settings   *SettingsConfig // Live user settings — pointer so changes propagate instantly
}
