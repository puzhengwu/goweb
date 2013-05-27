package goweb

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Context struct {
	Request *http.Request
	Params  map[string]string
	Server  *Server
	http.ResponseWriter
}

func (ctx *Context) WriteString(content string) {
	ctx.ResponseWriter.Write([]byte(content))
}

func (ctx *Context) Abort(status int, body string) {
	ctx.ResponseWriter.WriteHeader(status)
	ctx.ResponseWriter.Write([]byte(body))
}

func (ctx *Context) Redirect(status int, url_ string) {
	ctx.ResponseWriter.Header().Set("Location", url_)
	ctx.ResponseWriter.WriteHeader(status)
	ctx.ResponseWriter.Write([]byte("Redirecting to: " + url_))
}

func (ctx *Context) NotModified() {
	ctx.ResponseWriter.WriteHeader(304)
}

func (ctx *Context) NotFound(message string) {
	ctx.ResponseWriter.WriteHeader(404)
	ctx.ResponseWriter.Write([]byte(message))
}

func (ctx *Context) ToJson(o interface{}) {
	content, err := json.Marshal(o)
	if err != nil {
		ctx.Server.Logger.Println("json error")
		return
	}
	jsoncallback := ctx.Params["jsoncallback"]
	if jsoncallback == "" {
		ctx.SetHeader("Content-Length", strconv.Itoa(len(content)), true)
		ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
		ctx.ResponseWriter.Write(content)
	} else {
		callback := jsoncallback + "(" + string(content) + ")"
		ctx.SetHeader("Content-Length", strconv.Itoa(len(callback)), true)
		ctx.ResponseWriter.Header().Set("Content-Type", "application/javascript")
		ctx.ResponseWriter.Write([]byte(callback))
	}
}

func (ctx *Context) ToXml(o interface{}) {
	content, err := xml.Marshal(o)
	if err != nil {
		ctx.Server.Logger.Println("xml error")
		return
	}
	ctx.SetHeader("Content-Length", strconv.Itoa(len(content)), true)
	ctx.ResponseWriter.Header().Set("Content-Type", "application/xml")
	ctx.ResponseWriter.Write(content)
}

func (ctx *Context) ContentType(val string) string {
	var ctype string
	if strings.ContainsRune(val, '/') {
		ctype = val
	} else {
		if !strings.HasPrefix(val, ".") {
			val = "." + val
		}
		ctype = mime.TypeByExtension(val)
	}
	if ctype != "" {
		ctx.Header().Set("Content-Type", ctype)
	}
	return ctype
}

func (ctx *Context) SetHeader(hdr string, val string, unique bool) {
	if unique {
		ctx.Header().Set(hdr, val)
	} else {
		ctx.Header().Add(hdr, val)
	}
}

func (ctx *Context) setCookie(cookie *http.Cookie) {
	ctx.SetHeader("Set-Cookie", cookie.String(), false)
}

func getCookieSig(key string, val []byte, timestamp string) string {
	hm := hmac.New(sha1.New, []byte(key))

	hm.Write(val)
	hm.Write([]byte(timestamp))

	hex := fmt.Sprintf("%02x", hm.Sum(nil))
	return hex
}

func (ctx *Context) SetCookie(name string, val string, age int64) {
	ctx.setCookie(NewCookie(name, val, age, ctx.Server.Config.CookieDomain))
}

func (ctx *Context) SetSecureCookie(name string, val string, age int64) {
	//base64 encode the val
	if len(ctx.Server.Config.CookieSecret) == 0 {
		ctx.Server.Logger.Println("Secret Key for secure cookies has not been set. Please assign a cookie secret to web.Config.CookieSecret.")
		return
	}
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write([]byte(val))
	encoder.Close()
	vs := buf.String()
	vb := buf.Bytes()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	sig := getCookieSig(ctx.Server.Config.CookieSecret, vb, timestamp)
	cookie := strings.Join([]string{vs, timestamp, sig}, "|")
	ctx.setCookie(NewCookie(name, cookie, age, ctx.Server.Config.CookieDomain))
}

func (ctx *Context) GetSecureCookie(name string) (string, bool) {
	for _, cookie := range ctx.Request.Cookies() {
		if cookie.Name != name {
			continue
		}

		parts := strings.SplitN(cookie.Value, "|", 3)

		val := parts[0]
		timestamp := parts[1]
		sig := parts[2]

		if getCookieSig(ctx.Server.Config.CookieSecret, []byte(val), timestamp) != sig {
			return "", false
		}

		ts, _ := strconv.ParseInt(timestamp, 0, 64)

		if time.Now().Unix()-31*86400 > ts {
			return "", false
		}

		buf := bytes.NewBufferString(val)
		encoder := base64.NewDecoder(base64.StdEncoding, buf)

		res, _ := ioutil.ReadAll(encoder)
		return string(res), true
	}
	return "", false
}

var contextType reflect.Type
var defaultStaticDirs []string

func init() {
	contextType = reflect.TypeOf(Context{})

	wd, _ := os.Getwd()
	arg0 := path.Clean(os.Args[0])

	var exeFile string
	if strings.HasPrefix(arg0, "/") {
		exeFile = arg0
	} else {
		exeFile = path.Join(wd, arg0)
	}
	parent, _ := path.Split(exeFile)
	fmt.Println(parent, wd)
	defaultStaticDirs = append(defaultStaticDirs, path.Join(parent, "static"))
	defaultStaticDirs = append(defaultStaticDirs, path.Join(wd, "static"))
	return
}

func Process(c http.ResponseWriter, req *http.Request) {
	mainServer.Process(c, req)
}

func Run(addr string) {
	mainServer.Run(addr)
}

func RunTLS(addr string, config *tls.Config) {
	mainServer.RunTLS(addr, config)
}

func RunScgi(addr string) {
	mainServer.RunScgi(addr)
}

func RunFcgi(addr string) {
	mainServer.RunFcgi(addr)
}

func Close() {
	mainServer.Close()
}

func Get(route string, handler interface{}) {
	mainServer.Get(route, handler)
}

func Post(route string, handler interface{}) {
	mainServer.addRoute(route, "POST", handler)
}

func Put(route string, handler interface{}) {
	mainServer.addRoute(route, "PUT", handler)
}

func Delete(route string, handler interface{}) {
	mainServer.addRoute(route, "DELETE", handler)
}

func Match(method string, route string, handler interface{}) {
	mainServer.addRoute(route, method, handler)
}

func SetLogger(logger *log.Logger) {
	mainServer.Logger = logger
}

var mainServer = NewServer()
