package main

import (
    "fmt"
)
// Use for main memory and io memory
type MemoryMap struct {
    DBus *uint8
    ABus *uint16
    Write bool
    RamSz, RomSz uint
    Rom, Ram *[]uint8
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
}
