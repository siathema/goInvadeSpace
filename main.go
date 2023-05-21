package main

import (
    "fmt"
    "os"
)

func Kilobytes(kb uint) uint {
    return kb * 1024
}
// Use for main memory and io memory
type MemoryMap struct {
    DBus *uint8
    ABus *uint16
    Write bool
    Rom, Ram []uint8
}

type MainMemory MemoryMap

func (mem *MainMemory) Init(romData []uint8) {
   mem.Rom = romData
   mem.Ram = make([]uint8, Kilobytes(8)) 
   fmt.Println(len(mem.Ram))
}


type core8080 struct {
    A, B, C, D, E, H, L uint8
    SP, PC uint16
    DBus *uint8
    ABus *uint16
    Irq, Write, Sync bool
    cycles uint64
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
