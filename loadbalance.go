package main

type LoadBalance interface {
	Init([]interface{})
	Pick() interface{}
	Reset()
}

func NewLB(string) {

}
