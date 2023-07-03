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

func testCoreStateEq(a *Core8080, b *Core8080) bool {
    if a.A != b.A && a.B != b.B && a.C != b.C && a.D != b.D && a.E != b.E &&
        a.H != b.H && a.L != b.L && a.cycles != b.cycles && a.SP != b.SP &&
        a.PC != b.PC && a.Flags != b.Flags && a.Irq != b.Irq && a.Sync != b.Sync {
       return false 
    }
    return true
}
