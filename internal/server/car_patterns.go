package service

import (
	"go.uber.org/zap"
	"strconv"
	"strings"
)

const (
	NUM_GRID_MEGA = 25
	NUM_GRID_RUSH = 15
)

var (
	BinGrid        = [NUM_GRID_MEGA]int32{} //每个格子对应的要标1的数
	PatternListNew []int32

	Patterns     [][]int
	P0           = []int{0, 6, 12, 18, 24}
	P1           = []int{4, 8, 12, 16, 20}
	P2           = []int{0, 4, 12, 20, 24}
	P3L1         = []int{0, 5, 10, 15, 20, 1, 6, 11, 16, 21} //12
	P4L2         = []int{1, 6, 11, 16, 21, 2, 7, 12, 17, 22} //23
	P5L3         = []int{2, 7, 12, 17, 22, 3, 8, 13, 18, 23} //34
	P6L4         = []int{3, 8, 13, 18, 23, 4, 9, 14, 19, 24} //45
	P7L5         = []int{0, 5, 10, 15, 20, 2, 7, 12, 17, 22} //13
	P8L6         = []int{1, 6, 11, 16, 21, 3, 8, 13, 18, 23} //24
	P9L7         = []int{2, 7, 12, 17, 22, 4, 9, 14, 19, 24} //35
	P10L8        = []int{0, 5, 10, 15, 20, 3, 8, 13, 18, 23} //14
	P11L9        = []int{1, 6, 11, 16, 21, 4, 9, 14, 19, 24} //25
	P12L10       = []int{0, 5, 10, 15, 20, 4, 9, 14, 19, 24} //15
	P13E         = []int{0, 5, 10, 15, 20, 11, 12, 13, 14}
	P14F         = []int{0, 6, 12, 18, 24, 4, 8, 16, 20}                       //
	P15G1        = []int{0, 5, 10, 15, 20, 1, 6, 11, 16, 21, 2, 7, 12, 17, 22} // 123
	P16G2        = []int{1, 6, 11, 16, 21, 2, 7, 12, 17, 22, 3, 8, 13, 18, 23} //234
	P17G3        = []int{2, 7, 12, 17, 22, 3, 8, 13, 18, 23, 4, 9, 14, 19, 24} //345
	P18G4        = []int{0, 5, 10, 15, 20, 1, 6, 11, 16, 21, 3, 8, 13, 18, 23} //124
	P19G5        = []int{0, 5, 10, 15, 20, 1, 6, 11, 16, 21, 4, 9, 14, 19, 24} //125
	P20G6        = []int{0, 5, 10, 15, 20, 2, 7, 12, 17, 22, 3, 8, 13, 18, 23} //134
	P21G7        = []int{0, 5, 10, 15, 20, 3, 8, 13, 18, 23, 4, 9, 14, 19, 24} //145
	P22G8        = []int{1, 6, 11, 16, 21, 2, 7, 12, 17, 22, 4, 9, 14, 19, 24} //235
	P23G9        = []int{1, 6, 11, 16, 21, 3, 8, 13, 18, 23, 4, 9, 14, 19, 24} //245
	P24G10       = []int{0, 5, 10, 15, 20, 2, 7, 12, 17, 22, 4, 9, 14, 19, 24} //135
	P25H         = []int{0, 1, 2, 3, 5, 10, 15, 20, 4, 9, 14, 19, 24, 21, 22, 23, 12}
	P26M1        = []int{0, 5, 10, 15, 20, 1, 6, 11, 16, 21, 2, 7, 12, 17, 22, 3, 8, 13, 18, 23} //1234
	P27M2        = []int{0, 5, 10, 15, 20, 1, 6, 11, 16, 21, 2, 7, 12, 17, 22, 4, 9, 14, 19, 24} //1235
	P28M3        = []int{0, 5, 10, 15, 20, 1, 6, 11, 16, 21, 3, 8, 13, 18, 23, 4, 9, 14, 19, 24} //1245
	P29M4        = []int{0, 5, 10, 15, 20, 2, 7, 12, 17, 22, 3, 8, 13, 18, 23, 4, 9, 14, 19, 24} // 1345                                                     //1345                                                        //1345
	P30M5        = []int{1, 6, 11, 16, 21, 2, 7, 12, 17, 22, 3, 8, 13, 18, 23, 4, 9, 14, 19, 24} //2345
	P31N         = []int{0, 5, 10, 15, 20, 4, 9, 14, 19, 24, 1, 2, 3, 21, 22, 23, 6, 8, 16, 18, 12}
	Bingo        = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24}
	PrizeTypeMap = []string{"A", "B", "C",
		"D1", "D2", "D3", "D4", "D5", "D6", "D7", "D8", "D9", "D10",
		"E", "F",
		"G1", "G2", "G3", "G4", "G5", "G6", "G7", "G8", "G9", "G10",
		"H",
		"M1", "M2", "M3", "M4", "M5",
		"N"}
)

