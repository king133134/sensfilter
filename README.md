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
*   支持快速很多快捷方式使用：字符串数组、文件、MySQL、网页.详情使用请看[example](https://github.com/king133134/sensfilter/blob/master/example/main.go)
*   支持当成一个单独http服务器启动，是基于[gin](https://github.com/gin-gonic/gin)

快速开始
----

### 安装

使用如下命令安装：

shell

```shell
go get github.com/king133134/sensfilter
```

### 快速开始

以下是一个简单的示例代码：

```go
package main

import (
	"fmt"
	"github.com/king133134/sensfilter"
)

func main() {
	words := []string{"林茹", "林如", "临蓐", "空子", "霸王龙", "我是个SB", "是我", "TMD", "他妈的", "他妈"}
	filter := sensfilter.Strings(words)
	str := []byte("我空ss子sss我是霸**王*龙,我是我我是个(S)(B)真的,TMD，他妈的")
	fmt.Println(filter.Find(str))
	fmt.Println(string(filter.Replace(str, '*')))
	fmt.Println(string(filter.ReplaceRune(str, '*')))
}
```

输出：
```text
[word:霸王龙 mathced:霸**王*龙 start:20 end:31; word:是我 mathced:是我 start:36 end:41; word:我是个SB mathced:我是个(S)(B start:42 end:55; word:TMD mathced:TMD start:64 end:66; word:他妈的 mathced:他妈的 start:70 end:78;]
我空ss子sss我是************,我********************)真的,***，*********
我空ss子sss我是******,我**********)真的,***，***
```

### 带有自定义跳过字符列表

以下是一个简单的示例代码：

```go
package main

import (
	"fmt"
	"github.com/king133134/sensfilter"
)

func main() {
	words := []string{"林茹", "林如", "临蓐", "空子", "霸王龙", "我是个SB", "是我", "TMD", "他妈的", "他妈"}
	filter := sensfilter.Strings(words, "*!")
	str := []byte("我空ss子sss我是霸**王*龙,我是我我是个(S)(B)真的,TMD，他妈的")
	fmt.Println(filter.Find(str))
	fmt.Println(string(filter.Replace(str, '*')))
	fmt.Println(string(filter.ReplaceRune(str, '*')))
}
```

上面的"我是个(S)(B)"就无法匹配，默认跳过的字符列表是包含"()"，默认的字符列表[trie.go](https://github.com/king133134/sensfilter/blob/master/skip.go)的sortedSkipList  
输出：
```text
[word:霸王龙 mathced:霸**王*龙 start:20 end:31; word:是我 mathced:是我 start:36 end:41; word:TMD mathced:TMD start:64 end:66; word:他妈的 mathced:他妈的 start:70 end:78;]
我空ss子sss我是************,我******我是个(S)(B)真的,***，*********
我空ss子sss我是******,我**我是个(S)(B)真的,***，***
```

许可证
---

[MIT](https://github.com/king133134/sensfilter/blob/master/LICENSE)