package treetop

import (
	"errors"
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"
	"text/template/parse"
)

type TemplateLoader struct {
	Load  func(string) (string, error)
	Funcs template.FuncMap
}

func NewTemplateLoader(funcs template.FuncMap, load func(string) (string, error)) *TemplateLoader {
	return &TemplateLoader{
		Load:  load,
		Funcs: funcs,
	}
}

func (tl TemplateLoader) ViewTemplate(view *View) (*template.Template, error) {
	if view == nil {
		return nil, nil
	}
	var out *template.Template

	queue := viewQueue{}
	queue.add(view)

	for !queue.empty() {
		v, _ := queue.next()
		var t *template.Template
		if out == nil {
			out = template.New(v.Defines).Funcs(tl.Funcs)
			t = out
		} else {
			t = out.New(v.Defines)
		}
		templateString, err := tl.Load(v.Template)
		if err != nil {
			return nil, err
		}

		if _, err := t.Parse(templateString); err != nil {
			return nil, err
		}
		// require template to declare a template/block node for each direct subview name
		if err := checkTemplateForBlockNames(t, v); err != nil {
			return nil, err
		}
		for _, sub := range v.SubViews {
			if sub != nil {
				queue.add(sub)
			}
		}
	}
	return out, nil
}

// utilities ---

var errEmptyViewQueue = errors.New("empty view queue")

// viewQueue simple queue implementation used for breath first traversal
//
// NB: this is only suitable for localized short-lived queues since the underlying
// array will not deallocate pointers
type viewQueue struct {
	offset int
	items  []*View
}

func (q *viewQueue) add(v *View) {
	q.items = append(q.items, v)
}

func (q *viewQueue) next() (*View, error) {
	if q.empty() {
		return nil, errEmptyViewQueue
	}
	next := q.items[q.offset]
	q.offset++
	return next, nil
}

func (q *viewQueue) empty() bool {
	return q.offset >= len(q.items)
}

// checkTemplateForBlockNames will scan the parsed templates for blocks/template slots
// that match the declared block names. If a block naming is not present, return an error
func checkTemplateForBlockNames(tmpl *template.Template, v *View) error {
	parsedBlocks := make(map[string]bool)
	for _, tplName := range listTemplateNodeName(tmpl.Tree.Root) {
		parsedBlocks[tplName] = true
	}

	var missing []string
	for blockName := range v.SubViews {
		if _, ok := parsedBlocks[blockName]; !ok {
			missing = append(missing, strconv.Quote(blockName))
		}
	}
	if len(missing) == 0 {
		return nil
	}
	sort.Strings(missing)
	return fmt.Errorf("%s is missing template declaration(s) for sub view blocks: %s", v.Template, strings.Join(missing, ", "))
}

// listTemplateNodeName will scan a parsed template tree for template nodes
// and list all template names found
func listTemplateNodeName(list *parse.ListNode) (names []string) {
	if list == nil {
		return
	}
	for _, node := range list.Nodes {
		switch n := node.(type) {
		case *parse.TemplateNode:
			names = append(names, n.Name)
		case *parse.IfNode:
			names = append(names, listTemplateNodeName(n.List)...)
			names = append(names, listTemplateNodeName(n.ElseList)...)
		case *parse.RangeNode:
			names = append(names, listTemplateNodeName(n.List)...)
			names = append(names, listTemplateNodeName(n.ElseList)...)
		case *parse.WithNode:
			names = append(names, listTemplateNodeName(n.List)...)
			names = append(names, listTemplateNodeName(n.ElseList)...)
		case *parse.ListNode:
			names = append(names, listTemplateNodeName(n)...)
		}
	}
	return
}
