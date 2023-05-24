package main

import (
    "fmt"
    "os"
)

func Kilobytes(kb uint) uint {
    return kb * 1024
}

type Bus8 uint8
type Bus16 uint16

// Use for main memory and io memory
type MemoryMap struct {
    DBus *Bus8
    ABus *Bus16
    Write bool
    Rom, Ram []uint8
}

type MainMemory MemoryMap

func (mem *MainMemory) Init(romData []uint8) {
   mem.Rom = romData
   mem.Ram = make([]uint8, Kilobytes(8)) 
   fmt.Println(len(mem.Ram))
}

func (mem *MainMemory) Read(addr uint16) uint8 {
    return 0
}

type Core8080 struct {
    A, B, C, D, E, H, L uint8
    SP, PC uint16
    DBus *Bus8
    ABus *Bus16
    Irq, Write, Sync bool
    cycles uint64
}

func (core *Core8080) Init(dataBus *Bus8, addrBus *Bus16) {
    core.A = 0
    core.B = 0
    core.C = 0
    core.D = 0
    core.E = 0
    core.H = 0
    core.L = 0
    core.DBus = dataBus
    core.ABus = addrBus
    core.Irq = false
    core.Write = false
    core.Sync = false
}

func (core *Core8080) Run(mem *MainMemory) {
    // Read opcode from memory
    core.Write = false
    opcode := make([]uint8, 3)
    for i :=0; i < 3; i++ {
        opcode[i] = mem.Read(core.PC + uint16(i))
    }
    core.ExecuteOpcode(opcode)

}

func (core *Core8080) ExecuteOpcode(opcode []uint8) {
    return
}

func main() {
    fmt.Println("Hello weeb!")
    romData, err := os.ReadFile("roms/invaders.rom")
    if err != nil {
        panic(err)
    }


    memory := new(MainMemory)
    memory.Init(romData)
    fmt.Println(memory.Rom)
    fmt.Println(len(romData))
    
    fmt.Printf("Memory Initialized with %dK of rom and %dK of ram!\n", len(memory.Rom)/1024, len(memory.Ram)/1024)
}
