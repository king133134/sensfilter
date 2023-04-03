package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"log"
	"os"
	"runtime"
	"sensfilter"
	"sensfilter/http"
	"time"
)

func stringsNew() {
	// 待过滤的敏感词
	words := []string{"林茹", "林如", "临蓐", "空子", "霸王龙", "我是个SB", "是我", "TMD", "他妈的", "他妈"}
	// 创建敏感词过滤器
	obj := sensfilter.Strings(words)
	// 待过滤的字符串
	str := []byte("我空ss子sss我是霸**王*龙,我是我我是个(S)(B)真的,TMD，他妈的")
	// 查找敏感词
	fmt.Println(obj.Find(str))
	// 将敏感词替换成 '*'
	fmt.Println(string(obj.Replace(str, '*')))
	// 将敏感词替换成指定的rune
	fmt.Println(string(obj.ReplaceRune(str, '*')))
}

func fileNew() {
	// 读取敏感词文件的路径
	fileName := sensfilter.GetCurrentAbPath() + "/example/sens_words.txt"
	// 创建敏感词过滤器
	obj, err := sensfilter.File(fileName)
	if err != nil {
		log.Fatal(err)
	}
	// 待过滤的字符串
	str := []byte("我空ss子sss我是霸**王*龙,我是我我是个(S)(B)真的TMD")
	// 查找敏感词
	res := obj.Find(str)
	// 判断字符串是否包含敏感词
	fmt.Println(res, obj.HasSens(str), obj.ReplaceRune(str, rune('*')))
}

func networkNew() {
	fmt.Println("network")
	// 敏感词文件的url
	url := "https://raw.githubusercontent.com/king133134/test/main/sensi_words.txt"
	// 创建敏感词过滤器
	obj, err := sensfilter.Network(url)
	if err != nil {
		log.Fatal(err)
	}
	// 待过滤的字符串
	str := "我空ss子sss我是霸**王*龙,我是我我是个(S)(B)真的TMD"
	// 查找敏感词
	res := obj.Find([]byte(str))
	// 判断字符串是否包含敏感词
	fmt.Println(res, obj.HasSens([]byte("我空ss子sss我是霸**王*龙,我是我我是个(S)(B)真的")))
}

func mysqlNew() {
	// 定义 MySQL 数据库的配置
	conf := &sensfilter.DatabaseConf{
		TableName: "word_test",
		DSN:       "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
	}

	// 创建 MySQL 表格
	err := sensfilter.CreateMySQLTable(conf)
	if err != nil {
		log.Fatal(err)
	}

	// 连接 MySQL 数据库
	db, err := gorm.Open(mysql.Open(conf.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 读取敏感词列表
	data, err := os.ReadFile(sensfilter.GetCurrentAbPath() + "/word")
	if err != nil {
		log.Fatal(err)
	}

	// 定义敏感词结构体
	type SensitiveWord struct {
		Word string
	}

	words := []SensitiveWord{}
	for i, n := 0, len(data); i < n; i++ {
		j := i
		for ; j < n && data[j] != '\n'; j++ {
		}
		words = append(words, SensitiveWord{string(data[i:j])})
		i = j
	}

	// 批量插入敏感词到 MySQL 表格中
	db.Table(conf.TableName).Select("word").CreateInBatches(words, 1000)

	// 获取敏感词过滤器对象
	obj, err := sensfilter.MySQL(conf)
	if err != nil {
		log.Fatal(err)
	}

	// 测试敏感词过滤
	str := "我空ss子sss我是霸**王*龙,我是我我是个(S)(B)真的TMD"
	res := obj.Find([]byte(str))
	fmt.Println(res)
}

// originNew 创建 SensFilter 实例并进行基础设置，测试敏感词匹配
func originNew() {
	// 创建 SensFilter 实例，设置 TrieWriter 为默认的 TrieWriter 实例
	// 设置 Skip 为 "*!"，表示跳过所有以 * 或 ! 开头的敏感词
	obj := sensfilter.NewSearch(sensfilter.SetWriter(sensfilter.NewTrieWriter()), sensfilter.SetSkip("*!"))
	// 向 TrieWriter 中插入敏感词
	obj.TrieWriter().InsertWords([]string{"www"})
	// 构建 Trie 失败指针
	obj.TrieWriter().BuildFail()
	// 测试敏感词匹配
	fmt.Println(obj.HasSens([]byte("wwwww")))
}

// Handler 是 HTTP 请求处理程序
type Handler struct {
}

// Check 处理包含文本内容的 POST 请求
func (h *Handler) Check(server *http.Server, context *gin.Context) {
	// 从 POST 请求中获取文本内容
	text := context.PostForm("text")
	// 在 SensFilter 中查找敏感词
	res := server.Search().Find([]byte(text))
	// 返回结果
	context.JSON(200, res)
}

// getMemStats 获取当前程序占用的内存大小
func getMemStats() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// Rebuild 处理重建敏感词过滤器的 POST 请求
func (h *Handler) Rebuild(server *http.Server, context *gin.Context) {
	// 记录开始时间
	start := time.Now()
	var err error
	// 重建敏感词过滤器
	err = server.BuildSearch()
	if err != nil {
		log.Fatal(err)
	}
	// 返回重建结果和统计信息
	context.HTML(200, "form.html", gin.H{
		"title":  "Rebuild Test",
		"time":   time.Now().UnixMilli() - start.UnixMilli(),
		"memory": fmt.Sprintf("%.2f", float64(getMemStats())/1024/1024),
		"size":   server.Search().TrieWriter().Size(),
		"skip":   server.Search().TrieWriter().Skip(),
	})
}

// mySourceFunc 是自定义的 TrieWriter 源函数，用于向 TrieWriter 中添加敏感词
func mySourceFunc(w *sensfilter.TrieWriter) {
	w.InsertWords([]string{"www", "TMD", "LGD", "妹妹的"})
}

// myFileSourceFunc 是自定义的 TrieWriter 源函数，从文件中读取敏感词添加到 TrieWriter 中
func myFileSourceFunc(w *sensfilter.TrieWriter) {
	// 打开敏感词文件
	f, err := os.Open(sensfilter.GetCurrentAbPath() + "/example/sens_words.txt")
	if err != nil {
		log.Fatal(err)
	}
	// 将文件内容写入 TrieWriter
	_, err = io.Copy(w, f)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// 创建 HTTP 服务器，设置 TrieWriter 源函数为 myFileSourceFunc
	http.NewServer(http.SetSourceFunc(myFileSourceFunc)).Start()
}
