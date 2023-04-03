package sensfilter

import (
	"bytes"
	"github.com/goccy/go-json"
)

type ResultWriter interface {
	// Write 写入结果，stop是否停止写入
	Write(res *Result) (stop bool)
	Len() int
}

type DefaultResultWriter struct {
	list []*Result
}

func (_this *DefaultResultWriter) List() []*Result {
	return _this.list
}

func (_this *DefaultResultWriter) Write(res *Result) (stop bool) {
	_this.list = append(_this.list, res)
	return false
}

func (_this *DefaultResultWriter) Len() int {
	return len(_this.list)
}

func (_this *DefaultResultWriter) MarshalJSON() ([]byte, error) {
	return json.Marshal(_this.list)
}

func (_this *DefaultResultWriter) String() string {
	if len(_this.list) == 0 {
		return ""
	}
	buf := bytes.Buffer{}
	buf.Write([]byte{'['})
	last := _this.Len() - 1
	for i, v := range _this.list {
		buf.WriteString(v.String())
		if i == last {
			buf.WriteByte(']')
		} else {
			buf.Write([]byte{',', '\n'})
		}
	}
	return buf.String()
}
