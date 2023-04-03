敏感词过滤
=====

简介
--

敏感词过滤是一个使用 Go 语言编写基于[AC自动机算法](https://en.wikipedia.org/wiki/Aho%E2%80%93Corasick_algorithm)简单易用的敏感词过滤库。
该库支持敏感词搜索和替换功能，可以用于敏感信息过滤、垃圾邮件过滤等场景。

功能特点
----

*   支持中英文敏感词过滤
*   支持敏感词搜索和替换
*   支持用户自定义跳过字符列表
*   支持快速很多快捷方式使用：字符串数组、文件、MySQL、网页.详情使用请看[example](github.com/king133134/sensfilter/example/main.go)
*   支持当成一个单独http服务器启动，是基于[gin](https://github.com/gin-gonic/gin)

快速开始
----

### 安装

使用如下命令安装：

shell

```shell
go get github.com/king133134/sensfilter
```

### 使用

以下是一个简单的示例代码：

go

```go
package main

import (
	"fmt"
	"github.com/king133134/sensfilter"
)

func main() {
	words := []string{"林茹", "林如", "临蓐", "空子", "霸王龙", "我是个SB", "是我", "TMD", "他妈的", "他妈"}
	obj := sensfilter.Strings(words)
	str := []byte("我空ss子sss我是霸**王*龙,我是我我是个(S)(B)真的,TMD，他妈的")
	fmt.Println(obj.Find(str))
	fmt.Println(string(obj.Replace(str, '*')))
	fmt.Println(string(obj.ReplaceRune(str, '*')))
}
```

许可证
---

[MIT](https://github.com/king133134/sensfilter/blob/main/LICENSE)