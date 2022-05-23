package pyaml

import (
	"fmt"
	"github.com/dengpju/higo-utils/utils"
	"io/ioutil"
	"strings"
)

type Groups struct {
	group []*Group
}

type Group struct {
	line  int
	group []*Raw
}

type Raw struct {
	parent         *Raw
	prefixBlankNum int
	key            string
	value          interface{}
	child          []*Raw
}

func Unmarshal(filename string, out interface{}) (err error)  {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	group := &Groups{}
	var currentGroup *Group
	_ = utils.File.Read(filename).ForEach(func(line int, raw []byte) bool {
		var (
			prefixBlankNum []int32
			unblankNum     []int32
			rawKey         []int32
			rawValue       []int32
			currentRaw     *Raw
		)
		currentRaw = &Raw{}
		rowset := make(map[int32]int32)
		for i, b := range []rune(string(raw)) {
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
		newGroup:
			if currentGroup == nil {
				currentGroup = &Group{line: line, group: make([]*Raw, 0)}
				group.group = append(group.group, currentGroup)
			}
			if currentRaw.prefixBlankNum == 0 && currentGroup.line != line { // 开新组
				if currentRaw.key != "" {
					currentGroup = nil
					goto newGroup
				}
			}
			if currentRaw.key != "" {
				currentGroup.group = append(currentGroup.group, currentRaw)
			}
		}
		return true
	})
	fmt.Println(group)
	for _, g := range group.group {
		for _, gg := range g.group {
			fmt.Println(gg)
		}
	}
	return nil
}