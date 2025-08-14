package core

import (
	"fmt"
	"strings"

	"github.com/glycerine/zygomys/v9/zygo"
)

var htmlTags = []string{
	"a", "abbr", "address", "area", "article", "aside", "audio",
	"b", "base", "bdi", "bdo", "blockquote", "body", "br", "button",
	"canvas", "caption", "cite", "code", "col", "colgroup",
	"data", "datalist", "dd", "del", "details", "dfn", "dialog", "div", "dl", "dt",
	"em", "embed",
	"fieldset", "figcaption", "figure", "footer", "form",
	"h1", "h2", "h3", "h4", "h5", "h6", "head", "header", "hgroup", "hr", "html",
	"i", "iframe", "img", "input", "ins",
	"kbd",
	"label", "legend", "li", "link",
	"main", "map", "mark", "meta", "meter",
	"nav", "noscript",
	"object", "ol", "optgroup", "option", "output",
	"p", "picture", "pre", "progress",
	"q",
	"rp", "rt", "ruby",
	"s", "samp", "script", "section", "select", "small", "source", "span", "strong", "style", "sub", "summary", "sup", "svg",
	"table", "tbody", "td", "template", "textarea", "tfoot", "th", "thead", "time", "title", "tr", "track",
	"u", "ul",
	"var", "video",
	"wbr",
}

// UseClientModule registers client functions.
func (vm *VM) UseClientModule() *VM {
	for _, tag := range htmlTags {
		vm.environment.AddFunction(tag, fnGenericHtmlHandler(tag))
	}

	return vm
}

// fnGenericHtmlHandler generate a generic html handler
// Lisp: (tag prop: "value" "children")
func fnGenericHtmlHandler(tag string) func(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	return func(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
		var sb strings.Builder
		sb.WriteString("<" + tag)

		// props parsing
		propCount := 0
		for i := 0; i+1 < len(args); i += 2 {
			if sym, ok := args[i].(*zygo.SexpSymbol); ok {
				if val, ok := args[i+1].(*zygo.SexpStr); ok {
					sb.WriteString(fmt.Sprintf(` %s="%s"`, strings.TrimSuffix(sym.Name(), ":"), val.S))
					propCount += 2
				}
			}
		}

		// if even args => self closing
		if len(args)%2 == 0 {
			sb.WriteString(" />")
			return &zygo.SexpStr{S: sb.String()}, nil
		}

		// last arg is children
		sb.WriteString(">")
		if last, ok := args[len(args)-1].(*zygo.SexpStr); ok {
			sb.WriteString(last.S)
		}
		sb.WriteString("</" + tag + ">")

		return &zygo.SexpStr{S: sb.String()}, nil
	}
}
