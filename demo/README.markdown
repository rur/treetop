# Demos

A demo space for mini web apps that use the [Treetop library](https://github.com/rur/treetop) with HTML template
requests.

#### Links

* [Online Demo](https://treetop-demo.herokuapp.com/) - hosted by [Heroku](https://www.heroku.com/)
* Client library for template requests, [treetop client](https://github.com/rur/treetop-client)
* Styling for components, [Bootstrap 4](https://getbootstrap.com/docs/4.0)

## Greeter App

Given a name, the server responds with a greeting message which is routed to the "message" div on the page.
Not a real scenario, but it's nice to be nice!

It demonstrates: XHR form submit, browser history control and multi-fragment template response.

Review source: [greeter/greeter.go](greeter/greeter.go)

#### Template Hierarchy

Full page view hierarchy for the `"/greeter/greet"` path.

    - View("local://base.html", github.com/rur/treetop.Delegate.func1)
    |- content: SubView("content", "demo/greeter/templates/content.html", github.com/rur/treet……emo/greeter.contentViewHandler)
    |  |- message: SubView("message", "demo/greeter/templates/greeting.html", github.com/rur/treet……mo/greeter.greetingViewHandler)
    |  '- notes: SubView("notes", "demo/greeter/templates/notes.html", github.com/rur/treetop/demo/greeter.notesHandler)
    |
    '- nav: SubView("nav", "local://nav.html", github.com/rur/treetop.Noop)


## Inline Edit App

Inline editing is commonly used when many different elements can be modified independently.
A user profile is the scenario chosen for the app. The goal is to eliminate the
need for client code that 'knows what's going on'.

Review source: [inline/setup.go](inline/setup.go)

#### Template Hierarchy

Full page view hierarchy for user profile page

    - View("local://base.html", github.com/rur/treetop.Delegate.func1)
    |- content: SubView("content", "demo/inline/templates/content.html.tmpl", github.com/rur/treet……o/inline.profileContentHandler)
    |  |- country: SubView("country", "demo/inline/templates/select.html.tmpl", github.com/rur/treet……ine.(*cookieServer).bind.func1)
    |  |- description: SubView("description", "demo/inline/templates/textarea.html.tmpl", github.com/rur/treet……ine.(*cookieServer).bind.func1)
    |  |- email: SubView("email", "demo/inline/templates/email.html.tmpl", github.com/rur/treet……ine.(*cookieServer).bind.func1)
    |  |- first-name: SubView("first-name", "demo/inline/templates/input.html.tmpl", github.com/rur/treet……ine.(*cookieServer).bind.func1)
    |  '- surname: SubView("surname", "demo/inline/templates/input.html.tmpl", github.com/rur/treet……ine.(*cookieServer).bind.func1)
    |
    '- nav: SubView("nav", "local://nav.html", github.com/rur/treetop.Noop)


## Ticket Wizard App

A multi-stage workflow with branch points and conditions based upon user input.
The forms includes several input components that require server IO:
* temporarily store files,
* search the backend for input values,
* show new options conditioned on a previous one,
* flash messages, and
* redirects.

This app makes use of a 'route map' tool* to generate the router setup code.

Review source: [ticket/routemap.toml](ticket/routemap.toml), is used to generate, [ticket/routes.go](ticket/routes.go)

#### Template Hierarchy

The full page hierarchy for the `"/ticket/helpdesk/new"` endpoint.

    - View("local://base.html", github.com/rur/treetop.Delegate.func1)
    |- content: SubView("content", "demo/ticket/template……nt/ticketFormContent.html.tmpl", github.com/rur/treet……/ticket/handlers.TicketHandler)
    |  '- form: SubView("form", "demo/ticket/template……rm/newHelpdeskTicket.html.tmpl", github.com/rur/treet……dlers.NewHelpdeskTicketHandler)
    |     |- attachment-list: SubView("attachment-list", "demo/ticket/template……ntList/uploadedFiles.html.tmpl", github.com/rur/treet……lers.AttachmentFileListHandler)
    |     |- form-message: nil
    |     |- notes: SubView("notes", "demo/ticket/templates/content/notes.html.tmpl", github.com/rur/treetop.Noop)
    |     '- reported-by: SubView("reported-by", "demo/ticket/template……y/helpdeskReportedBy.html.tmpl", github.com/rur/treet……lers.HelpdeskReportedByHandler)
    |        '- find-user: SubView("find-user", "demo/ticket/template……y/findReportedByUser.html.tmpl", github.com/rur/treet……lers.FindReportedByUserHandler)
    |
    '- nav: SubView("nav", "local://nav.html", github.com/rur/treetop.Noop)

_* [ttroutes](https://github.com/rur/ttgen) command line tool is a prototype for generating routing code for nested templates and Treetop view handlers_