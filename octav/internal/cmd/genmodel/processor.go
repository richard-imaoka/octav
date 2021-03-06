package main

// Given model name X
// 1) Look at its struct representation. This MUST exist
// 2) Look at its database counterpart under db/*. This MUST exist

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lestrrat/go-pdebug"
)

var ErrAnnotatedStructNotFound = errors.New("annotated struct was not found")

func snakeCase(s string) string {
	ret := []rune{}
	wasLower := false
	for len(s) > 0 {
		r, n := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError {
			panic("yikes")
		}

		s = s[n:]
		if unicode.IsUpper(r) {
			if wasLower {
				ret = append(ret, '_')
			}
			wasLower = false
		} else {
			wasLower = true
		}

		ret = append(ret, unicode.ToLower(r))
	}
	return string(ret)
}

type Processor struct {
	Types []string
	Dir   string
}

func skipGenerated(fi os.FileInfo) bool {
	switch {
	case strings.HasSuffix(fi.Name(), "_gen.go"):
		return false
	case strings.HasSuffix(fi.Name(), "_gen.go"):
		return false
	}
	return true
}

type genctx struct {
	Dir         string
	DBRows      map[string]DBRow
	Models      []Model
	PkgName     string
	Services    map[string]Service
	TargetTypes []string
}

type Service struct {
	Name                         string
	HasPostLookupFromPayloadHook bool
	HasPreUpdateFromPayloadHook  bool
	HasPostUpdateFromPayloadHook bool
	HasPostCreateHook            bool
	HasPostUpdateHook            bool
	HasPostDeleteHook            bool
}

type Field struct {
	Convert  bool
	Decorate bool
	JSONName string
	L10N     bool
	Name     string
	Tag      reflect.StructTag
	Type     string
}

type Model struct {
	Fields        []Field
	HasL10N       bool
	Name          string
	PkgName       string
	Lookup        bool
	Delete        bool
	Update        bool
	LookupRequest bool
	CreateRequest bool
	UpdateRequest bool
}

type DBColumn struct {
	BaseType   string
	IsNullType bool
	Name       string
	Type       string
}

type DBRow struct {
	Columns map[string]DBColumn
	Name    string
	PkgName string
}

func (p *Processor) Do() error {
	ctx := genctx{
		Dir:         p.Dir,
		DBRows:      make(map[string]DBRow),
		Services:    make(map[string]Service),
		TargetTypes: p.Types,
	}
	if err := parseModelDir(&ctx, filepath.Join(ctx.Dir, "model")); err != nil {
		return err
	}

	if err := parseDBDir(&ctx, filepath.Join(ctx.Dir, "db")); err != nil {
		return err
	}

	if err := parseServiceDir(&ctx, filepath.Join(ctx.Dir, "service")); err != nil {
		return err
	}

	if err := generateFiles(&ctx); err != nil {
		return err
	}

	return nil
}

func parseDir(ctx *genctx, dir string, cb func(*genctx, *ast.Package) error) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, skipGenerated, parser.ParseComments)
	if err != nil {
		return err
	}

	if len(pkgs) == 0 {
		return errors.New("no packages to process...")
	}

	for _, pkg := range pkgs {
		if strings.HasSuffix(pkg.Name, "_test") {
			continue
		}

		ctx.PkgName = pkg.Name
		if err := cb(ctx, pkg); err != nil {
			return err
		}
		return nil
	}

	return errors.New("only found test package")
}

func parseModelDir(ctx *genctx, dir string) error {
	return parseDir(ctx, dir, processModelPkg)
}

func parseServiceDir(ctx *genctx, dir string) error {
	return parseDir(ctx, dir, processServicePkg)
}

func parseDBDir(ctx *genctx, dir string) error {
	return parseDir(ctx, dir, processDBPkg)
}

func processPkg(ctx *genctx, pkg *ast.Package, cb func(ast.Node) bool) error {
	if pdebug.Enabled {
		g := pdebug.Marker("processPkg %s", pkg.Name)
		defer g.End()
	}

	for fn, f := range pkg.Files {
		if pdebug.Enabled {
			pdebug.Printf("Checking file %s", fn)
		}

		ast.Inspect(f, cb)
	}
	return nil
}

func processModelPkg(ctx *genctx, pkg *ast.Package) error {
	if pdebug.Enabled {
		g := pdebug.Marker("processModelPkg %s", pkg.Name)
		defer g.End()
	}

	if err := processPkg(ctx, pkg, ctx.extractModelStructs); err != nil {
		return err
	}
	return nil
}

func processServicePkg(ctx *genctx, pkg *ast.Package) error {
	if pdebug.Enabled {
		g := pdebug.Marker("processServicePkg %s", pkg.Name)
		defer g.End()
	}

	if err := processPkg(ctx, pkg, ctx.extractServiceStructs); err != nil {
		return err
	}
	return nil
}

func processDBPkg(ctx *genctx, pkg *ast.Package) error {
	if pdebug.Enabled {
		g := pdebug.Marker("processDBPkg %s", pkg.Name)
		defer g.End()
	}

	if err := processPkg(ctx, pkg, ctx.extractDBStructs); err != nil {
		return err
	}
	return nil
}

func shouldProceed(ctx *genctx, name string) bool {
	if len(ctx.TargetTypes) == 0 {
		return true
	}

	for _, t := range ctx.TargetTypes {
		if t == name {
			return true
		}
	}
	return false
}

