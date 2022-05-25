package main

import "github.com/dengpju/higo-pyaml/pyaml"

func main()  {
	_ = pyaml.Unmarshal("./app.yaml", nil)
}
