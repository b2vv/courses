package openapigen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// ErrorSite records every production WriteError invocation, including forwarding
// wrappers. Codes are sorted; Location and Sources retain reviewable provenance.
type ErrorSite struct {
	Location string
	Codes    []string
}
type ErrorCensus struct {
	Codes  []string
	Domain map[string][]string
	Sites  []ErrorSite
}
type errorFile struct {
	path, pkg string
	tree      *ast.File
	imports   map[string]string
}
type errorFunction struct {
	file *errorFile
	decl *ast.FuncDecl
}
type errorSymbol struct {
	code  string
	typed bool
}
type errorScanner struct {
	fset    *token.FileSet
	files   []*errorFile
	symbols map[string]errorSymbol
	parents map[ast.Node]ast.Node
	active  map[string]bool
}

// ScanErrorCodes reads source, never imports business modules. Test files are not
// production emissions. Unsupported code expressions and ambiguous flows fail.
// transport maps each non-domain code to a production emission file (root-relative).
func ScanErrorCodes(root string, transport map[string]string) (*ErrorCensus, error) {
	s := &errorScanner{fset: token.NewFileSet(), symbols: map[string]errorSymbol{}, parents: map[ast.Node]ast.Node{}, active: map[string]bool{}}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "vendor" || d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		tree, err := parser.ParseFile(s.fset, path, nil, 0)
		if err != nil {
			return err
		}
		f := &errorFile{rel, "courses/" + filepath.ToSlash(filepath.Dir(rel)), tree, map[string]string{}}
		for _, imp := range tree.Imports {
			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			name := filepath.Base(p)
			if imp.Name != nil {
				name = imp.Name.Name
			}
			f.imports[name] = p
		}
		s.files = append(s.files, f)
		var stack []ast.Node
		ast.Inspect(tree, func(n ast.Node) bool {
			if n == nil {
				stack = stack[:len(stack)-1]
				return false
			}
			if len(stack) > 0 {
				s.parents[n] = stack[len(stack)-1]
			}
			stack = append(stack, n)
			return true
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := &ErrorCensus{Domain: map[string][]string{}}
	for _, f := range s.files {
		parts := strings.Split(f.path, "/")
		if len(parts) < 4 || parts[0] != "internal" || parts[2] != "domain" {
			continue
		}
		if err = s.domain(f, result); err != nil {
			return nil, err
		}
	}
	emitted := map[string]bool{}
	used := map[string]bool{}
	for _, f := range s.files {
		var failure error
		ast.Inspect(f.tree, func(n ast.Node) bool {
			if failure != nil {
				return false
			}
			var name string
			switch x := n.(type) {
			case *ast.SelectorExpr:
				name = x.Sel.Name
			case *ast.Ident:
				if _, ok := s.parents[n].(*ast.SelectorExpr); ok {
					return true
				}
				name = x.Name
			default:
				return true
			}
			if name != "WriteError" {
				return true
			}
			// Declaration is not an invocation. Every other reference must be a call;
			// assigning the function to an alias must not evade the census.
			if fn, ok := s.parents[n].(*ast.FuncDecl); ok && fn.Name == n {
				return true
			}
			call, ok := s.parents[n].(*ast.CallExpr)
			if !ok || call.Fun != n {
				failure = fmt.Errorf("%s: unsupported WriteError reference", s.location(f, n))
				return false
			}
			if s.functionKey(f, call.Fun) != "courses/platform/httpserver.WriteError" {
				failure = fmt.Errorf("%s: unresolved WriteError target", s.location(f, n))
				return false
			}
			if len(call.Args) != 5 {
				failure = fmt.Errorf("%s: WriteError requires five arguments", s.location(f, n))
				return false
			}
			fn := s.enclosing(f, n)
			if fn == nil {
				failure = fmt.Errorf("%s: WriteError outside function", s.location(f, n))
				return false
			}
			codes, e := s.resolve(fn, call.Args[3], call.Pos())
			if e != nil {
				failure = fmt.Errorf("%s: %w", s.location(f, n), e)
				return false
			}
			codes = uniqueCodes(codes)
			if len(codes) == 0 {
				failure = fmt.Errorf("%s: empty resolution", s.location(f, n))
				return false
			}
			for _, code := range codes {
				if len(result.Domain[code]) == 0 {
					provenance, ok := transport[code]
					if !ok {
						failure = fmt.Errorf("%s: code %q has no domain or transport provenance", s.location(f, n), code)
						return false
					}
					if provenance == f.path {
						used[code] = true
					}
				}
				emitted[code] = true
			}
			result.Sites = append(result.Sites, ErrorSite{s.location(f, n), codes})
			return true
		})
		if failure != nil {
			return nil, failure
		}
	}
	for code, path := range transport {
		if !used[code] {
			return nil, fmt.Errorf("unused transport provenance %q at %s", code, path)
		}
	}
	for code := range emitted {
		result.Codes = append(result.Codes, code)
	}
	slices.Sort(result.Codes)
	if len(result.Sites) == 0 {
		return nil, fmt.Errorf("no WriteError calls found")
	}
	return result, nil
}
func uniqueCodes(in []string) []string { slices.Sort(in); return slices.Compact(in) }
func (s *errorScanner) location(f *errorFile, n ast.Node) string {
	return fmt.Sprintf("%s:%d", f.path, s.fset.Position(n.Pos()).Line)
}
func (s *errorScanner) domain(f *errorFile, out *ErrorCensus) error {
	for _, decl := range f.tree.Decls {
		// Error() values are finite only while the defined Error string type's
		// method is the canonical identity conversion. A changed method must fail.
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "Error" && fn.Recv != nil && len(fn.Recv.List) == 1 {
			receiver := fn.Recv.List[0]
			if typ, ok := receiver.Type.(*ast.Ident); ok && typ.Name == "Error" {
				valid := false
				if len(receiver.Names) == 1 && fn.Body != nil && len(fn.Body.List) == 1 {
					if ret, ok := fn.Body.List[0].(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
						if call, ok := ret.Results[0].(*ast.CallExpr); ok && len(call.Args) == 1 {
							cast, castOK := call.Fun.(*ast.Ident)
							value, valueOK := call.Args[0].(*ast.Ident)
							valid = castOK && cast.Name == "string" && cast.Obj == nil && valueOK && value.Obj == receiver.Names[0].Obj
						}
					}
				}
				if !valid {
					return fmt.Errorf("%s: noncanonical Error method", s.location(f, fn))
				}
			}
		}
		if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
			for _, spec := range gen.Specs {
				t := spec.(*ast.TypeSpec)
				if t.Name.Name == "Error" {
					underlying, ok := t.Type.(*ast.Ident)
					if !ok || underlying.Name != "string" || t.Assign.IsValid() {
						return fmt.Errorf("%s: Error must be a defined string type", s.location(f, t))
					}
				}
			}
		}
		gen, ok := decl.(*ast.GenDecl)
		if !ok || (gen.Tok != token.CONST && gen.Tok != token.VAR) {
			continue
		}
		for _, spec := range gen.Specs {
			v := spec.(*ast.ValueSpec)
			typed := false
			if t, ok := v.Type.(*ast.Ident); ok && t.Name == "Error" {
				typed = true
			}
			for i, name := range v.Names {
				var value ast.Expr
				if i < len(v.Values) {
					value = v.Values[i]
				}
				isNew := false
				if c, ok := value.(*ast.CallExpr); ok {
					isNew = s.functionKey(f, c.Fun) == "errors.New"
				}
				if !typed && !isNew {
					if strings.HasPrefix(name.Name, "Err") {
						return fmt.Errorf("%s: unsupported error declaration %s", s.location(f, v), name.Name)
					}
					continue
				}
				if typed && gen.Tok != token.CONST {
					return fmt.Errorf("%s: Error must be a constant", s.location(f, v))
				}
				if isNew {
					c := value.(*ast.CallExpr)
					if len(c.Args) != 1 {
						return fmt.Errorf("%s: invalid errors.New", s.location(f, v))
					}
					value = c.Args[0]
				}
				code, ok := stringLiteral(value)
				if !ok || code == "" {
					return fmt.Errorf("%s: error declaration requires nonempty literal", s.location(f, v))
				}
				key := f.pkg + "." + name.Name
				if _, ok := s.symbols[key]; ok {
					return fmt.Errorf("duplicate error symbol %s", key)
				}
				site := s.location(f, v)
				s.symbols[key] = errorSymbol{code, typed}
				out.Domain[code] = append(out.Domain[code], site)
			}
		}
	}
	return nil
}
func stringLiteral(e ast.Expr) (string, bool) {
	v, ok := e.(*ast.BasicLit)
	if !ok || v.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(v.Value)
	return s, err == nil
}
func (s *errorScanner) functionKey(f *errorFile, e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		if x.Obj == nil || x.Obj.Kind == ast.Fun {
			return f.pkg + "." + x.Name
		}
	case *ast.SelectorExpr:
		if id, ok := x.X.(*ast.Ident); ok && id.Obj == nil {
			if p := f.imports[id.Name]; p != "" {
				return p + "." + x.Sel.Name
			}
		}
	}
	return ""
}
func (s *errorScanner) enclosing(f *errorFile, n ast.Node) *errorFunction {
	for n != nil {
		if fn, ok := n.(*ast.FuncDecl); ok {
			return &errorFunction{f, fn}
		}
		n = s.parents[n]
	}
	return nil
}
func (s *errorScanner) resolve(fn *errorFunction, e ast.Expr, at token.Pos) ([]string, error) {
	return s.resolveValue(fn, e, at, false)
}
func (s *errorScanner) resolveValue(fn *errorFunction, e ast.Expr, at token.Pos, errorValue bool) ([]string, error) {
	key := fmt.Sprintf("%s:%p:%d:%t", fn.file.path, e, at, errorValue)
	if s.active[key] {
		return nil, fmt.Errorf("cyclic code flow at %s", s.location(fn.file, e))
	}
	s.active[key] = true
	defer delete(s.active, key)
	if value, ok := stringLiteral(e); ok {
		if value == "" {
			return nil, fmt.Errorf("empty error code")
		}
		return []string{value}, nil
	}
	switch x := e.(type) {
	case *ast.ParenExpr:
		return s.resolveValue(fn, x.X, at, errorValue)
	case *ast.SelectorExpr:
		if symbol, ok := s.symbols[s.functionKey(fn.file, x)]; ok && (symbol.typed || errorValue) {
			return []string{symbol.code}, nil
		}
	case *ast.CallExpr:
		if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "string" && id.Obj == nil && len(x.Args) == 1 {
			return s.resolve(fn, x.Args[0], at)
		}
		if errorValue && s.functionKey(fn.file, x.Fun) == "errors.New" && len(x.Args) == 1 {
			return s.resolve(fn, x.Args[0], at)
		}
		if sel, ok := x.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Error" && len(x.Args) == 0 {
			if symbol, ok := s.symbols[s.functionKey(fn.file, sel.X)]; ok {
				return []string{symbol.code}, nil
			}
			return s.resolveValue(fn, sel.X, at, true)
		}
	case *ast.Ident:
		return s.variable(fn, x, at, errorValue)
	}
	return nil, fmt.Errorf("unresolved code expression %T at %s", e, s.location(fn.file, e))
}

