package main

import (
	
	"log"
	
	"golang.org/x/net/websocket"
	"os"
	"io"
	"strings"
	"net"
	"net/http"
	"net/url"
	"net/http/httputil"
	
)

func main() {

	log.Printf("main\n")

	
/*
	proxy := os.Getenv("HTTP_PROXY")
	proxyUrl, err := url.Parse(proxy)

	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}

	resp, err := myClient.Get("http://13.124.64.10/ws")
	if err != nil {
		log.Fatal("Get:", err)
	}
	log.Printf("resp{%s}\n", resp)
*/

	//ProxyDial("ws://13.124.64.10/ws", "", "ws://13.124.64.10/ws")
	//ProxyDial("ws://13.124.64.10/ws", "tcp", "http://13.124.64.10")
	//ProxyDial("ws://echo.websocket.org", "tcp", "ws://echo.websocket.org")

	// no-proxy localhost
	//ws, err := ProxyDial("ws://localhost:4000", "tcp", "ws://localhost:4000")
	// proxy 10.113.76.39
	//ws, err := ProxyDial("ws://10.113.76.39:4000", "tcp", "ws://10.113.76.39:4000")
	// proxy 13.124.64.10
	ws, err := ProxyDial("ws://13.124.64.10/ws", "tcp", "ws://13.124.64.10/ws")

	if err != nil {
		log.Fatal("ProxyDial : ", err)
	}
	log.Printf("ws", ws)

	defer ws.Close()

	for {
		var message string
		if err := websocket.Message.Receive(ws, &message); err != nil {
			log.Fatal(err)
		}
		log.Printf("message: %s", message)
	}
}


func ProxyDial(url_, protocol, origin string) (ws *websocket.Conn, err error) {

	log.Printf("http_proxy {%s}\n", os.Getenv("HTTP_PROXY"))

	// comment out in case of testing without proxy
	if strings.Contains(url_, "localhost") {
		return websocket.Dial(url_, protocol, origin)
	}

	if os.Getenv("HTTP_PROXY") == "" {
		return websocket.Dial(url_, protocol, origin)
	}

	purl, err := url.Parse(os.Getenv("HTTP_PROXY"))
	if err != nil {
		log.Fatal("Parse : ", err)
		return nil, err
	}

	log.Printf("====================================")
	log.Printf("    websocket.NewConfig")
	log.Printf("====================================")
	config, err := websocket.NewConfig(url_, origin)
	if err != nil {
		log.Fatal("NewConfig : ", err)
		return nil, err
	}

	if protocol != "" {
		config.Protocol = []string{protocol}
	}

	log.Printf("====================================")
	log.Printf("    HttpConnect")
	log.Printf("====================================")
	client, err := HttpConnect(purl.Host, url_)
	if err != nil {
		log.Fatal("HttpConnect : ", err)
		return nil, err
	}

	log.Printf("====================================")
	log.Printf("    websocket.NewClient")
	log.Printf("====================================")
	return websocket.NewClient(config, client)
}


func HttpConnect(proxy, url_ string) (io.ReadWriteCloser, error) {
	log.Printf("proxy =", proxy)
	proxy_tcp_conn, err := net.Dial("tcp", proxy)
	if err != nil {
		return nil, err
	}
	log.Printf("proxy_tcp_conn =", proxy_tcp_conn)
	log.Printf("url_ =", url_)

	turl, err := url.Parse(url_)
	if err != nil {
		log.Fatal("Parse : ", err)
		return nil, err
	}
	
	log.Printf("proxy turl.Host =", string(turl.Host))


	req := http.Request{
		Method: "CONNECT",
		URL:    &url.URL{},
		Host:   turl.Host,
	}

	/*
	// origin
	req := http.Request{
		Method: "CONNECT",
		URL:    &url.URL{},
		Host:   turl.Host,
	}
	*/

	proxy_http_conn := httputil.NewProxyClientConn(proxy_tcp_conn, nil)
	//cc := http.NewClientConn(proxy_tcp_conn, nil)

	log.Printf("proxy_http_conn =", proxy_http_conn)	

	resp, err := proxy_http_conn.Do(&req)
	if err != nil && err != httputil.ErrPersistEOF {
		log.Fatal("ErrPersistEOF : ", err)
		return nil, err
	}
	log.Printf("proxy_http_conn<resp> =", (resp))

	rwc, _ := proxy_http_conn.Hijack()

	return rwc, nil
}


// return Handler (A Handler reponds to an HTTP request)
func websocketProxy(target string) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                d, err := net.Dial("tcp", target)
                if err != nil {
                        http.Error(w, "Error contacting backend server.", 500)
                        log.Printf("Error dialing websocket backend %s: %v", target, err)
                        return
                }
                hj, ok := w.(http.Hijacker)
                if !ok {
                        http.Error(w, "Not a hijacker?", 500)
                        return
                }
                nc, _, err := hj.Hijack()
                if err != nil {
                        log.Printf("Hijack error: %v", err)
                        return
                }
                defer nc.Close()
                defer d.Close()

                err = r.Write(d)
                if err != nil {
                        log.Printf("Error copying request to target: %v", err)
                        return
                }

                errc := make(chan error, 2)
                cp := func(dst io.Writer, src io.Reader) {
                        _, err := io.Copy(dst, src)
                        errc <- err
                }
                go cp(d, nc)
                go cp(nc, d)
                <-errc
        })
    }

