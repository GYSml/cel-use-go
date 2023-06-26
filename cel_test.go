package main

import "testing"

func TestInitCleConf(t *testing.T) {
	InitCleConf()
	val := make(map[string]interface{})
	// 变量之间的加减乘除
	val["M"] = 5
	val["N"] = 2
	result,err := calculate("(M+N)*M",val)
	t.Errorf("result:%v err:%v \n",result,err)
}

