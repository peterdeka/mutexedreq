# mutexedreq
A simple middleware to mutex requests based on an arbitrary gorilla/context key.
##Why mutex?
If you want critical/atomic requests handling for some resources, for example db payment objects or critical userinfo editing that rely on external services responses (Stripe in our case) and you have a middleware that can provide a unique `reqId` bound to that resource, just place a *mutexedreq* handler in the request serving stack.

`ctxKey` is the name of the gorilla/context key that will be used to retrieve the key used to mutex the requests and must be defined at  middleware creation.
For any value set by a preceding middleware to `ctxKey`, only one request will be accepted at a time.
