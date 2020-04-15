// Code generated by go generate; DO NOT EDIT.

// This file was created by github.com/rur/ttgen/cmd/ttroutes
// Map file: ./routemap.toml
// Template file: ./routes.go.templ

package ticket

import (
	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/ticket/handlers"
)

func Routes(m Mux, exec treetop.ViewExecutor) {

	baseView := treetop.NewView(
		"local://base.html",
		treetop.Delegate("content"),
	)

	// content
	issuePreview := baseView.NewSubView(
		"content",
		"examples/ticket/templates/content/issuePreview.html.tmpl",
		handlers.IssuePreviewHandler,
	)

	// content -> notes
	_ = issuePreview.NewDefaultSubView(
		"notes",
		"examples/ticket/templates/content/notes.html.tmpl",
		treetop.Noop,
	)

	// content -> preview
	previewHelpdeskTicket := issuePreview.NewSubView(
		"preview",
		"examples/ticket/templates/content/previewHelpdeskTicket.html.tmpl",
		handlers.PreviewHelpdeskTicketHandler,
	)

	// content
	ticketFormContent := baseView.NewDefaultSubView(
		"content",
		"examples/ticket/templates/content/ticketFormContent.html.tmpl",
		handlers.TicketHandler,
	)

	// content -> form
	newHelpdeskTicket := ticketFormContent.NewSubView(
		"form",
		"examples/ticket/templates/content/form/newHelpdeskTicket.html.tmpl",
		handlers.NewHelpdeskTicketHandler,
	)

	// content -> form -> attachment-list
	_ = newHelpdeskTicket.NewDefaultSubView(
		"attachment-list",
		"examples/ticket/templates/content/form/attachmentList/uploadedHelpdeskFiles.html.tmpl",
		handlers.HelpdeskAttachmentFileListHandler,
	)
	uploadedHelpdeskFiles := newHelpdeskTicket.NewSubView(
		"attachment-list",
		"examples/ticket/templates/content/form/attachmentList/uploadedHelpdeskFiles.html.tmpl",
		handlers.UploadedFilesHandler,
	)

	// content -> form -> form-message
	submitHelpDeskTicket := newHelpdeskTicket.NewSubView(
		"form-message",
		"examples/ticket/templates/content/form/formMessage/submitHelpDeskTicket.html.tmpl",
		handlers.SubmitHelpDeskTicketHandler,
	)

	// content -> form -> notes
	_ = newHelpdeskTicket.NewDefaultSubView(
		"notes",
		"examples/ticket/templates/content/notes.html.tmpl",
		treetop.Noop,
	)

	// content -> form -> reported-by
	helpdeskReportedBy := newHelpdeskTicket.NewDefaultSubView(
		"reported-by",
		"examples/ticket/templates/content/form/reportedBy/helpdeskReportedBy.html.tmpl",
		handlers.HelpdeskReportedByHandler,
	)

	// content -> form -> reported-by -> find-reported-by
	findHelpdeskReportedBy := helpdeskReportedBy.NewSubView(
		"find-reported-by",
		"examples/ticket/templates/content/form/reportedBy/findReportedBy/findHelpdeskReportedBy.html.tmpl",
		handlers.FindTeamMemberHandler,
	)
	newSoftwareTicket := ticketFormContent.NewSubView(
		"form",
		"examples/ticket/templates/content/form/newSoftwareTicket.html.tmpl",
		handlers.NewSoftwareTicketHandler,
	)
	newSystemsTicket := ticketFormContent.NewSubView(
		"form",
		"examples/ticket/templates/content/form/newSystemsTicket.html.tmpl",
		handlers.NewSystemsTicketHandler,
	)

	// nav
	_ = baseView.NewDefaultSubView(
		"nav",
		"local://nav.html",
		treetop.Noop,
	)

	m.HandleGET("/ticket/helpdesk/preview",
		exec.NewViewHandler(previewHelpdeskTicket))
	m.HandlePOST("/ticket/helpdesk/upload-attachment",
		exec.NewViewHandler(uploadedHelpdeskFiles).FragmentOnly())
	m.HandlePOST("/ticket/helpdesk/submit",
		exec.NewViewHandler(submitHelpDeskTicket).FragmentOnly())
	m.HandleGET("/ticket/helpdesk/find-reported-by",
		exec.NewViewHandler(findHelpdeskReportedBy).FragmentOnly())
	m.HandleGET("/ticket/helpdesk/update-reported-by",
		exec.NewViewHandler(helpdeskReportedBy).FragmentOnly())
	m.HandleGET("/ticket/helpdesk/new",
		exec.NewViewHandler(newHelpdeskTicket))
	m.HandleGET("/ticket/software/new",
		exec.NewViewHandler(newSoftwareTicket))
	m.HandleGET("/ticket/systems/new",
		exec.NewViewHandler(newSystemsTicket))
	m.HandleGET("/ticket",
		exec.NewViewHandler(ticketFormContent))

}
