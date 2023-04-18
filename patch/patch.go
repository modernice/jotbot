package patch

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/modernice/opendocs/git"
	"github.com/modernice/opendocs/internal"
	"github.com/modernice/opendocs/internal/slice"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slog"
)

type Patch struct {
	mux         sync.RWMutex
	repo        fs.FS
	fset        *token.FileSet
	files       map[string]*dst.File
	fileLocks   map[string]*sync.Mutex
	identifiers map[string][]string
	log         *slog.Logger
}

type Option func(*Patch)

func WithLogger(h slog.Handler) Option {
	return func(p *Patch) {
		p.log = slog.New(h)
	}
}

func New(repo fs.FS, opts ...Option) *Patch {
	p := &Patch{
		repo:        repo,
		fset:        token.NewFileSet(),
		files:       make(map[string]*dst.File),
		fileLocks:   make(map[string]*sync.Mutex),
		identifiers: make(map[string][]string),
	}
	for _, opt := range opts {
		opt(p)
	}
	if p.log == nil {
		p.log = internal.NopLogger()
	}
	return p
}

func (p *Patch) Identifiers() map[string][]string {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return maps.Clone(p.identifiers)
}

// func (p *Patch) Merge(p2 *Patch) error {
// 	// Lock p2 until we're done merging.
// 	p2.mux.Lock()
// 	defer p2.mux.Unlock()

// 	for file, identifiers := range p2.identifiers {
// 		for _, identifier := range identifiers {
// 			f, ok := p2.files[file]
// 			if !ok {
// 				continue
// 			}

// 			node, ok := nodes.Find(identifier, f)
// 			if !ok {
// 				continue
// 			}

// 			if err := p.Comment(file, identifier, nodes.Doc(node, true)); err != nil {
// 				return fmt.Errorf("add comment for %s@%s: %w", file, identifier, err)
// 			}
// 		}
// 	}

// 	return nil
// }

func (p *Patch) Comment(file, identifier, comment string) (rerr error) {
	defer func() {
		if rerr == nil {
			p.mux.Lock()
			defer p.mux.Unlock()
			p.identifiers[file] = append(p.identifiers[file], identifier)
		}
	}()

	if comment == "" {
		return nil
	}

	p.log.Debug("Adding comment ...", "identifier", identifier, "file", file)

	{
		spec, decl, ok, err := p.findType(file, identifier)
		if err != nil {
			return fmt.Errorf("look for type %s in %s: %w", identifier, file, err)
		}
		if ok {
			return p.commentType(file, decl, spec, comment)
		}
	}

	{
		decl, ok, err := p.findFunction(file, identifier)
		if err != nil {
			return fmt.Errorf("look for function %s in %s: %w", identifier, file, err)
		}
		if ok {
			return p.commentFunction(file, decl, comment)
		}
	}

	{
		decl, ok, err := p.findMethod(file, identifier)
		if err != nil {
			return fmt.Errorf("look for method %s in %s: %w", identifier, file, err)
		}
		if ok {
			return p.commentFunction(file, decl, comment)
		}
	}

	return fmt.Errorf("could not find %s in %s", identifier, file)
}

func (p *Patch) parseFile(path string) (*dst.File, error) {
	if node, ok := p.cached(path); ok {
		return node, nil
	}

	p.mux.Lock()
	defer p.mux.Unlock()

	if node, ok := p.files[path]; ok {
		return node, nil
	}

	f, err := p.repo.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	code, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	node, err := decorator.ParseFile(p.fset, "", code, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	p.files[path] = node

	return node, nil
}

func (p *Patch) cached(file string) (*dst.File, bool) {
	p.mux.RLock()
	defer p.mux.RUnlock()
	node, ok := p.files[file]
	return node, ok
}

func (p *Patch) acquireFile(file string) (*dst.File, func()) {
	p.mux.Lock()
	defer p.mux.Unlock()

	if _, ok := p.fileLocks[file]; !ok {
		p.fileLocks[file] = &sync.Mutex{}
	}

	p.fileLocks[file].Lock()
	return p.files[file], p.fileLocks[file].Unlock
}

func (p *Patch) findFunction(file, identifier string) (*dst.FuncDecl, bool, error) {
	node, err := p.parseFile(file)
	if err != nil {
		return nil, false, err
	}

	for _, astDecl := range node.Decls {
		if fd, ok := astDecl.(*dst.FuncDecl); ok && fd.Name.Name == identifier {
			return fd, true, nil
		}
	}

	return nil, false, nil
}

func (p *Patch) commentFunction(file string, decl *dst.FuncDecl, comment string) error {
	_, unlock := p.acquireFile(file)
	defer unlock()

	if len(decl.Decs.Start.All()) > 0 {
		return fmt.Errorf("function %s already has documentation", decl.Name.Name)
	}

	decl.Decs.Start.Append(formatComment(comment))

	return nil
}

func (p *Patch) findMethod(file, identifier string) (*dst.FuncDecl, bool, error) {
	parts := strings.Split(identifier, ".")
	if len(parts) != 2 {
		return nil, false, nil
	}

	node, err := p.parseFile(file)
	if err != nil {
		return nil, false, nil
	}

	name, method := parts[0], parts[1]
	name = strings.TrimPrefix(name, "*")

	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*dst.FuncDecl)
		if !ok {
			continue
		}

		if funcDecl.Name.Name != method {
			continue
		}

		if funcDecl.Recv == nil {
			continue
		}

		if len(funcDecl.Recv.List) == 0 {
			continue
		}

		rcv := funcDecl.Recv.List[0].Type
		if star, ok := rcv.(*dst.StarExpr); ok {
			rcv = star.X
		}

		if ident, ok := rcv.(*dst.Ident); ok && ident.Name == name {
			return funcDecl, true, nil
		}
	}

	return nil, false, nil
}