func (ctx *genctx) extractModelStructs(n ast.Node) bool {
	decl, ok := n.(*ast.GenDecl)
	if !ok {
		return true
	}

	if decl.Tok != token.TYPE {
		return true
	}

	for _, spec := range decl.Specs {
		var t *ast.TypeSpec
		var s *ast.StructType
		var ok bool

		if t, ok = spec.(*ast.TypeSpec); !ok {
			continue
		}

		if !shouldProceed(ctx, t.Name.Name) {
			continue
		}

		if s, ok = t.Type.(*ast.StructType); !ok {
			continue
		}

		cgroup := decl.Doc
		if cgroup == nil {
			continue
		}
		ismodel := false
		var tags string
		for _, c := range cgroup.List {
			if strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(c.Text, "//")), "+model") {
				ismodel = true
				tags = c.Text
				break
			}
		}
		if !ismodel {
			if pdebug.Enabled {
				pdebug.Printf("%s is NOT a valid model", t.Name.Name)
			}
			continue
		}

		st := Model{
			Fields:        make([]Field, 0, len(s.Fields.List)),
			Update:        true,
			Delete:        true,
			Lookup:        true,
			LookupRequest: true,
			UpdateRequest: true,
			CreateRequest: true,
			Name:          t.Name.Name,
			PkgName:       ctx.PkgName,
		}

		if i := strings.IndexByte(tags, '`'); i > 0 {
			if len(tags) > i {
				tags = tags[i+1:]
			}
		}
		if i := strings.IndexByte(tags, '`'); i > 0 {
			tags = tags[:i]
		}

		tag := reflect.StructTag(tags)

		if b, ok := tag.Lookup("Update"); ok {
			if ok, err := strconv.ParseBool(b); err == nil {
				st.Update = ok
			}
		}
		if b, ok := tag.Lookup("Delete"); ok {
			if ok, err := strconv.ParseBool(b); err == nil {
				st.Delete = ok
			}
		}
		if b, ok := tag.Lookup("Lookup"); ok {
			if ok, err := strconv.ParseBool(b); err == nil {
				st.Lookup = ok
			}
		}
		if b, ok := tag.Lookup("LookupRequest"); ok {
			if ok, err := strconv.ParseBool(b); err == nil {
				st.LookupRequest = ok
			}
		}
		if b, ok := tag.Lookup("UpdateRequest"); ok {
			if ok, err := strconv.ParseBool(b); err == nil {
				st.UpdateRequest = ok
			}
		}
		if b, ok := tag.Lookup("CreateRequest"); ok {
			if ok, err := strconv.ParseBool(b); err == nil {
				st.CreateRequest = ok
			}
		}

	LoopFields:
		for _, f := range s.Fields.List {
			if len(f.Names) == 0 {
				continue
			}

			if unicode.IsLower(rune(f.Names[0].Name[0])) {
				continue
			}

			var jsname string
			var l10n bool
			var decorate bool
			var convert bool
			var ft reflect.StructTag
			if f.Tag != nil {
				v := f.Tag.Value
				if len(v) >= 2 {
					if v[0] == '`' {
						v = v[1:]
					}
					if v[len(v)-1] == '`' {
						v = v[:len(v)-1]
					}
				}

				ft = reflect.StructTag(v)
				tag := ft.Get("json")
				if tag == "-" {
					continue LoopFields
				}
				if tag == "" || tag[0] == ',' {
					jsname = f.Names[0].Name
				} else {
					tl := strings.SplitN(tag, ",", 2)
					jsname = tl[0]
				}

				tag = ft.Get("l10n")
				if b, err := strconv.ParseBool(tag); err == nil && b {
					l10n = true
				}

				tag = ft.Get("decorate")
				if b, err := strconv.ParseBool(tag); err == nil && b {
					decorate = true
				}

				if tag = ft.Get("assign"); tag == "convert" {
					convert = true
				}
			}

			typ, err := getTypeName(f.Type)
			if err != nil {
				return true
			}

			field := Field{
				L10N:     l10n,
				Convert:  convert,
				Decorate: decorate,
				Name:     f.Names[0].Name,
				JSONName: jsname,
				Tag:      ft,
				Type:     typ,
			}

			st.Fields = append(st.Fields, field)
		}
		ctx.Models = append(ctx.Models, st)
	}

	return true
}

func (ctx *genctx) extractServiceStructs(n ast.Node) bool {
	decl, ok := n.(*ast.GenDecl)
	if !ok {
		return true
	}

	if decl.Tok != token.TYPE {
		return true
	}

	for _, spec := range decl.Specs {
		var t *ast.TypeSpec
		var ok bool

		if t, ok = spec.(*ast.TypeSpec); !ok {
			continue
		}

		if !shouldProceed(ctx, t.Name.Name) {
			continue
		}

		_, ok = t.Type.(*ast.StructType)
		if !ok {
			continue
		}

		cgroup := decl.Doc
		if cgroup == nil {
			continue
		}

		var svc Service
		svc.Name = t.Name.Name

		for _, c := range cgroup.List {
			if strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(c.Text, "//")), "+PostLookupFromPayloadHook") {
				svc.HasPostLookupFromPayloadHook = true
			}
			if strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(c.Text, "//")), "+PreUpdateFromPayloadHook") {
				svc.HasPreUpdateFromPayloadHook = true
			}
			if strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(c.Text, "//")), "+PostUpdateFromPayloadHook") {
				svc.HasPostUpdateFromPayloadHook = true
			}
			if strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(c.Text, "//")), "+PostCreateHook") {
				svc.HasPostCreateHook = true
			}
			if strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(c.Text, "//")), "+PostUpdateHook") {
				svc.HasPostUpdateHook = true
			}
			if strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(c.Text, "//")), "+PostDeleteHook") {
				svc.HasPostDeleteHook = true
			}
		}

		ctx.Services[svc.Name] = svc
	}

	return true
}

