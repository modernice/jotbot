package golang

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
	"github.com/modernice/jotbot/git"
	"github.com/modernice/jotbot/internal"
	"github.com/modernice/jotbot/internal/slice"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slog"
)

// Patch represents a code patch for adding or updating documentation comments
// in Go files. It provides methods for commenting functions, type declarations,
// and variables/constants. The Patch struct is created with the New function,
// which takes a file system and options. The Identifiers method returns a map
// of all the identifiers in the patch that have been commented. The Comment
// method comments the identifier with the given comment string. The Apply
// method applies all the patches to the files in a given repository directory.
// The File method returns the source code of a single file from the patch.
type Patch struct {
	mux         sync.RWMutex
	repo        fs.FS
	fset        *token.FileSet
	files       map[string]*dst.File
	fileLocks   map[string]*sync.Mutex
	identifiers map[string][]string
	override    bool
	log         *slog.Logger
}

// PatchOption represents a functional option that can be used when creating a new
// Patch instance. It allows customization of the new instance by setting
// various options such as a logger or an override flag. Use WithLogger to set a
// logger and Override to set the override flag.
type PatchOption interface {
	applyPatch(*Patch)
}

type patchOptionFunc func(*Patch)

func (opt patchOptionFunc) applyPatch(p *Patch) {
	opt(p)
}

// Override sets the Patch's override flag to the provided boolean value. When
// the override flag is set to true, existing documentation for a function or
// type will be overwritten when new documentation is added. When it is false,
// an error will be returned if there is already documentation for a function or
// type.
func Override(override bool) PatchOption {
	return patchOptionFunc(func(p *Patch) {
		p.override = override
	})
}

// NewPatch returns a new *Patch that can be used to update documentation comments in
// Go source files. The repo argument is a filesystem to read the source files
// from. Options may be provided to modify the behavior of the patcher.
// WithLogger sets the logger for the patcher. Override allows existing comments
// to be overwritten.
func NewPatch(repo fs.FS, opts ...PatchOption) *Patch {
	p := &Patch{
		repo:        repo,
		fset:        token.NewFileSet(),
		files:       make(map[string]*dst.File),
		fileLocks:   make(map[string]*sync.Mutex),
		identifiers: make(map[string][]string),
	}
	for _, opt := range opts {
		opt.applyPatch(p)
	}
	if p.log == nil {
		p.log = internal.NopLogger()
	}
	return p
}

// Identifiers returns a map where the keys are file paths and the values are
// slices of identifiers for functions, types, variables or constants that have
// been commented using the Comment function.
func (p *Patch) Identifiers() map[string][]string {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return maps.Clone(p.identifiers)
}

