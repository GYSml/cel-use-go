package cel

import (
	"testing"
)

func TestInitCleConf(t *testing.T) {
	InitCleConf()
	val := make(map[string]interface{})
	data :=  Student{Age: 12,AgeRes: "user.Age * user.Age"}
	val["user"] = data
	result,err := calculate("user.Age",val)
	t.Errorf("result:%v err:%v \n",result,err)
	result,err = calculate("[1,2,3].sum()",val)
	t.Errorf("result:%v err:%v \n",result,err)
	result,err = calculate("[1,2,3].map(x,x*x)",val)
	t.Errorf("result:%v err:%v \n",result,err)
}