func getTypeName(ref ast.Expr) (string, error) {
	var typ string
	var err error
	switch ref.(type) {
	case *ast.Ident:
		typ = ref.(*ast.Ident).Name
	case *ast.SelectorExpr:
		se := ref.(*ast.SelectorExpr)
		typ = se.X.(*ast.Ident).Name + "." + se.Sel.Name
	case *ast.StarExpr:
		typ, err = getTypeName(ref.(*ast.StarExpr).X)
		if err != nil {
			return "", err
		}
		return "*" + typ, nil
	case *ast.ArrayType:
		typ = "[]" + ref.(*ast.ArrayType).Elt.(*ast.Ident).Name
	case *ast.MapType:
		mt := ref.(*ast.MapType)
		typ = "map[" + mt.Key.(*ast.Ident).Name + "]" + mt.Value.(*ast.Ident).Name
	default:
		fmt.Printf("%#v\n", ref)
		return "", errors.New("field type not supported")
	}
	return typ, nil
}

func (ctx *genctx) extractDBStructs(n ast.Node) bool {
	decl, ok := n.(*ast.GenDecl)
	if !ok {
		return true
	}

	if decl.Tok != token.TYPE {
		return true
	}

	for _, spec := range decl.Specs {
		var t *ast.TypeSpec
		var s *ast.StructType
		var ok bool

		if t, ok = spec.(*ast.TypeSpec); !ok {
			continue
		}

		if !shouldProceed(ctx, t.Name.Name) {
			continue
		}

		if s, ok = t.Type.(*ast.StructType); !ok {
			continue
		}

		st := DBRow{
			Columns: make(map[string]DBColumn),
			Name:    t.Name.Name,
			PkgName: ctx.PkgName,
		}

		for _, f := range s.Fields.List {
			if len(f.Names) == 0 {
				continue
			}

			if unicode.IsLower(rune(f.Names[0].Name[0])) {
				continue
			}

			typ, err := getTypeName(f.Type)
			if err != nil {
				return true
			}

			// If this is a Null* field, record that
			var nulltyp bool
			var basetyp string

			{
				// extract the package portion
				var prefix string
				if dotpos := strings.IndexRune(typ, '.'); dotpos > -1 {
					prefix = typ[:dotpos+1]
				}

				if i := strings.Index(typ, prefix+"Null"); i > -1 {
					nulltyp = true
					basetyp = typ[len(prefix)+i+4:]
				}
			}

			pdebug.Printf("--------> typ: %s", typ)
			pdebug.Printf("----> nulltyp: %t", nulltyp)
			pdebug.Printf("----> basetyp: %s", basetyp)

			column := DBColumn{
				BaseType:   basetyp,
				IsNullType: nulltyp,
				Name:       f.Names[0].Name,
				Type:       typ,
			}

			st.Columns[column.Name] = column
		}
		ctx.DBRows[st.Name] = st
	}

	return true
}

