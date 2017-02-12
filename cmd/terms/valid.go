package main

import (
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/exp/ebnf"
)

// validate validates the given grammar, and asserts that:
//
//    - all productions used are defined
//    - all productions defined are used when beginning at start
//    - lexical productions refer only to other lexical productions
//    - ranges are only used in lexical productions
func validate(grammar ebnf.Grammar, start string) error {
	if err := ebnf.Verify(grammar, start); err != nil {
		return errors.WithStack(err)
	}
	// Verify that ranges are only used in lexical productions.
	for name, prod := range grammar {
		if !isLexical(name) {
			if r, ok := hasRange(prod.Expr); ok {
				return errors.Errorf("non-terminal production rule %q containing range `%q … %q`", name, r.Begin.String, r.End.String)
			}
		}
	}
	return nil
}

// hasRange reports whether the given expression contains a range expression.
func hasRange(expr ebnf.Expression) (*ebnf.Range, bool) {
	switch expr := expr.(type) {
	case nil:
		// empty expression
		return nil, false
	case ebnf.Alternative:
		for _, e := range expr {
			if r, ok := hasRange(e); ok {
				return r, true
			}
		}
		return nil, false
	case ebnf.Sequence:
		for _, e := range expr {
			if r, ok := hasRange(e); ok {
				return r, true
			}
		}
		return nil, false
	case *ebnf.Name:
		return nil, false
	case *ebnf.Token:
		return nil, false
	case *ebnf.Range:
		return expr, true
	case *ebnf.Group:
		return hasRange(expr.Body)
	case *ebnf.Option:
		return hasRange(expr.Body)
	case *ebnf.Repetition:
		return hasRange(expr.Body)
	default:
		panic(fmt.Sprintf("internal error: unexpected type %T", expr))
	}
}