// Comment is a method of the Patch type. It adds or removes documentation
// comments to an identifier in a Go source file. The identifier can be a
// function, variable, constant, or type. If the given comment is an empty
// string, the existing documentation for that identifier will be removed. If
// the identifier cannot be found in the file, an error will be returned.
func (p *Patch) Comment(file, identifier, comment string) (rerr error) {
	defer func() {
		if rerr == nil {
			p.mux.Lock()
			defer p.mux.Unlock()
			p.identifiers[file] = append(p.identifiers[file], identifier)
		}
	}()

	if comment == "" {
		p.log.Debug("Removing comment ...", "file", file, "identifier", identifier)
	} else {
		p.log.Debug("Adding comment ...", "file", file, "identifier", identifier, "comment", comment)
	}

	if recv, method, isMethod := splitMethodIdentifier(identifier); isMethod {
		decl, ok, err := p.findMethod(file, recv, method)
		if err != nil {
			return fmt.Errorf("look for method %s in %s: %w", identifier, file, err)
		}
		if !ok {
			return fmt.Errorf("could not find method %s in %s", identifier, file)
		}

		return p.commentFunction(file, decl, comment)
	}

	{
		spec, decl, ok, err := p.findType(file, identifier)
		if err != nil {
			return fmt.Errorf("look for type %s in %s: %w", identifier, file, err)
		}
		if ok {
			return p.commentGenDecl(file, spec.Name.Name, comment, decl)
		}
	}

	{
		spec, decl, ok, err := p.findVarOrConst(file, identifier)
		if err != nil {
			return fmt.Errorf("look for var/const %s in %s: %w", identifier, file, err)
		}
		if ok {
			return p.commentGenDecl(file, spec.Names[0].Name, comment, decl)
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

	return fmt.Errorf("could not find %s in %s", identifier, file)
}

func splitMethodIdentifier(identifier string) (recv, method string, ok bool) {
	parts := strings.Split(identifier, ".")
	if len(parts) != 2 {
		return "", "", false
	}
	recv = parts[0]
	if strings.HasPrefix(recv, "(*") {
		recv = recv[2 : len(recv)-1]
	}

	return recv, parts[1], true
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

func (p *Patch) commentGenDecl(file, identifier string, comment string, decl *dst.GenDecl) error {
	_, unlock := p.acquireFile(file)
	defer unlock()

	if !p.override && len(decl.Decs.Start.All()) > 0 {
		return fmt.Errorf("%s already has documentation", identifier)
	}

	decl.Decs.Start.Clear()
	if comment != "" {
		decl.Decs.Start.Append(formatComment(comment))
	}

	return nil
}

func (p *Patch) findFunction(file, identifier string) (*dst.FuncDecl, bool, error) {
	node, err := p.parseFile(file)
	if err != nil {
		return nil, false, err
	}

	for _, astDecl := range node.Decls {
		if fd, ok := astDecl.(*dst.FuncDecl); ok && fd.Name.Name == identifier && fd.Recv == nil {
			return fd, true, nil
		}
	}

	return nil, false, nil
}

func (p *Patch) commentFunction(file string, decl *dst.FuncDecl, comment string) error {
	_, unlock := p.acquireFile(file)
	defer unlock()

	if !p.override && len(decl.Decs.Start.All()) > 0 {
		return fmt.Errorf("function %s already has documentation", decl.Name.Name)
	}

	decl.Decs.Start.Clear()
	if comment != "" {
		decl.Decs.Start.Append(formatComment(comment))
	}

	return nil
}

func (p *Patch) findMethod(file, name, method string) (*dst.FuncDecl, bool, error) {
	node, err := p.parseFile(file)
	if err != nil {
		return nil, false, nil
	}

	name = strings.TrimPrefix(name, "*")
	if name == "" {
		return nil, false, fmt.Errorf("empty struct name")
	}

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

		if list, ok := rcv.(*dst.IndexListExpr); ok {
			rcv = list.X
		}

		if idx, ok := rcv.(*dst.IndexExpr); ok {
			rcv = idx.X
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

func (p *Patch) findVarOrConst(file, identifier string) (*dst.ValueSpec, *dst.GenDecl, bool, error) {
	node, err := p.parseFile(file)
	if err != nil {
		return nil, nil, false, err
	}

	var (
		spec *dst.ValueSpec
		decl *dst.GenDecl
	)

	dst.Inspect(node, func(node dst.Node) bool {
		switch node := node.(type) {
		case *dst.GenDecl:
			for _, sp := range node.Specs {
				if vs, ok := sp.(*dst.ValueSpec); ok {
					for _, ident := range vs.Names {
						if ident.Name == identifier {
							spec = vs
							decl = node
							return false
						}
					}
				}
			}
		}

		return true
	})

	return spec, decl, spec != nil, nil
}

// Commit commits the changes made by the Patch to the repository as a new
// git.Commit. The commit message will include a list of all files and
// identifiers that were updated with documentation.
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

// Apply applies the documentation patches to the source files of the patch. It
// updates the files with new comments or removes existing ones.
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

// File "*Patch.File" returns the source code of the specified file in bytes. It
// takes a string argument representing the path to the file. If the file is not
// found in the patch, it returns an error.
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

// DryRun returns a map of file paths to their corresponding bytes in the
// current state of the patch. No changes are made to the files on disk.
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
	words := strings.Split(str, " ")

	var lines []string
	var line string
	for _, word := range words {
		if len(line)+len(word) > maxLen {
			lines = append(lines, line)
			line = ""
		}
		line += word + " "
	}
	lines = append(lines, line)

	return lines
}
