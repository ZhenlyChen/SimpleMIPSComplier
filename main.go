package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var myFlagSet = flag.NewFlagSet("myflagset", flag.ExitOnError)
var mapF map[string]string
var mapT map[string]string

var currentIns string
var currentNo int

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func cErr(e error, msg string) {
	fmt.Println("[Error]")
	fmt.Println(msg)
	fmt.Println(currentNo, ": ", currentIns)
	if e != nil {
		panic(e)
	} else {
		os.Exit(1)
	}
}

func main() {
	initMap()
	currentNo = 0
	myFlagSet.Parse(os.Args)
	var sourceFile = myFlagSet.Arg(1)
	var targetFile = myFlagSet.Arg(2)
	// test code ------
	//dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	//var sourceFile = dir + "/source.asm"
	//var targetFile = dir+ "/data.txt"
	// ----------------

	fi, err := os.Open(sourceFile)
	check(err)
	defer fi.Close()

	br := bufio.NewReader(fi)
	var rs = new(string)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		currentIns = string(a)
		currentNo++
		*rs += currentIns
	}
	err = ioutil.WriteFile("doc.csv", doc([]byte(*rs)), 0644)
	check(err)
	err = ioutil.WriteFile(targetFile, format([]byte(*rs)), 0644)
	check(err)
}

func format(old []byte) (res []byte) {
	count := 0
	for i := range old {
		if count > 7 {
			res = append(res, ' ')
			count = 0
		}
		count++
		res = append(res, old[i])
	}
	return
}

func doc(old []byte) (res []byte) {
	var buffer, currentBuffer bytes.Buffer
	count := 0
	for i := range old {
		count++
		buffer.WriteByte(old[i])
		currentBuffer.WriteByte(old[i])
		if count == 6 {
			buffer.WriteByte('&')
			buffer.WriteByte(',')
		} else if count == 11 {
			buffer.WriteByte('&')
			buffer.WriteByte(',')
		} else if count == 16 {
			buffer.WriteByte('&')
			buffer.WriteByte(',')
		} else if count == 20 {
			buffer.WriteByte(' ')
		} else if count == 24 {
			buffer.WriteByte(' ')
		} else if count == 28 {
			buffer.WriteByte(' ')
		} else if count == 32 {
			buffer.Write([]byte(",=,"))
			num, _ := strconv.ParseInt(currentBuffer.String(), 2, 64)
			// hexStr := strconv.FormatInt(num, 16)
			hexStr := fmt.Sprintf("%08X&", num)
			buffer.WriteString(hexStr)
			currentBuffer.Reset()
			buffer.Write([]byte("\r\n"))
			count = 0
		}
	}
	res = buffer.Bytes()
	return

}

func compile(s string) (ins string) {
	ins = ""
	a := strings.Fields(s)
	if len(a) < 1 {
		return
	}
	other := ""
	for i := 1; i < len(a); i++ {
		other += a[i]
	}
	op, ok := mapF[a[0]]
	if !ok {
		cErr(nil, "There is no a instruction")
	}
	typeIns := mapT[a[0]]
	switch typeIns {
	case "I":
		regs := ParseReg(other, 3)
		ins = toInsI(op, regs[1], regs[0], regs[2])
	case "I1":
		rs, rt, imm := ParseI(other)
		ins = toInsI(op, rs, rt, imm)
	case "I2":
		rs, imm := ParseII(other)
		ins = toInsI(op, rs, 0, imm)
	case "R":
		regs := ParseReg(other, 3)
		ins = toInsR(op, regs[1], regs[2], regs[0], 0, 0)
	case "R1":
		regs := ParseReg(other, 3)
		ins = toInsR(op, 0, regs[1], regs[0], regs[2], 0)
	case "R2":
		regs := ParseReg(other, 1)
		ins = toInsR(op, regs[0], 0, 0, 0, 0)
	case "J":
		imm := ParseJ(other)
		ins = toInsJ(op, imm)
	}
	return
}

func ParseReg(ins string, count int) (rs []int64) {
	regs := strings.Split(ins, ",")
	if len(regs) != count {
		cErr(nil, "Illegal instruction")
	}
	rs = make([]int64, count)
	for i := 0; i < count; i++ {
		rs[i], _ = strconv.ParseInt(strings.Replace(regs[i], "$", "", -1), 10, 64)
	}
	return
}

