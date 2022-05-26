package pyaml

import (
	"fmt"
	"github.com/dengpju/higo-utils/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

type groups struct {
	group []*group
}

type group struct {
	line int
	raws []*raw
}

type raw struct {
	parent         *raw
	prefixBlankNum int
	line           int
	key            string
	value          interface{}
	child          []*raw
}

func Unmarshal(filename string, out interface{}) (err error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	yamlMap := make(map[interface{}]interface{})
	yamlFileErr := yaml.Unmarshal(yamlFile, yamlMap)
	fmt.Println(yamlFileErr, yamlMap)
	gs := &groups{}
	var (
		currentGroup              *group
		fileValidFirstRawBlankNum int
	)
	_ = utils.File.Read(filename).ForEach(func(line int, r []byte) bool {
		var (
			prefixBlankNum []int32
			unblankNum     []int32
			rawKey         []int32
			rawValue       []int32
			currentRaw     *raw
		)
		currentRaw = &raw{}
		rowset := make(map[int32]int32)
		for i, b := range []rune(string(r)) {
			if b == 35 { // 行开头注释标记
				if i == 0 {
					currentRaw = nil
				}
				if _, ok := rowset[32]; ok && len(rowset) == 1 { // 注释行 35 -> #
					currentRaw = nil
				}
				break
			}
			if b == 58 { // 58 -> :
				if currentRaw.key == "" {
					currentRaw.prefixBlankNum = len(prefixBlankNum)
					if fileValidFirstRawBlankNum > currentRaw.prefixBlankNum { // 当前行前缀空格数量,比文件有效第一行空格数量小
						return false // 结束文件读取
					}
					currentRaw.key = strings.TrimSuffix(strings.TrimPrefix(string(string(rawKey)), " "), " ")
					currentRaw.key = strings.TrimSuffix(strings.TrimPrefix(string(string(currentRaw.key)), `"`), `"`)
					currentRaw.key = strings.TrimSuffix(strings.TrimPrefix(string(string(currentRaw.key)), `'`), `'`)
					continue
				}
			}
			if currentRaw.key != "" { // 计算value
				rawValue = append(rawValue, b)
			}
			rowset[b] = b
			if b == 32 { // 前缀空格  32 -> 空格
				if len(unblankNum) == 0 {
					prefixBlankNum = append(prefixBlankNum, b)
				}
			} else {
				unblankNum = append(unblankNum, b)
				rawKey = append(rawKey, b)
			}
		}
		if currentRaw != nil {
			value := strings.TrimSuffix(strings.TrimPrefix(string(rawValue), " "), " ")
			value = strings.TrimSuffix(strings.TrimPrefix(value, `"`), `"`)
			currentRaw.value = strings.TrimSuffix(strings.TrimPrefix(value, `'`), `'`)
			currentRaw.line = line
			if len(gs.group) == 0 {
				fileValidFirstRawBlankNum = currentRaw.prefixBlankNum
			}
		newGroup:
			if currentGroup == nil {
				currentGroup = &group{line: line, raws: make([]*raw, 0)}
				gs.group = append(gs.group, currentGroup)
			}
			if currentRaw.prefixBlankNum == fileValidFirstRawBlankNum && currentGroup.line != line { // 开新组
				if currentRaw.key != "" {
					currentGroup = nil
					goto newGroup
				}
			}
			if currentRaw.key != "" {
				currentGroup.raws = append(currentGroup.raws, currentRaw)
			}
		}
		return true
	})
	fmt.Println(gs)
	rs := make([]*raw, 0)
	var currentChildMapGroup interface{}
	var currentMapValue interface{}
	for _, g := range gs.group {
		for i, r := range g.raws {
			if r.prefixBlankNum == fileValidFirstRawBlankNum || len(rs) == 0 {
				currentMapValue = yamlMap[r.key]
				fmt.Println(currentMapValue)
				if r.value != nil && r.value != "" {
					r.value = currentMapValue
				} else {
					currentChildMapGroup = yamlMap[r.key]
				}
				rs = append(rs, r)
			} else if len(rs) > 0 {
				if r.prefixBlankNum > g.raws[i-1].prefixBlankNum {
					r.parent = g.raws[i-1]
					if r.value != nil && r.value != "" {
						fmt.Printf("%T\n", r.value)
						fmt.Println(*r, r.value)
						r.value = currentChildMapGroup.(map[interface{}]interface{})[r.key]
					}
					g.raws[i-1].child = append(g.raws[i-1].child, r)
				} else {
					j := i
				forBegin:
					if r.prefixBlankNum == g.raws[j-1].prefixBlankNum {
						r.parent = g.raws[j-1].parent
						g.raws[j-1].parent.child = append(g.raws[j-1].parent.child, r)
					} else if r.prefixBlankNum < g.raws[j-1].prefixBlankNum {
						j--
						goto forBegin
					}
				}
			}
		}
	}
	fmt.Println(rs)
	fmt.Println(currentChildMapGroup)
	fmt.Println(*rs[0].child[0].child[0].child[0])
	fmt.Printf("%p\n", rs[0].child[0].child[0].child[0].parent)
	return nil
}
