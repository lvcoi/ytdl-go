package downloader

import (
	"embed"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dop251/goja"
)

//go:embed assets/bgutils.js
var assets embed.FS

type BgUtils struct {
	vm *goja.Runtime
}

func NewBgUtils() (*BgUtils, error) {
	script, err := assets.ReadFile("assets/bgutils.js")
	if err != nil {
		return nil, fmt.Errorf("failed to read bgutils.js: %w", err)
	}

	vm := goja.New()
	
	module := vm.NewObject()
	exports := vm.NewObject()
	module.Set("exports", exports)
	vm.Set("module", module)
	vm.Set("exports", exports)

	vm.Set("global", vm.GlobalObject())
	vm.Set("window", vm.GlobalObject())

	vm.Set("TextEncoder", func(call goja.ConstructorCall) *goja.Object {
		obj := vm.NewObject()
		obj.Set("encode", func(s string) []byte {
			return []byte(s)
		})
		return obj
	})

	vm.Set("btoa", func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	})
	vm.Set("atob", func(s string) (string, error) {
		b, err := base64.StdEncoding.DecodeString(s)
		return string(b), err
	})

	vm.Set("fetch", func(url string, options map[string]interface{}) *goja.Object {
		promise, resolve, reject := vm.NewPromise()
		go func() {
			method := "GET"
			if m, ok := options["method"].(string); ok { method = m }
			var body io.Reader
			if b, ok := options["body"].(string); ok { body = strings.NewReader(b) }
			req, err := http.NewRequest(method, url, body)
			if err != nil { reject(vm.ToValue(err.Error())); return }
			if h, ok := options["headers"].(map[string]interface{}); ok {
				for k, v := range h { req.Header.Set(k, fmt.Sprint(v)) }
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil { reject(vm.ToValue(err.Error())); return }
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)
			responseObj := vm.NewObject()
			responseObj.Set("status", resp.StatusCode)
			responseObj.Set("json", func() *goja.Object {
				jPromise, jResolve, _ := vm.NewPromise()
				jResolve(vm.ToValue(string(respBody)))
				return vm.ToValue(jPromise).ToObject(vm)
			})
			resolve(responseObj)
		}()
		return vm.ToValue(promise).ToObject(vm)
	})

	_, err = vm.RunString(string(script))
	if err != nil { return nil, err }
	return &BgUtils{vm: vm}, nil
}

func (b *BgUtils) getBG() *goja.Object {
	return b.vm.Get("module").ToObject(b.vm).Get("exports").ToObject(b.vm).Get("BG").ToObject(b.vm)
}

func (b *BgUtils) GeneratePlaceholder(identifier string) (string, error) {
	bg := b.getBG()
	fn, ok := goja.AssertFunction(bg.Get("PoToken").ToObject(b.vm).Get("generatePlaceholder"))
	if !ok { return "", fmt.Errorf("no generatePlaceholder") }
	result, err := fn(goja.Undefined(), b.vm.ToValue(identifier))
	if err != nil { return "", err }
	return result.String(), nil
}
