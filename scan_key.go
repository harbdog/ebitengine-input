package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// KeyScanStatus represents the KeyScanner.Scan operation result.
type KeyScanStatus int

const (
	KeyScanUnchanged KeyScanStatus = iota
	KeyScanCompleted
)

// KeyScanner checks the currently pressed keys and buttons and tries to map them
// to a local Key type that can be used in a Keymap.
//
// Use NewKeyScanner to create a usable object of this type.
//
// Experimental: this is a part of a key remapping API, which is not stable yet.
type KeyScanner struct {
	canScan bool
	h       *Handler

	// uses special handlers to detect certain events for key and axes scanning purposes
	_scanMouseKeyHelper *Handler
	_scanAxesHelper     *Handler
}

// NewKeyScanner creates a key scanner for the specifier input Handler.
//
// You don't have to create a new scanner for every remap; they can be reused.
//
// It's important to have the correct Handler though: their ID is used to
// check the appropriate device keys.
//
// Experimental: this is a part of a key remapping API, which is not stable yet.
func NewKeyScanner(h *Handler) *KeyScanner {
	return &KeyScanner{h: h}
}

// Scan reads the buttons state and tries to map them to a Key.
//
// It's intended to work with keyboard keys as well as mouse and gamepad buttons.
//
// This function should be called on every frame where you're reading
// the new keybind combination.
// See the remap example for more info.
//
// The function can return these result statuses:
// * Unchanged - nothing updated since the last Scan() operation
// * Completed - the user finished specifying the keys combination, you can use the Key as a new binding
func (s *KeyScanner) Scan() (Key, KeyScanStatus) {
	if s == nil || s.h == nil || s.h.sys == nil {
		panic("KeyScanner must be initialized using: NewKeyScanner(*Handler)")
	}

	// Note that this function may not be needed by some users,
	// so we're better of making it as independent as possible, so it
	// doesn't make the package more expensive if you don't use it.
	//
	// This function doesn't have to be very fast, but it should be relatively
	// inexpensive for the "no keys were pressed" case.
	// When some keys combo is being pressed, it's OK to spend some resources.
	k, status := s.scanKeyboard(nil, nil)
	if status == KeyScanUnchanged {
		// scan for mouse buttons
		k, status = s.scanMouse(nil, nil)
	}
	if status == KeyScanUnchanged {
		// scan for gamepad buttons
		k, status = s.scanGamepad(nil)
	}

	if !s.canScan {
		if k.name != "" {
			// do not start scanning until preexisting button presses are no longer held
			return Key{}, KeyScanUnchanged
		}
		s.canScan = true
	}

	switch status {
	case KeyScanCompleted:
		s.canScan = false

	}
	return k, status
}

const (
	scanKeyMouseWheelDown Action = iota
	scanKeyMouseWheelUp
)

var (
	specialMouseKeyKeymap = Keymap{
		scanKeyMouseWheelDown: {KeyWheelDown},
		scanKeyMouseWheelUp:   {KeyWheelUp},
	}
)

// newMouseKeyScanHandler creates a new handler used to help scan for
// special mouse keys to be able to detect them.
func newMouseKeyScanHandler(scanHandler *Handler) *Handler {
	return scanHandler.sys.NewHandler(scanHandler.id, specialMouseKeyKeymap)
}

func (s *KeyScanner) scanSpecialMouseKeys() (Key, KeyScanStatus) {
	for a := range specialMouseKeyKeymap {
		if _, ok := s._scanMouseKeyHelper.JustPressedActionInfo(a); ok {
			k := specialMouseKeyKeymap[a][0]
			return k, KeyScanCompleted
		}
	}
	return Key{}, KeyScanUnchanged
}

func (s *KeyScanner) scanMouse(mouseKeys []ebiten.MouseButton, heldKeys []ebiten.Key) (Key, KeyScanStatus) {
	if !s.h.MouseEventsEnabled() {
		return Key{}, KeyScanUnchanged
	}

	if s._scanMouseKeyHelper == nil {
		// special Handler needed to determine certain events using special keymap
		s._scanMouseKeyHelper = newMouseKeyScanHandler(s.h)
	}

	// We will need to do our own "AppendJustReleased" for mouse button presses
	for k := ebiten.MouseButton(0); k < ebiten.MouseButtonMax; k++ {
		if inpututil.IsMouseButtonJustReleased(k) {
			mouseKeys = append(mouseKeys, k)
		}
	}

	containsButtonCode := func(keys []ebiten.MouseButton, code int) bool {
		for _, k := range keys {
			if int(k) == code {
				return true
			}
		}
		return false
	}

	var mappedKey Key

	// map the Ebitengine keys to the local types.
Loop:
	for _, k := range allKeys {
		switch k.kind {
		case keyMouse:
			if containsButtonCode(mouseKeys, k.code) {
				mappedKey = k
				break Loop
			}
		}
	}

	if mappedKey.name == "" {
		// check special mouse key presses which cannot be detected via inpututil funcs (e.g. mouse wheel up/down)
		mappedKey, _ = s.scanSpecialMouseKeys()
	}

	// attach any held key modifiers
	keymod := s.scanKeyModifiers(heldKeys)
	if keymod != ModUnknown {
		switch mappedKey.kind {
		case keyMouse:
			mappedKey = KeyWithModifier(mappedKey, keymod)
		}
	}

	status := KeyScanUnchanged
	if mappedKey.name != "" {
		// mouse button pressed, scan complete
		status = KeyScanCompleted
		s._scanMouseKeyHelper = nil
	}
	return mappedKey, status
}

