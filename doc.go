/*
Package treetop implements tools for constructing HTTP handlers for nested templates

To read about nested template support in Go see https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions

Multi-page web apps require a lot of endpoints. Template inheritance
is commonly used to reduce HTML boilerplate and improve reuse. Treetop views incorporate
request handlers into the hierarchy to gain the same advantage.

A 'View' is a template string (usually file path) paired with a handler function.
Go templates can contain named nested blocks. Defining a 'SubView' associates
a handler and a template with a block embedded within a parent template.
HTTP handlers can then be constructed for various page configurations.

Example of a basic template hierarchy

                 baseHandler(...)
               | base.html ========================|
               | …                                 |
               | {{ template "content" .Content }} |
               | …               ^                 |
               |_________________|_________________|
                                 |
                          ______/ \______
     contentAHandler(...)               contentBHandler(...)
   | contentA.html ========== |        | contentB.html ========== |
   |                          |        |                          |
   | {{ block "content" . }}… |        | {{ block "content" . }}… |
   |__________________________|        |__________________________|

Example of using the library to constructs handlers for HTTP routes.

	base := treetop.NewView(
		"base.html",
		baseHandler,
	)

	contentA := base.NewSubView(
		"content",
		"contentA.html",
		contentAHandler,
	)

	contentB := base.NewSubView(
		"content",
		"contentB.html",
		contentBHandler,
	)

	exec := treetop.FileExecutor{}
	mymux.Handle("/path/to/a", exec.ViewHandler(contentA))
	mymux.Handle("/path/to/b", exec.ViewHandler(contentB))

The generated handlers bind togeather related views. Thus views can be mixed and matched
to create many endpoints.

    GET /path/to/a
    > HTTP/1.1 200 OK
    > <!-- base.html --><html>
    > ...
    > <!-- contentA.html --><div id="content"> Content A </div>
    > ...
    > </html>

    GET /path/to/b
    > HTTP/1.1 200 OK
    > <!-- base.html --><html>
    > ...
    > <!-- contentB.html --><div id="content"> Content B </div>
    > ...
    > </html>

Note, many levels of nesting are possible once block names remain unique.

HTML Template Protocol

The constructed handlers are capable of rendering just sections of the page depending
upon the request headers. See the Treetop JS library for more details. (https://github.com/rur/treetop-client)

*/
package treetop
