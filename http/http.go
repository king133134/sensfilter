package http

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"runtime"
	"sensfilter"
	"sync"
	"time"
)

// SourceStrings 定义 SourceStrings 类型，表示敏感词来源为字符串切片
type SourceStrings []string

// SourceFilename 定义 SourceFilename 类型，表示敏感词来源为文件路径
type SourceFilename string

// SourceUrl 定义 SourceUrl 类型，表示敏感词来源为 URL
type SourceUrl string

// SourceMySQL 定义 SourceMySQL 类型，表示敏感词来源为 MySQL 数据库
type SourceMySQL *sensfilter.DatabaseConf

// SourceFunc 定义 SourceFunc 类型，表示敏感词来源为自定义函数
type SourceFunc func(w *sensfilter.TrieWriter)

// SourceAddFunc 定义增量更新类型，表示可以在原来的敏感词基础上做增量更新
type SourceAddFunc func(w *sensfilter.TrieWriter)

// resultWriter 结构体继承自 sensfilter.DefaultResultWriter，用于响应检查请求
type resultWriter struct {
	sensfilter.DefaultResultWriter
}

// DefaultHttpHandler 结构体，用于处理 HTTP 请求
type DefaultHttpHandler struct{}

// Check 方法用于检查输入文本是否包含敏感词，如果有则返回结果
func (h *DefaultHttpHandler) Check(server *Server, context *gin.Context) {
	text := context.PostForm("text")
	res := server.Search().Find([]byte(text))
	context.JSON(http.StatusOK, res)
}

// 获取内存使用情况
func getMemStats() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// Rebuild 方法用于重建敏感词索引
func (h *DefaultHttpHandler) Rebuild(server *Server, context *gin.Context) {
	start := time.Now()
	var err error
	err = server.BuildSearch()
	if err != nil {
		log.Fatal(err)
	}
	context.HTML(http.StatusOK, "form.html", gin.H{
		"time":   time.Now().UnixMilli() - start.UnixMilli(),
		"memory": fmt.Sprintf("%.2f", float64(getMemStats())/1024/1024),
		"size":   server.Search().TrieWriter().Size(),
		"skip":   server.Search().TrieWriter().Skip(),
	})
}

// Handler 接口定义了 Check 和 Rebuild 两个方法
type Handler interface {
	Check(server *Server, context *gin.Context)
	Rebuild(server *Server, context *gin.Context)
}

type StartBefore func(server *Server)

// Server 结构体表示敏感词服务器，包含端口号、敏感词来源、处理 HTTP 请求的 Handler 和跳过的字符列表
type Server struct {
	port        int                // 服务器端口号
	source      interface{}        // 敏感词来源
	handler     Handler            // 处理 HTTP 请求的 Handler
	skip        string             // 跳过的字符列表
	startBefore StartBefore        // http服务启动之前回调函数,例如：可以用来增加路由，监控，日志等等
	lock        sync.Mutex         // 定义读写锁，用于保护重建敏感词索引的过程
	search      *sensfilter.Search // 敏感词搜索器
	gin         *gin.Engine        // gin
}

// BuildSearch 方法用于根据敏感词来源建立敏感词索引
func (_this *Server) BuildSearch() (err error) {
	defer _this.lock.Unlock()
	_this.lock.Lock()
	switch val := _this.source.(type) {
	case SourceStrings: // 如果敏感词来源为字符串切片
		_this.search, err = sensfilter.Strings(val, _this.skip), nil // 调用 sensfilter.Strings 函数建立索引
	case SourceFilename: // 如果敏感词来源为文件路径
		_this.search, err = sensfilter.File(string(val), _this.skip) // 调用 sensfilter.File 函数建立索引
	case SourceMySQL: // 如果敏感词来源为 MySQL 数据库
		_this.search, err = sensfilter.MySQL(val, _this.skip) // 调用 sensfilter.MySQL 函数建立索引
	case SourceUrl: // 如果敏感词来源为 Url 网页
		_this.search, err = sensfilter.Network(string(val), _this.skip)
	case SourceFunc: // 如果敏感词来源为自定义函数
		search := sensfilter.NewSearch(sensfilter.SetWriter(sensfilter.NewTrieWriter()), sensfilter.SetSkip(_this.skip))
		val(search.TrieWriter())
		search.TrieWriter().BuildFail()
		_this.search = search
	default:
		err = errors.New("unknown type can't build search")
	}
	return
}

