/*
Package treetop is a library for incorporating HTML fragment requests into Go
web applications.

Fragments requests are specified by a protocol with the goal of removing some of the common
causes of JavaScript boilerplate in modern web applications. Reliance on data APIs can also be greatly reduced.

For documentation and examples see https://github.com/rur/treetop and https://github.com/rur/treetop-recipes

Introduction

In the spirit of opt-in integration this package supports two use cases.
The first is the handler scoped 'Writer' which is useful for supporting
ad-hoc fragments in an existing application. The view builder abstraction is
the second, it is designed for constructing a UI with many
cooperating endpoints.

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
		treetop.DefaultTemplateExec,
		"base.html.tmpl",
		baseHandler,
	)

	greeting := base.NewSubView(
		"greeting",
		"greeting.html.tmpl",
		greetingHandler,
	)

	mymux.Handle("/", treetop.ViewHandler(greeting))

See the documentation of the View type for details.

Browser History - Partials vs Fragments

This can cause confussion, but it is a very useful distinction. A partial is a
_part_ of an HTML document, a fragment is a general purpose HTML
snippet. Both have a URL, but only the former should be considered 'navigation'
by the user agent. This allows browser history to be handled correctly so that back,
forward and refresh behavior work as expected.

Note: The client relies upon the HTML 5 history API to support Treetop partials.

*/
package treetop