func (p *Patch) findType(file, identifier string) (*dst.TypeSpec, *dst.GenDecl, bool, error) {
	node, err := p.parseFile(file)
	if err != nil {
		return nil, nil, false, err
	}

	for _, astDecl := range node.Decls {
		if decl, ok := astDecl.(*dst.GenDecl); ok {
			for _, spec := range decl.Specs {
				if ts, ok := spec.(*dst.TypeSpec); ok && ts.Name.Name == identifier {
					return ts, decl, true, nil
				}
			}
		}
	}
	return nil, nil, false, nil
}

func (p *Patch) commentType(file string, decl *dst.GenDecl, spec *dst.TypeSpec, comment string) error {
	_, unlock := p.acquireFile(file)
	defer unlock()

	if len(decl.Decs.Start.All()) > 0 {
		return fmt.Errorf("type %s already has documentation", spec.Name.Name)
	}

	decl.Decs.Start.Append(formatComment(comment))

	return nil
}

func (p *Patch) Commit() git.Commit {
	c := git.DefaultCommit()
	if len(p.files) == 0 {
		return c
	}

	c.Desc = append(c.Desc, "Updated docs:")

	for file, identifiers := range p.identifiers {
		for _, ident := range identifiers {
			c.Desc = append(c.Desc, fmt.Sprintf("  - %s@%s", file, ident))
		}
	}

	return c
}

func (p *Patch) Apply(repo string) error {
	p.log.Info("Applying patches ...", "files", len(p.files))

	p.mux.RLock()
	defer p.mux.RUnlock()

	for file, node := range p.files {
		restorer := decorator.NewRestorer()
		restorer.Fset = p.fset

		var buf bytes.Buffer
		if err := restorer.Fprint(&buf, node); err != nil {
			return fmt.Errorf("format %s: %w", file, err)
		}

		fullpath := filepath.Join(repo, file)
		if err := p.patchFile(fullpath, &buf); err != nil {
			return fmt.Errorf("patch %s: %w", file, err)
		}
	}

	return nil
}

func (p *Patch) patchFile(path string, buf *bytes.Buffer) error {
	p.log.Info(fmt.Sprintf("Patching file %s ...", path))

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	_, err = io.Copy(f, buf)
	return err
}

func (p *Patch) File(file string) ([]byte, error) {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return p.printFile(file)
}

func (p *Patch) printFile(file string) ([]byte, error) {
	node, ok := p.files[file]
	if !ok {
		return nil, fmt.Errorf("file %s not found in patch", file)
	}

	restorer := decorator.NewRestorer()
	restorer.Fset = p.fset

	var buf bytes.Buffer
	if err := restorer.Fprint(&buf, node); err != nil {
		return buf.Bytes(), fmt.Errorf("format %s in %s: %w", node.Name.Name, file, err)
	}

	return buf.Bytes(), nil
}

func (p *Patch) DryRun() (map[string][]byte, error) {
	result := make(map[string][]byte)

	p.mux.RLock()
	defer p.mux.RUnlock()

	for path := range p.files {
		b, err := p.printFile(path)
		if err != nil {
			return result, err
		}
		result[path] = b
	}

	return result, nil
}

func formatComment(doc string) string {
	lines := splitString(doc, 77)
	lines = slice.Map(lines, func(s string) string {
		return "// " + s
	})
	return strings.Join(lines, "\n")
}

func splitString(str string, maxLen int) []string {
	var out []string

	paras := strings.Split(str, "\n\n")
	for i, para := range paras {
		lines := splitByWords(para, maxLen)
		out = append(out, lines...)
		if i < len(paras)-1 {
			out = append(out, "")
		}
	}

	return out
}

func splitByWords(str string, maxLen int) []string {
	words := strings.Fields(str)

	var lines []string
	var line string
	for _, word := range words {
		if len(line)+len(word) > maxLen {
			lines = append(lines, line)
			line = ""
		}
		line += word + " "
	}
	if line = strings.TrimSpace(line); line != "" {
		lines = append(lines, line)
	}

	return lines
}