func (s *KeyScanner) scanGamepad(gamepadKeys []ebiten.StandardGamepadButton) (Key, KeyScanStatus) {
	if !s.h.GamepadEventsEnabled() {
		return Key{}, KeyScanUnchanged
	}

	var handlerID uint8
	if s.h != nil {
		handlerID = s.h.id
	}
	gamepadKeys = inpututil.AppendJustReleasedStandardGamepadButtons(ebiten.GamepadID(handlerID), gamepadKeys)

	if len(gamepadKeys) == 0 {
		return Key{}, KeyScanUnchanged
	}

	containsButtonCode := func(keys []ebiten.StandardGamepadButton, code int) bool {
		for _, k := range keys {
			if int(k) == code {
				return true
			}
		}
		return false
	}

	var mappedKey Key

	// map the Ebitengine keys to the local types.
Loop:
	for _, k := range allKeys {
		switch k.kind {
		case keyGamepad:
			if containsButtonCode(gamepadKeys, k.code) {
				mappedKey = k
				break Loop
			}
		}
	}

	status := KeyScanUnchanged
	if mappedKey.name != "" {
		status = KeyScanCompleted
	}
	return mappedKey, status
}

func (s *KeyScanner) scanKeyboard(keys []ebiten.Key, heldKeys []ebiten.Key) (Key, KeyScanStatus) {
	if !s.h.KeyboardEventsEnabled() {
		return Key{}, KeyScanUnchanged
	}

	keys = inpututil.AppendJustReleasedKeys(keys)
	if len(keys) == 0 {
		// We're still collecting the keys.
		return Key{}, KeyScanUnchanged
	}

	containsKeyCode := func(keys []ebiten.Key, code int) bool {
		for _, k := range keys {
			if int(k) == code {
				return true
			}
		}
		return false
	}

	// map the Ebitengine keys to the local types.
	// In theory, we could generate a big LUT to make this mapping very fast.
	// But this would mean more data reserved for this package.
	// Since this part of the code is not that performance-sensitive,
	// we'll handle it in a less efficient, but less memory-hungry way.
	var mappedKey Key

Loop:
	for _, k := range allKeys {
		switch k.kind {
		case keyKeyboard:
			if containsKeyCode(keys, k.code) {
				mappedKey = k
				break Loop
			}
		}
	}

	// attach any held key modifiers
	keymod := s.scanKeyModifiers(heldKeys)
	if keymod != ModUnknown {
		switch mappedKey.kind {
		case keyKeyboard:
			mappedKey = KeyWithModifier(mappedKey, keymod)
		}
	}

	status := KeyScanUnchanged
	if mappedKey.name != "" {
		status = KeyScanCompleted
	}
	return mappedKey, status
}

func (s *KeyScanner) scanKeyModifiers(heldKeys []ebiten.Key) KeyModifier {
	if !s.h.KeyboardEventsEnabled() {
		return 0
	}
	heldKeys = inpututil.AppendPressedKeys(heldKeys)

	var ctrlKey Key
	var shiftKey Key
	for _, k := range heldKeys {
		switch k {
		case ebiten.KeyControl:
			ctrlKey = KeyControl
		case ebiten.KeyControlLeft:
			ctrlKey = KeyControlLeft
		case ebiten.KeyControlRight:
			ctrlKey = KeyControlRight
		case ebiten.KeyShift:
			shiftKey = KeyShift
		case ebiten.KeyShiftLeft:
			shiftKey = KeyShiftLeft
		case ebiten.KeyShiftRight:
			shiftKey = KeyShiftRight
		}
	}
	hasCtrl := ctrlKey.name != ""
	hasShift := shiftKey.name != ""

	var keymod KeyModifier
	switch {
	case hasCtrl && hasShift:
		keymod = ModControlShift
	case hasCtrl:
		keymod = ModControl
	case hasShift:
		keymod = ModShift
	}
	return keymod
}