func generateFiles(ctx *genctx) error {
	for _, m := range ctx.Models {
		if pdebug.Enabled {
			pdebug.Printf("Checking model %s", m.Name)
		}
		if !shouldProceed(ctx, m.Name) {
			if pdebug.Enabled {
				pdebug.Printf("Skipping model %s", m.Name)
			}
			continue
		}

		if err := generateModelFile(ctx, m); err != nil {
			return err
		}

		if err := generateServiceFile(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func wrapConvertIf(b bool, out io.Writer, t, expr string) {
	if b {
		fmt.Fprintf(out, "%s(%s)", t, expr)
	} else {
		fmt.Fprint(out, expr)
	}
}

func generateModelFile(ctx *genctx, m Model) error {
	if pdebug.Enabled {
		g := pdebug.Marker("generateModelFile %s", m.Name)
		defer g.End()
	}

	row, ok := ctx.DBRows[m.Name]
	if !ok {
		return errors.New("could not find matching row for " + m.Name)
	}
	varname := 'v'
	hasL10N := false
	var l10nfields bytes.Buffer
	hasID := false
	for _, f := range m.Fields {
		if f.Name == "ID" {
			hasID = true
		}
		if f.L10N {
			hasL10N = true
			l10nfields.WriteString(strconv.Quote(f.JSONName))
			l10nfields.WriteString(",")
		}
	}
	var buf bytes.Buffer

	buf.WriteString("package model")
	buf.WriteString("\n\n// Automatically generated by genmodel utility. DO NOT EDIT!")
	buf.WriteString("\n\nimport (")
	if hasL10N {
		buf.WriteString("\n" + strconv.Quote("encoding/json"))
	}
	buf.WriteString("\n" + strconv.Quote("time"))
	buf.WriteString("\n\n" + strconv.Quote("github.com/builderscon/octav/octav/db"))
	buf.WriteString("\n" + strconv.Quote("github.com/lestrrat/go-pdebug"))
	buf.WriteString("\n)")
	buf.WriteString("\n\nvar _ = pdebug.Enabled")
	buf.WriteString("\nvar _ = time.Time{}")

	if hasL10N {
		fmt.Fprintf(&buf, "\n\ntype raw%s struct {", m.Name)
		for _, f := range m.Fields {
			if f.Type == "time.Time" && strings.Contains(f.Tag.Get("json"), "omitempty") {
				fmt.Fprintf(&buf, "\n%s *time.Time `%s`", f.Name, f.Tag)
			} else {
				fmt.Fprintf(&buf, "\n%s %s", f.Name, f.Type)
				if f.Tag != "" {
					fmt.Fprintf(&buf, "`%s`", f.Tag)
				}
			}
		}
		buf.WriteString("\n}")

		fmt.Fprintf(&buf, "\n\nfunc (%c %s) MarshalJSON() ([]byte, error) {", varname, m.Name)
		fmt.Fprintf(&buf, "\nvar raw raw%s", m.Name)
		for _, f := range m.Fields {
			if f.Type == "time.Time" && strings.Contains(f.Tag.Get("json"), "omitempty") {
				fmt.Fprintf(&buf, "\nif !v.%s.IsZero() {", f.Name)
				fmt.Fprintf(&buf, "\nraw.%s = &v.%s", f.Name, f.Name)
				buf.WriteString("\n}")
			} else {
				fmt.Fprintf(&buf, "\nraw.%s = v.%s", f.Name, f.Name)
			}
		}
		buf.WriteString("\nbuf, err := json.Marshal(raw)")
		buf.WriteString("\nif err != nil {")
		buf.WriteString("\nreturn nil, err")
		buf.WriteString("\n}")
		buf.WriteString("\nreturn MarshalJSONWithL10N(buf, v.LocalizedFields)")
		buf.WriteString("\n}")
	}

	if hasID {
		fmt.Fprintf(&buf, "\n\nfunc (%c *%s) Load(tx *db.Tx, id string) (err error) {", varname, m.Name)
		buf.WriteString("\nif pdebug.Enabled {")
		fmt.Fprintf(&buf, "\n"+`g := pdebug.Marker("model.%s.Load %%s", id).BindError(&err)`, m.Name)
		buf.WriteString("\ndefer g.End()")
		buf.WriteString("\n}")
		fmt.Fprintf(&buf, "\nvdb := db.%s{}", m.Name)
		buf.WriteString("\nif err := vdb.LoadByEID(tx, id); err != nil {")
		buf.WriteString("\nreturn err")
		buf.WriteString("\n}\n")
		buf.WriteString("\nif err := v.FromRow(&vdb); err != nil {")
		buf.WriteString("\nreturn err")
		buf.WriteString("\n}")
		buf.WriteString("\nreturn nil")
		buf.WriteString("\n}")

		fmt.Fprintf(&buf, "\n\nfunc (%c *%s) FromRow(vdb *db.%s) error {", varname, m.Name, m.Name)
		buf.WriteString("\nv.ID = vdb.EID")
		for _, f := range m.Fields {
			if f.Name == "ID" {
				continue
			}

			c, ok := row.Columns[f.Name]
			if !ok {
				continue
			}

			if c.IsNullType {
				fmt.Fprintf(&buf, "\nif vdb.%s.Valid {", f.Name)
				fmt.Fprintf(&buf, "\nv.%s = ", f.Name)
				wrapConvertIf(f.Convert, &buf, f.Type, fmt.Sprintf("vdb.%s.%s", f.Name, c.BaseType))
				buf.WriteString("\n}")
			} else {
				fmt.Fprintf(&buf, "\nv.%s = ", f.Name)
				wrapConvertIf(f.Convert, &buf, f.Type, fmt.Sprintf("vdb.%s", f.Name))
			}
		}
		buf.WriteString("\nreturn nil")
		buf.WriteString("\n}")
	}

	fmt.Fprintf(&buf, "\n\nfunc (%c *%s) ToRow(vdb *db.%s) error {", varname, m.Name, m.Name)
	for _, f := range m.Fields {
		if f.Name == "ID" {
			buf.WriteString("\nvdb.EID = v.ID")
		} else {
			c, ok := row.Columns[f.Name]
			if !ok {
				continue
			}

			if c.IsNullType {
				fmt.Fprintf(&buf, "\nvdb.%s.Valid = true", f.Name)
				fmt.Fprintf(&buf, "\nvdb.%s.%s = ", f.Name, c.BaseType)
				wrapConvertIf(f.Convert, &buf, strings.ToLower(c.BaseType), "v."+f.Name)
			} else {
				fmt.Fprintf(&buf, "\nvdb.%s = ", f.Name)
				wrapConvertIf(f.Convert, &buf, strings.ToLower(c.BaseType), "v."+f.Name)
			}
		}
	}
	buf.WriteString("\nreturn nil")
	buf.WriteString("\n}")
	/*
		if l10nfields.Len() > 0 {
			fmt.Fprintf(&buf, "\n\nfunc (%c %sL10N) GetPropNames() ([]string, error) {", varname, m.Name)
			fmt.Fprintf(&buf, "\nl, _ := %c.L10N.GetPropNames()", varname)
			buf.WriteString("\nreturn append(l, ")
			buf.WriteString(l10nfields.String())
			buf.WriteString("), nil")
			buf.WriteString("\n}")

			fmt.Fprintf(&buf, "\n\nfunc (%c %sL10N) GetPropValue(s string) (interface{}, error) {", varname, m.Name)
			buf.WriteString("\nswitch s {")
			for _, f := range m.Fields {
				fmt.Fprintf(&buf, "\ncase %s:", strconv.Quote(f.JSONName))
				fmt.Fprintf(&buf, "\nreturn %c.%s, nil", varname, f.Name)
			}
			buf.WriteString("\ndefault:")
			fmt.Fprintf(&buf, "\nreturn %c.L10N.GetPropValue(s)", varname)
			buf.WriteString("\n}\n}")

			fmt.Fprintf(&buf, "\n\nfunc (v *%sL10N) UnmarshalJSON(data []byte) error {", m.Name)
			fmt.Fprintf(&buf, "\nvar s %s", m.Name)
			buf.WriteString("\nif err := json.Unmarshal(data, &s); err != nil {")
			buf.WriteString("\nreturn err")
			buf.WriteString("\n}")
			fmt.Fprintf(&buf, "\n\nv.%s = s", m.Name)
			buf.WriteString("\nm := make(map[string]interface{})")
			buf.WriteString("\nif err := json.Unmarshal(data, &m); err != nil {")
			buf.WriteString("\nreturn err")
			buf.WriteString("\n}")
			fmt.Fprintf(&buf, "\n\nif err := tools.ExtractL10NFields(m, &v.L10N, []string{%s}); err != nil {", l10nfields.String())
			buf.WriteString("\nreturn err")
			buf.WriteString("\n}")
			buf.WriteString("\n\nreturn nil")
			buf.WriteString("\n}")

			fmt.Fprintf(&buf, "\n\nfunc (v *%sL10N) LoadLocalizedFields(tx *db.Tx) error {", m.Name)
			fmt.Fprintf(&buf, "\nls, err := db.LoadLocalizedStringsForParent(tx, v.%s.ID, %s)", m.Name, strconv.Quote(m.Name))
			buf.WriteString("\nif err != nil {")
			buf.WriteString("\nreturn err")
			buf.WriteString("\n}")
			buf.WriteString("\n\nif len(ls) > 0 {")
			buf.WriteString("\nv.L10N = tools.LocalizedFields{}")
			buf.WriteString("\nfor _, l := range ls {")
			buf.WriteString("\nv.L10N.Set(l.Language, l.Name, l.Localized)")
			buf.WriteString("\n}")
			buf.WriteString("\n}")
			buf.WriteString("\nreturn nil")
			buf.WriteString("\n}")
		}
	*/

	fsrc, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("%s", buf.Bytes())
		return err
	}

	fn := filepath.Join(ctx.Dir, "model", snakeCase(m.Name)+"_gen.go")
	if pdebug.Enabled {
		pdebug.Printf("Generating file %s", fn)
	}
	fi, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer fi.Close()

	if _, err := fi.Write(fsrc); err != nil {
		return err
	}

	return nil
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

func generateServiceFile(ctx *genctx, m Model) error {
	if pdebug.Enabled {
		g := pdebug.Marker("generateServiceFile %s", m.Name)
		defer g.End()
	}

	// Find the matching DBRow
	row, ok := ctx.DBRows[m.Name]
	if !ok {
		return errors.New("could not find matching row for " + m.Name)
	}

	svcname := m.Name + "Svc"
	svc := ctx.Services[svcname]

	hasID := false
	for _, f := range m.Fields {
		if f.Name == "ID" {
			hasID = true
		}
	}
	colnames := make([]string, 0, len(row.Columns))
	for k := range row.Columns {
		colnames = append(colnames, k)
	}
	sort.Strings(colnames)

	buf := bytes.Buffer{}

	hasL10N := false
	hasDecorate := false
	for _, f := range m.Fields {
		if f.L10N {
			hasL10N = true
		}
		if f.Decorate {
			hasDecorate = true
		}
	}

	buf.WriteString("package service")
	buf.WriteString("\n\n// Automatically generated by genmodel utility. DO NOT EDIT!")
	buf.WriteString("\n\nimport (")
	buf.WriteString("\n" + strconv.Quote("context"))
	buf.WriteString("\n" + strconv.Quote("sync"))
	buf.WriteString("\n" + strconv.Quote("time"))
	buf.WriteString("\n\n" + strconv.Quote("github.com/builderscon/octav/octav/cache"))
	buf.WriteString("\n\n" + strconv.Quote("github.com/builderscon/octav/octav/db"))
	buf.WriteString("\n" + strconv.Quote("github.com/builderscon/octav/octav/internal/errors"))
	buf.WriteString("\n" + strconv.Quote("github.com/builderscon/octav/octav/model"))
	buf.WriteString("\n" + strconv.Quote("github.com/lestrrat/go-pdebug"))
	buf.WriteString("\n)")
	buf.WriteString("\n\nvar _ = time.Time{}")
	buf.WriteString("\nvar _ = cache.WithExpires(time.Minute)")
	buf.WriteString("\nvar _ = context.Background")
	buf.WriteString("\nvar _ = errors.Wrap")
	fmt.Fprintf(&buf, "\nvar _ = model.%s{}", m.Name)
	fmt.Fprintf(&buf, "\nvar _ = db.%s{}", m.Name)
	buf.WriteString("\nvar _ = pdebug.Enabled")

	svcvarname := lowerFirst(svcname)
	oncename := lowerFirst(m.Name + "Once")
	fmt.Fprintf(&buf, "\n\nvar %s %s", svcvarname, svcname)
	fmt.Fprintf(&buf, "\nvar %s sync.Once", oncename)
	fmt.Fprintf(&buf, "\nfunc %s() *%s {", m.Name, svcname)
	fmt.Fprintf(&buf, "\n%s.Do(%s.Init)", oncename, svcvarname)
	fmt.Fprintf(&buf, "\nreturn &%s", svcvarname)
	buf.WriteString("\n}")

	if hasID && m.LookupRequest {
		fmt.Fprintf(&buf, "\n\nfunc (v *%s) LookupFromPayload(tx *db.Tx, m *model.%s, payload *model.Lookup%sRequest) (err error) {", svcname, m.Name, m.Name)
		buf.WriteString("\nif pdebug.Enabled {")
		fmt.Fprintf(&buf, "\n"+`g := pdebug.Marker("service.%s.LookupFromPayload").BindError(&err)`, m.Name)
		buf.WriteString("\ndefer g.End()")
		buf.WriteString("\n}")
		buf.WriteString("\nif err = v.Lookup(tx, m, payload.ID); err != nil {")
		fmt.Fprintf(&buf, "\n"+`return errors.Wrap(err, "failed to load model.%s from database")`, m.Name)
		buf.WriteString("\n}")
		if hasL10N || hasDecorate {
			buf.WriteString("\nif err := v.Decorate(tx, m, payload.TrustedCall, payload.Lang.String); err != nil {")
			fmt.Fprintf(&buf, "\n"+`return errors.Wrap(err, "failed to load associated data for model.%s from database")`, m.Name)
			buf.WriteString("\n}")
		}
		buf.WriteString("\nreturn nil")
		buf.WriteString("\n}")
	}

	if hasID && m.Lookup {
		fmt.Fprintf(&buf, "\n\nfunc (v *%s) Lookup(tx *db.Tx, m *model.%s, id string) (err error) {", svcname, m.Name)
		buf.WriteString("\nif pdebug.Enabled {")
		fmt.Fprintf(&buf, "\n"+`g := pdebug.Marker("service.%s.Lookup").BindError(&err)`, m.Name)
		buf.WriteString("\ndefer g.End()")
		buf.WriteString("\n}")
		fmt.Fprintf(&buf, "\n\nvar r model.%s", m.Name)
		buf.WriteString("\nc := Cache()")
		fmt.Fprintf(&buf, "\nkey := c.Key(%s, id)", strconv.Quote(m.Name))
		buf.WriteString("\nvar cacheMiss bool")
		buf.WriteString("\n_, err = c.GetOrSet(key, &r, func() (interface{}, error) {")
		buf.WriteString("\nif pdebug.Enabled {")
		buf.WriteString("\ncacheMiss = true")
		buf.WriteString("\n}")
		buf.WriteString("\nif err := r.Load(tx, id); err != nil {")
		fmt.Fprintf(&buf, "\n"+`return nil, errors.Wrap(err, "failed to load model.%s from database")`, m.Name)
		buf.WriteString("\n}")
		buf.WriteString("\nreturn &r, nil")
		buf.WriteString("\n}, cache.WithExpires(time.Hour))")
		buf.WriteString("\nif pdebug.Enabled {")
		buf.WriteString("\ncacheSt := `HIT`")
		buf.WriteString("\nif cacheMiss {")
		buf.WriteString("\ncacheSt = `MISS`")
		buf.WriteString("\n}")
		buf.WriteString("\npdebug.Printf(`CACHE %s: %s`, cacheSt, key)")
		buf.WriteString("\n}")

		if svc.HasPostLookupFromPayloadHook {
			buf.WriteString("\nif err = v.PostLookupFromPayloadHook(tx, &r); err != nil {")
			buf.WriteString("\nreturn errors.Wrap(err, \"failed to execute PostLookupFromPayloadHook\")")
			buf.WriteString("\n}")
		}

		buf.WriteString("\n*m = r")
		buf.WriteString("\nreturn nil")
		buf.WriteString("\n}")
	}

	if m.CreateRequest {
		buf.WriteString("\n\n// Create takes in the transaction, the incoming payload, and a reference to")
		buf.WriteString("\n// a database row. The database row is initialized/populated so that the")
		buf.WriteString("\n// caller can use it afterwards.")
		fmt.Fprintf(&buf, "\nfunc (v *%s) Create(tx *db.Tx, vdb *db.%s, payload *model.Create%sRequest) (err error) {", svcname, m.Name, m.Name)
		buf.WriteString("\nif pdebug.Enabled {")
		fmt.Fprintf(&buf, "\n"+`g := pdebug.Marker("service.%s.Create").BindError(&err)`, m.Name)
		buf.WriteString("\ndefer g.End()")
		buf.WriteString("\n}")
		buf.WriteString("\n\nif err := v.populateRowForCreate(vdb, payload); err != nil {")
		buf.WriteString("\nreturn errors.Wrap(err, `failed to populate row`)")
		buf.WriteString("\n}")
		buf.WriteString("\n\nif err := vdb.Create(tx, payload.DatabaseOptions...); err != nil {")
		buf.WriteString("\nreturn errors.Wrap(err, `failed to insert into database`)")
		buf.WriteString("\n}\n")
		if hasL10N {
			fmt.Fprintf(&buf, "\nif err := payload.LocalizedFields.CreateLocalizedStrings(tx, %s, vdb.EID); err != nil {", strconv.Quote(m.Name))
			buf.WriteString("\nreturn errors.Wrap(err, `failed to populate localized strings`)")
			buf.WriteString("\n}")
		}

		if svc.HasPostCreateHook {
			buf.WriteString("\nif err := v.PostCreateHook(tx, vdb); err != nil {")
			buf.WriteString("\nreturn errors.Wrap(err, `post create hook failed`)")
			buf.WriteString("\n}")
		}
		buf.WriteString("\nreturn nil")
		buf.WriteString("\n}")
	}

	if hasID && m.Update {
		fmt.Fprintf(&buf, "\n\nfunc (v *%s) Update(tx *db.Tx, vdb *db.%s) (err error) {", svcname, m.Name)
		buf.WriteString("\nif pdebug.Enabled {")
		fmt.Fprintf(&buf, "\n"+`g := pdebug.Marker("service.%s.Update (%%s)", vdb.EID).BindError(&err)`, m.Name)
		buf.WriteString("\ndefer g.End()")
		buf.WriteString("\n}")
		buf.WriteString("\n\nif vdb.EID == `` {")
		fmt.Fprintf(
			&buf,
			"\nreturn errors.New(%s)",
			strconv.Quote("vdb.EID is required (did you forget to call vdb.Load(tx) before hand?)"),
		)
		buf.WriteString("\n}")
		buf.WriteString("\n\nif err := vdb.Update(tx); err != nil {")
		buf.WriteString("\nreturn errors.Wrap(err, `failed to update database`)")
		buf.WriteString("\n}")
		buf.WriteString("\nc := Cache()")
		fmt.Fprintf(&buf, "\nkey := c.Key(%s, vdb.EID)", strconv.Quote(m.Name))
		buf.WriteString("\nif pdebug.Enabled {")
		buf.WriteString("\npdebug.Printf(`CACHE DEL %s`, key)")
		buf.WriteString("\n}")
		buf.WriteString("\ncerr := c.Delete(key)")
		buf.WriteString("\nif pdebug.Enabled {")
		buf.WriteString("\nif cerr != nil {")
		buf.WriteString("\npdebug.Printf(`CACHE ERR: %s`, cerr)")
		buf.WriteString("\n}")
		buf.WriteString("\n}")
		if svc.HasPostUpdateHook {
			buf.WriteString("\nif err := v.PostUpdateHook(tx, vdb); err != nil {")
			buf.WriteString("\nreturn errors.Wrap(err, `post update hook failed`)")
			buf.WriteString("\n}")
		}
		buf.WriteString("\nreturn nil")
		buf.WriteString("\n}")
	}

	if hasID && m.UpdateRequest {
		fmt.Fprintf(&buf, "\n\nfunc (v *%s) UpdateFromPayload(ctx context.Context, tx *db.Tx, payload *model.Update%sRequest) (err error) {", svcname, m.Name)
		buf.WriteString("\nif pdebug.Enabled {")
		fmt.Fprintf(&buf, "\n"+`g := pdebug.Marker("service.%s.UpdateFromPayload (%%s)", payload.ID).BindError(&err)`, m.Name)
		buf.WriteString("\ndefer g.End()")
		buf.WriteString("\n}")

		fmt.Fprintf(&buf, "\nvar vdb db.%s", m.Name)
		buf.WriteString("\nif err := vdb.LoadByEID(tx, payload.ID); err != nil {")
		buf.WriteString("\nreturn errors.Wrap(err, `failed to load from database`)")
		buf.WriteString("\n}")

		if svc.HasPreUpdateFromPayloadHook {
			// Note: must pass row from database
			buf.WriteString("\n\nif err := v.PreUpdateFromPayloadHook(ctx, tx, &vdb, payload); err != nil {")
			buf.WriteString("\nreturn errors.Wrap(err, `failed to execute PreUpdateFromPayloadHook`)")
			buf.WriteString("\n}")
		}

		buf.WriteString("\n\nif err := v.populateRowForUpdate(&vdb, payload); err != nil {")
		buf.WriteString("\nreturn errors.Wrap(err, `failed to populate row data`)")
		buf.WriteString("\n}")
		buf.WriteString("\n\nif err := v.Update(tx, &vdb); err != nil {")
		buf.WriteString("\nreturn errors.Wrap(err, `failed to update row in database`)")
		buf.WriteString("\n}")
		if svc.HasPostUpdateFromPayloadHook {
			buf.WriteString("\n\nif err := v.PostUpdateFromPayloadHook(ctx, tx, &vdb, payload); err != nil {")
			buf.WriteString("\nreturn errors.Wrap(err, `failed to execute PostUpdateFromPayloadHook`)")
			buf.WriteString("\n}")
		}
		if hasL10N {
			buf.WriteString("\n\nls := LocalizedString()")
			fmt.Fprintf(&buf, "\nif err := ls.UpdateFields(tx, %s, vdb.EID, payload.LocalizedFields); err != nil {", strconv.Quote(m.Name))
			buf.WriteString("\nreturn errors.Wrap(err, `failed to update localized fields`)")
			buf.WriteString("\n}")
		}
		buf.WriteString("\nreturn nil")
		buf.WriteString("\n}")
	}

	if hasL10N {
		fmt.Fprintf(&buf, "\n\nfunc (v *%s) ReplaceL10NStrings(tx *db.Tx, m *model.%s, lang string) error {", svcname, m.Name)
		buf.WriteString("\nif pdebug.Enabled {")
		fmt.Fprintf(&buf, "\n"+`g := pdebug.Marker("service.%s.ReplaceL10NStrings lang = %%s", lang)`, m.Name)
		buf.WriteString("\ndefer g.End()")
		buf.WriteString("\n}")

		var l10nf []string
		for _, f := range m.Fields {
			if !f.L10N {
				continue
			}
			l10nf = append(l10nf, "len(m."+f.Name+") > 0")
		}
		buf.WriteString("\nls := LocalizedString()")
		fmt.Fprintf(&buf, "\nlist := make([]db.LocalizedString, 0, %d)", len(l10nf))
		buf.WriteString("\nswitch lang {")
		buf.WriteString("\ncase \"\", \"en\":")
		buf.WriteString("\nif ")
		buf.WriteString(strings.Join(l10nf, " && "))
		buf.WriteString("{\nreturn nil\n}")

		buf.WriteString("\nfor _, extralang := range []string{`ja`} {")
		buf.WriteString("\nlist = list[:0]")
		fmt.Fprintf(&buf, "\nif err := ls.LookupFields(tx, %s, m.ID, extralang, &list); err != nil {", strconv.Quote(m.Name))
		buf.WriteString("\nreturn errors.Wrap(err, `failed to lookup localized fields`)")
		buf.WriteString("\n}")

		buf.WriteString("\n\nfor _, l := range list {")
		buf.WriteString("\nswitch l.Name {")
		for _, f := range m.Fields {
			if !f.L10N {
				continue
			}
			fmt.Fprintf(&buf, "\ncase %s:", strconv.Quote(f.JSONName))
			fmt.Fprintf(&buf, "\nif len(m.%s) == 0 {", f.Name)
			buf.WriteString("\nif pdebug.Enabled {")
			fmt.Fprintf(&buf, "\n"+`pdebug.Printf("Replacing for key '%s' (fallback en -> %%s", l.Language)`, f.JSONName)
			buf.WriteString("\n}")
			fmt.Fprintf(&buf, "\nm.%s = l.Localized", f.Name)
			buf.WriteString("\n}")
		}
		buf.WriteString("\n}")
		buf.WriteString("\n}")
		buf.WriteString("\n}")
		buf.WriteString("\nreturn nil")
		buf.WriteString("\ncase \"all\":")
		buf.WriteString("\nfor _, extralang := range []string{`ja`} {")
		buf.WriteString("\nlist = list[:0]")
		fmt.Fprintf(&buf, "\nif err := ls.LookupFields(tx, %s, m.ID, extralang, &list); err != nil {", strconv.Quote(m.Name))
		buf.WriteString("\nreturn errors.Wrap(err, `failed to lookup localized fields`)")
		buf.WriteString("\n}")

		buf.WriteString("\n\nfor _, l := range list {")
		buf.WriteString("\nif pdebug.Enabled {")
		buf.WriteString("\npdebug.Printf(\"Adding key '%s#%s'\", l.Name, l.Language)")
		buf.WriteString("\n}")
		buf.WriteString("\nm.LocalizedFields.Set(l.Language, l.Name, l.Localized)")
		buf.WriteString("\n}")
		buf.WriteString("\n}")
		buf.WriteString("\ndefault:")
		buf.WriteString("\nfor _, extralang := range []string{`ja`} {")
		buf.WriteString("\nlist = list[:0]")
		fmt.Fprintf(&buf, "\nif err := ls.LookupFields(tx, %s, m.ID, extralang, &list); err != nil {", strconv.Quote(m.Name))
		buf.WriteString("\nreturn errors.Wrap(err, `failed to lookup localized fields`)")
		buf.WriteString("\n}")

		buf.WriteString("\n\nfor _, l := range list {")
		buf.WriteString("\nswitch l.Name {")
		for _, f := range m.Fields {
			if !f.L10N {
				continue
			}
			fmt.Fprintf(&buf, "\ncase %s:", strconv.Quote(f.JSONName))
			buf.WriteString("\nif pdebug.Enabled {")
			fmt.Fprintf(&buf, "\n"+`pdebug.Printf("Replacing for key '%s'")`, f.JSONName)
			buf.WriteString("\n}")
			fmt.Fprintf(&buf, "\nm.%s = l.Localized", f.Name)
		}
		buf.WriteString("\n}")
		buf.WriteString("\n}")
		buf.WriteString("\n}")
		buf.WriteString("\n}")
		buf.WriteString("\nreturn nil")
		buf.WriteString("\n}")
	}

	if hasID && m.Delete {
		fmt.Fprintf(&buf, "\n\nfunc (v *%s) Delete(tx *db.Tx, id string) error {", svcname)
		buf.WriteString("\nif pdebug.Enabled {")
		fmt.Fprintf(&buf, "\n"+`g := pdebug.Marker("%s.Delete (%%s)", id)`, m.Name)
		buf.WriteString("\ndefer g.End()")
		buf.WriteString("\n}")
		if svc.HasPostDeleteHook {
			// If there's a delete hook, we need the original contents to
			// pass to the hook. therefore we must make this one extra call
			fmt.Fprintf(&buf, "\noriginal := db.%s{EID: id}", m.Name)
			buf.WriteString("\nif err := original.LoadByEID(tx, id); err != nil {")
			buf.WriteString("\nreturn errors.Wrap(err, `failed load before delete`)")
			buf.WriteString("\n}")
		}
		fmt.Fprintf(&buf, "\n\nvdb := db.%s{EID: id}", m.Name)
		buf.WriteString("\nif err := vdb.Delete(tx); err != nil {")
		buf.WriteString("\nreturn errors.Wrap(err, `failed to delete from database`)")
		buf.WriteString("\n}")

		buf.WriteString("\nc := Cache()")
		fmt.Fprintf(&buf, "\nkey := c.Key(%s, id)", strconv.Quote(m.Name))
		buf.WriteString("\nc.Delete(key)")
		buf.WriteString("\nif pdebug.Enabled {")
		buf.WriteString("\npdebug.Printf(`CACHE DEL %s`, key)")
		buf.WriteString("\n}")
		if hasL10N {
			fmt.Fprintf(&buf, "\nif err := db.DeleteLocalizedStringsForParent(tx, id, %s); err != nil {", strconv.Quote(m.Name))
			buf.WriteString("\nreturn errors.Wrap(err, `failed to delete localized strings`)")
			buf.WriteString("\n}")
		}
		if svc.HasPostDeleteHook {
			buf.WriteString("\nif err := v.PostDeleteHook(tx, &original); err != nil {")
			buf.WriteString("\nreturn errors.Wrap(err, `post delete hook failed`)")
			buf.WriteString("\n}")
		}
		buf.WriteString("\nreturn nil")
		buf.WriteString("\n}")
	}

	fsrc, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("%s", buf.Bytes())
		return err
	}

	fn := filepath.Join(ctx.Dir, "service", snakeCase(m.Name)+"_gen.go")
	if pdebug.Enabled {
		pdebug.Printf("Generating file %s", fn)
	}
	fi, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer fi.Close()

	if _, err := fi.Write(fsrc); err != nil {
		return err
	}

	return nil
}
