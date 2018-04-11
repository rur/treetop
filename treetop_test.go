package treetop

import (
	"net/http"
	"reflect"
	"testing"
)

type TestBase struct {
	Title   string
	Content interface{}
	Footer  interface{}
}

func handler(w DataWriter, r *http.Request) {
	content, ok := w.Delegate("content", r)
	if !ok {
		content = 999
	}
	footer, ok := w.Delegate("footer", r)
	if !ok {
		footer = 999
	}
	w.Data(TestBase{
		Title:   "Base!!",
		Content: content,
		Footer:  footer,
	})
}

func contentHandler(w DataWriter, r *http.Request) {
	w.Data(123)
}

func footerHandler(w DataWriter, r *http.Request) {
	w.Data("This is the footer")
}

func Test_resolveTemplatesForHandler(t *testing.T) {
	base := NewPage("base.templ.html", handler)
	content := base.DefineBlock("content")
	footer := base.DefineBlock("footer").WithDefault("footer.templ.html", footerHandler)
	sub := content.Extend("sub.templ.html", contentHandler)
	otherFooter := footer.Extend("other_footer.templ.html", footerHandler)
	subWithOther := sub.Includes(otherFooter)
	subWithOther2 := subWithOther.Includes(
		NewPage("orphan.templ.html", handler),
	)

	type args struct {
		block   Block
		primary Handler
	}
	tests := []struct {
		name       string
		args       args
		handlerMap map[Block]Handler
		templates  []string
	}{
		{
			name: "basic",
			args: args{
				block:   base.Extends(),
				primary: base,
			},
			handlerMap: map[Block]Handler{
				base.Extends(): base,
				footer:         footer.Default(),
			},
			templates: []string{
				"base.templ.html",
				"footer.templ.html",
			},
		},
		{
			name: "sub",
			args: args{
				block:   base.Extends(),
				primary: sub,
			},
			handlerMap: map[Block]Handler{
				base.Extends(): base,
				footer:         footer.Default(),
				content:        sub,
			},
			templates: []string{
				"base.templ.html",
				"sub.templ.html",
				"footer.templ.html",
			},
		},
		{
			name: "includes",
			args: args{
				block:   base.Extends(),
				primary: subWithOther,
			},
			handlerMap: map[Block]Handler{
				base.Extends(): base,
				footer:         otherFooter,
				content:        subWithOther,
			},
			templates: []string{
				"base.templ.html",
				"sub.templ.html",
				"other_footer.templ.html",
			},
		},
		{
			name: "ignore orphan",
			args: args{
				block:   base.Extends(),
				primary: subWithOther2,
			},
			handlerMap: map[Block]Handler{
				base.Extends(): base,
				footer:         otherFooter,
				content:        subWithOther2,
			},
			templates: []string{
				"base.templ.html",
				"sub.templ.html",
				"other_footer.templ.html",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := resolveTemplatesForHandler(tt.args.block, tt.args.primary)
			if !reflect.DeepEqual(got, tt.handlerMap) {
				t.Errorf("resolveTemplatesForHandler() got = %v, want %v", got, tt.handlerMap)
			}
			if !reflect.DeepEqual(got1, tt.templates) {
				t.Errorf("resolveTemplatesForHandler() got1 = %v, want %v", got1, tt.templates)
			}
		})
	}
}
