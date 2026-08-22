package input

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestScanKeyboard(t *testing.T) {
	testHandler := &Handler{id: 0}
	testScanner := NewKeyScanner(testHandler)
	testScanner.canScan = true

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
		{[]ebiten.Key{ebiten.KeyE}, []ebiten.Key{ebiten.KeyControlRight}, KeyWithModifier(KeyE, ModControl)},

		// Shift modifiers.
		{[]ebiten.Key{ebiten.KeyF}, []ebiten.Key{ebiten.KeyShiftLeft, ebiten.KeyShift}, KeyWithModifier(KeyF, ModShift)},

		// Control+Shift modifiers.
		{[]ebiten.Key{ebiten.KeyC}, []ebiten.Key{ebiten.KeyControlRight, ebiten.KeyShiftLeft}, KeyWithModifier(KeyC, ModControlShift)},
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
}