// Search 返回当前Server中的 敏感词过滤器 实例
func (_this *Server) Search() *sensfilter.Search {
	return _this.search
}

// Gin 返回当前Server中的 gin 实例
func (_this *Server) Gin() *gin.Engine {
	return _this.gin
}

// Start 用于创建HTTP服务器并启动它。它会将HTML模板加载到gin中，并注册两个处理函数check和rebuild。
func (_this *Server) Start() {
	// 创建一个默认的gin路由器
	router := _this.gin
	// 处理HTTP POST请求
	router.POST("/check", func(context *gin.Context) {
		_this.handler.Check(_this, context)
	})
	// 处理HTTP GET请求
	router.GET("/rebuild", func(context *gin.Context) {
		_this.handler.Rebuild(_this, context)
	})

	// http服务启动前
	_this.startBefore(_this)
	// 监听并启动HTTP服务器
	_ = router.Run(fmt.Sprintf(":%d", _this.port))
}

// Option 是一个可选参数的类型
type Option func(opt *Server)

// SetPort 设置HTTP服务器的端口
func SetPort(port int) Option {
	return func(opt *Server) {
		opt.port = port
	}
}

// SetSourceStrings 设置一个字符串数组作为敏感词过滤源
func SetSourceStrings(source []string) Option {
	return func(opt *Server) {
		opt.source = SourceStrings(source)
	}
}

// SetSourceMysql 设置MySQL作为敏感词过滤源
func SetSourceMysql(source *sensfilter.DatabaseConf) Option {
	return func(opt *Server) {
		opt.source = SourceMySQL(source)
	}
}

// SetNetworkSource 设置一个URL作为敏感词过滤源
func SetNetworkSource(source string) Option {
	return func(opt *Server) {
		opt.source = SourceUrl(source)
	}
}

// SetSourceFile 设置一个文件作为敏感词过滤源
func SetSourceFile(source string) Option {
	return func(opt *Server) {
		opt.source = SourceFilename(source)
	}
}

// SetSourceFunc 设置自定义函数为名词过滤源
func SetSourceFunc(source SourceFunc) Option {
	return func(opt *Server) {
		opt.source = source
	}
}

// SetHandler 设置HTTP请求处理器
func SetHandler(handler Handler) Option {
	return func(opt *Server) {
		opt.handler = handler
	}
}

// SetSkip 设置跳过字符
func SetSkip(s string) Option {
	return func(http *Server) {
		http.skip = s
	}
}

func SetGin(gin *gin.Engine) Option {
	return func(http *Server) {
		http.gin = gin
	}
}

func SetStartBefore(f StartBefore) Option {
	return func(http *Server) {
		http.startBefore = f
	}
}

// NewServer 创建一个新的Server实例
func NewServer(opts ...Option) *Server {
	// 创建一个默认的Server实例
	s := &Server{
		port:    8080,
		source:  SourceFilename(sensfilter.GetCurrentAbPath() + "/example/sens_words.txt"),
		skip:    sensfilter.SortedSkipList(),
		handler: &DefaultHttpHandler{},
		gin:     nil,
		startBefore: func(ser *Server) {
			// 加载HTML文件
			ser.Gin().LoadHTMLFiles(sensfilter.GetCurrentAbPath() + "/example/form.html")
		},
	}

	// 遍历所有的可选参数并将它们应用到Server实例中
	for _, o := range opts {
		o(s)
	}

	// 构建敏感词过滤器
	err := s.BuildSearch()
	if err != nil {
		log.Fatal(err)
	}

	// 如果gin没有被赋值，使用默认配置gin
	if s.gin == nil {
		s.gin = gin.Default()
	}

	return s
}