func ParseI(ins string) (rs, rt, imm int64) {
	regs := strings.Split(ins, ",")
	if len(regs) != 2 {
		cErr(nil, "Illegal instruction")
	}
	rt, _ = strconv.ParseInt(strings.Replace(regs[0], "$", "", -1), 10, 64)
	regs = strings.Split(regs[1], "$")
	if len(regs) != 2 {
		cErr(nil, "Illegal instruction")
	}
	rs, _ = strconv.ParseInt(strings.Replace(regs[1], ")", "", -1), 10, 64)
	imm, _ = strconv.ParseInt(strings.Replace(regs[0], "(", "", -1), 10, 64)
	return
}

func ParseJ(ins string) (imm int64) {
	if ins == "" {
		imm = 0
	} else {
		imm, _ = strconv.ParseInt(ins, 0, 64)
	}
	return
}

func ParseII(ins string) (rs, imm int64) {
	regs := strings.Split(ins, ",")
	if len(regs) != 2 {
		cErr(nil, "Illegal instruction")
	}
	rs, _ = strconv.ParseInt(strings.Replace(regs[0], "$", "", -1), 10, 64)
	imm, _ = strconv.ParseInt(regs[1], 10, 64)
	return
}

func Int64toBinStr(num, lenght int64) (ins string) {
	r, _ := strconv.Atoi(strconv.FormatInt(num, 2))
	ins = fmt.Sprintf("%0"+strconv.FormatInt(lenght, 10)+"d", r)
	return
}

func toInsR(op string, rs, rt, rd, sa, funct int64) (ins string) {
	if rs > 31 || rt > 31 || rd > 31 || sa > 31 || funct > 63 {
		cErr(nil, "Illegal register")
	}
	ins = op
	ins += Int64toBinStr(rs, 5)
	ins += Int64toBinStr(rt, 5)
	ins += Int64toBinStr(rd, 5)
	ins += Int64toBinStr(sa, 5)
	ins += Int64toBinStr(funct, 6)
	return
}

func toImm(imm, lenght int64) (str string) {
	if imm >= 0 {
		str = Int64toBinStr(imm, lenght)
	} else {
		tmp := []byte(Int64toBinStr(-imm, lenght))
		found := false
		for i := len(tmp) - 1; i >= 0; i-- {
			if found {
				if tmp[i] == '1' {
					tmp[i] = '0'
				} else {
					tmp[i] = '1'
				}
			}
			if tmp[i] == '1' {
				found = true
			}
		}
		str = string(tmp)
	}
	return
}

func toInsI(op string, rs, rt, imm int64) (ins string) {
	if rs > 31 || rt > 31 {
		cErr(nil, "Illegal register")
	}
	ins = op
	ins += Int64toBinStr(rs, 5)
	ins += Int64toBinStr(rt, 5)
	ins += toImm(imm, 16)
	return
}

func toInsJ(op string, imm int64) (ins string) {
	ins = op
	if imm < 0 {
		cErr(nil, "Illegal instruction")
	}
	imm = imm >> 2
	r, _ := strconv.Atoi(strconv.FormatInt(imm, 2))
	ins += fmt.Sprintf("%026d", r)
	return
}

func initMap() {
	mapF = make(map[string]string)
	mapF["add"] = "000000"
	mapF["sub"] = "000001"
	mapF["addi"] = "000010"
	mapF["or"] = "010000"
	mapF["and"] = "010001"
	mapF["ori"] = "010010"
	mapF["sll"] = "011000"
	mapF["slt"] = "100110"
	mapF["sltiu"] = "100111"
	mapF["sw"] = "110000"
	mapF["lw"] = "110001"
	mapF["beq"] = "110100"
	mapF["bltz"] = "110110"
	mapF["j"] = "111000"
	mapF["jr"] = "111001"
	mapF["jal"] = "111010"
	mapF["halt"] = "111111"
	mapT = make(map[string]string)
	mapT["add"] = "R"
	mapT["sub"] = "R"
	mapT["addi"] = "I"
	mapT["or"] = "R"
	mapT["and"] = "R"
	mapT["ori"] = "I"
	mapT["sll"] = "R1"
	mapT["slt"] = "R"
	mapT["sltiu"] = "I"
	mapT["sw"] = "I1"
	mapT["lw"] = "I1"
	mapT["beq"] = "I"
	mapT["bltz"] = "I2"
	mapT["j"] = "J"
	mapT["jr"] = "R2"
	mapT["jal"] = "J"
	mapT["halt"] = "J"
}
