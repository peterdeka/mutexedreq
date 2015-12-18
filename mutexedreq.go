package mutexedreq

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"net/http"
	"sync"
)

type lockedMap struct {
	*sync.Mutex
	locksByReqId map[interface{}]chan bool
}

var locksByCtxKey = map[interface{}]lockedMap{}
var mutex = sync.Mutex{}

//returns an handler that mutexes requests on a Gorilla context key ctxKey
func MutexedRequestMiddleware(ctxKey interface{}) negroni.HandlerFunc {
	//we exploit the closure so the handler knows its ctxKey
	return func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		reqId := context.Get(req, ctxKey)
		if reqId == nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		//create or get the lockedMap for this context key
		mutex.Lock()
		if _, ok := locksByCtxKey[ctxKey]; !ok {
			locksByCtxKey[ctxKey] = lockedMap{
				&sync.Mutex{},
				map[interface{}]chan bool{},
			}
		}
		mutex.Unlock()

		//create or get the channel to mutex the requests on
		locksByCtxKey[ctxKey].Lock()
		if _, ok := locksByCtxKey[ctxKey].locksByReqId[reqId]; !ok {
			locksByCtxKey[ctxKey].locksByReqId[reqId] = make(chan bool, 1)
			locksByCtxKey[ctxKey].locksByReqId[reqId] <- true
		}
		locksByCtxKey[ctxKey].Unlock()
		select {
		case <-locksByCtxKey[ctxKey].locksByReqId[reqId]:
			defer func() {
				locksByCtxKey[ctxKey].locksByReqId[reqId] <- true
				/*mutex.Lock()
				delete(locks, k)
				mutex.Unlock()*/
			}()
			next(rw, req)

		default:
			rw.WriteHeader(http.StatusConflict)
			return
		}
	}
}