var (
	RushPatterns     [][]int
	R0               = []int{2, 4, 6, 10, 14}
	R1               = []int{0, 1, 2, 12, 13, 14}
	R2               = []int{2, 5, 8, 11, 14, 4, 10, 6}
	R3               = []int{0, 1, 4, 7, 10, 13, 12, 8}
	R4               = []int{0, 2, 4, 6, 8, 10, 12, 14}
	R5               = []int{0, 3, 6, 9, 12, 4, 5, 10, 11}
	R6               = []int{0, 3, 6, 9, 12, 4, 5, 8, 11, 10}
	R7               = []int{0, 1, 2, 4, 6, 7, 8, 10, 12, 13, 14}
	R8               = []int{0, 3, 6, 9, 12, 1, 13, 2, 5, 8, 11, 14}
	R9               = []int{0, 3, 6, 9, 12, 1, 4, 7, 10, 13, 2, 5, 8, 11, 14}
	R10L1            = []int{0, 3, 6, 9, 12}
	R11L2            = []int{1, 4, 7, 10, 13}
	R12L3            = []int{2, 5, 8, 11, 14}
	R13LL1           = []int{0, 3, 6, 9, 12, 1, 4, 7, 10, 13}
	R14LL2           = []int{1, 4, 7, 10, 13, 2, 5, 8, 11, 14}
	R15LL3           = []int{0, 3, 6, 9, 12, 2, 5, 8, 11, 14}
	RushPrizeTypeMap = []string{
		"A", "B", "C", "D", "E",
		"H", "M", "N", "O", "J",
		"1L1", "1L2", "1L3",
		"2L1", "2L2", "2L3",
	}
)

func InitPatterns() {
	Patterns = append(Patterns, P0)
	Patterns = append(Patterns, P1)
	Patterns = append(Patterns, P2)
	Patterns = append(Patterns, P3L1)
	Patterns = append(Patterns, P4L2)
	Patterns = append(Patterns, P5L3)
	Patterns = append(Patterns, P6L4)
	Patterns = append(Patterns, P7L5)
	Patterns = append(Patterns, P8L6)
	Patterns = append(Patterns, P9L7)
	Patterns = append(Patterns, P10L8)
	Patterns = append(Patterns, P11L9)
	Patterns = append(Patterns, P12L10)
	Patterns = append(Patterns, P13E)
	Patterns = append(Patterns, P14F)
	Patterns = append(Patterns, P15G1)
	Patterns = append(Patterns, P16G2)
	Patterns = append(Patterns, P17G3)
	Patterns = append(Patterns, P18G4)
	Patterns = append(Patterns, P19G5)
	Patterns = append(Patterns, P20G6)
	Patterns = append(Patterns, P21G7)
	Patterns = append(Patterns, P22G8)
	Patterns = append(Patterns, P23G9)
	Patterns = append(Patterns, P24G10)
	Patterns = append(Patterns, P25H)
	Patterns = append(Patterns, P26M1)
	Patterns = append(Patterns, P27M2)
	Patterns = append(Patterns, P28M3)
	Patterns = append(Patterns, P29M4)
	Patterns = append(Patterns, P30M5)
	Patterns = append(Patterns, P31N)
	InitBinGrid()
}