// variable deliberately accepts only reviewable finite flows: a single local
// initialization, an exhaustive switch assigning every arm, or a closed wrapper
// parameter whose callers all resolve. It is not an optimistic Go evaluator.
func (s *errorScanner) variable(fn *errorFunction, id *ast.Ident, at token.Pos, errorValue bool) ([]string, error) {
	escaped := false
	ast.Inspect(fn.decl.Body, func(n ast.Node) bool {
		if unary, ok := n.(*ast.UnaryExpr); ok && unary.Op == token.AND {
			if v, ok := unary.X.(*ast.Ident); ok && v.Obj == id.Obj {
				escaped = true
			}
		}
		return true
	})
	if escaped {
		return nil, fmt.Errorf("address escape of code variable %s", id.Name)
	}
	// A branch equality proves a parameter's value at this use (identity override).
	for n := ast.Node(id); n != nil && n != fn.decl; n = s.parents[n] {
		if clause, ok := n.(*ast.CaseClause); ok && len(clause.List) == 1 {
			sw, ok := s.parents[s.parents[clause]].(*ast.SwitchStmt)
			if ok && sw.Tag == nil {
				if eq, ok := clause.List[0].(*ast.BinaryExpr); ok && eq.Op == token.EQL {
					if lhs, ok := eq.X.(*ast.Ident); ok && lhs.Obj == id.Obj && lhs.Name == id.Name {
						if value, ok := stringLiteral(eq.Y); ok && !s.assignedBetween(fn, id, eq.End(), at) {
							return []string{value}, nil
						}
					}
				}
			}
		}
	}
	var defs []ast.Expr
	var writes []*ast.AssignStmt
	ast.Inspect(fn.decl.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ValueSpec:
			for i, name := range x.Names {
				if name.Obj == id.Obj && id.Obj != nil && i < len(x.Values) {
					defs = append(defs, x.Values[i])
				}
			}
		case *ast.AssignStmt:
			for i, lhs := range x.Lhs {
				if v, ok := lhs.(*ast.Ident); ok && v.Obj == id.Obj && id.Obj != nil {
					writes = append(writes, x)
					if len(x.Lhs) == len(x.Rhs) && i < len(x.Rhs) {
						defs = append(defs, x.Rhs[i])
					}
				}
			}
		}
		return true
	})
	if len(writes) > 0 || len(defs) > 0 {
		for _, w := range writes {
			if w.End() > at || (w.Tok != token.ASSIGN && w.Tok != token.DEFINE) || len(w.Lhs) != len(w.Rhs) {
				return nil, fmt.Errorf("ambiguous code assignment to %s", id.Name)
			}
		}
		if len(defs) == 1 && len(writes) <= 1 {
			// A conditional single assignment is not guaranteed to initialize the code.
			if len(writes) == 1 && s.parents[writes[0]] != fn.decl.Body {
				return nil, fmt.Errorf("conditional code assignment to %s", id.Name)
			}
			return s.resolveValue(fn, defs[0], defs[0].Pos(), errorValue)
		}
		if len(defs) != len(writes) || len(writes) == 0 {
			return nil, fmt.Errorf("ambiguous definitions of %s", id.Name)
		}
		var sw *ast.SwitchStmt
		for _, w := range writes {
			clause, ok := s.parents[w].(*ast.CaseClause)
			if !ok {
				return nil, fmt.Errorf("unsupported assignments to %s", id.Name)
			}
			parent, ok := s.parents[s.parents[clause]].(*ast.SwitchStmt)
			if !ok || (sw != nil && sw != parent) {
				return nil, fmt.Errorf("multiple switches assign %s", id.Name)
			}
			sw = parent
		}
		if s.parents[sw] != fn.decl.Body || sw.End() > at {
			return nil, fmt.Errorf("conditional switch for %s", id.Name)
		}
		hasDefault := false
		for _, n := range sw.Body.List {
			c := n.(*ast.CaseClause)
			if c.List == nil {
				hasDefault = true
			}
			count := 0
			for _, stmt := range c.Body {
				if w, ok := stmt.(*ast.AssignStmt); ok {
					for _, candidate := range writes {
						if w == candidate {
							count++
						}
					}
				} else {
					return nil, fmt.Errorf("unsupported control flow in code switch")
				}
			}
			if count != 1 {
				return nil, fmt.Errorf("code switch arm must assign %s exactly once", id.Name)
			}
		}
		if !hasDefault {
			return nil, fmt.Errorf("code switch missing default")
		}
		var out []string
		for _, d := range defs {
			values, err := s.resolveValue(fn, d, d.Pos(), errorValue)
			if err != nil {
				return nil, err
			}
			out = append(out, values...)
		}
		return uniqueCodes(out), nil
	}
	index := 0
	for _, field := range fn.decl.Type.Params.List {
		for _, name := range field.Names {
			if name.Obj == id.Obj && id.Obj != nil {
				return s.parameter(fn, index, errorValue)
			}
			index++
		}
	}
	if symbol, ok := s.symbols[fn.file.pkg+"."+id.Name]; ok && (symbol.typed || errorValue) {
		return []string{symbol.code}, nil
	}
	return nil, fmt.Errorf("unresolved variable %s at %s", id.Name, s.location(fn.file, id))
}
func (s *errorScanner) assignedBetween(fn *errorFunction, id *ast.Ident, start, end token.Pos) bool {
	assigned := false
	ast.Inspect(fn.decl.Body, func(n ast.Node) bool {
		if a, ok := n.(*ast.AssignStmt); ok && a.Pos() > start && a.Pos() < end {
			for _, lhs := range a.Lhs {
				if v, ok := lhs.(*ast.Ident); ok && v.Obj == id.Obj {
					assigned = true
				}
			}
		}
		return true
	})
	return assigned
}
func (s *errorScanner) parameter(fn *errorFunction, index int, errorValue bool) ([]string, error) {
	target := fn.file.pkg + "." + fn.decl.Name.Name
	if fn.decl.Recv != nil || ast.IsExported(fn.decl.Name.Name) {
		return nil, fmt.Errorf("open caller set for %s", target)
	}
	var out []string
	for _, f := range s.files {
		var failure error
		ast.Inspect(f.tree, func(n ast.Node) bool {
			if failure != nil {
				return false
			}
			expr, ok := n.(ast.Expr)
			if !ok || s.functionKey(f, expr) != target {
				return true
			}
			if n == fn.decl.Name {
				return true
			}
			call, ok := s.parents[n].(*ast.CallExpr)
			if !ok || call.Fun != n {
				failure = fmt.Errorf("escaping wrapper %s", target)
				return false
			}
			if index >= len(call.Args) {
				failure = fmt.Errorf("missing code argument to %s", target)
				return false
			}
			caller := s.enclosing(f, call)
			if caller == nil {
				failure = fmt.Errorf("wrapper call outside function")
				return false
			}
			codes, err := s.resolveValue(caller, call.Args[index], call.Pos(), errorValue)
			if err != nil {
				failure = err
				return false
			}
			out = append(out, codes...)
			return true
		})
		if failure != nil {
			return nil, failure
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no closed callers for %s", target)
	}
	return uniqueCodes(out), nil
}
