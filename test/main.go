package main

import (
	"fmt"
	"github.com/dengpju/higo-pyaml/pyaml"
)

func main()  {
	pya, _ := pyaml.Unmarshal("./app.yaml")
	fmt.Println(pya.Get("gg.hh.y1").Value())
	fmt.Printf("%T\n",pya.Get("gg.hh.y1").Value())
	fmt.Println(pya.Get("yyy").Value())
	fmt.Printf("%T\n", pya.Get("yyy").Value())
	//fmt.Println(raws.Get("gg").Get("hh").Get("y1").Get("o").Get("yy").Get("tt"))
	//pya.Each(func(raw *pyaml.Raw) bool {
	//	fmt.Println(raw)
	//	return true
	//})
}