func InitBinGrid() {
	for i := 0; i < NUM_GRID_MEGA; i++ {
		BinGrid[i] = 1 << i
	}
	for _, pattern := range Patterns {
		var d int32
		for _, i2 := range pattern {
			d |= BinGrid[i2]
		}
		PatternListNew = append(PatternListNew, d)
	}
}
func InitPatternsRush() {
	RushPatterns = append(RushPatterns, R0)
	RushPatterns = append(RushPatterns, R1)
	RushPatterns = append(RushPatterns, R2)
	RushPatterns = append(RushPatterns, R3)
	RushPatterns = append(RushPatterns, R4)
	RushPatterns = append(RushPatterns, R5)
	RushPatterns = append(RushPatterns, R6)
	RushPatterns = append(RushPatterns, R7)
	RushPatterns = append(RushPatterns, R8)
	RushPatterns = append(RushPatterns, R9)
	RushPatterns = append(RushPatterns, R10L1)
	RushPatterns = append(RushPatterns, R11L2)
	RushPatterns = append(RushPatterns, R12L3)
	RushPatterns = append(RushPatterns, R13LL1)
	RushPatterns = append(RushPatterns, R14LL2)
	RushPatterns = append(RushPatterns, R15LL3)
}

type CardN struct {
	Id      string
	Numbers []int
	Index   []int
	Prize   []int
	Jackpot bool
	Bingo   bool
	cmodel.TbCard
	//////////////////////////-新算法的辅助结构-///////////////////
	//表示数字n的index值，有点像倒排
	IndexOfN map[int]int
	PrizeN   map[int]struct{} //
	/////////////////////////////////////////
	//false表示这个数没有开起来,最大值为25格，中间的空白格不算
	Marked [NUM_GRID_MEGA]bool
	//25位二进制，为1的位置表示有开出来
	BoardStatus int32
}

func NewCardFromTbCard(numbers string, isMega bool) *CardN {
	card := &CardN{
		Index:       []int{},
		PrizeN:      map[int]struct{}{},
		Marked:      [NUM_GRID_MEGA]bool{},
		IndexOfN:    map[int]int{},
		BoardStatus: 0,
	}
	card.Str2Numbers(numbers, isMega) //把号码放进numbers
	if isMega {
		card.BoardStatus = card.BoardStatus | BinGrid[12]
		card.Marked[12] = true
	}
	return card
}

// 把已出的球给划掉
func (c *CardN) MarkNumber(i int) bool {
	position, ok := c.IndexOfN[i]
	if ok {
		c.Marked[position] = true                         //划掉这个球（开起来了）
		c.BoardStatus = c.BoardStatus | BinGrid[position] //把当前这个格子记为1
		c.Index = append(c.Index, position)
	}
	return ok
}

func (c *CardN) CheckExtraPatternNew() {
	for i, i2 := range PatternListNew {
		if _, ok := c.PrizeN[i]; !ok {
			if c.BoardStatus&i2 == i2 {
				c.PrizeN[i] = struct{}{}
				c.Prize = append(c.Prize, i)
			}
		}
	}
}
func (c *CardN) Str2Numbers(str string, isMega bool) {
	arrSplit := strings.Split(str, ",")
	if len(arrSplit) < 1 {
		return
	}
	nLen := NUM_GRID_MEGA
	if !isMega {
		nLen = NUM_GRID_RUSH
	}
	iBallList := make([]int, 0, nLen)
	for i, s := range arrSplit {
		if isMega && i == 12 { //如果是25格玩法
			iBallList = append(iBallList, 0)
			continue
		}
		gridNum, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		c.IndexOfN[gridNum] = i //倒排索引
		iBallList = append(iBallList, gridNum)
	}
	c.Numbers = iBallList
}

