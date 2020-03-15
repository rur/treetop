package writer

import (
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/greeter"
)

//Routes registers writer greeter endpoints
func Routes(mux *http.ServeMux) {
	mux.HandleFunc("/writer/greet", greetingWriteHandler)
	mux.HandleFunc("/writer", landingWriteHandler)
}

func landingWriteHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	d := struct {
		Value   string
		Message interface{}
	}{
		Value: req.URL.Query().Get("name"),
	}
	w.Header().Set("Vary", "Accept")
	if pw, ok := treetop.NewPartialWriter(w, req); ok {
		// template request
		err = landingTemplate.ExecuteTemplate(pw, "message", d)
		return
	}
	t, _ := baseTemplate.Clone()
	t.AddParseTree("content", contentTemplate.Tree)
	t.AddParseTree("message", landingTemplate.Tree)
	err = t.ExecuteTemplate(w, "base", d)
}

func greetingWriteHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()

	// use the greet parsing logic from greeting page
	msg := greeter.GetGreetingQuery(req)

	w.Header().Set("Vary", "Accept")
	if pw, ok := treetop.NewPartialWriter(w, req); ok {
		// template request
		err = greetingTemplate.ExecuteTemplate(pw, "message", msg)
		return
	}

	// return full page instead
	t, _ := baseTemplate.Clone()
	t.AddParseTree("content", contentTemplate.Tree)
	t.AddParseTree("message", greetingTemplate.Tree)
	err = t.ExecuteTemplate(w, "base", struct {
		Value   string
		Message interface{}
	}{
		Value:   msg.Who,
		Message: msg,
	})
}
