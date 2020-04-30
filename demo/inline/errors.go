package inline

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/rur/treetop"
)

// pageErrorMessage will show an error message
func pageErrorMessage(w http.ResponseWriter, req *http.Request, msg string, status int) {
	if ttW, ok := treetop.NewFragmentWriter(w, req); ok {
		if status >= 200 {
			ttW.Status(status)
		}
		fmt.Fprint(ttW,
			`<div id="error-message" class="alert alert-danger fade show" role="alert">
			<strong>Error!</strong> `+
				template.HTMLEscapeString(msg)+
				`
			<button treetop-link="/inline" type="button" class="close" data-dismiss="alert" aria-label="Close">
				<span aria-hidden="true">&times;</span>
			</button>
			</div>

			</div>`)
		return
	}
	w.WriteHeader(status)
	fmt.Fprint(w, `<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Error</title>
	</head>
	<body>
		<h1>Example Error</h1>
		<div class="alert alert-danger" role="alert">
		`, template.HTMLEscapeString(msg), `
		</div>
	</body>
	</html>`)
}
