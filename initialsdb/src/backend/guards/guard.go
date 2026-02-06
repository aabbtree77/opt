package guards

import "net/http"

/*
Guards return true when request processing should continue.

Guards MAY READ:

- headers
- URL
- remote address
- already-parsed form values

Guards MUST NOT:

- ever touch r.Body, r.Form as this will drain the request stream
- call io.ReadAll, ParseForm or FormValue (same problem)
- write responses
- redirect

These should ALWAYS be empty:

grep -R "r.Body" backend/
grep -R "ReadAll" backend/
grep -R "FormValue" backend/guards

Guards MAY:

- attach values to r.Context()

*/

type Guard interface {
	Check(r *http.Request) bool
}
