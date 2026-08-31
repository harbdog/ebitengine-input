package input

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func newTestKeyScanner(config SystemConfig) *KeyScanner {
	var sys System
	sys.Init(config)
	testHandler := sys.NewHandler(0, Keymap{})
	testScanner := NewKeyScanner(testHandler)
	testScanner.canScan = true
	return testScanner
}

func TestScanKeyboard(t *testing.T) {
	testScanner := newTestKeyScanner(SystemConfig{DevicesEnabled: KeyboardDevice})

	tests := []struct {
		keys     []ebiten.Key
		heldKeys []ebiten.Key
		want     Key
	}{
		// Sanity tests.
		{[]ebiten.Key{}, nil, Key{}},

		// The simple cases with a single key.
		{[]ebiten.Key{ebiten.KeyB}, nil, KeyB},
		{[]ebiten.Key{ebiten.KeyEnter}, nil, KeyEnter},
		{[]ebiten.Key{ebiten.KeyControlLeft}, nil, KeyControlLeft},
		{[]ebiten.Key{ebiten.KeyControlRight}, nil, KeyControlRight},
		{[]ebiten.Key{ebiten.KeyControl}, nil, KeyControl},

		// Multiple key candidates without a way to merge them into a single Key.
		{[]ebiten.Key{ebiten.KeyB, ebiten.KeyA}, nil, KeyA},
		{[]ebiten.Key{ebiten.KeyA, ebiten.KeyB}, nil, KeyA},

		// Control modifiers.
		{[]ebiten.Key{ebiten.KeyC}, []ebiten.Key{ebiten.KeyControlLeft}, KeyWithModifier(KeyC, ModControl)},
		{[]ebiten.Key{ebiten.KeyC}, []ebiten.Key{ebiten.KeyControlRight}, KeyWithModifier(KeyC, ModControl)},
		{[]ebiten.Key{ebiten.KeyC}, []ebiten.Key{ebiten.KeyControl}, KeyWithModifier(KeyC, ModControl)},
		{[]ebiten.Key{ebiten.KeyE}, []ebiten.Key{ebiten.KeyControlLeft}, KeyWithModifier(KeyE, ModControl)},
		{[]ebiten.Key{ebiten.KeyE}, []ebiten.Key{ebiten.KeyControlRight, ebiten.KeyA}, KeyWithModifier(KeyE, ModControl)},

		// Shift modifiers.
		{[]ebiten.Key{ebiten.KeyF}, []ebiten.Key{ebiten.KeyShiftLeft}, KeyWithModifier(KeyF, ModShift)},
		{[]ebiten.Key{ebiten.KeyF}, []ebiten.Key{ebiten.KeyShiftRight}, KeyWithModifier(KeyF, ModShift)},
		{[]ebiten.Key{ebiten.KeyF}, []ebiten.Key{ebiten.KeyShift}, KeyWithModifier(KeyF, ModShift)},

		// Control+Shift modifiers.
		{[]ebiten.Key{ebiten.KeyC}, []ebiten.Key{ebiten.KeyControlLeft, ebiten.KeyControl, ebiten.KeyShiftLeft, ebiten.KeyShift}, KeyWithModifier(KeyC, ModControlShift)},
		{[]ebiten.Key{ebiten.KeyA}, []ebiten.Key{ebiten.KeyControl, ebiten.KeyShift}, KeyWithModifier(KeyA, ModControlShift)},
		{[]ebiten.Key{ebiten.KeyA}, []ebiten.Key{ebiten.KeyControlLeft, ebiten.KeyShiftLeft}, KeyWithModifier(KeyA, ModControlShift)},
		{[]ebiten.Key{ebiten.KeyA}, []ebiten.Key{ebiten.KeyControlRight, ebiten.KeyShiftRight}, KeyWithModifier(KeyA, ModControlShift)},
		{[]ebiten.Key{ebiten.KeyA}, []ebiten.Key{ebiten.KeyControlLeft, ebiten.KeyShiftRight}, KeyWithModifier(KeyA, ModControlShift)},
	}

	for i, test := range tests {
		have, _ := testScanner.scanKeyboard(test.keys, test.heldKeys)
		if have != test.want {
			t.Fatalf("test[%d] failed:\nhave: %s (%#v)\nwant: %s (%#v)",
				i, have, have, test.want, test.want)
		}
	}

	// test mouse press denied due to SystemConfig.DevicesEnabled mask
	key, _ := testScanner.scanMouse([]ebiten.MouseButton{ebiten.MouseButtonMiddle}, nil)
	if key.name != "" {
		t.Fatalf("test device mask failed:\nkey: %s (%#v)", key, key)
	}
}

