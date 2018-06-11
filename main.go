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
	"github.com/kataras/iris/core/errors"
)

var mapF map[string]string
var mapT map[string]string
var currentIns string
var currentNo = 0
var dev *bool
var errG error

func check(e error, msg string) {
	if e == nil {
		return
	}
	fmt.Println("[Error]")
	fmt.Println(msg)
	if currentNo != 0 {
		fmt.Println(currentNo, ": ", currentIns)
	}
	if *dev {
		panic(e)
	}
	os.Exit(1)
}

func main() {
	// 解析参数
	sourceFile := flag.String("s", "null", "source file path.")
	targetFile := flag.String("o", "res.txt", "result file path.")
	docFile := flag.String("d", "null", "doc file path.")
	dev = flag.Bool("dev", false, "Debug model")
	flag.Parse()
	// 打开文件
	if *sourceFile == "null" {
		check(errors.New(""), "Please choose the source file. For example: -s source.asm")
	}
	fi, err := os.Open(*sourceFile)
	check(err, "Can't open " + *sourceFile)
	defer fi.Close()
	// 循环处理
	br := bufio.NewReader(fi)
	var rs = new(string)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		currentIns = string(a)
		currentNo++
		*rs += compile(currentIns)
	}
	// 生成结果文件
	err = ioutil.WriteFile(*targetFile, format([]byte(*rs)), 0644)
	check(err, "Can't write result in " + *targetFile)
	// 生成文档
	if *docFile != "null" {
		err = ioutil.WriteFile(*docFile, doc([]byte(*rs)), 0644)
		check(err, "Can't write doc in " + *docFile)
	}
}

func format(old []byte) (res []byte) {
	var buffer bytes.Buffer
	count := 0
	for i := range old {
		if count > 7 {
			buffer.WriteByte(' ')
			count = 0
		}
		count++
		buffer.WriteByte(old[i])
	}
	res = buffer.Bytes()
	return
}

func doc(old []byte) (res []byte) {
	var buffer, currentBuffer bytes.Buffer
	count := 0
	for i := range old {
		count++
		buffer.WriteByte(old[i])
		currentBuffer.WriteByte(old[i])
		if count == 6 || count == 11 || count == 16 {
			buffer.WriteByte('&')
			buffer.WriteByte(',')
		} else if count == 20 || count == 24 || count == 28  {
			buffer.WriteByte(' ')
		} else if count == 32 {
			buffer.Write([]byte(",=,"))
			num, _ := strconv.ParseInt(currentBuffer.String(), 2, 64)
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
		check(nil, "There is no a instruction")
	}
	typeIns := mapT[a[0]]
	switch typeIns {
	case "I":
		regs := parseReg(other, 3)
		ins = toInsI(op, regs[1], regs[0], regs[2])
	case "I1":
		rs, rt, imm := parseI(other)
		ins = toInsI(op, rs, rt, imm)
	case "I2":
		rs, imm := parseII(other)
		ins = toInsI(op, rs, 0, imm)
	case "R":
		regs := parseReg(other, 3)
		ins = toInsR(op, regs[1], regs[2], regs[0], 0, 0)
	case "R1":
		regs := parseReg(other, 3)
		ins = toInsR(op, 0, regs[1], regs[0], regs[2], 0)
	case "R2":
		regs := parseReg(other, 1)
		ins = toInsR(op, regs[0], 0, 0, 0, 0)
	case "J":
		imm := parseJ(other)
		ins = toInsJ(op, imm)
	}
	return
}

func parseReg(ins string, count int) (rs []int64) {
	regs := strings.Split(ins, ",")
	if len(regs) != count {
		check(errors.New(""), "Illegal instruction")
	}
	rs = make([]int64, count)
	for i := 0; i < count; i++ {
		rs[i], errG = strconv.ParseInt(strings.Replace(regs[i], "$", "", -1), 10, 64)
		check(errG, "Illegal operand")
	}
	return
}

func parseI(ins string) (rs, rt, imm int64) {
	regs := strings.Split(ins, ",")
	if len(regs) != 2 {
		check(errors.New(""), "Illegal instruction")
	}
	rt, errG = strconv.ParseInt(strings.Replace(regs[0], "$", "", -1), 10, 64)
	check(errG, "Illegal operand")
	regs = strings.Split(regs[1], "$")
	if len(regs) != 2 {
		check(errors.New(""), "Illegal instruction")
	}
	rs, errG = strconv.ParseInt(strings.Replace(regs[1], ")", "", -1), 10, 64)
	check(errG, "Illegal operand")
	imm, errG = strconv.ParseInt(strings.Replace(regs[0], "(", "", -1), 10, 64)
	check(errG, "Illegal operand")
	return
}

func parseJ(ins string) (imm int64) {
	if ins == "" {
		imm = 0
	} else {
		imm, errG = strconv.ParseInt(ins, 0, 64)
		check(errG, "Illegal operand")
	}
	return
}

func parseII(ins string) (rs, imm int64) {
	regs := strings.Split(ins, ",")
	if len(regs) != 2 {
		check(errors.New(""), "Illegal instruction")
	}
	rs, errG = strconv.ParseInt(strings.Replace(regs[0], "$", "", -1), 10, 64)
	check(errG, "Illegal operand")
	imm, errG = strconv.ParseInt(regs[1], 10, 64)
	check(errG, "Illegal operand")
	return
}

func int64toBinStr(num, lenght int64) (ins string) {
	r, err := strconv.Atoi(strconv.FormatInt(num, 2))
	check(err, "Illegal operand")
	ins = fmt.Sprintf("%0"+strconv.FormatInt(lenght, 10)+"d", r)
	return
}

func toInsR(op string, rs, rt, rd, sa, funct int64) (ins string) {
	if rs > 31 || rt > 31 || rd > 31 || sa > 31 || funct > 63 {
		check(errors.New(""), "Illegal register")
	}
	ins = op
	ins += int64toBinStr(rs, 5)
	ins += int64toBinStr(rt, 5)
	ins += int64toBinStr(rd, 5)
	ins += int64toBinStr(sa, 5)
	ins += int64toBinStr(funct, 6)
	return
}

func toImm(imm, length int64) (str string) {
	if imm >= 0 {
		str = int64toBinStr(imm, length)
	} else {
		tmp := []byte(int64toBinStr(-imm, length))
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
		check(errors.New(""), "Illegal register")
	}
	ins = op
	ins += int64toBinStr(rs, 5)
	ins += int64toBinStr(rt, 5)
	ins += toImm(imm, 16)
	return
}

func toInsJ(op string, imm int64) (ins string) {
	ins = op
	if imm < 0 {
		check(errors.New(""), "Illegal instruction")
	}
	imm = imm >> 2
	r, err := strconv.Atoi(strconv.FormatInt(imm, 2))
	check(err, "Illegal operand")
	ins += fmt.Sprintf("%026d", r)
	return
}

func init() {
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
