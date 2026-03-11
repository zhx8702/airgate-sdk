package devserver

import (
	"embed"
	"encoding/json"
	"flag"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	sdk "github.com/DouDOU-start/airgate-sdk"
)

//go:embed static
var staticFiles embed.FS

// Config devserver 配置
type Config struct {
	Plugin      sdk.GatewayPlugin                             // 必填：网关插件实例
	Addr        string                                        // 监听地址，默认 ":18080"
	DataDir     string                                        // 数据目录，默认 "./devdata"
	ExtraRoutes func(mux *http.ServeMux, store *AccountStore) // 插件自定义路由（如 OAuth）
}

// Run 启动 devserver（阻塞运行）
func Run(cfg Config) error {
	// 支持命令行覆盖
	addr := flag.String("addr", cfg.Addr, "监听地址")
	dataDir := flag.String("data", cfg.DataDir, "数据目录")
	logFile := flag.String("log", "", "日志文件路径")
	flag.Parse()

	if *addr == "" {
		*addr = ":18080"
	}
	if *dataDir == "" {
		*dataDir = "./devdata"
	}
	if *logFile == "" {
		*logFile = filepath.Join(*dataDir, "debug.log")
	}

	// 初始化日志：控制台 INFO + 文件 DEBUG
	if err := os.MkdirAll(filepath.Dir(*logFile), 0o755); err == nil {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			consoleHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
			fileHandler := slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})
			slog.SetDefault(slog.New(&multiHandler{handlers: []slog.Handler{consoleHandler, fileHandler}}))
			log.SetOutput(io.MultiWriter(os.Stderr, f))
			log.Printf("日志文件: %s", *logFile)
		}
	}

	// 初始化插件
	ctx := &devPluginContext{logger: slog.Default()}
	if err := cfg.Plugin.Init(ctx); err != nil {
		return err
	}

	// 初始化账号存储
	store := NewAccountStore(filepath.Join(*dataDir, "accounts.json"))

	// 路由
	mux := http.NewServeMux()

	// 插件信息 API
	mux.HandleFunc("/api/plugin/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		info := cfg.Plugin.Info()
		json.NewEncoder(w).Encode(info)
	})

	// 账号管理 API
	accountHandler := &AccountHandler{store: store}
	mux.Handle("/api/accounts/", accountHandler)
	mux.Handle("/api/accounts", accountHandler)

	// 插件自定义路由
	if cfg.ExtraRoutes != nil {
		cfg.ExtraRoutes(mux, store)
	}

	// 从 plugin.Routes() 提取路径前缀注册代理
	proxy := &ProxyHandler{plugin: cfg.Plugin, store: store}
	prefixes := routePrefixes(cfg.Plugin.Routes())
	for _, prefix := range prefixes {
		mux.Handle(prefix, proxy)
	}

	// 插件前端资源（如果实现了 WebAssetsProvider）
	if wap, ok := cfg.Plugin.(sdk.WebAssetsProvider); ok {
		assets := wap.GetWebAssets()
		if len(assets) > 0 {
			mux.HandleFunc("/plugin-assets/", func(w http.ResponseWriter, r *http.Request) {
				name := strings.TrimPrefix(r.URL.Path, "/plugin-assets/")
				data, exists := assets[name]
				if !exists {
					http.NotFound(w, r)
					return
				}
				if strings.HasSuffix(name, ".js") {
					w.Header().Set("Content-Type", "application/javascript")
				} else if strings.HasSuffix(name, ".css") {
					w.Header().Set("Content-Type", "text/css")
				}
				w.Write(data)
			})
		}
	}

	// 内嵌管理页面
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return err
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	info := cfg.Plugin.Info()
	log.Printf("devserver 启动: http://localhost%s", *addr)
	log.Printf("插件: %s v%s", info.Name, info.Version)
	log.Printf("管理页面: http://localhost%s", *addr)
	return http.ListenAndServe(*addr, mux)
}

// routePrefixes 从路由声明中提取不重复的路径前缀（如 /v1/）
func routePrefixes(routes []sdk.RouteDefinition) []string {
	seen := make(map[string]bool)
	var prefixes []string
	for _, r := range routes {
		// 取第二个 / 之前的部分作为前缀
		parts := strings.SplitN(strings.TrimPrefix(r.Path, "/"), "/", 2)
		prefix := "/" + parts[0] + "/"
		if !seen[prefix] {
			seen[prefix] = true
			prefixes = append(prefixes, prefix)
		}
	}
	return prefixes
}