func TestScanMouse(t *testing.T) {
	testScanner := newTestKeyScanner(
		// enabling mouse with keyboard device for modifiers
		SystemConfig{DevicesEnabled: MouseDevice | KeyboardDevice},
	)

	tests := []struct {
		keys     []ebiten.MouseButton
		heldKeys []ebiten.Key
		want     Key
	}{
		// Sanity tests.
		{[]ebiten.MouseButton{}, nil, Key{}},

		// The simple cases with a single key.
		{[]ebiten.MouseButton{ebiten.MouseButtonLeft}, nil, KeyMouseLeft},
		{[]ebiten.MouseButton{ebiten.MouseButtonRight}, nil, KeyMouseRight},
		{[]ebiten.MouseButton{ebiten.MouseButtonMiddle}, nil, KeyMouseMiddle},
		{[]ebiten.MouseButton{ebiten.MouseButton3}, nil, KeyMouseBack},
		{[]ebiten.MouseButton{ebiten.MouseButton4}, nil, KeyMouseForward},

		// Multiple key candidates without a way to merge them into a single Key.
		{[]ebiten.MouseButton{ebiten.MouseButtonLeft, ebiten.MouseButtonRight}, nil, KeyMouseLeft},
		{[]ebiten.MouseButton{ebiten.MouseButtonRight, ebiten.MouseButtonLeft}, nil, KeyMouseLeft},

		// Control modifiers.
		{[]ebiten.MouseButton{ebiten.MouseButtonLeft}, []ebiten.Key{ebiten.KeyControl}, KeyWithModifier(KeyMouseLeft, ModControl)},

		// Shift modifiers.
		{[]ebiten.MouseButton{ebiten.MouseButtonRight}, []ebiten.Key{ebiten.KeyShift}, KeyWithModifier(KeyMouseRight, ModShift)},

		// Control+Shift modifiers.
		{[]ebiten.MouseButton{ebiten.MouseButtonLeft}, []ebiten.Key{ebiten.KeyControlLeft, ebiten.KeyShiftLeft}, KeyWithModifier(KeyMouseLeft, ModControlShift)},
		{[]ebiten.MouseButton{ebiten.MouseButtonRight}, []ebiten.Key{ebiten.KeyControlRight, ebiten.KeyShiftRight}, KeyWithModifier(KeyMouseRight, ModControlShift)},
		{[]ebiten.MouseButton{ebiten.MouseButtonMiddle}, []ebiten.Key{ebiten.KeyControl, ebiten.KeyShift}, KeyWithModifier(KeyMouseMiddle, ModControlShift)},
	}

	for i, test := range tests {
		have, _ := testScanner.scanMouse(test.keys, test.heldKeys)
		if have != test.want {
			t.Fatalf("test[%d] failed:\nhave: %s (%#v)\nwant: %s (%#v)",
				i, have, have, test.want, test.want)
		}
	}

	// test gamepad press denied due to SystemConfig.DevicesEnabled mask
	key, _ := testScanner.scanGamepad([]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightBottom})
	if key.name != "" {
		t.Fatalf("test device mask failed:\nkey: %s (%#v)", key, key)
	}
}

func TestScanGamepad(t *testing.T) {
	testScanner := newTestKeyScanner(SystemConfig{DevicesEnabled: GamepadDevice})

	tests := []struct {
		keys []ebiten.StandardGamepadButton
		want Key
	}{
		// Sanity tests.
		{[]ebiten.StandardGamepadButton{}, Key{}},

		// The simple cases with a single button.
		{[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightBottom}, KeyGamepadA},
		{[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightRight}, KeyGamepadB},
		{[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightLeft}, KeyGamepadX},
		{[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightTop}, KeyGamepadY},

		// Multiple button candidates without a way to merge them into a single Key.
		{[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightBottom, ebiten.StandardGamepadButtonLeftStick}, KeyGamepadA},
		{[]ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftStick, ebiten.StandardGamepadButtonRightBottom}, KeyGamepadA},
	}

	for i, test := range tests {
		have, _ := testScanner.scanGamepad(test.keys)
		if have != test.want {
			t.Fatalf("test[%d] failed:\nhave: %s (%#v)\nwant: %s (%#v)",
				i, have, have, test.want, test.want)
		}
	}

	// test keyboard press denied due to SystemConfig.DevicesEnabled mask
	key, _ := testScanner.scanKeyboard([]ebiten.Key{ebiten.KeyEscape}, nil)
	if key.name != "" {
		t.Fatalf("test device mask failed:\nkey: %s (%#v)", key, key)
	}
}
