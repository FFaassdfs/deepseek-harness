package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// startLinkBridge 启动本地桥接服务器：harness 页面里的外链通过它跳转到系统浏览器。
// harness 页面（127.0.0.1:3080）没有 Wails 运行时，无法直接调用 BrowserOpenURL，
// 所以注入的脚本用 fetch 请求本桥接服务器，由 Go 侧打开系统浏览器。
func (a *App) startLinkBridge() {
	mux := http.NewServeMux()
	mux.HandleFunc("/open", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		url := r.URL.Query().Get("url")
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			runtime.BrowserOpenURL(a.ctx, url)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return
	}
	a.bridgePort = ln.Addr().(*net.TCPAddr).Port
	go func() {
		_ = http.Serve(ln, mux)
	}()
}

// startLinkInterceptor 定期向 harness 页面注入外链拦截脚本（幂等），
// 使外链跳转到系统浏览器，而不是在桌面壳内把页面跳走。
func (a *App) startLinkInterceptor(ctx context.Context) {
	go func() {
		time.Sleep(3 * time.Second) // 等待 harness 页面加载
		for {
			a.injectLinkInterceptor(ctx)
			select {
			case <-a.healthStop:
				return
			case <-time.After(30 * time.Second):
			}
		}
	}()
}

func (a *App) injectLinkInterceptor(ctx context.Context) {
	if a.bridgePort == 0 {
		return
	}
	runtime.WindowExecJS(ctx, fmt.Sprintf(linkInterceptorJS, a.bridgePort))
}

const linkInterceptorJS = `(function(){
  if (window.__dshLinkInterceptor) return;
  window.__dshLinkInterceptor = true;
  var bridge = 'http://127.0.0.1:%d/open?url=';
  function isExternal(href){
    return href.indexOf('http://') === 0 || href.indexOf('https://') === 0;
  }
  function isHarness(href){
    if (href.charAt(0) === '/' || href.charAt(0) === '#') return true;
    try {
      var h = new URL(href).hostname;
      return h === '127.0.0.1' || h === 'localhost';
    } catch (e) { return false; }
  }
  document.addEventListener('click', function(e){
    var a = e.target;
    while (a && a.tagName !== 'A') a = a.parentElement;
    if (!a) return;
    var href = a.getAttribute('href');
    if (!href || !isExternal(href) || isHarness(href)) return;
    if (a.target === '_blank') return;
    e.preventDefault();
    e.stopPropagation();
    fetch(bridge + encodeURIComponent(href), { mode: 'no-cors' }).catch(function(){});
  }, true);
})();`
