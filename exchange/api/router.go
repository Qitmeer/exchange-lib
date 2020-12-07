package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bCoder778/log"
	"net/http"
	"strings"
)

var restApi *RestApi

type RestValues map[string]string
type Context struct {
	request  *http.Request
	response http.ResponseWriter
	Query    map[string]string
	Form     RestValues
	Body     []byte
}

type RestApi struct {
	Address   string
	routerMap map[string]RouteOption
	Auth      bool
	Serv      *http.Server
}
type RestSet struct {
	Path   string
	Prefix string
	Auth   bool
}

type RouteOption struct {
	Method  string
	Auth    bool
	Special bool
	Handler func(ct *Context) (interface{}, *Error)
}

type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Result  interface{} `json:"rs"`
}

type Token struct {
	Token string `json:"token"`
}

func NewRestApi(addr string) *RestApi {
	if restApi == nil {
		restApi = &RestApi{Address: addr, routerMap: map[string]RouteOption{}, Auth: false}
	}
	return restApi
}

func (rest *RestApi) SetAuth(auth bool) {
	rest.Auth = auth
}

func (rest *RestApi) AuthRouteSet(path string) *RestSet {
	res := RestSet{Path: path, Auth: rest.Auth}
	return &res
}

func (rest *RestApi) RouteSet(path string) *RestSet {
	res := RestSet{Path: path, Auth: false}
	return &res
}

func (rest *RestApi) Start() error {
	http.HandleFunc("/", Handle)
	rest.Serv = &http.Server{Addr: rest.Address}
	log.Debugf("Listen:%s", rest.Address)
	for k := range restApi.routerMap {
		log.Debug(k)
	}
	if err := rest.Serv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (rest *RestApi) Stop() {
	if rest.Serv != nil {
		rest.Serv.Shutdown(context.Background())
		log.Debug("Shutdown")
	}
}

func (r *RestSet) GetSub(name string, handler func(ct *Context) (interface{}, *Error)) *RestSet {
	r.addToRouterMap("GET", name, handler)
	return r
}

func (r *RestSet) GetSpecialSub(name string, handler func(ct *Context) (interface{}, *Error)) *RestSet {
	r.addToSpecialRouterMap("GET", name, handler)
	return r
}

func (r *RestSet) PostSub(name string, handler func(ct *Context) (interface{}, *Error)) *RestSet {
	r.addToRouterMap("POST", name, handler)
	return r
}
func (r *RestSet) Get(handler func(ct *Context) (interface{}, *Error)) *RestSet {
	r.addToRouterMap("GET", "", handler)
	return r
}

func (r *RestSet) Post(handler func(ct *Context) (interface{}, *Error)) *RestSet {
	r.addToRouterMap("POST", "", handler)
	return r
}

func (r *RestSet) addToRouterMap(method string, name string, handler func(ct *Context) (interface{}, *Error)) {
	key := r.getRouterKey(method, name)
	if _, exist := restApi.routerMap[key]; !exist {
		restApi.routerMap[key] = RouteOption{Method: method, Handler: handler, Auth: r.Auth}
	} else {
		log.Errorf("the name %s has existed", name)
	}
}

func (r *RestSet) addToSpecialRouterMap(method string, name string, handler func(ct *Context) (interface{}, *Error)) {
	key := r.getRouterKey(method, name)
	if _, exist := restApi.routerMap[key]; !exist {
		restApi.routerMap[key] = RouteOption{Method: method, Handler: handler, Special: true, Auth: r.Auth}
	} else {
		log.Errorf("the name %s has existed", name)
	}
}

func (r *RestSet) getRouterKey(method string, name string) string {
	if len(name) > 0 && !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	return fmt.Sprintf("%s-%s", method, r.Path+name)
}
func (ct *Context) initQuery() {
	if ct.Query == nil {
		ct.Query = map[string]string{}
	}
	for k, it := range ct.request.URL.Query() {
		ct.Query[k] = it[0]
	}
}
func (ct *Context) initForm() error {
	length := ct.request.ContentLength
	if length < 1 {
		return nil
	}
	var countSum = 0
	var body []byte
	for {
		readOne := make([]byte, 1024*2)
		n, err := ct.request.Body.Read(readOne)
		if err != nil {
			body = append(body, readOne[0:n]...)
			break
		}
		body = append(body, readOne[0:n]...)
		countSum += n
		if int64(countSum) >= length {
			break
		}
	}

	ct.Body = body
	if ct.Form == nil {
		ct.Form = RestValues{}
	}
	ctType := ct.request.Header.Get("Content-Type")
	switch ctType {
	case "application/json":
		if err := json.Unmarshal(body, &ct.Form); err != nil {
			return errors.New(fmt.Sprintf("Json parsing failed.%s", err.Error()))
		}
		break
	case "application/x-www-form-urlencoded":
		_ = ct.request.ParseForm()
		for k, it := range ct.request.Form {
			ct.Form[k] = it[0]
		}
		break
	}
	return nil
}

func (rp *RouteOption) Handle(ct *Context) (interface{}, *Error) {
	if rp.Auth {
		authorizations := strings.Split(ct.request.Header.Get("Authorization"), " ")
		if len(authorizations) > 1 {
			if !validateToken(authorizations[1]) {
				ct.response.WriteHeader(ERROR_REQUEST_UNAUTHPRIZED)
				return nil, &Error{ERROR_REQUEST_UNAUTHPRIZED, "Without authorization"}
			}
		} else {
			ct.response.WriteHeader(ERROR_REQUEST_UNAUTHPRIZED)
			return nil, &Error{ERROR_REQUEST_UNAUTHPRIZED, "Without authorization"}
		}
	}
	return rp.Handler(ct)
}

func validateToken(string) bool {
	return true
}
