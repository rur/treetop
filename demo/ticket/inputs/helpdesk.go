package inputs

import (
	"log"
	"net/url"
	"strings"
)

// HelpDeskTicket is the model for a helpdesk ticket
type HelpDeskTicket struct {
	Summary            string
	ReportedBy         string
	ReportedByCustomer string
	ReportedByUser     string
	CustomerContact    string
	Urgency            string
	Description        string
	Attachments        []*FileInfo
}

// HelpdeskTicketFromQuery descodes a help desk ticket model from url query parameters
// Note, unicode normalization omitted for brevity
func HelpdeskTicketFromQuery(query url.Values) *HelpDeskTicket {
	ticket := &HelpDeskTicket{
		Summary:     strings.TrimSpace(query.Get("summary")),
		ReportedBy:  query.Get("reported-by"),
		Urgency:     query.Get("urgency"),
		Description: strings.TrimSpace(query.Get("description")),
	}

	switch ticket.ReportedBy {
	case "team-member":
		ticket.ReportedByUser = strings.TrimSpace(query.Get("reported-by-user"))
	case "customer":
		ticket.ReportedByCustomer = strings.TrimSpace(query.Get("reported-by-customer"))
		ticket.CustomerContact = strings.TrimSpace(query.Get("customer-contact"))
	case "myself":
	default:
		ticket.ReportedBy = ""
	}

	switch ticket.Urgency {
	case "critical", "major", "normal", "minor":
	default:
		ticket.Urgency = ""
	}

	for _, enc := range query["attachment"] {
		info := &FileInfo{}
		if err := info.UnmarshalBase64([]byte(enc)); err != nil {
			log.Println("Error parsing encoded attachmet", err)
			continue
		}
		ticket.Attachments = append(ticket.Attachments, info)
	}

	return ticket
}

func (t *HelpDeskTicket) RawQuery() string {
	query := url.Values{}
	query.Set("department", "helpdesk")
	query.Set("summary", t.Summary)
	query.Set("reported-by", t.ReportedBy)
	query.Set("reported-by-user", t.ReportedByUser)
	query.Set("reported-by-customer", t.ReportedByCustomer)
	query.Set("customer-contact", t.CustomerContact)
	query.Set("urgency", t.Urgency)
	query.Set("description", t.Description)
	for _, att := range t.Attachments {
		enc, err := att.MarshalBase64()
		if err != nil {
			log.Println("Failed to encode attachment", err)
			// skip it
			continue
		}
		query["attachment"] = append(query["attachment"], string(enc))
	}
	return query.Encode()
}
