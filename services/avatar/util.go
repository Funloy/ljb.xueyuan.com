package avatar

import (
	"math/rand"
	"regexp"
	"strings"
)

// 返回在[min, max]区间的随机整数
func randInt(rnd *rand.Rand, min int, max int) int {
	return min + rnd.Intn(max-min)
}

// 从字符串切片数组中随机返回某个字符串
func randSliceString(rnd *rand.Rand, slice []string) string {
	return slice[randInt(rnd, 0, len(slice))]
}

type naturalSort []string

var r = regexp.MustCompile(`[^0-9]+|[0-9]+`)

func (s naturalSort) Len() int {
	return len(s)
}
func (s naturalSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s naturalSort) Less(i, j int) bool {

	spliti := r.FindAllString(strings.Replace(s[i], " ", "", -1), -1)
	splitj := r.FindAllString(strings.Replace(s[j], " ", "", -1), -1)

	for index := 0; index < len(spliti) && index < len(splitj); index++ {
		if spliti[index] != splitj[index] {
			//全为0
			if isNumber(spliti[index][0]) && isNumber(splitj[index][0]) {
				// 删除前面的0
				stringi := strings.TrimLeft(spliti[index], "0")
				stringj := strings.TrimLeft(splitj[index], "0")
				if len(stringi) == len(stringj) {
					for indexchar := 0; indexchar < len(stringi); indexchar++ {
						if stringi[indexchar] != stringj[indexchar] {
							return stringi[indexchar] < stringj[indexchar]
						}
					}
					return len(spliti[index]) < len(splitj[index])
				}
				return len(stringi) < len(stringj)
			}
			// One of the slices is a number (we give precedence to numbers regardless of ASCII table position)
			if isNumber(spliti[index][0]) || isNumber(splitj[index][0]) {
				return isNumber(spliti[index][0])
			}
			// 全不为0
			return spliti[index] < splitj[index]
		}

	}
	// Fall back for cases where space characters have been annihilated by the replacement call
	// Here we iterate over the unmolested string and prioritize numbers
	for index := 0; index < len(s[i]) && index < len(s[j]); index++ {
		if isNumber(s[i][index]) || isNumber(s[j][index]) {
			return isNumber(s[i][index])
		}
	}
	return s[i] < s[j]
}
func isNumber(input uint8) bool {
	return input >= '0' && input <= '9'
}