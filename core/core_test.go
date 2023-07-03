package core

import (
	"testing"

	"github.com/siathema/goInvadeSpace/memory"
)


func TestCore(t *testing.T) {
    m := memory.NewMainMemory(nil)
    c := New()

    testOp := []uint8{0x00, 0x00, 0x00}
    c.ExecuteOpcode(testOp, m)
    if  c.PC == 0x00 {
        t.Errorf("Expected c.PC == 1, got=%d", c.PC)
    }
    if !testCoreStateEq(c, c) {
        t.Errorf("Don't work Jack")
    }
}

func Testu16u8Conversion(t *testing.T) {
    testsU16 := []struct {
        input       uint16 
        expectedHi  uint8
        expectedLow uint8
    }{
        {0x0000, 0x00, 0x00},
        {0xFFFF, 0xFF, 0xFF},
        {0xDEAD, 0xDE, 0xAD},
        {0x0050, 0x00, 0x50},
        {0xFF00, 0xFF, 0x00},
    }
    
    for _, n := range testsU16 {
        if hi, low := u16ToHiLowU8(n.input); hi != n.expectedHi &&
            low != n.expectedLow {
                t.Errorf("Expected n.expectedHi=%d, n.expectedLow=%d, got=(hi=%d, low=%d)",
                    n.expectedHi, n.expectedLow, hi, low)
            }
    }
    for _, n := range testsU16 {
        if r := u8HiLowRoU16(n.expectedHi, n.expectedLow); r != n.input {
                t.Errorf("Expected n.input=%d, got=%d)",
                    n.input, r)
            }
    }

}


func testCoreStateEq(a *Core8080, b *Core8080) bool {
    if a.A != b.A && a.B != b.B && a.C != b.C && a.D != b.D && a.E != b.E &&
        a.H != b.H && a.L != b.L && a.cycles != b.cycles && a.SP != b.SP &&
        a.PC != b.PC && a.Flags != b.Flags && a.Irq != b.Irq && a.Sync != b.Sync {
       return false 
    }
    return true
}
