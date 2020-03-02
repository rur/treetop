/*
Package treetop provides tools for building HTML template handlers.

So your webpage does IO; and you need to show updates without clobbering the interface.
The common approach is to expose a data API and dispatch JavaScript to micromanage the client.
That will work, but it is pretty heavy duty for what seems like a simple problem.

Conventional HTTP works very well for navigation and web forms alike, no micromanagement required.
Perhaps it could be extended to solve our dynamic update problem. That is the starting point for Treetop,
to see how far we can get with a simple protocol.

Treetop is unique because it puts the server-side hander in complete control of how the page will be updated
following a request.

For documentation and examples see https://github.com/rur/treetop and https://github.com/rur/treetop-recipes

Introduction

In the spirit of opt-in integration this package supports two use cases.
The first is the handler scoped 'Writer' which is useful for supporting
ad-hoc fragments in an existing application. The view builder abstraction is
the second, it is designed for constructing a UI with many
cooperating endpoints.

Example of ad-hoc template writer

  import (
	  "fmt"
	  "github.com/rur/treetop"
	  "net/http"
  )
  ...
  func MyHandlerFunc(w http.ResponseWriter, req *http.Request) {
	  if pw, ok := treetop.NewPartialWriter(w, req); ok {
		  fmt.Fprint(pw, `<p id="greeting">Hello Treetop!</p>`)
		  return
	  }
	  // otherwise render a full page
	  ...
  }

For the example, the 'full page' document will include: the Treetop client library, an element
with an ID of "greeting" and an anchor element with a 'treetop' attribute.

Example HTML document

	<!DOCTYPE html>
	<html>
	<head>
		<title>My Page</title>
	</head>
	<body>

		<h1>Message</h1>

		<p id="greeting">Hello Page!</p>

		<div><a treetop href=".">Fetch greeting</a></div>

		<script>
		// use default config, global variable will signal treetop client to initialize
		TREETOP_CONFIG = {}
		</script>
		<script src="/lib/treetop.js" async></script>
	</body>
	</html>

When the 'Fetch greeting' anchor is clicked, the client library will issue an XHR fetch
with the appropriate headers. Elements on the DOM will then be replaced with fragments from
the response body based upon their ID attribute.

Views and Templates

Many Go web applications take advantage of the HTML template support
in the Go standard library. Package treetop includes a view builder
that works with the template system to support pages, partials and fragments.

Example

	base := treetop.NewView(
		"base.html.tmpl",
		baseHandler,
	)

	greeting := base.NewSubView(
		"greeting",
		"greeting.html.tmpl",
		greetingHandler,
	)

	exec := treetop.FileExecutor{}
	mymux.Handle("/", exec.NewViewHandler(greeting))

See the documentation of the View type for details.

Browser History - PartialWriter vs FragmentWriter

This can cause confussion, but it is a very useful distinction. A partial is a
_part_ of an HTML document, a fragment is a general purpose HTML
snippet. Both have a URL, but only the former should be considered 'navigation'
by the user agent. This allows browser history to be handled correctly so that back,
forward and refresh behavior work as expected.

The ViewHandler interface has the FragmentOnly and PageOnly qualifiers for this purpose.

Note: The client relies upon the HTML 5 history API to support Treetop partials.

*/
package treetop
