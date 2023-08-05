package pinyin

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	tones = [][]rune{
		{'ā', 'ē', 'ī', 'ō', 'ū', 'ǖ', 'Ā', 'Ē', 'Ī', 'Ō', 'Ū', 'Ǖ'},
		{'á', 'é', 'í', 'ó', 'ú', 'ǘ', 'Á', 'É', 'Í', 'Ó', 'Ú', 'Ǘ'},
		{'ǎ', 'ě', 'ǐ', 'ǒ', 'ǔ', 'ǚ', 'Ǎ', 'Ě', 'Ǐ', 'Ǒ', 'Ǔ', 'Ǚ'},
		{'à', 'è', 'ì', 'ò', 'ù', 'ǜ', 'À', 'È', 'Ì', 'Ò', 'Ù', 'Ǜ'},
	}
	neutrals = []rune{'a', 'e', 'i', 'o', 'u', 'v', 'A', 'E', 'I', 'O', 'U', 'V'}
	dim      = map[string]string{
		"eng":  "en",
		"ing":  "in",
		"yong": "yun",
	}
	numberSound = map[rune]string{
		48: "lin",
		49: "yi",
		50: "er",
		51: "san",
		52: "si",
		53: "wu",
		54: "liu",
		55: "qi",
		56: "ba",
		57: "jiu",
	}
)

var (
	// 从带声调的声母到对应的英文字符的映射
	tonesMap map[rune]rune

	// 从汉字到声调的映射
	numericTonesMap map[rune]int

	// 从汉字到拼音的映射（带声调）
	pinyinMap map[rune]string
)

// Mode 返回输出模式
type Mode int

const (
	// WithoutTone 默认模式，例如：guo
	WithoutTone Mode = iota + 1
	//Tone 带声调的拼音 例如：guó
	Tone
	//InitialsInCapitals 首字母大写不带声调，例如：Guo
	InitialsInCapitals
)

// Pinyin 定义Pinyin结构
type Pinyin struct {
}

// New 实例化一个Pinyin对象
func New() *Pinyin {
	return &Pinyin{}
}

// Init 根据文件地址初始化数据
func (py *Pinyin) Init(file string) error {
	tonesMap = make(map[rune]rune)
	numericTonesMap = make(map[rune]int)
	pinyinMap = make(map[rune]string)
	for i, runes := range tones {
		for j, tone := range runes {
			tonesMap[tone] = neutrals[j]
			numericTonesMap[tone] = i + 1
		}
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		strs := strings.Split(scanner.Text(), "=>")
		if len(strs) < 2 {
			continue
		}
		i, err := strconv.ParseInt(strs[0], 16, 32)
		if err != nil {
			continue
		}
		pinyinMap[rune(i)] = strs[1]
	}
	return nil
}

// ConvertWithoutTone 返回不带声调的拼音
func (py *Pinyin) ConvertWithoutTone(origin string) ([]string, error) {
	return py.Convert(origin, WithoutTone)
}

// ConvertWithoutToneToSame 返回不带声调的拼音
func (py *Pinyin) ConvertWithoutToneToSame(origin string) ([]string, error) {
	words, err := py.Convert(origin, WithoutTone)
	if err != nil {
		return words, err
	}
	for i := 0; i < len(words); i++ {
		for y, to := range dim {
			if strings.HasSuffix(words[i], y) {
				words[i] = strings.Replace(words[i], y, to, 1)
				break
			}
		}
	}
	return words, err
}

// Convert 自己设定返回模式
func (py *Pinyin) Convert(origin string, mode Mode) ([]string, error) {
	//sr := []rune(origin)
	words := make([]string, 0)
	buffer := &bytes.Buffer{}
	for _, s := range origin {
		if unicode.IsNumber(s) {
			if sound, ok := numberSound[s]; ok {
				buffer.WriteString(sound)
			} else {
				buffer.WriteRune(s)
			}
			continue
		} else if unicode.IsLetter(s) {
			if s >= 'A' && s < 'Z' {
				buffer.WriteRune(s)
				continue
			} else if s >= 'a' && s < 'z' {
				buffer.WriteRune(s)
				continue
			} else {
				if buffer.Len() > 0 {
					words = append(words, buffer.String())
					buffer.Reset()
				}
				word, err := getPinyin(s, mode)
				if err != nil {
					return words, err
				}
				if len(word) > 0 {
					words = append(words, word)
				} else {
					words = append(words, fmt.Sprintf("%c", s))
				}
			}
		}
	}
	if buffer.Len() > 0 {
		words = append(words, buffer.String())
		buffer.Reset()
	}
	return words, nil
}

func getPinyin(hanzi rune, mode Mode) (string, error) {
	switch mode {
	case Tone:
		return getTone(hanzi), nil
	case InitialsInCapitals:
		return getInitialsInCapitals(hanzi), nil
	default:
		return getDefault(hanzi), nil
	}
}

func getTone(hanzi rune) string {
	return pinyinMap[hanzi]
}

func getDefault(hanzi rune) string {
	tone := getTone(hanzi)

	if tone == "" {
		return tone
	}

	output := make([]rune, utf8.RuneCountInString(tone))

	count := 0
	for _, t := range tone {
		neutral, found := tonesMap[t]
		if found {
			output[count] = neutral
		} else {
			output[count] = t
		}
		count++
	}
	return string(output)
}

func getInitialsInCapitals(hanzi rune) string {
	def := getDefault(hanzi)
	if def == "" {
		return def
	}
	sr := []rune(def)
	if sr[0] > 32 {
		sr[0] = sr[0] - 32
	}
	return string(sr)
}
