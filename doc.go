/*
Package treetop includes tools for incorporating HTML fragment requests into Go
web applications.

Fragments requests are specified by a protocol with the goal removing some of the common
causes of JavaScript boilerplate in modern web applications. Reliance on data APIs can also be greatly reduced.

For documentation and examples see https://github.com/rur/treetop and https://github.com/rur/treetop-recipes

Introduction

In the spirit of opt-in integration, this package supports two use cases.
Writers are handler scoped helpers that are useful for incorporating
ad-hoc fragments with an existing application. The view builder abstraction
is designed for constructing an integrated UI with many cooperating endpoints.

Example of ad-hoc partial writer

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
	  // otherwise render a full page containing an element with id="greeting"
	  ...
  }

For this example the full page document will include the standard Treetop client library, an element
with the "greeting" ID attribute value and an anchor element that has a 'treetop' attribute.

HTML document for the example,

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
			TREETOP_CONFIG = {}
		</script>
		<script src="/lib/treetop.js" async></script>
	</body>
	</html>

When the 'Fetch greeting' anchor is clicked, the client library will issue an XHR fetch
with the appropriate headers. Elements on the DOM will be replaced with fragments from
the response body based upon their ID attribute.


Views and Templates

More realistic web applications will take advantage of the HTML template support
in the Go standard library. Treetop includes a View builder to makes it more convenient
to define handlers that support partials and fragments.

See the documentation of the View type for details.

Browser History - Partials vs Fragments

This can cause confussion, but it is a very useful distinction. A partial is a
_part_ of an HTML document, a fragment is a general purpose HTML
snippet. Both have a URL, but only the former should be considered 'navigation'
by the user agent. This allows browser history to be handled correctly so that the back,
forward and refresh behavior work as expected.

You should note that the client relies upon the HTML 5 history API to support
Treetop partials.

*/
package treetop
