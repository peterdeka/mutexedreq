package mutexedreq

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

var authKey = "authmutex"

func init() {

}

//tests that two requests with different reqIds can concur without mutexing
func TestDifferentReqId(t *testing.T) {
	require := require.New(t)
	router := mux.NewRouter()
	//get a new server
	router.Handle("/test",
		negroni.New(
			negroni.HandlerFunc(authReq("authmutex")),
			negroni.HandlerFunc(MutexedRequestMiddleware(authKey)),
			negroni.HandlerFunc(slowController),
		)).Methods("GET")
	n := negroni.Classic()
	n.UseHandler(router)
	r1, _ := http.NewRequest("GET", "/test", nil)
	AuthRequest(r1, "A")
	r2, _ := http.NewRequest("GET", "/test", nil)
	AuthRequest(r2, "B")
	res1 := httptest.NewRecorder()
	res2 := httptest.NewRecorder()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go serve(n, res1, r1, &wg)
	go serve(n, res2, r2, &wg)
	wg.Wait()
	require.Equal(res1.Code, http.StatusOK)
	require.Equal(res2.Code, http.StatusOK)
}

//tests that two requests with the same reqId can not concur
func TestRaceReqId(t *testing.T) {
	require := require.New(t)
	router := mux.NewRouter()
	//get a new server
	router.Handle("/test",
		negroni.New(
			negroni.HandlerFunc(authReq("authmutex")),
			negroni.HandlerFunc(MutexedRequestMiddleware("authmutex")),
			negroni.HandlerFunc(slowController),
		)).Methods("GET")
	n := negroni.Classic()
	n.UseHandler(router)
	r1, _ := http.NewRequest("GET", "/test", nil)
	AuthRequest(r1, "A")
	r2, _ := http.NewRequest("GET", "/test", nil)
	AuthRequest(r2, "A")
	res1 := httptest.NewRecorder()
	res2 := httptest.NewRecorder()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go serve(n, res1, r1, &wg)
	go serve(n, res2, r2, &wg)
	wg.Wait()
	res2Failed := res1.Code == http.StatusOK && res2.Code == http.StatusConflict
	res1Failed := res2.Code == http.StatusOK && res1.Code == http.StatusConflict
	require.Equal(true, res2Failed || res1Failed)
}

//tests that two requests with different ctxKey can concur without mutexing
func TestDifferentCtxKey(t *testing.T) {
	require := require.New(t)
	router := mux.NewRouter()
	//get a new server
	router.Handle("/testone",
		negroni.New(
			negroni.HandlerFunc(authReq("ctx_one")),
			negroni.HandlerFunc(MutexedRequestMiddleware("ctx_one")),
			negroni.HandlerFunc(slowController),
		)).Methods("GET")
	n := negroni.Classic()
	n.UseHandler(router)
	//
	router.Handle("/testtwo",
		negroni.New(
			negroni.HandlerFunc(authReq("ctx_two")),
			negroni.HandlerFunc(MutexedRequestMiddleware("ctx_two")),
			negroni.HandlerFunc(slowController),
		)).Methods("GET")
	n.UseHandler(router)
	///
	r1, _ := http.NewRequest("GET", "/testone", nil)
	AuthRequest(r1, "A")
	r2, _ := http.NewRequest("GET", "/testtwo", nil)
	AuthRequest(r2, "B")
	res1 := httptest.NewRecorder()
	res2 := httptest.NewRecorder()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go serve(n, res1, r1, &wg)
	go serve(n, res2, r2, &wg)
	wg.Wait()
	require.Equal(res1.Code, http.StatusOK)
	require.Equal(res2.Code, http.StatusOK)
}

//utility middlewares

func authReq(ctxKey string) negroni.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		context.Set(req, ctxKey, req.Header.Get("X-Auth-Test"))
		next(rw, req)
	}
}
func slowController(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	time.Sleep(1 * time.Second)
	//TODO return reqId for verification
}

func AuthRequest(req *http.Request, id string) {
	req.Header.Set("X-Auth-Test", id)
}

//utilty goroutine to serve concurrently
func serve(n *negroni.Negroni, rw http.ResponseWriter, r *http.Request, wg *sync.WaitGroup) {
	n.ServeHTTP(rw, r)
	wg.Done()
}
