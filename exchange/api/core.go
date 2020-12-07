package api

import (
	"encoding/json"
	"fmt"
	"github.com/bCoder778/log"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Debug("Request:%s", r.RequestURI)
	reqPath := strings.TrimLeft(r.URL.Path, "/")
	var result interface{}
	var err *Error //errors.New(fmt.Sprintf("%s is not exist.", reqPath))
	var cont []byte
	var opt RouteOption
	var exist bool
	//var ok bool
	key := fmt.Sprintf("%s-%s", r.Method, reqPath)
	log.Debugf("Key:%s", key)
	if opt, exist = restApi.routerMap[key]; exist {
		log.Debugf("K:%s", key)
		ct := Context{request: r, response: w}
		ct.initQuery()
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			if e := ct.initForm(); e != nil {
				err = &Error{ERROR_FORM_INIT_FAILED, e.Error()}
			}
		}
		if err == nil {
			result, err = opt.Handle(&ct)
		}
	} else {
		result, err = handleFile(reqPath, w)
		return
	}

	contentType := r.Header.Get("Content-Type")
	w.Header().Set("Content-Type", contentType)
	if opt.Special {
		_, _ = w.Write([]byte(result.(string)))
		return
	}

	if err != nil {
		log.Warn(err.Message)
		rs, e := json.Marshal(*err.DealError())
		if e != nil {
			log.Error(e.Error())
		}
		cont = rs
		_, _ = w.Write(cont)
		return
	}

	log.Debugf("content-type:%s", contentType)
	rs, e := json.Marshal(Result{Code: 0, Message: "ok", Result: result})
	if e != nil {
		log.Error(e.Error())
	}
	cont = rs

	if _, e := w.Write(cont); e != nil {
		log.Error(e.Error())
	}
}

func handleFile(path string, w http.ResponseWriter) (interface{}, *Error) {
	if strings.HasSuffix(path, ".txt") ||
		strings.HasSuffix(path, ".html") ||
		strings.HasSuffix(path, ".dat") ||
		strings.HasSuffix(path, ".csv") {
		log.Debugf("Path:%s", path)

		f, e := os.Open(path)
		if e != nil {
			return nil, &Error{ERROR_REQUEST_NODFOUND, "Not found"}
		}
		defer f.Close()

		r, e := ioutil.ReadAll(f)
		if e != nil {
			return nil, &Error{ERROR_REQUEST_NODFOUND, e.Error()}
		}
		_, err := w.Write(r)
		if err != nil {
			return nil, &Error{ERROR_UNKNOWN, e.Error()}
		}
		return nil, nil
	}
	return nil, &Error{ERROR_REQUEST_NODFOUND, "Not found"}
}
