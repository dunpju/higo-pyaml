package pyaml

import (
	"github.com/dengpju/higo-utils/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

type groups struct {
	group []*group
}

type Pyaml struct {
	ymap map[interface{}]interface{}
	keys map[string]int
	raws []*Raw
}

func NewPyaml() *Pyaml {
	return &Pyaml{ymap: make(map[interface{}]interface{}), keys: make(map[string]int), raws: make([]*Raw, 0)}
}

func (this *Pyaml) Each(fn func(raw *Raw) bool) {
	each(this.raws, fn)
}

func (this *Pyaml) Get(key string) *Raw {
	if strings.Contains(key, ".") {
		keys := strings.Split(key, ".")
		var raw *Raw
		for i, k := range keys {
			if i == 0 {
				if index, ok := this.keys[k]; ok {
					raw = this.raws[index]
				}
			} else {
				raw = raw.Get(k)
			}
		}
		return raw
	} else {
		if index, ok := this.keys[key]; ok {
			return this.raws[index]
		}
	}
	return nil
}

func (this *Pyaml) Map() map[interface{}]interface{} {
	return this.ymap
}

type Raws []*Raw

func each(rs Raws, fn func(raw *Raw) bool) {
	for _, r := range rs {
		if fn != nil {
			if !fn(r) {
				break
			}
		}
		if len(r.child) > 0 {
			each(r.child, fn)
		}
	}
}

type group struct {
	line int
	raws []*Raw
}

type Raw struct {
	parent         *Raw
	prefixBlankNum int
	line           int
	path           string
	key            string
	value          interface{}
	keys           map[string]int
	child          []*Raw
}

func NewRaw() *Raw {
	return &Raw{keys: make(map[string]int), child: make([]*Raw, 0)}
}

func (this *Raw) Get(key string) *Raw {
	if index, ok := this.keys[key]; ok {
		return this.child[index]
	}
	return nil
}

func (this *Raw) Value() interface{} {
	return this.value
}

func Unmarshal(filename string) (pya *Pyaml, err error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	yamlMap := make(map[interface{}]interface{})
	yamlFileErr := yaml.Unmarshal(yamlFile, yamlMap)
	if yamlFileErr != nil {
		return nil, yamlFileErr
	}
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
			currentRaw     *Raw
		)
		currentRaw = NewRaw()
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
				currentGroup = &group{line: line, raws: make([]*Raw, 0)}
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
	rs := make([]*Raw, 0)
	pya = NewPyaml()
	for _, g := range gs.group {
		for i, r := range g.raws {
			if r.prefixBlankNum == fileValidFirstRawBlankNum || len(rs) == 0 {
				pya.keys[r.key] = len(rs)
				r.path = r.key
				rs = append(rs, r)
			} else if len(rs) > 0 {
				if r.prefixBlankNum > g.raws[i-1].prefixBlankNum {
					r.parent = g.raws[i-1]
					r.path = r.parent.path + "`" + r.key
					g.raws[i-1].child = append(g.raws[i-1].child, r)
					g.raws[i-1].keys[r.key] = len(g.raws[i-1].child) - 1
				} else {
					j := i
				forBegin:
					if r.prefixBlankNum == g.raws[j-1].prefixBlankNum {
						r.parent = g.raws[j-1].parent
						r.path = r.parent.path + "`" + r.key
						g.raws[j-1].parent.child = append(g.raws[j-1].parent.child, r)
						g.raws[j-1].parent.keys[r.key] = len(g.raws[j-1].parent.child) - 1
					} else if r.prefixBlankNum < g.raws[j-1].prefixBlankNum {
						j--
						goto forBegin
					}
				}
			}
		}
	}
	pya.ymap = yamlMap
	pya.raws = rs
	return pya, nil
}
