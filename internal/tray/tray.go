package tray

// Action is a user intent from the tray icon or menu.
// Pure mapping, unit-tested without Win32.
type Action int

const (
	ActionNone Action = iota
	ActionShow
	ActionSettings
	ActionHide
	ActionExit
)

// Menu IDs shown in the popup menu.
const (
	menuShow     = 1
	menuSettings = 2
	menuHide     = 3
	menuExit     = 4
)

// ActionForMenuID maps a popup menu command ID to an Action.
// Unknown IDs yield ActionNone and must be ignored.
func ActionForMenuID(id uint32) Action {
	switch id {
	case menuShow:
		return ActionShow
	case menuSettings:
		return ActionSettings
	case menuHide:
		return ActionHide
	case menuExit:
		return ActionExit
	default:
		return ActionNone
	}
}

// Callbacks wires tray intents to the app. All run on the tray thread;
// implementations must be thread-safe (runtime calls are).
type Callbacks struct {
	OnShow     func()
	OnSettings func()
	OnHide     func()
	OnExit     func()
}

func (c Callbacks) dispatch(a Action) {
	switch a {
	case ActionShow:
		if c.OnShow != nil {
			c.OnShow()
		}
	case ActionSettings:
		if c.OnSettings != nil {
			c.OnSettings()
		}
	case ActionHide:
		if c.OnHide != nil {
			c.OnHide()
		}
	case ActionExit:
		if c.OnExit != nil {
			c.OnExit()
		}
	}
}
