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
	previewSoftwareTicket := issuePreview.NewSubView(
		"preview",
		"examples/ticket/templates/content/previewSoftwareTicket.html.tmpl",
		handlers.PreviewSoftwareTicketHandler,
	)
	previewHelpdeskTicket := issuePreview.NewSubView(
		"preview",
		"examples/ticket/templates/content/previewHelpdeskTicket.html.tmpl",
		handlers.PreviewHelpdeskTicketHandler,
	)
	previewSystemsTicket := issuePreview.NewSubView(
		"preview",
		"examples/ticket/templates/content/previewSystemsTicket.html.tmpl",
		handlers.PreviewSystemsTicketHandler,
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
		"examples/ticket/templates/content/form/attachmentList/uploadedFiles.html.tmpl",
		handlers.AttachmentFileListHandler,
	)
	uploadedHelpdeskFiles := newHelpdeskTicket.NewSubView(
		"attachment-list",
		"examples/ticket/templates/content/form/attachmentList/uploadedFiles.html.tmpl",
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

	// content -> form -> reported-by -> find-user
	findHelpdeskReportedBy := helpdeskReportedBy.NewDefaultSubView(
		"find-user",
		"examples/ticket/templates/content/form/reportedBy/findReportedByUser.html.tmpl",
		handlers.FindReportedByUserHandler,
	)
	newSoftwareTicket := ticketFormContent.NewSubView(
		"form",
		"examples/ticket/templates/content/form/newSoftwareTicket.html.tmpl",
		handlers.NewSoftwareTicketHandler,
	)

	// content -> form -> assignee
	viewSoftwareAssignee := newSoftwareTicket.NewDefaultSubView(
		"assignee",
		"examples/ticket/templates/content/form/assignee/assignee-input.html.tmpl",
		handlers.SoftwareAssigneeHandler,
	)

	// content -> form -> assignee -> find-user
	findSoftwareAssignee := viewSoftwareAssignee.NewDefaultSubView(
		"find-user",
		"examples/ticket/templates/content/form/assignee/find-assignee.html.tmpl",
		handlers.SoftwareFindAssigneeHandler,
	)

	// content -> form -> attachment-list
	_ = newSoftwareTicket.NewDefaultSubView(
		"attachment-list",
		"examples/ticket/templates/content/form/attachmentList/uploadedFiles.html.tmpl",
		handlers.AttachmentFileListHandler,
	)
	uploadedSoftwareFiles := newSoftwareTicket.NewSubView(
		"attachment-list",
		"examples/ticket/templates/content/form/attachmentList/uploadedFiles.html.tmpl",
		handlers.UploadedFilesHandler,
	)

	// content -> form -> form-message
	submitSoftwareTicket := newSoftwareTicket.NewSubView(
		"form-message",
		"examples/ticket/templates/content/form/formMessage/submitSoftwareTicket.html.tmpl",
		handlers.SubmitSoftwareTicketHandler,
	)

	// content -> form -> notes
	_ = newSoftwareTicket.NewDefaultSubView(
		"notes",
		"examples/ticket/templates/content/notes.html.tmpl",
		treetop.Noop,
	)
	newSystemsTicket := ticketFormContent.NewSubView(
		"form",
		"examples/ticket/templates/content/form/newSystemsTicket.html.tmpl",
		handlers.NewSystemsTicketHandler,
	)

	// content -> form -> attachment-list
	_ = newSystemsTicket.NewDefaultSubView(
		"attachment-list",
		"examples/ticket/templates/content/form/attachmentList/uploadedFiles.html.tmpl",
		handlers.AttachmentFileListHandler,
	)
	uploadedSystemsFiles := newSystemsTicket.NewSubView(
		"attachment-list",
		"examples/ticket/templates/content/form/attachmentList/uploadedFiles.html.tmpl",
		handlers.UploadedFilesHandler,
	)

	// content -> form -> component-tags
	systemsComponentTagsInputGroup := newSystemsTicket.NewDefaultSubView(
		"component-tags",
		"examples/ticket/templates/content/form/componentTags/systemsComponentTagsInputGroup.html.tmpl",
		handlers.SystemsComponentTagsInputGroupHandler,
	)

	// content -> form -> component-tags -> tag-search
	systemsComponentTagSearch := systemsComponentTagsInputGroup.NewSubView(
		"tag-search",
		"examples/ticket/templates/content/form/componentTags/systemsComponentTagSearch.html.tmpl",
		handlers.SystemsComponentTagSearchHandler,
	)

	// content -> form -> form-message
	submitSystemsTicket := newSystemsTicket.NewSubView(
		"form-message",
		"examples/ticket/templates/content/form/formMessage/submitSystemsTicket.html.tmpl",
		handlers.SubmitSystemsTicketHandler,
	)

	// content -> form -> notes
	_ = newSystemsTicket.NewDefaultSubView(
		"notes",
		"examples/ticket/templates/content/notes.html.tmpl",
		treetop.Noop,
	)

	// nav
	_ = baseView.NewDefaultSubView(
		"nav",
		"local://nav.html",
		treetop.Noop,
	)

	m.HandleGET("/ticket/software/preview",
		exec.NewViewHandler(previewSoftwareTicket))
	m.HandleGET("/ticket/helpdesk/preview",
		exec.NewViewHandler(previewHelpdeskTicket))
	m.HandleGET("/ticket/systems/preview",
		exec.NewViewHandler(previewSystemsTicket))
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
	m.HandleGET("/ticket/software/find-assignee",
		exec.NewViewHandler(findSoftwareAssignee).FragmentOnly())
	m.HandleGET("/ticket/software/update-assignee",
		exec.NewViewHandler(viewSoftwareAssignee).FragmentOnly())
	m.HandlePOST("/ticket/software/upload-attachment",
		exec.NewViewHandler(uploadedSoftwareFiles).FragmentOnly())
	m.HandlePOST("/ticket/software/submit",
		exec.NewViewHandler(submitSoftwareTicket).FragmentOnly())
	m.HandleGET("/ticket/software/new",
		exec.NewViewHandler(newSoftwareTicket))
	m.HandlePOST("/ticket/systems/upload-attachment",
		exec.NewViewHandler(uploadedSystemsFiles).FragmentOnly())
	m.HandleGET("/ticket/systems/find-tag",
		exec.NewViewHandler(systemsComponentTagSearch).FragmentOnly())
	m.HandleGET("/ticket/systems/update-tags",
		exec.NewViewHandler(systemsComponentTagsInputGroup).FragmentOnly())
	m.HandlePOST("/ticket/systems/submit",
		exec.NewViewHandler(submitSystemsTicket).FragmentOnly())
	m.HandleGET("/ticket/systems/new",
		exec.NewViewHandler(newSystemsTicket))
	m.HandleGET("/ticket",
		exec.NewViewHandler(ticketFormContent))

}