func (c *CardN) Str2NumbersRush(str string) {
	split := strings.Split(str, ",")
	if len(split) < 1 {
		return
	}
	ints := make([]int, 0, 15)
	for _, s := range split {
		atoi, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		ints = append(ints, atoi)
	}
	c.Numbers = ints
}

//func (c *CardN) GenIndex(i int) bool {
//	i2 := i / 15
//	if i%15 > 0 {
//		i2 = i2 + 1
//	}
//	i2 = i2 - 1
//	//0-4
//	//5-9
//	//10-14
//	//15-19
//	//20-24
//	i3 := i2 * 5
//	i4 := (i2+1)*5 - 1
//	for j := i3; j <= i4; j++ {
//		if i == c.Numbers[j] {
//			//fmt.Printf("%d<=>%d\t", i, j)
//			c.Index = append(c.Index, j)
//			return true
//		}
//	}
//	return false
//}

func (c *CardN) GenIndexRush(i int) bool {
	i2 := i / 12
	if i%12 > 0 {
		i2 = i2 + 1
	}
	i2 = i2 - 1
	i3 := i2 * 3
	i4 := (i2+1)*3 - 1
	for j := i3; j <= i4; j++ {
		if j < len(c.Numbers) && i == c.Numbers[j] {
			c.Index = append(c.Index, j)
			return true
		}
	}
	return false
}

//
//func (c *CardN) ExtraPatternsSettle() {
//	for i, pattern := range Patterns {
//		if c.PrizeContain(i) { // contain
//			continue
//		}
//		if Compare(pattern, c.Index) {
//			c.Prize = append(c.Prize, i)
//		}
//	}
//}

func (c *CardN) BingoJackPotSettle(bingo bool) {
	if len(c.Index) >= 24 && !bingo {
		c.Jackpot = true
		c.Bingo = false
	}
	if len(c.Index) >= 24 && bingo && !c.Jackpot {
		c.Bingo = true
		zap.S().Info("中奖bingo")
	}
}
func (c *CardN) BingoJackPotRushSettle() {
	if !c.Jackpot && len(c.Index) == 15 {
		c.Jackpot = true
		zap.S().Info("有人中rush jackpot")
	}
}
func (c *CardN) ExtraPatternsSettleRush() {
	if len(RushPatterns) == 0 {
		InitPatternsRush()
	}
	for i, pattern := range RushPatterns {
		if c.PrizeContain(i) { // contain
			continue
		}
		if RushCompare(pattern, c.Index) {
			c.Prize = append(c.Prize, i)
		}
	}
}
func (c *CardN) PrizeContain(i int) bool {
	if c.Prize != nil {
		for _, i3 := range c.Prize {
			if i == i3 {
				return true
			}
		}
	}
	return false
}

func Compare(p, c []int) bool {
	var q int
	var l = len(p)
	for _, i := range p {
		if 12 == i {
			q++
			if q == l {
				return true
			}
			continue
		}
		for _, i2 := range c {
			if i == i2 {
				q++
				if q == l {
					return true
				}
				continue
			}
		}
	}
	return false
}
func RushCompare(p, c []int) bool {
	var q int
	var l = len(p)
	for _, i := range p {
		for _, i2 := range c {
			if i == i2 {
				q++
				if q == l {
					return true
				}
				continue
			}
		}
	}
	return false
}
func (c *CardN) IsOneTG() bool {
	return !c.Jackpot && !c.Bingo && len(c.Index) == 23
}
func (c *CardN) IsTwoTG() bool {
	return !c.Jackpot && !c.Bingo && len(c.Index) == 22
}
