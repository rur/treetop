// Code generated by go generate; DO NOT EDIT.

// This file was created by github.com/rur/ttgen/cmd/ttroutes
// Map file: ./routemap.toml
// Template file: ./routes.go.templ

package ticket

import (
	"github.com/rur/treetop"
)

func Routes(m Mux, exec treetop.ViewExecutor) {

	baseView := treetop.NewView(
		"local://base.html",
		treetop.Delegate("content"),
	)

	// content
	ticketFormContent := baseView.NewDefaultSubView(
		"content",
		"examples/ticket/templates/content/ticketFormContent.html.tmpl",
		ticketHandler,
	)

	// content -> form
	newHelpdeskTicket := ticketFormContent.NewSubView(
		"form",
		"examples/ticket/templates/content/form/newHelpdeskTicket.html.tmpl",
		newHelpdeskTicketHandler,
	)

	// content -> form -> reported-by
	helpdeskReportedBy := newHelpdeskTicket.NewDefaultSubView(
		"reported-by",
		"examples/ticket/templates/content/form/reportedBy/helpdeskReportedBy.html.tmpl",
		helpdeskReportedByHandler,
	)

	// content -> form -> reported-by -> find-reported-by
	findHelpdeskReportedBy := helpdeskReportedBy.NewSubView(
		"find-reported-by",
		"examples/ticket/templates/content/form/reportedBy/findReportedBy/findHelpdeskReportedBy.html.tmpl",
		findTeamMemberHandler,
	)

	// content -> form -> upload-file-list
	uploadedHelpdeskFiles := newHelpdeskTicket.NewDefaultSubView(
		"upload-file-list",
		"examples/ticket/templates/content/form/uploadFileList/uploadedHelpdeskFiles.html.tmpl",
		uploadedHelpdeskFilesHandler,
	)
	newSoftwareTicket := ticketFormContent.NewSubView(
		"form",
		"examples/ticket/templates/content/form/newSoftwareTicket.html.tmpl",
		newSoftwareTicketHandler,
	)
	newSystemsTicket := ticketFormContent.NewSubView(
		"form",
		"examples/ticket/templates/content/form/newSystemsTicket.html.tmpl",
		newSystemsTicketHandler,
	)

	// nav
	_ = baseView.NewDefaultSubView(
		"nav",
		"local://nav.html",
		treetop.Noop,
	)

	m.HandleGET("/ticket/helpdesk/find-reported-by",
		exec.NewViewHandler(findHelpdeskReportedBy).FragmentOnly())
	m.HandleGET("/ticket/helpdesk/update-reported-by",
		exec.NewViewHandler(helpdeskReportedBy).FragmentOnly())
	m.HandlePOST("/ticket/helpdesk/upload-attachment",
		exec.NewViewHandler(uploadedHelpdeskFiles).FragmentOnly())
	m.HandleGET("/ticket/helpdesk/new",
		exec.NewViewHandler(newHelpdeskTicket))
	m.HandleGET("/ticket/software/new",
		exec.NewViewHandler(newSoftwareTicket))
	m.HandleGET("/ticket/systems/new",
		exec.NewViewHandler(newSystemsTicket))
	m.HandleGET("/ticket",
		exec.NewViewHandler(ticketFormContent))
}
