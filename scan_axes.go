package input

const (
	scanAxesMouseWheel Action = iota
	scanAxesGamepadLStick
	scanAxesGamepadRStick
	scanAxesActionCount // always last to keep accurate count to iterate over
)

var (
	// use special handler and keymap to detect axes action events for key scanning purposes
	scanAxesHandler *Handler
	scanAxesKeymap  = Keymap{
		scanAxesMouseWheel:    {KeyWheelVertical},
		scanAxesGamepadLStick: {KeyGamepadLStickMotion},
		scanAxesGamepadRStick: {KeyGamepadRStickMotion},
	}
)

func newScanAxesHandler(scanHandler *Handler) *Handler {
	return scanHandler.sys.NewHandler(scanHandler.id, scanAxesKeymap)
}

// ScanAxes reads the axis changed state and tries to map them to a Key.
//
// It's intended to work with mouse, mouse wheel, and gamepad stick axes.
//
// This function should be called on every frame where you're reading
// the new keybind combination.
// See the remap example for more info.
//
// The function can return these result statuses:
// * Unchanged - nothing updated since the last Scan() operation
// * Completed - the user finished specifying the keys combination, you can use the Key as a new binding
func (s *KeyScanner) ScanAxes() (Key, KeyScanStatus) {

	// FIXME: if s.h == nil, panic or err because this method requires handler instance to be provided?

	if scanAxesHandler == nil {
		// special Handler is needed to determine axis events using special keymap
		scanAxesHandler = newScanAxesHandler(s.h)
	}

	k, status := s.scanAxesEvents()

	// FIXME: mouse motion axes do not appear to be available in ebitengine-input, can we add it?

	switch status {
	case KeyScanCompleted:
		s.canScan = false
		scanAxesHandler = nil
	}
	return k, status
}

func (s *KeyScanner) scanAxesEvents() (Key, KeyScanStatus) {
	for a := Action(0); a < scanAxesActionCount; a++ {
		if _, ok := scanAxesHandler.JustPressedActionInfo(a); ok {
			k := scanAxesKeymap[a][0]
			return k, KeyScanCompleted
		}
	}
	return Key{}, KeyScanUnchanged
}
