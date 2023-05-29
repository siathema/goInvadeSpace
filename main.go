package main

import (
	"errors"
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
    WriteEnable bool
    Rom, Ram []uint8
}

type MainMemory MemoryMap

func (mem *MainMemory) Init(romData []uint8) {
   mem.Rom = romData
   mem.Ram = make([]uint8, Kilobytes(8)) 
   fmt.Println(len(mem.Ram))
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

type Core8080 struct {
    A, B, C, D, E, H, L, Flags uint8
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
    core.Flags = 0
    core.DBus = dataBus
    core.ABus = addrBus
    core.Irq = false
    core.Write = false
    core.Sync = false
    core.PC = 0
    core.SP = 0
}

func (core *Core8080) RunTick(mem *MainMemory) {
    // Read opcode from memory
    core.Write = false
    opcode := make([]uint8, 3)
    for i :=0; i < 3; i++ {
        opcode[i] = mem.Read(core.PC + uint16(i))
    }
    core.ExecuteOpcode(opcode, mem)

}

func (core *Core8080) UpdateFlags(after uint8, carry bool) {
    // Zero flag
    if(after == 0) {
        core.Flags |= 0x40
    } else {
        core.Flags &= 0xBF
    }
    // Negative flag
    if((after & 0x80) != 0) {
        core.Flags |= 0x80
    } else {
        core.Flags &= 0x70
    }
    // Parity flag
    p := 0
    for i := 0; i < 8; i++ {
        if((after >> i) & 0x01 != 0) {
            p++
        }
    }
    if (p % 2 == 0) {
        core.Flags |= 0x04
    } else {
        core.Flags &= 0xfb
    }
    //carry
    if carry {
        core.Flags |= 0x01
    } else {
        core.Flags &= 0xFE
    }
}

func (core *Core8080) ExecuteOpcode(opcode []uint8, mem *MainMemory) {
    switch opcode[0] {
    case 0x00: 	   //NOP	1		
        fmt.Printf("NOP\n")
        core.PC++
    case 0x01://LXI B,D16	3		B <- byte 3, C <- byte 2
        fmt.Printf("NLXI B,D16\n")
        core.B = opcode[2]
        core.C = opcode[1]
        core.PC += 3
    case 0x02://STAX B	1		(BC) <- A
        core.PC++
        addr := uint16(core.B) << 8
        addr |= uint16(core.C)
        mem.Write(addr, core.A)
        fmt.Printf("STAX B\n")
    case 0x03://INX B	1		BC <- BC+1
        core.PC++
        item := uint16(core.B) << 8
        item |= uint16(core.C)
        item++
        core.C = uint8(item)
        core.B = uint8(item >> 8)
        fmt.Printf("INX B\n")
    case 0x04://INR B	1	Z, S, P, AC	B <- B+1
        fmt.Printf("INR B\n")
        core.PC++
    case 0x05://DCR B	1	Z, S, P, AC	B <- B-1
        fmt.Printf("DCR B\n")
        core.PC++
    case 0x06://MVI B, D8	2		B <- byte 2
        fmt.Printf("MVI B\n")
        core.B = opcode[1]
        core.PC += 2
    case 0x07://RLC	1	CY	A = A << 1; bit 0 = prev bit 7; CY = prev bit 7
        fmt.Printf("RLC\n")
        core.PC++
    case 0x08://-			
        fmt.Printf("Invalid Instruction\n")
        core.PC++
    case 0x09://DAD B	1	CY	HL = HL + BC
        fmt.Printf("DAD B\n")
        core.PC++
    case 0x0a://LDAX B	1		A <- (BC)
        fmt.Printf("LDAX B\n")
        addr := uint16(core.B) << 8
        addr |= uint16(core.C)
        core.A = mem.Read(addr)
        core.PC++
    case 0x0b://DCX B	1		BC = BC-1
        fmt.Printf("DCX B\n")
        item := uint16(core.B) << 8
        item |= uint16(core.C)
        item--
        core.C = uint8(item)
        core.B = uint8(item >> 8)
        core.PC++
    case 0x0c://INR C	1	Z, S, P, AC	C <- C+1
        fmt.Printf("INR C\n")
        temp := core.C + 1
        carry := temp < core.C
        core.UpdateFlags(temp, carry)
        core.C = temp 
        core.PC++
    case 0x0d://DCR C	1	Z, S, P, AC	C <-C-1
        fmt.Printf("DCR C\n")
        temp := core.C - 1
        carry := temp > core.C
        core.UpdateFlags(temp, carry)
        core.C = temp 
        core.PC++
    case 0x0e://MVI C,D8	2		C <- byte 2
        fmt.Printf("MVI C, D8\n")
        core.C = opcode[1]
        core.PC += 2
    case 0x0f://RRC	1	CY	A = A >> 1; bit 7 = prev bit 0; CY = prev bit 0
        fmt.Printf("RRC\n")
        carry := (core.A & 0x01) != 0
        core.A >>= 1
        if carry {
            core.A |= 0x80
            core.Flags |= 0x01
        } else {
            core.Flags &= 0xFE
        } 
        core.PC++
    case 0x10://-			
        fmt.Printf("Invalid Instruction\n")
        core.PC++
    case 0x11://LXI D,D16	3		D <- byte 3, E <- byte 2
        fmt.Printf("LXI D, D16\n")
        core.D = opcode[1]
        core.E = opcode[2]
        core.PC += 3
    case 0x12://STAX D	1		(DE) <- A
        fmt.Printf("STAX D\n")
        addr := uint16(core.D) << 8
        addr |= uint16(core.E)
        mem.Write(addr, core.A)
        core.PC++
    case 0x13://INX D	1		DE <- DE + 1
        fmt.Printf("INX D\n")
        temp := uint16(core.D) << 8
        temp |= uint16(core.E)
        temp++
        core.E = uint8(core.E)
        core.D = uint8(core.D >> 8)
        core.PC++
    case 0x14://INR D	1	Z, S, P, AC	D <- D+1
        fmt.Printf("INR D\n")
        temp := core.D + 1
        carry := temp < core.D
        core.UpdateFlags(temp, carry)
        core.D = temp 
        core.PC++
    case 0x15://DCR D	1	Z, S, P, AC	D <- D-1
        fmt.Printf("DCR D\n")
        temp := core.D - 1
        carry := temp > core.D
        core.UpdateFlags(temp, carry)
        core.D = temp 
        core.PC++
    case 0x16://MVI D, D8	2		D <- byte 2
        fmt.Printf("MVI D, D8\n")
        core.D = opcode[1]
        core.PC += 2
    case 0x17://RAL	1	CY	A = A << 1; bit 0 = prev CY; CY = prev bit 7
        fmt.Printf("RAL\n")
        carry := (core.A & 0x80) != 0
        core.A <<= 1
        if carry {
            core.A |= 0x80
            core.Flags |= 0x01
        } else {
            core.Flags &= 0xFE
        } 
        core.PC++
    case 0x18://-			
        fmt.Printf("Invalid Instruction\n")
        core.PC++
    case 0x19://DAD D	1	CY	HL = HL + DE
        fmt.Printf("DAD D\n")
        temp1 := uint16(core.H) << 8
        temp1 |= uint16(core.L)
        temp2 := uint16(core.D) << 8
        temp2 |= uint16(core.E)
        temp3 := temp1 + temp2
        carry := temp3 <= temp1 || temp3 <= temp2
        if carry {
            core.A |= 0x80
            core.Flags |= 0x01
        } else {
            core.Flags &= 0xFE
        } 
        core.L = uint8(temp3)
        core.H = uint8(temp3 >> 8)
        core.PC++
    case 0x1a://LDAX D	1		A <- (DE)
        fmt.Printf("LDAX D\n")
        addr := uint16(core.D) << 8
        addr |= uint16(core.E)
        core.A = mem.Read(addr)
        core.PC++
    case 0x1b://DCX D	1		DE = DE-1
        fmt.Printf("DCX D\n")
        core.PC++
    case 0x1c://INR E	1	Z, S, P, AC	E <-E+1
        fmt.Printf("INR E\n")
        core.PC++
    case 0x1d://DCR E	1	Z, S, P, AC	E <- E-1
        fmt.Printf("DCR E\n")
        core.PC++
    case 0x1e://MVI E,D8	2		E <- byte 2
        fmt.Printf("MVI E, D8\n")
        core.PC += 2
    case 0x1f://RAR	1	CY	A = A >> 1; bit 7 = prev bit 7; CY = prev bit 0
        fmt.Printf("RAR\n")
        core.PC++
    case 0x20://-			
        fmt.Printf("Invalid Instruction\n")
        core.PC++
    case 0x21://LXI H,D16	3		H <- byte 3, L <- byte 2
        fmt.Printf("LXI H, D16\n")
        core.PC += 3
    case 0x22://SHLD adr	3		(adr) <-L; (adr+1)<-H
        fmt.Printf("SHLD adr\n")
        core.PC += 3
    case 0x23://INX H	1		HL <- HL + 1
        fmt.Printf("INX H\n")
        core.PC++
    case 0x24://INR H	1	Z, S, P, AC	H <- H+1
        fmt.Printf("INR H\n")
        core.PC++
    case 0x25://DCR H	1	Z, S, P, AC	H <- H-1
        fmt.Printf("DCR H\n")
        core.PC++
    case 0x26://MVI H,D8	2		H <- byte 2
        fmt.Printf("MVI H, D8\n")
        core.PC += 2
    case 0x27://DAA	1		special
        fmt.Printf("DAA\n")
        core.PC++
    case 0x28://-			
        fmt.Printf("Invalid Instruction\n")
        core.PC++
    case 0x29://DAD H	1	CY	HL = HL + HI
        fmt.Printf("DAD H\n")
        core.PC++
    case 0x2a://LHLD adr	3		L <- (adr); H<-(adr+1)
        fmt.Printf("LHLD adr\n")
        core.PC += 3
    case 0x2b://DCX H	1		HL = HL-1
        fmt.Printf("DCX H\n")
        core.PC++
    case 0x2c://INR L	1	Z, S, P, AC	L <- L+1
        fmt.Printf("INR L\n")
        core.PC++
    case 0x2d://DCR L	1	Z, S, P, AC	L <- L-1
        fmt.Printf("DCR L\n")
        core.PC++
    case 0x2e://MVI L, D8	2		L <- byte 2
        fmt.Printf("MVI L, D*\n")
        core.PC += 2
    case 0x2f://CMA	1		A <- !A
        fmt.Printf("CMA 1\n")
        core.PC++
    case 0x30://-			
        fmt.Printf("Invalid Instruction\n")
        core.PC++
    case 0x31://LXI SP, D16	3		SP.hi <- byte 3, SP.lo <- byte 2
        fmt.Printf("LXI SP, D16\n")
        core.PC += 3
    case 0x32://STA adr	3		(adr) <- A
        fmt.Printf("STA adr\n")
        core.PC += 3
    case 0x33://INX SP	1		SP = SP + 1
        fmt.Printf("INX SP\n")
        core.PC++
    case 0x34://INR M	1	Z, S, P, AC	(HL) <- (HL)+1
        fmt.Printf("INR M\n")
        core.PC++
    case 0x35://DCR M	1	Z, S, P, AC	(HL) <- (HL)-1
        fmt.Printf("DCR M\n")
        core.PC++
    case 0x36://MVI M,D8	2		(HL) <- byte 2
        fmt.Printf("MVI M, M8\n")
        core.PC += 2
    case 0x37://STC	1	CY	CY = 1
        fmt.Printf("STC\n")
        core.PC++
    case 0x38://-			
        fmt.Printf("Invalid Instruction\n")
        core.PC++
    case 0x39://DAD SP	1	CY	HL = HL + SP
        fmt.Printf("DAD SP\n")
        core.PC++
    case 0x3a://LDA adr	3		A <- (adr)
        fmt.Printf("LDA adr\n")
        core.PC += 3
    case 0x3b://DCX SP	1		SP = SP-1
        fmt.Printf("DCX SP\n")
        core.PC++
    case 0x3c://INR A	1	Z, S, P, AC	A <- A+1
        fmt.Printf("INR A\n")
        core.PC++
    case 0x3d://DCR A	1	Z, S, P, AC	A <- A-1
        fmt.Printf("DCR A\n")
        core.PC++
    case 0x3e://MVI A,D8	2		A <- byte 2
        fmt.Printf("MVI A, D8\n")
        core.PC += 2
    case 0x3f://CMC	1	CY	CY=!CY
        fmt.Printf("CMC\n")
        core.PC++
    case 0x40://MOV B,B	1		B <- B
        fmt.Printf("MOV B, B\n")
        core.PC++
    case 0x41://MOV B,C	1		B <- C
        fmt.Printf("MOV B, C\n")
        core.B = core.C
        core.PC++
    case 0x42://MOV B,D	1		B <- D
        fmt.Printf("MOV B, D\n")
        core.B = core.D
        core.PC++
    case 0x43://MOV B,E	1		B <- E
        fmt.Printf("MOV B, E\n")
        core.B = core.E
        core.PC++
    case 0x44://MOV B,H	1		B <- H
        fmt.Printf("MOV B, H\n")
        core.B = core.H
        core.PC++
    case 0x45://MOV B,L	1		B <- L
        fmt.Printf("MOV B, L\n")
        core.B = core.L
        core.PC++
    case 0x46://MOV B,M	1		B <- (HL)
        fmt.Printf("MOV B, M\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        core.B = mem.Read(addr)
        core.PC++
    case 0x47://MOV B,A	1		B <- A
        fmt.Printf("MOV B, A\n")
        core.B = core.A
        core.PC++
    case 0x48://MOV C,B	1		C <- B
        fmt.Printf("MOV C, B\n")
        core.C = core.B
        core.PC++
    case 0x49://MOV C,C	1		C <- C
        fmt.Printf("MOV C, C\n")
        core.PC++
    case 0x4a://MOV C,D	1		C <- D
        fmt.Printf("MOV C, D\n")
        core.C = core.D
        core.PC++
    case 0x4b://MOV C,E	1		C <- E
        fmt.Printf("MOV C, E\n")
        core.C = core.E
        core.PC++
    case 0x4c://MOV C,H	1		C <- H
        fmt.Printf("MOV C, H\n")
        core.C = core.H
        core.PC++
    case 0x4d://MOV C,L	1		C <- L
        fmt.Printf("MOV C, L\n")
        core.C = core.L
        core.PC++
    case 0x4e://MOV C,M	1		C <- (HL)
        fmt.Printf("MOV C, M\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        core.C = mem.Read(addr)
        core.PC++
    case 0x4f://MOV C,A	1		C <- A
        fmt.Printf("MOV C, A\n")
        core.C = core.A
        core.PC++
    case 0x50://MOV D,B	1		D <- B
        fmt.Printf("MOV D, B\n")
        core.D = core.B
        core.PC++
    case 0x51://MOV D,C	1		D <- C
        fmt.Printf("MOV D, C\n")
        core.D = core.C
        core.PC++
    case 0x52://MOV D,D	1		D <- D
        fmt.Printf("MOV D, D\n")
        core.D = core.D
        core.PC++
    case 0x53://MOV D,E	1		D <- E
        fmt.Printf("MOV D, E\n")
        core.D = core.E
        core.PC++
    case 0x54://MOV D,H	1		D <- H
        fmt.Printf("MOV D, H\n")
        core.D = core.H
        core.PC++
    case 0x55://MOV D,L	1		D <- L
        fmt.Printf("MOV D, L\n")
        core.D = core.L
        core.PC++
    case 0x56://MOV D,M	1		D <- (HL)
        fmt.Printf("MOV D, M\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        core.D = mem.Read(addr)
        core.PC++
    case 0x57://MOV D,A	1		D <- A
        fmt.Printf("MOV D, A\n")
        core.D = core.A
        core.PC++
    case 0x58://MOV E,B	1		E <- B
        fmt.Printf("MOV E, B\n")
        core.E = core.B
        core.PC++
    case 0x59://MOV E,C	1		E <- C
        fmt.Printf("MOV E, C\n")
        core.E = core.C
        core.PC++
    case 0x5a://MOV E,D	1		E <- D
        fmt.Printf("MOV E, D\n")
        core.E = core.D
        core.PC++
    case 0x5b://MOV E,E	1		E <- E
        fmt.Printf("MOV E, E\n")
        core.PC++
    case 0x5c://MOV E,H	1		E <- H
        fmt.Printf("MOV E, H\n")
        core.E = core.H
        core.PC++
    case 0x5d://MOV E,L	1		E <- L
        fmt.Printf("MOV E, L\n")
        core.E = core.L
        core.PC++
    case 0x5e://MOV E,M	1		E <- (HL)
        fmt.Printf("MOV E, M\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        core.E = mem.Read(addr)
        core.PC++
    case 0x5f://MOV E,A	1		E <- A
        fmt.Printf("MOV E, A\n")
        core.E = core.A
        core.PC++
    case 0x60://MOV H,B	1		H <- B
        fmt.Printf("MOV H, B\n")
        core.H = core.B
        core.PC++
    case 0x61://MOV H,C	1		H <- C
        fmt.Printf("MOV H, C\n")
        core.H = core.B
        core.PC++
    case 0x62://MOV H,D	1		H <- D
        fmt.Printf("MOV H, D\n")
        core.H = core.D
        core.PC++
    case 0x63://MOV H,E	1		H <- E
        fmt.Printf("MOV H, E\n")
        core.H = core.E
        core.PC++
    case 0x64://MOV H,H	1		H <- H
        fmt.Printf("MOV H, H\n")
        core.PC++
    case 0x65://MOV H,L	1		H <- L
        fmt.Printf("MOV H, L\n")
        core.H = core.L
        core.PC++
    case 0x66://MOV H,M	1		H <- (HL)
        fmt.Printf("MOV H, M\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        core.H = mem.Read(addr)
        core.PC++
    case 0x67://MOV H,A	1		H <- A
        fmt.Printf("MOV H, A\n")
        core.H = core.A
        core.PC++
    case 0x68://MOV L,B	1		L <- B
        fmt.Printf("MOV L, B\n")
        core.L = core.B
        core.PC++
    case 0x69://MOV L,C	1		L <- C
        fmt.Printf("MOV L, C\n")
        core.L = core.C
        core.PC++
    case 0x6a://MOV L,D	1		L <- D
        fmt.Printf("MOV L, D\n")
        core.L = core.D
        core.PC++
    case 0x6b://MOV L,E	1		L <- E
        fmt.Printf("MOV L, E\n")
        core.L = core.E
        core.PC++
    case 0x6c://MOV L,H	1		L <- H
        fmt.Printf("MOV L, H\n")
        core.L = core.H
        core.PC++
    case 0x6d://MOV L,L	1		L <- L
        fmt.Printf("MOV L, L\n")
        core.PC++
    case 0x6e://MOV L,M	1		L <- (HL)
        fmt.Printf("MOV L, M\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        core.L = mem.Read(addr)
        core.PC++
    case 0x6f://MOV L,A	1		L <- A
        fmt.Printf("MOV L, A\n")
        core.L = core.A
        core.PC++
    case 0x70://MOV M,B	1		(HL) <- B
        fmt.Printf("MOV M, B\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        mem.Write(addr, core.B)
        core.PC++
    case 0x71://MOV M,C	1		(HL) <- C
        fmt.Printf("MOV M, C\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        mem.Write(addr, core.C)
        core.PC++
    case 0x72://MOV M,D	1		(HL) <- D
        fmt.Printf("MOV M, D\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        mem.Write(addr, core.D)
        core.PC++
    case 0x73://MOV M,E	1		(HL) <- E
        fmt.Printf("MOV M, E\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        mem.Write(addr, core.E)
        core.PC++
    case 0x74://MOV M,H	1		(HL) <- H
        fmt.Printf("MOV M, H\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        mem.Write(addr, core.H)
        core.PC++
    case 0x75://MOV M,L	1		(HL) <- L
        fmt.Printf("MOV M, L\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        mem.Write(addr, core.L)
        core.PC++
    case 0x76://HLT	1		special
        fmt.Printf("HLT\n")
        core.PC++
    case 0x77://MOV M,A	1		(HL) <- A
        fmt.Printf("MOV M, A\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        mem.Write(addr, core.A)
        core.PC++
    case 0x78://MOV A,B	1		A <- B
        fmt.Printf("MOV A, B\n")
        core.A = core.B
        core.PC++
    case 0x79://MOV A,C	1		A <- C
        fmt.Printf("MOV A, C\n")
        core.A = core.C
        core.PC++
    case 0x7a://MOV A,D	1		A <- D
        fmt.Printf("MOV A, D\n")
        core.A = core.D
        core.PC++
    case 0x7b://MOV A,E	1		A <- E
        fmt.Printf("MOV A, E\n")
        core.A = core.E
        core.PC++
    case 0x7c://MOV A,H	1		A <- H
        fmt.Printf("MOV A, H\n")
        core.A = core.L
        core.PC++
    case 0x7d://MOV A,L	1		A <- L
        fmt.Printf("MOV A, L\n")
        core.A = core.L
        core.PC++
    case 0x7e://MOV A,M	1		A <- (HL)
        fmt.Printf("MOV A, M\n")
        addr := uint16(core.H) << 8
        addr |= uint16(core.L)
        core.A = mem.Read(addr)
        core.PC++
    case 0x7f://MOV A,A	1		A <- A
        fmt.Printf("MOV A, A\n")
        core.PC++
    case 0x80://ADD B	1	Z, S, P, CY, AC	A <- A + B
        fmt.Printf("ADD B\n")
        core.PC++
    case 0x81://ADD C	1	Z, S, P, CY, AC	A <- A + C
        fmt.Printf("ADD C\n")
        core.PC++
    case 0x82://ADD D	1	Z, S, P, CY, AC	A <- A + D
        fmt.Printf("ADD D\n")
        core.PC++
    case 0x83://ADD E	1	Z, S, P, CY, AC	A <- A + E
        fmt.Printf("ADD E\n")
        core.PC++
    case 0x84://ADD H	1	Z, S, P, CY, AC	A <- A + H
        fmt.Printf("ADD H\n")
        core.PC++
    case 0x85://ADD L	1	Z, S, P, CY, AC	A <- A + L
        fmt.Printf("ADD L\n")
        core.PC++
    case 0x86://ADD M	1	Z, S, P, CY, AC	A <- A + (HL)
        fmt.Printf("ADD M\n")
        core.PC++
    case 0x87://ADD A	1	Z, S, P, CY, AC	A <- A + A
        fmt.Printf("ADD A\n")
        core.PC++
    case 0x88://ADC B	1	Z, S, P, CY, AC	A <- A + B + CY
        fmt.Printf("ADC B\n")
        core.PC++
    case 0x89://ADC C	1	Z, S, P, CY, AC	A <- A + C + CY
        fmt.Printf("ADC C\n")
        core.PC++
    case 0x8a://ADC D	1	Z, S, P, CY, AC	A <- A + D + CY
        fmt.Printf("ADC D\n")
        core.PC++
    case 0x8b://ADC E	1	Z, S, P, CY, AC	A <- A + E + CY
        fmt.Printf("ADC E\n")
        core.PC++
    case 0x8c://ADC H	1	Z, S, P, CY, AC	A <- A + H + CY
        fmt.Printf("ADC H\n")
        core.PC++
    case 0x8d://ADC L	1	Z, S, P, CY, AC	A <- A + L + CY
        fmt.Printf("ADC L\n")
        core.PC++
    case 0x8e://ADC M	1	Z, S, P, CY, AC	A <- A + (HL) + CY
        fmt.Printf("ADC M\n")
        core.PC++
    case 0x8f://ADC A	1	Z, S, P, CY, AC	A <- A + A + CY
        fmt.Printf("ADC A\n")
        core.PC++
    case 0x90://SUB B	1	Z, S, P, CY, AC	A <- A - B
        fmt.Printf("SUB B\n")
        core.PC++
    case 0x91://SUB C	1	Z, S, P, CY, AC	A <- A - C
        fmt.Printf("SUB C\n")
        core.PC++
    case 0x92://SUB D	1	Z, S, P, CY, AC	A <- A + D
        fmt.Printf("SUB D\n")
        core.PC++
    case 0x93://SUB E	1	Z, S, P, CY, AC	A <- A - E
        fmt.Printf("SUB E\n")
        core.PC++
    case 0x94://SUB H	1	Z, S, P, CY, AC	A <- A + H
        fmt.Printf("SUB H\n")
        core.PC++
    case 0x95://SUB L	1	Z, S, P, CY, AC	A <- A - L
        fmt.Printf("SUB L\n")
        core.PC++
    case 0x96://SUB M	1	Z, S, P, CY, AC	A <- A + (HL)
        fmt.Printf("SUB M\n")
        core.PC++
    case 0x97://SUB A	1	Z, S, P, CY, AC	A <- A - A
        fmt.Printf("SUB A\n")
        core.PC++
    case 0x98://SBB B	1	Z, S, P, CY, AC	A <- A - B - CY
        fmt.Printf("SBB B\n")
        core.PC++
    case 0x99://SBB C	1	Z, S, P, CY, AC	A <- A - C - CY
        fmt.Printf("SBB C\n")
        core.PC++
    case 0x9a://SBB D	1	Z, S, P, CY, AC	A <- A - D - CY
        fmt.Printf("SBB D\n")
        core.PC++
    case 0x9b://SBB E	1	Z, S, P, CY, AC	A <- A - E - CY
        fmt.Printf("SBB E\n")
        core.PC++
    case 0x9c://SBB H	1	Z, S, P, CY, AC	A <- A - H - CY
        fmt.Printf("SBB H\n")
        core.PC++
    case 0x9d://SBB L	1	Z, S, P, CY, AC	A <- A - L - CY
        fmt.Printf("SBB L\n")
        core.PC++
    case 0x9e://SBB M	1	Z, S, P, CY, AC	A <- A - (HL) - CY
        fmt.Printf("SBB M\n")
        core.PC++
    case 0x9f://SBB A	1	Z, S, P, CY, AC	A <- A - A - CY
        fmt.Printf("SBB A\n")
        core.PC++
    case 0xa0://ANA B	1	Z, S, P, CY, AC	A <- A & B
        fmt.Printf("ANA B\n")
        core.PC++
    case 0xa1://ANA C	1	Z, S, P, CY, AC	A <- A & C
        fmt.Printf("ANA C\n")
        core.PC++
    case 0xa2://ANA D	1	Z, S, P, CY, AC	A <- A & D
        fmt.Printf("ANA D\n")
        core.PC++
    case 0xa3://ANA E	1	Z, S, P, CY, AC	A <- A & E
        fmt.Printf("ANA E\n")
        core.PC++
    case 0xa4://ANA H	1	Z, S, P, CY, AC	A <- A & H
        fmt.Printf("ANA\n")
        core.PC++
    case 0xa5://ANA L	1	Z, S, P, CY, AC	A <- A & L
        fmt.Printf("ANA L\n")
        core.PC++
    case 0xa6://ANA M	1	Z, S, P, CY, AC	A <- A & (HL)
        fmt.Printf("ANA M\n")
        core.PC++
    case 0xa7://ANA A	1	Z, S, P, CY, AC	A <- A & A
        fmt.Printf("ANA A\n")
        core.PC++
    case 0xa8://XRA B	1	Z, S, P, CY, AC	A <- A ^ B
        fmt.Printf("XRA B\n")
        core.PC++
    case 0xa9://XRA C	1	Z, S, P, CY, AC	A <- A ^ C
        fmt.Printf("XRA C\n")
        core.PC++
    case 0xaa://XRA D	1	Z, S, P, CY, AC	A <- A ^ D
        fmt.Printf("XRA D\n")
        core.PC++
    case 0xab://XRA E	1	Z, S, P, CY, AC	A <- A ^ E
        fmt.Printf("XRA E\n")
        core.PC++
    case 0xac://XRA H	1	Z, S, P, CY, AC	A <- A ^ H
        fmt.Printf("XRA H\n")
        core.PC++
    case 0xad://XRA L	1	Z, S, P, CY, AC	A <- A ^ L
        fmt.Printf("XRA L\n")
        core.PC++
    case 0xae://XRA M	1	Z, S, P, CY, AC	A <- A ^ (HL)
        fmt.Printf("XRA M\n")
        core.PC++
    case 0xaf://XRA A	1	Z, S, P, CY, AC	A <- A ^ A
        fmt.Printf("XRA A\n")
        core.PC++
    case 0xb0://ORA B	1	Z, S, P, CY, AC	A <- A | B
        fmt.Printf("ORA B\n")
        core.PC++
    case 0xb1://ORA C	1	Z, S, P, CY, AC	A <- A | C
        fmt.Printf("ORA C\n")
        core.PC++
    case 0xb2://ORA D	1	Z, S, P, CY, AC	A <- A | D
        fmt.Printf("ORA D\n")
        core.PC++
    case 0xb3://ORA E	1	Z, S, P, CY, AC	A <- A | E
        fmt.Printf("ORA E\n")
        core.PC++
    case 0xb4://ORA H	1	Z, S, P, CY, AC	A <- A | H
        fmt.Printf("ORA H\n")
        core.PC++
    case 0xb5://ORA L	1	Z, S, P, CY, AC	A <- A | L
        fmt.Printf("ORA L\n")
        core.PC++
    case 0xb6://ORA M	1	Z, S, P, CY, AC	A <- A | (HL)
        fmt.Printf("ORA M\n")
        core.PC++
    case 0xb7://ORA A	1	Z, S, P, CY, AC	A <- A | A
        fmt.Printf("ORA A\n")
        core.PC++
    case 0xb8://CMP B	1	Z, S, P, CY, AC	A - B
        fmt.Printf("CMP B\n")
        core.PC++
    case 0xb9://CMP C	1	Z, S, P, CY, AC	A - C
        fmt.Printf("CMP C\n")
        core.PC++
    case 0xba://CMP D	1	Z, S, P, CY, AC	A - D
        fmt.Printf("CMP D\n")
        core.PC++
    case 0xbb://CMP E	1	Z, S, P, CY, AC	A - E
        fmt.Printf("CMP E\n")
        core.PC++
    case 0xbc://CMP H	1	Z, S, P, CY, AC	A - H
        fmt.Printf("CMP H\n")
        core.PC++
    case 0xbd://CMP L	1	Z, S, P, CY, AC	A - L
        fmt.Printf("CMP L\n")
        core.PC++
    case 0xbe://CMP M	1	Z, S, P, CY, AC	A - (HL)
        fmt.Printf("CMP M\n")
        core.PC++
    case 0xbf://CMP A	1	Z, S, P, CY, AC	A - A
        fmt.Printf("CMP A\n")
        core.PC++
    case 0xc0://RNZ	1		if NZ, RET
        fmt.Printf("RNZ\n")
        core.PC++
    case 0xc1://POP B	1		C <- (sp); B <- (sp+1); sp <- sp+2
        fmt.Printf("POP B\n")
        core.PC++
    case 0xc2://JNZ adr	3		if NZ, PC <- adr
        fmt.Printf("JNZ adr\n")
        core.PC += 3
    case 0xc3://JMP adr	3		PC <= adr
        fmt.Printf("JMP adr\n")
        addr := uint16(opcode[2]) << 8
        addr |= uint16(opcode[1])
        core.PC += addr
    case 0xc4://CNZ adr	3		if NZ, CALL adr
        fmt.Printf("CNZ adr\n")
        core.PC += 3
    case 0xc5://PUSH B	1		(sp-2)<-C; (sp-1)<-B; sp <- sp - 2
        fmt.Printf("PUSH B\n")
        core.PC++
    case 0xc6://ADI D8	2	Z, S, P, CY, AC	A <- A + byte
        fmt.Printf("ADI D8\n")
        core.PC += 2
    case 0xc7://RST 0	1		CALL $0
        fmt.Printf("RST\n")
        core.PC++
    case 0xc8://RZ	1		if Z, RET
        fmt.Printf("RZ\n")
        core.PC++
    case 0xc9://RET	1		PC.lo <- (sp); PC.hi<-(sp+1); SP <- SP+2
        fmt.Printf("RET\n")
        core.PC++
    case 0xca://JZ 3		if Z, PC <- adr
        fmt.Printf("JZ\n")
        core.PC += 3
    case 0xcb://-			
        fmt.Printf("Invalid Instruction\n")
        core.PC++
    case 0xcc://CZ adr	3		if Z, CALL adr
        fmt.Printf("CZ adr\n")
        core.PC += 3
    case 0xcd://CALL adr	3		(SP-1)<-PC.hi;(SP-2)<-PC.lo;SP<-SP-2;PC=adr
        fmt.Printf("CALL adr\n")
        core.PC += 3
    case 0xce://ACI D8	2	Z, S, P, CY, AC	A <- A + data + CY
        fmt.Printf("ACI D8\n")
        core.PC += 2
    case 0xcf://RST 1	1		CALL $8
        fmt.Printf("RST\n")
        core.PC++
    case 0xd0://RNC	1		if NCY, RET
        fmt.Printf("RNC\n")
        core.PC++
    case 0xd1://POP D	1		E <- (sp); D <- (sp+1); sp <- sp+2
        fmt.Printf("POP D\n")
        core.PC++
    case 0xd2://JNC adr	3		if NCY, PC<-adr
        fmt.Printf("JNC\n")
        core.PC += 3
    case 0xd3://OUT D8	2		special
        fmt.Printf("OUT\n")
        core.PC +=2
    case 0xd4://CNC adr	3		if NCY, CALL adr
        fmt.Printf("CNC adr\n")
        core.PC += 3
    case 0xd5://PUSH D	1		(sp-2)<-E; (sp-1)<-D; sp <- sp - 2
        fmt.Printf("PUSH D\n")
        core.PC++
    case 0xd6://SUI D8	2	Z, S, P, CY, AC	A <- A - data
        fmt.Printf("SUI D8\n")
        core.PC += 2
    case 0xd7://RST 2	1		CALL $10
        fmt.Printf("RST 2\n")
        core.PC++
    case 0xd8://RC	1		if CY, RET
        fmt.Printf("RC\n")
        core.PC++
    case 0xd9://-			
        fmt.Printf("Invalid Instruction\n")
        core.PC++
    case 0xda://JC adr	3		if CY, PC<-adr
        fmt.Printf("JC adr\n")
        core.PC += 3
    case 0xdb://IN D8	2		special
        fmt.Printf("IN D8\n")
        core.PC += 2
    case 0xdc://CC adr	3		if CY, CALL adr
        fmt.Printf("CC adr\n")
        core.PC += 3
    case 0xdd://-			
        fmt.Printf("Invalid Instruction\n")
        core.PC++
    case 0xde://SBI D8	2	Z, S, P, CY, AC	A <- A - data - CY
        fmt.Printf("SBI D8\n")
        core.PC += 2
    case 0xdf://RST 3	1		CALL $18
        fmt.Printf("RST\n")
        core.PC++
    case 0xe0://RPO	1		if PO, RET
        fmt.Printf("RPO\n")
        core.PC++
    case 0xe1://POP H	1		L <- (sp); H <- (sp+1); sp <- sp+2
        fmt.Printf("POP\n")
        core.PC++
    case 0xe2://JPO adr	3		if PO, PC <- adr
        fmt.Printf("JPO adr\n")
        core.PC += 3
    case 0xe3://XTHL	1		L <-> (SP); H <-> (SP+1)
        fmt.Printf("XTHL\n")
        core.PC++
    case 0xe4://CPO adr	3		if PO, CALL adr
        fmt.Printf("CPO adr\n")
        core.PC += 3
    case 0xe5://PUSH H	1		(sp-2)<-L; (sp-1)<-H; sp <- sp - 2
        fmt.Printf("PUSH H\n")
        core.PC++
    case 0xe6://ANI D8	2	Z, S, P, CY, AC	A <- A & data
        fmt.Printf("ANI D8\n")
        core.PC += 2
    case 0xe7://RST 4	1		CALL $20
        fmt.Printf("RST 4\n")
        core.PC++
    case 0xe8://RPE	1		if PE, RET
        fmt.Printf("RPE 1\n")
        core.PC++
    case 0xe9://PCHL	1		PC.hi <- H; PC.lo <- L
        fmt.Printf("PCHL\n")
        core.PC++
    case 0xea://JPE adr	3		if PE, PC <- adr
        fmt.Printf("JPE adr\n")
        core.PC += 3
    case 0xeb://XCHG	1		H <-> D; L <-> E
        fmt.Printf("XCHG\n")
        core.PC++
    case 0xec://CPE adr	3		if PE, CALL adr
        fmt.Printf("CPE adr\n")
        core.PC += 3
    case 0xed://-			
        fmt.Printf("Invalid Instruction\n")
        core.PC++
    case 0xee://XRI D8	2	Z, S, P, CY, AC	A <- A ^ data
        fmt.Printf("XRI D8\n")
        core.PC += 2
    case 0xef://RST 5	1		CALL $28
        fmt.Printf("RST 5\n")
        core.PC++
    case 0xf0://RP	1		if P, RET
        fmt.Printf("RP 1\n")
        core.PC++
    case 0xf1://POP PSW	1		flags <- (sp); A <- (sp+1); sp <- sp+2
        fmt.Printf("POP PSW 1\n")
        core.PC++
    case 0xf2://JP adr	3		if P=1 PC <- adr
        fmt.Printf("JP adr\n")
        core.PC += 3
    case 0xf3://DI	1		special
        fmt.Printf("DI\n")
        core.PC++
    case 0xf4://CP adr	3		if P, PC <- adr
        fmt.Printf("CP adr\n")
        core.PC += 3
    case 0xf5://PUSH PSW	1		(sp-2)<-flags; (sp-1)<-A; sp <- sp - 2
        fmt.Printf("PUSH PSW\n")
        core.PC++
    case 0xf6://ORI D8	2	Z, S, P, CY, AC	A <- A | data
        fmt.Printf("ORI D8\n")
        core.PC += 2
    case 0xf7://RST 6	1		CALL $30
        fmt.Printf("RST 6\n")
        core.PC++
    case 0xf8://RM	1		if M, RET
        fmt.Printf("RM\n")
        core.PC++
    case 0xf9://SPHL	1		SP=HL
        fmt.Printf("SPHL\n")
        core.PC++
    case 0xfa://JM adr	3		if M, PC <- adr
        fmt.Printf("JM adr\n")
        core.PC += 3
    case 0xfb:	//EI	1		special
        fmt.Printf("EI\n")
        core.PC++
    case 0xfc:	//CM adr	3		if M, CALL adr
        fmt.Printf("CM adr\n")
        core.PC += 3
    case 0xfd:	//-			
        fmt.Printf("Invalid Instruction\n")
        core.PC++
    case 0xfe:	//CPI D8	2	Z, S, P, CY, AC	A - data
        fmt.Printf("CPI D8\n")
        core.PC += 2
    case 0xff:	//RST 7	1		CALL $38
        fmt.Printf("RST 7\n")
        core.PC++
    }
}

func main() {
    fmt.Println("Hello weeb!")
    romData, err := os.ReadFile("roms/invaders.rom")
    if err != nil {
        panic(err)
    }

    core := new(Core8080)
    core.Init(nil, nil)
    memory := new(MainMemory)
    memory.Init(romData)
    fmt.Printf("Memory Initialized with %dK of rom and %dK of ram!\n", len(memory.Rom)/1024, len(memory.Ram)/1024)

    for {
        core.RunTick(memory)
        if core.PC == uint16(Kilobytes(8)) {
            break
        }
    }
    //fmt.Println(memory.Rom)
    //fmt.Println(len(romData))
}
