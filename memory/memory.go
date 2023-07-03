package memory

import(
    "errors"
)

func Kilobytes(kb uint) uint {
    return kb * 1024
}

// Use for main memory and io memory
type MemoryMap struct {
    WriteEnable bool
    Rom, Ram []uint8
}

type MainMemory MemoryMap

func NewMainMemory(romData []uint8) *MainMemory {
    m := &MainMemory{
        WriteEnable: false,
    }
    if romData == nil {
        m.Rom = make([]uint8, Kilobytes(8))
    } else {
        m.Rom = romData
    }
    m.Ram = make([]uint8, Kilobytes(8)) 

    return m
}

func (mem *MainMemory) Read(addr uint16) uint8 {
    if uint(addr) < Kilobytes(8) {
        return mem.Rom[addr]
    } else if uint(addr) < Kilobytes(16){
        return mem.Ram[addr - uint16(Kilobytes(8))] 
    } else {
        // this needs to error out
        return 0
    }
}

func (mem *MainMemory) Write(addr uint16, data uint8) error {
    if uint(addr) < Kilobytes(8) {
        mem.Rom[addr] = data    
        return nil
    }  else if uint(addr) < Kilobytes(16) {
        mem.Ram[addr - uint16(Kilobytes(8))] = data
        return nil
    } else {
        return errors.New("Address out of bounds!")
    }
}
