# mutexedreq [![Coverage Status](https://coveralls.io/repos/peterdeka/mutexedreq/badge.svg?branch=master&service=github)](https://coveralls.io/github/peterdeka/mutexedreq?branch=master)
A simple middleware to mutex requests based on an arbitrary gorilla/context key.
##Why mutexedreq?
If you want critical/atomic requests handling for some resources, for example db payment objects or critical userinfo editing that rely on external services responses (Stripe in our case) and you have a middleware that can provide a unique `reqId` bound to that resource, just place a *mutexedreq* handler in the request serving stack.

`ctxKey` is the name of the gorilla/context key that will be used to retrieve the key used to mutex the requests and must be defined at  middleware creation.
For any value set by a preceding middleware to `ctxKey`, only one request will be accepted at a time.

##Example
This example is taken from the tests, simply pass the gorilla/context key that you want to mutex on:
```
router.Handle("/testone",
		negroni.New(
			negroni.HandlerFunc(authReq("ctx_one")),
			negroni.HandlerFunc(MutexedRequestMiddleware("ctx_one")),
			negroni.HandlerFunc(slowController),
		)).Methods("GET")
	n := negroni.Classic()
```
In this case authReq is a middleware that injects the user authentication id in the request context key `ctx_one`. When a request arrives, MutexedRequest will ensure that only one request per user id will be served at a time, returning a `409 Conflict` to all the others.

##For the future
Implement an alternative mutex provider based on Redis or another key/value, in order to be able to scale the mutexing across multiple services.
