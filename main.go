package main

import (
	"fmt"
	"os"

    "github.com/siathema/goInvadeSpace/core"
    "github.com/siathema/goInvadeSpace/memory"
)

func main() {
    fmt.Println("Hello weeb!")
    romData, err := os.ReadFile("roms/invaders.rom")
    if err != nil {
        panic(err)
    }

    core.New()
    mem := new(memory.MainMemory)
    mem.Init(romData)
    fmt.Printf("Memory Initialized with %dK of rom and %dK of ram!\n",
        len(mem.Rom)/1024, len(mem.Ram)/1024)

    for {
        core.RunTick(mem)
        if core.PC == uint16(memory.Kilobytes(8)) {
            break
        }
    }
    //fmt.Println(memory.Rom)
    //fmt.Println(len(romData))
}
