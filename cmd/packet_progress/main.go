package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	_ "github.com/ggmolly/belfast/internal/protobuf"
)

const (
	statusImplemented = "implemented"
	statusPartial     = "partial"
	statusStub        = "stub"
	statusPanic       = "panic"
	statusMissing     = "missing"
)

type heuristicsConfig struct {
	Weights    heuristicWeights    `json:"weights"`
	Thresholds heuristicThresholds `json:"thresholds"`
}

type heuristicWeights struct {
	SendMessage  int `json:"send_message"`
	ResponseType int `json:"response_struct"`
	RequestType  int `json:"request_struct"`
	ProtoSetter  int `json:"proto_setter"`
	RequestParse int `json:"request_parse"`
	ClientUsage  int `json:"client_usage"`
	CommanderUse int `json:"commander_usage"`
	ORMUsage     int `json:"orm_usage"`
	MiscUsage    int `json:"misc_usage"`
	DBWrite      int `json:"db_write"`
}

type heuristicThresholds struct {
	ImplementedMin int `json:"implemented_min"`
}

type packetReport struct {
	ID             int             `json:"id"`
	Status         string          `json:"status"`
	ComputedStatus string          `json:"computed_status"`
	Score          int             `json:"score"`
	Signals        []string        `json:"signals"`
	Handlers       []handlerReport `json:"handlers"`
	Override       string          `json:"override,omitempty"`
}

type handlerReport struct {
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Score   int      `json:"score"`
	Signals []string `json:"signals"`
	File    string   `json:"file"`
	Line    int      `json:"line"`
}

type responseReport struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Files []string `json:"files"`
}

type report struct {
	GeneratedAt  string            `json:"generated_at"`
	Total        int               `json:"total_registered"`
	TotalKnown   int               `json:"total_known"`
	TotalKnownCS int               `json:"total_known_cs,omitempty"`
	TotalKnownSC int               `json:"total_known_sc,omitempty"`
	Missing      int               `json:"missing"`
	MissingCS    int               `json:"missing_cs,omitempty"`
	MissingSC    int               `json:"missing_sc,omitempty"`
	MissingIDs   []int             `json:"missing_ids"`
	MissingCSIDs []int             `json:"missing_cs_ids,omitempty"`
	MissingSCIDs []int             `json:"missing_sc_ids,omitempty"`
	Counts       map[string]int    `json:"counts"`
	Packets      []packetReport    `json:"packets"`
	Responses    []responseReport  `json:"responses"`
	Overrides    map[string]string `json:"overrides"`
}

type importAliases struct {
	Protobuf   []string
	Proto      []string
	ORM        []string
	Misc       []string
	Connection []string
}

type handlerSource struct {
	Name    string
	File    string
	Line    int
	Decl    *ast.FuncDecl
	Imports importAliases
	FileSet *token.FileSet
}

type handlerExpr struct {
	Name   string
	Inline *ast.FuncLit
	File   string
	Line   int
}

type packetRegistration struct {
	ID       int
	Handlers []handlerExpr
	File     string
	Line     int
	Imports  importAliases
}

type analysisResult struct {
	Status  string
	Score   int
	Signals map[string]bool
}

type analysisContext struct {
	Imports          importAliases
	ClientParamNames map[string]bool
}

func defaultHeuristics() heuristicsConfig {
	return heuristicsConfig{
		Weights: heuristicWeights{
			SendMessage:  3,
			ResponseType: 2,
			RequestType:  1,
			ProtoSetter:  1,
			RequestParse: 1,
			ClientUsage:  1,
			CommanderUse: 2,
			ORMUsage:     2,
			MiscUsage:    1,
			DBWrite:      2,
		},
		Thresholds: heuristicThresholds{
			ImplementedMin: 4,
		},
	}
}

func main() {
	pngMode := false
	if len(os.Args) > 1 && os.Args[1] == "png" {
		pngMode = true
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	mainPath := flag.String("main", "cmd/belfast/main.go", "path to a file in the repo (used to locate go.mod)")
	outJSON := flag.String("out-json", "docs/packet-progress.json", "output json report")
	outSVG := flag.String("out-svg", "docs/packet-progress.svg", "output svg progress bar")
	outPNG := flag.String("out-png", "", "output png progress (requires rsvg-convert)")
	pngScale := flag.Float64("png-scale", 1, "png scale factor")
	fontFamily := flag.String("font-family", "Verdana, Arial, sans-serif", "svg font family")
	overridesPath := flag.String("overrides", "cmd/packet_progress/overrides.json", "override status map")
	heuristicsPath := flag.String("heuristics", "cmd/packet_progress/heuristics.json", "heuristics config")
	includeCS := flag.Bool("cs", false, "track CS_ packet types (commands)")
	includeSC := flag.Bool("sc", false, "track SC_ packet types (responses)")
	includeBoth := flag.Bool("both", false, "track both CS_ and SC_ packet types")
	flag.Parse()

	// Default behavior: track both CS_ and SC_ unless explicitly requested otherwise.
	if *includeBoth {
		*includeCS = true
		*includeSC = true
	}
	if !*includeCS && !*includeSC {
		*includeCS = true
		*includeSC = true
	}

	if pngMode && *outPNG == "" {
		*outPNG = replaceExt(*outSVG, ".png")
	}

	cfg := defaultHeuristics()
	if err := loadJSONIfExists(*heuristicsPath, &cfg); err != nil {
		exitWithError("failed to load heuristics config", err)
	}

	overrides, err := loadOverrides(*overridesPath)
	if err != nil {
		exitWithError("failed to load overrides", err)
	}

	repoRoot, err := findRepoRoot(filepath.Dir(*mainPath))
	if err != nil {
		exitWithError("failed to locate repo root", err)
	}

	modulePath, err := loadModulePath(repoRoot)
	if err != nil {
		exitWithError("failed to load module path", err)
	}

	registrations, err := collectRegistrations(repoRoot)
	if err != nil {
		exitWithError("failed to collect registrations", err)
	}

	handlers, err := loadAnswerHandlers("internal/answer")
	if err != nil {
		exitWithError("failed to load handlers", err)
	}

	constValues, err := collectConstValues(repoRoot, modulePath)
	if err != nil {
		exitWithError("failed to collect constants", err)
	}
	responsePackets, err := collectResponsePackets(repoRoot, modulePath, constValues)
	if err != nil {
		exitWithError("failed to collect response packets", err)
	}
	responseReports := buildResponseReports(responsePackets, buildPacketTypeNameMap("SC_"))

	packetReports := make([]packetReport, 0, len(registrations))

	for _, registration := range registrations {
		handlerReports := make([]handlerReport, 0, len(registration.Handlers))
		for _, handler := range registration.Handlers {
			if handler.Inline != nil {
				result := analyzeFunction(handler.Inline.Type, handler.Inline.Body, analysisContext{
					Imports:          registration.Imports,
					ClientParamNames: clientParamNames(handler.Inline.Type, registration.Imports),
				}, cfg)
				handlerReports = append(handlerReports, handlerReport{
					Name:    handler.Name,
					Status:  result.Status,
					Score:   result.Score,
					Signals: sortedSignals(result.Signals),
					File:    handler.File,
					Line:    handler.Line,
				})
				continue
			}

			lookupName := handlerLookupName(handler.Name)
			source, ok := handlers[lookupName]
			if !ok {
				handlerReports = append(handlerReports, handlerReport{
					Name:    handler.Name,
					Status:  statusStub,
					Score:   0,
					Signals: []string{"missing_handler"},
					File:    handler.File,
					Line:    handler.Line,
				})
				continue
			}
			result := analyzeFunction(source.Decl.Type, source.Decl.Body, analysisContext{
				Imports:          source.Imports,
				ClientParamNames: clientParamNames(source.Decl.Type, source.Imports),
			}, cfg)
			handlerReports = append(handlerReports, handlerReport{
				Name:    handler.Name,
				Status:  result.Status,
				Score:   result.Score,
				Signals: sortedSignals(result.Signals),
				File:    source.File,
				Line:    source.Line,
			})
		}

		combined := combineHandlerReports(handlerReports)
		packet := packetReport{
			ID:             registration.ID,
			Status:         combined.Status,
			ComputedStatus: combined.Status,
			Score:          combined.Score,
			Signals:        sortedSignals(combined.Signals),
			Handlers:       handlerReports,
		}
		if override, ok := overrides[strconv.Itoa(registration.ID)]; ok {
			packet.Override = override
			packet.Status = override
		}
		packetReports = append(packetReports, packet)
	}

	sort.Slice(packetReports, func(i, j int) bool {
		return packetReports[i].ID < packetReports[j].ID
	})

	registeredIDs := packetIDSet(packetReports)
	coveredResponseIDs := responseIDSet(responseReports)

	totalKnownCS := 0
	totalKnownSC := 0
	missingCSIDs := []int{}
	missingSCIDs := []int{}
	coveredSCCount := 0

	if *includeCS {
		totalKnownCS = countKnownPacketTypes("CS_")
		missingCSIDs = missingIDsForPrefixes(registeredIDs, "CS_")
	}
	if *includeSC {
		totalKnownSC = countKnownPacketTypes("SC_")
		missingSCIDs = missingIDsForPrefixes(coveredResponseIDs, "SC_")
		coveredSCCount = totalKnownSC - len(missingSCIDs)
		if coveredSCCount < 0 {
			coveredSCCount = 0
		}
	}

	totalKnown := totalKnownCS + totalKnownSC
	missingIDs := unionSortedInts(missingCSIDs, missingSCIDs)
	missing := len(missingCSIDs) + len(missingSCIDs)
	if missing < 0 {
		missing = 0
	}

	counts := map[string]int{
		statusImplemented: 0,
		statusPartial:     0,
		statusStub:        0,
		statusPanic:       0,
		statusMissing:     missing,
	}
	if *includeCS {
		knownCSIDs := knownPacketIDs("CS_")
		statusByID := combinePacketStatuses(packetReports)
		for id, status := range statusByID {
			if !knownCSIDs[id] {
				continue
			}
			counts[status]++
		}
	}
	if *includeSC {
		// For now, treat observed response packets as implemented for progress.
		counts[statusImplemented] += coveredSCCount
	}

	generated := report{
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		Total:        len(packetReports),
		TotalKnown:   totalKnown,
		TotalKnownCS: totalKnownCS,
		TotalKnownSC: totalKnownSC,
		Missing:      missing,
		MissingCS:    len(missingCSIDs),
		MissingSC:    len(missingSCIDs),
		MissingIDs:   missingIDs,
		MissingCSIDs: missingCSIDs,
		MissingSCIDs: missingSCIDs,
		Counts:       counts,
		Packets:      packetReports,
		Responses:    responseReports,
		Overrides:    overrides,
	}

	if err := writeJSON(*outJSON, generated); err != nil {
		exitWithError("failed to write json", err)
	}
	if err := writeSVG(*outSVG, counts, totalKnown, *fontFamily); err != nil {
		exitWithError("failed to write svg", err)
	}
	if *outPNG != "" {
		if err := writePNG(*outSVG, *outPNG, *pngScale); err != nil {
			exitWithError("failed to write png", err)
		}
	}

	outputs := []string{*outJSON, *outSVG}
	if *outPNG != "" {
		outputs = append(outputs, *outPNG)
	}
	fmt.Printf("wrote %s\n", strings.Join(outputs, ", "))
}

func parseFile(path string) (*ast.File, *token.FileSet, importAliases, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, nil, importAliases{}, err
	}
	return file, fset, buildImportAliases(file), nil
}

func collectRegistrations(root string) ([]packetRegistration, error) {
	registrations := []packetRegistration{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if shouldSkipDir(root, path, entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, fset, imports, err := parseFile(path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		reportPath := path
		if rel, err := filepath.Rel(root, path); err == nil {
			reportPath = rel
		}
		fileRegistrations, err := extractRegistrations(file, fset, reportPath, imports)
		if err != nil {
			return fmt.Errorf("extract registrations from %s: %w", path, err)
		}
		registrations = append(registrations, fileRegistrations...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return registrations, nil
}

func findRepoRoot(start string) (string, error) {
	absolute, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	current := absolute
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("go.mod not found from %s", start)
		}
		current = parent
	}
}

func buildImportAliases(file *ast.File) importAliases {
	aliases := importAliases{}
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		name := importName(spec, path)
		switch path {
		case "github.com/ggmolly/belfast/internal/protobuf":
			aliases.Protobuf = append(aliases.Protobuf, name)
		case "google.golang.org/protobuf/proto":
			aliases.Proto = append(aliases.Proto, name)
		case "github.com/ggmolly/belfast/internal/orm":
			aliases.ORM = append(aliases.ORM, name)
		case "github.com/ggmolly/belfast/internal/misc":
			aliases.Misc = append(aliases.Misc, name)
		case "github.com/ggmolly/belfast/internal/connection":
			aliases.Connection = append(aliases.Connection, name)
		}
	}
	return aliases
}

func loadModulePath(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module path not found in go.mod")
}

func collectConstValues(root string, modulePath string) (map[string]int, error) {
	values := map[string]int{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if shouldSkipDir(root, path, entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		importPath := packageImportPath(root, modulePath, path)
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}
			for _, spec := range gen.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range valueSpec.Names {
					if i >= len(valueSpec.Values) {
						continue
					}
					value, ok := parseIntLiteral(valueSpec.Values[i])
					if !ok {
						continue
					}
					values[importPath+"."+name.Name] = value
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}

func collectResponsePackets(root string, modulePath string, constValues map[string]int) (map[int]map[string]bool, error) {
	responses := map[int]map[string]bool{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if shouldSkipDir(root, path, entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		imports := buildImportMap(file)
		importPath := packageImportPath(root, modulePath, path)
		relativePath := path
		if rel, err := filepath.Rel(root, path); err == nil {
			relativePath = rel
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if !isSendCall(call) {
				return true
			}
			if len(call.Args) == 0 {
				return true
			}
			packetID, ok := resolvePacketID(call.Args[0], imports, importPath, constValues)
			if !ok {
				return true
			}
			addResponseUsage(responses, packetID, relativePath)
			return true
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return responses, nil
}

func shouldSkipDir(root string, path string, name string) bool {
	if name == ".git" || name == "vendor" {
		return true
	}
	if path == filepath.Join(root, "internal", "protobuf") {
		return true
	}
	return false
}

func addResponseUsage(responses map[int]map[string]bool, packetID int, file string) {
	files, ok := responses[packetID]
	if !ok {
		files = map[string]bool{}
		responses[packetID] = files
	}
	files[file] = true
}

func buildResponseReports(responses map[int]map[string]bool, nameMap map[int]string) []responseReport {
	reports := make([]responseReport, 0, len(responses))
	for id, files := range responses {
		name := nameMap[id]
		if name == "" {
			name = fmt.Sprintf("SC_%d", id)
		}
		fileList := make([]string, 0, len(files))
		for file := range files {
			fileList = append(fileList, file)
		}
		sort.Strings(fileList)
		reports = append(reports, responseReport{
			ID:    id,
			Name:  name,
			Files: fileList,
		})
	}
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].ID < reports[j].ID
	})
	return reports
}

func countResponseOnly(responses []responseReport, packets []packetReport) int {
	if len(responses) == 0 {
		return 0
	}
	registered := make(map[int]bool, len(packets))
	for _, packet := range packets {
		registered[packet.ID] = true
	}
	count := 0
	for _, response := range responses {
		if registered[response.ID] {
			continue
		}
		count++
	}
	return count
}

func buildPacketTypeNameMap(prefix string) map[int]string {
	names := map[int]string{}
	protoregistry.GlobalTypes.RangeMessages(func(desc protoreflect.MessageType) bool {
		name := string(desc.Descriptor().Name())
		if !strings.HasPrefix(name, prefix) {
			return true
		}
		id, ok := parsePacketTypeID(name)
		if !ok {
			return true
		}
		names[id] = name
		return true
	})
	return names
}

func parsePacketTypeID(name string) (int, bool) {
	value := strings.TrimPrefix(name, "SC_")
	id, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return id, true
}

func buildImportMap(file *ast.File) map[string]string {
	imports := map[string]string{}
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		imports[importName(spec, path)] = path
	}
	return imports
}

func packageImportPath(root string, modulePath string, filePath string) string {
	relativeDir, err := filepath.Rel(root, filepath.Dir(filePath))
	if err != nil || relativeDir == "." {
		return modulePath
	}
	return filepath.ToSlash(filepath.Join(modulePath, relativeDir))
}

func resolvePacketID(expr ast.Expr, imports map[string]string, importPath string, constValues map[string]int) (int, bool) {
	switch value := expr.(type) {
	case *ast.BasicLit:
		return parseBasicInt(value)
	case *ast.UnaryExpr:
		return parseUnaryInt(value)
	case *ast.Ident:
		if id, ok := constValues[importPath+"."+value.Name]; ok {
			return id, true
		}
	case *ast.SelectorExpr:
		ident, ok := value.X.(*ast.Ident)
		if !ok {
			return 0, false
		}
		path, ok := imports[ident.Name]
		if !ok {
			return 0, false
		}
		if id, ok := constValues[path+"."+value.Sel.Name]; ok {
			return id, true
		}
	}
	return 0, false
}

func parseIntLiteral(expr ast.Expr) (int, bool) {
	switch value := expr.(type) {
	case *ast.BasicLit:
		return parseBasicInt(value)
	case *ast.UnaryExpr:
		return parseUnaryInt(value)
	default:
		return 0, false
	}
}

func parseBasicInt(lit *ast.BasicLit) (int, bool) {
	if lit == nil || lit.Kind != token.INT {
		return 0, false
	}
	value, err := strconv.Atoi(lit.Value)
	if err != nil {
		return 0, false
	}
	return value, true
}

func parseUnaryInt(expr *ast.UnaryExpr) (int, bool) {
	if expr == nil {
		return 0, false
	}
	value, ok := expr.X.(*ast.BasicLit)
	if !ok || value.Kind != token.INT {
		return 0, false
	}
	parsed, err := strconv.Atoi(value.Value)
	if err != nil {
		return 0, false
	}
	if expr.Op == token.SUB {
		parsed = -parsed
	}
	if expr.Op != token.SUB && expr.Op != token.ADD {
		return 0, false
	}
	return parsed, true
}

func isSendCall(call *ast.CallExpr) bool {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return fun.Name == "SendProtoMessage"
	case *ast.SelectorExpr:
		return fun.Sel != nil && (fun.Sel.Name == "SendMessage" || fun.Sel.Name == "SendProtoMessage")
	default:
		return false
	}
}

func importName(spec *ast.ImportSpec, path string) string {
	if spec.Name != nil {
		return spec.Name.Name
	}
	return filepath.Base(path)
}

func loadAnswerHandlers(root string) (map[string]handlerSource, error) {
	handlers := map[string]handlerSource{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		imports := buildImportAliases(file)
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			name := fn.Name.Name
			if _, exists := handlers[name]; exists {
				continue
			}
			handlers[name] = handlerSource{
				Name:    name,
				File:    path,
				Line:    fset.Position(fn.Pos()).Line,
				Decl:    fn,
				Imports: imports,
				FileSet: fset,
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return handlers, nil
}

func extractRegistrations(file *ast.File, fset *token.FileSet, path string, imports importAliases) ([]packetRegistration, error) {
	var registrations []packetRegistration
	var parseErr error
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := selector.X.(*ast.Ident)
		if !ok || ident.Name != "packets" {
			return true
		}
		if selector.Sel == nil {
			return true
		}
		switch selector.Sel.Name {
		case "RegisterPacketHandler", "RegisterLocalizedPacketHandler":
			if len(call.Args) < 2 {
				return true
			}
			packetID, err := parsePacketID(call.Args[0])
			if err != nil {
				parseErr = err
				return false
			}
			handlers, err := handlersFromExpr(call.Args[1], fset, path)
			if err != nil {
				parseErr = err
				return false
			}
			registrations = append(registrations, packetRegistration{
				ID:       packetID,
				Handlers: handlers,
				File:     path,
				Line:     fset.Position(call.Pos()).Line,
				Imports:  imports,
			})
		}
		return true
	})
	if parseErr != nil {
		return nil, parseErr
	}
	return registrations, nil
}

func parsePacketID(expr ast.Expr) (int, error) {
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.INT {
		return 0, fmt.Errorf("packet id is not an int literal")
	}
	value, err := strconv.Atoi(literal.Value)
	if err != nil {
		return 0, fmt.Errorf("invalid packet id %q", literal.Value)
	}
	return value, nil
}

func handlersFromExpr(expr ast.Expr, fset *token.FileSet, path string) ([]handlerExpr, error) {
	switch value := expr.(type) {
	case *ast.UnaryExpr:
		if value.Op == token.AND {
			return handlersFromExpr(value.X, fset, path)
		}
	case *ast.CompositeLit:
		if isLocalizedHandlerType(value.Type) {
			handlers := []handlerExpr{}
			for _, elt := range value.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				nested, err := handlersFromExpr(kv.Value, fset, path)
				if err != nil {
					return nil, err
				}
				handlers = append(handlers, nested...)
			}
			return handlers, nil
		}
		if _, ok := value.Type.(*ast.ArrayType); ok {
			handlers := make([]handlerExpr, 0, len(value.Elts))
			for _, elt := range value.Elts {
				parsed, err := handlerFromExpr(elt, fset, path)
				if err != nil {
					return nil, err
				}
				if parsed != nil {
					handlers = append(handlers, *parsed)
				}
			}
			return handlers, nil
		}
	}
	return nil, fmt.Errorf("unexpected handlers expression")
}

func handlerFromExpr(expr ast.Expr, fset *token.FileSet, path string) (*handlerExpr, error) {
	switch value := expr.(type) {
	case *ast.Ident:
		if value.Name == "nil" {
			return nil, nil
		}
		return &handlerExpr{Name: value.Name, File: path, Line: fset.Position(value.Pos()).Line}, nil
	case *ast.SelectorExpr:
		return &handlerExpr{Name: selectorName(value), File: path, Line: fset.Position(value.Pos()).Line}, nil
	case *ast.FuncLit:
		line := fset.Position(value.Pos()).Line
		return &handlerExpr{Name: fmt.Sprintf("inline@%s:%d", filepath.Base(path), line), Inline: value, File: path, Line: line}, nil
	default:
		return nil, fmt.Errorf("unsupported handler expression")
	}
}

func selectorName(selector *ast.SelectorExpr) string {
	if ident, ok := selector.X.(*ast.Ident); ok {
		return ident.Name + "." + selector.Sel.Name
	}
	return selector.Sel.Name
}

func isLocalizedHandlerType(expr ast.Expr) bool {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name == "LocalizedHandler"
	case *ast.SelectorExpr:
		return value.Sel != nil && value.Sel.Name == "LocalizedHandler"
	default:
		return false
	}
}

func analyzeFunction(fnType *ast.FuncType, body *ast.BlockStmt, ctx analysisContext, cfg heuristicsConfig) analysisResult {
	signals := map[string]bool{}
	emptyBody, returnZeroOnly := analyzeReturns(body)
	if emptyBody {
		signals["empty_body"] = true
	}
	if returnZeroOnly {
		signals["return_zero"] = true
	}

	ast.Inspect(body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.CallExpr:
			inspectCallExpr(value, ctx, signals)
		case *ast.CompositeLit:
			markProtobufType(value.Type, ctx.Imports, signals)
		case *ast.ValueSpec:
			markProtobufType(value.Type, ctx.Imports, signals)
		case *ast.SelectorExpr:
			inspectSelectorExpr(value, ctx, signals)
		}
		return true
	})

	score := 0
	if signals["send_message"] {
		score += cfg.Weights.SendMessage
	}
	if signals["response_struct"] {
		score += cfg.Weights.ResponseType
	}
	if signals["request_struct"] {
		score += cfg.Weights.RequestType
	}
	if signals["proto_setter"] {
		score += cfg.Weights.ProtoSetter
	}
	if signals["request_parse"] {
		score += cfg.Weights.RequestParse
	}
	if signals["client_usage"] {
		score += cfg.Weights.ClientUsage
	}
	if signals["commander_usage"] {
		score += cfg.Weights.CommanderUse
	}
	if signals["orm_usage"] {
		score += cfg.Weights.ORMUsage
	}
	if signals["misc_usage"] {
		score += cfg.Weights.MiscUsage
	}
	if signals["db_write"] {
		score += cfg.Weights.DBWrite
	}

	if signals["panic"] {
		return analysisResult{Status: statusPanic, Score: score, Signals: signals}
	}
	if score == 0 {
		return analysisResult{Status: statusStub, Score: score, Signals: signals}
	}
	if score < cfg.Thresholds.ImplementedMin {
		return analysisResult{Status: statusPartial, Score: score, Signals: signals}
	}
	return analysisResult{Status: statusImplemented, Score: score, Signals: signals}
}

func inspectCallExpr(call *ast.CallExpr, ctx analysisContext, signals map[string]bool) {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		if fun.Name == "panic" {
			signals["panic"] = true
		}
	case *ast.SelectorExpr:
		if strings.HasPrefix(fun.Sel.Name, "SendMessage") {
			signals["send_message"] = true
		}
		if isAlias(fun.X, ctx.Imports.Proto) {
			if fun.Sel.Name == "Unmarshal" {
				signals["request_parse"] = true
			} else {
				signals["proto_setter"] = true
			}
		}
		if ident, ok := fun.X.(*ast.Ident); ok && ident.Name == "log" {
			if fun.Sel.Name == "Fatal" || fun.Sel.Name == "Fatalf" || fun.Sel.Name == "Fatalln" {
				signals["panic"] = true
			}
		}
		if isDBWriteMethod(fun.Sel.Name) {
			signals["db_write"] = true
		}
	}
}

func inspectSelectorExpr(selector *ast.SelectorExpr, ctx analysisContext, signals map[string]bool) {
	if ident, ok := selector.X.(*ast.Ident); ok {
		if ctx.ClientParamNames[ident.Name] {
			signals["client_usage"] = true
			if selector.Sel.Name == "Commander" {
				signals["commander_usage"] = true
			}
		}
		if isAlias(ident, ctx.Imports.ORM) {
			signals["orm_usage"] = true
		}
		if isAlias(ident, ctx.Imports.Misc) {
			signals["misc_usage"] = true
		}
	}
}

func markProtobufType(expr ast.Expr, imports importAliases, signals map[string]bool) {
	if isProtobufType(expr, imports.Protobuf, "SC_") {
		signals["response_struct"] = true
	}
	if isProtobufType(expr, imports.Protobuf, "CS_") {
		signals["request_struct"] = true
	}
}

func isProtobufType(expr ast.Expr, aliases []string, prefix string) bool {
	switch value := expr.(type) {
	case *ast.StarExpr:
		return isProtobufType(value.X, aliases, prefix)
	case *ast.SelectorExpr:
		if isAlias(value.X, aliases) && strings.HasPrefix(value.Sel.Name, prefix) {
			return true
		}
	}
	return false
}

func clientParamNames(fnType *ast.FuncType, imports importAliases) map[string]bool {
	names := map[string]bool{}
	if fnType == nil || fnType.Params == nil {
		return names
	}
	for _, field := range fnType.Params.List {
		if !isConnectionClientType(field.Type, imports.Connection) {
			continue
		}
		if len(field.Names) == 0 {
			continue
		}
		for _, name := range field.Names {
			names[name.Name] = true
		}
	}
	return names
}

func isConnectionClientType(expr ast.Expr, aliases []string) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if selector.Sel == nil || selector.Sel.Name != "Client" {
		return false
	}
	return isAlias(selector.X, aliases)
}

func isAlias(expr ast.Expr, aliases []string) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	for _, alias := range aliases {
		if ident.Name == alias {
			return true
		}
	}
	return false
}

func analyzeReturns(body *ast.BlockStmt) (bool, bool) {
	if body == nil || len(body.List) == 0 {
		return true, false
	}
	returnZeroOnly := true
	for _, stmt := range body.List {
		returnStmt, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			returnZeroOnly = false
			continue
		}
		if !returnsZero(returnStmt) {
			returnZeroOnly = false
		}
	}
	return false, returnZeroOnly
}

func returnsZero(stmt *ast.ReturnStmt) bool {
	if stmt == nil {
		return true
	}
	if len(stmt.Results) == 0 {
		return true
	}
	for _, expr := range stmt.Results {
		if !isZeroExpr(expr) {
			return false
		}
	}
	return true
}

func isZeroExpr(expr ast.Expr) bool {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name == "nil"
	case *ast.BasicLit:
		return value.Kind == token.INT && value.Value == "0"
	}
	return false
}

func isDBWriteMethod(name string) bool {
	switch name {
	case "Create", "Save", "Update", "Updates", "Delete":
		return true
	default:
		return false
	}
}

func combineHandlerReports(handlers []handlerReport) analysisResult {
	if len(handlers) == 0 {
		return analysisResult{Status: statusStub, Score: 0, Signals: map[string]bool{"no_handlers": true}}
	}
	scores := make([]int, 0, len(handlers))
	statuses := make([]string, 0, len(handlers))
	signals := map[string]bool{}
	for _, handler := range handlers {
		scores = append(scores, handler.Score)
		statuses = append(statuses, handler.Status)
		for _, signal := range handler.Signals {
			signals[signal] = true
		}
	}
	return analysisResult{
		Status:  combineStatuses(statuses),
		Score:   maxScore(scores),
		Signals: signals,
	}
}

func handlerLookupName(name string) string {
	if index := strings.LastIndex(name, "."); index != -1 {
		return name[index+1:]
	}
	return name
}

func combineStatuses(statuses []string) string {
	hasImplemented := false
	hasPartial := false
	hasStub := false
	for _, status := range statuses {
		if status == statusPanic {
			return statusPanic
		}
		switch status {
		case statusImplemented:
			hasImplemented = true
		case statusPartial:
			hasPartial = true
		case statusStub:
			hasStub = true
		}
	}
	if hasImplemented && (hasPartial || hasStub) {
		return statusPartial
	}
	if hasPartial {
		return statusPartial
	}
	if hasImplemented {
		return statusImplemented
	}
	return statusStub
}

func maxScore(scores []int) int {
	if len(scores) == 0 {
		return 0
	}
	max := scores[0]
	for _, score := range scores[1:] {
		if score > max {
			max = score
		}
	}
	return max
}

func sortedSignals(signals map[string]bool) []string {
	keys := make([]string, 0, len(signals))
	for key := range signals {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func countKnownPacketTypes(prefixes ...string) int {
	// Only count packet types with a numeric ID suffix.
	return len(knownPacketIDs(prefixes...))
}

func missingIDsForPrefixes(covered map[int]bool, prefixes ...string) []int {
	known := knownPacketIDs(prefixes...)
	for id := range covered {
		delete(known, id)
	}

	ids := make([]int, 0, len(known))
	for id := range known {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}

func knownPacketIDs(prefixes ...string) map[int]bool {
	ids := map[int]bool{}
	protoregistry.GlobalTypes.RangeMessages(func(desc protoreflect.MessageType) bool {
		name := string(desc.Descriptor().Name())
		id, ok := parsePacketTypeIDForPrefixes(name, prefixes)
		if !ok {
			return true
		}
		ids[id] = true
		return true
	})
	return ids
}

func parsePacketTypeIDForPrefixes(name string, prefixes []string) (int, bool) {
	for _, prefix := range prefixes {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		id, err := strconv.Atoi(strings.TrimPrefix(name, prefix))
		if err != nil {
			return 0, false
		}
		return id, true
	}
	return 0, false
}

func packetIDSet(packets []packetReport) map[int]bool {
	ids := make(map[int]bool, len(packets))
	for _, packet := range packets {
		ids[packet.ID] = true
	}
	return ids
}

func responseIDSet(responses []responseReport) map[int]bool {
	ids := make(map[int]bool, len(responses))
	for _, response := range responses {
		ids[response.ID] = true
	}
	return ids
}

func unionSortedInts(a []int, b []int) []int {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	seen := make(map[int]bool, len(a)+len(b))
	for _, v := range a {
		seen[v] = true
	}
	for _, v := range b {
		seen[v] = true
	}
	out := make([]int, 0, len(seen))
	for v := range seen {
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}

func combinePacketStatuses(packets []packetReport) map[int]string {
	statuses := map[int][]string{}
	for _, packet := range packets {
		statuses[packet.ID] = append(statuses[packet.ID], packet.Status)
	}
	combined := make(map[int]string, len(statuses))
	for id, list := range statuses {
		combined[id] = combineStatuses(list)
	}
	return combined
}

func loadOverrides(path string) (map[string]string, error) {
	overrides := map[string]string{}
	if err := loadJSONIfExists(path, &overrides); err != nil {
		return nil, err
	}
	for _, status := range overrides {
		if !isValidStatus(status) {
			return nil, fmt.Errorf("invalid override status %q", status)
		}
	}
	return overrides, nil
}

func isValidStatus(status string) bool {
	switch status {
	case statusImplemented, statusPartial, statusStub, statusPanic:
		return true
	default:
		return false
	}
}

func loadJSONIfExists(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return err
	}
	return nil
}

func replaceExt(path string, ext string) string {
	base := strings.TrimSuffix(path, filepath.Ext(path))
	return base + ext
}

func writeJSON(path string, data report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return os.WriteFile(path, payload, 0o644)
}

func writeSVG(path string, counts map[string]int, total int, fontFamily string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	svg := buildSVG(counts, total, fontFamily)
	return os.WriteFile(path, []byte(svg), 0o644)
}

func writePNG(svgPath string, pngPath string, scale float64) error {
	if scale <= 0 {
		scale = 1
	}
	if err := os.MkdirAll(filepath.Dir(pngPath), 0o755); err != nil {
		return err
	}
	converter, err := exec.LookPath("rsvg-convert")
	if err != nil {
		return fmt.Errorf("rsvg-convert not found; install librsvg")
	}
	cmd := exec.Command(converter, "--zoom", fmt.Sprintf("%.2f", scale), "-o", pngPath, svgPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func buildSVG(counts map[string]int, total int, fontFamily string) string {
	width := 1200
	height := 780
	centerX := width / 2
	centerY := 210
	radius := 150
	legendY := 420
	legendGap := 70
	safeFont := strings.ReplaceAll(fontFamily, "\"", "'")
	legendFontSize := 52
	textStroke := "#000000"
	textStrokeWidth := 4

	order := []string{statusImplemented, statusPartial, statusStub, statusPanic, statusMissing}
	colors := map[string]string{
		statusImplemented: "#00c853",
		statusPartial:     "#ffeb3b",
		statusStub:        "#8e8e8e",
		statusPanic:       "#ff1744",
		statusMissing:     "#424242",
	}

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%d\" height=\"%d\" viewBox=\"0 0 %d %d\">", width, height, width, height))

	if total > 0 {
		startAngle := -90.0
		for _, status := range order {
			count := counts[status]
			if count == 0 {
				continue
			}
			portion := float64(count) / float64(total)
			sweep := portion * 360.0
			endAngle := startAngle + sweep
			path := pieSlicePath(centerX, centerY, radius, startAngle, endAngle)
			builder.WriteString(fmt.Sprintf("<path d=\"%s\" fill=\"%s\"/>", path, colors[status]))
			startAngle = endAngle
		}
	} else {
		builder.WriteString(fmt.Sprintf("<circle cx=\"%d\" cy=\"%d\" r=\"%d\" fill=\"#2b2b2b\"/>", centerX, centerY, radius))
	}

	legendX := centerX - 300
	legendYPos := legendY
	for _, status := range order {
		label := statusLabel(status)
		count := counts[status]
		percentage := 0
		if total > 0 {
			percentage = int(math.Round(float64(count) / float64(total) * 100))
		}
		builder.WriteString(fmt.Sprintf("<rect x=\"%d\" y=\"%d\" width=\"32\" height=\"32\" fill=\"%s\"/>", legendX, legendYPos-32, colors[status]))
		builder.WriteString(fmt.Sprintf("<text x=\"%d\" y=\"%d\" fill=\"#ffffff\" font-family=\"%s\" font-size=\"%d\" stroke=\"%s\" stroke-width=\"%d\" paint-order=\"stroke fill\">%s %d (%d%%)</text>", legendX+48, legendYPos, safeFont, legendFontSize, textStroke, textStrokeWidth, label, count, percentage))
		legendYPos += legendGap
	}

	builder.WriteString(fmt.Sprintf("<title>implemented %d, partial %d, stub %d, panic %d, missing %d</title>", counts[statusImplemented], counts[statusPartial], counts[statusStub], counts[statusPanic], counts[statusMissing]))
	builder.WriteString("</svg>")
	return builder.String()
}

func statusLabel(status string) string {
	switch status {
	case statusImplemented:
		return "Implemented"
	case statusPartial:
		return "Partial"
	case statusStub:
		return "Stub"
	case statusPanic:
		return "Panic"
	case statusMissing:
		return "Missing"
	default:
		return status
	}
}

func pieSlicePath(cx int, cy int, radius int, startAngle float64, endAngle float64) string {
	start := polarToCartesian(cx, cy, radius, endAngle)
	end := polarToCartesian(cx, cy, radius, startAngle)
	largeArc := 0
	if endAngle-startAngle > 180 {
		largeArc = 1
	}
	return fmt.Sprintf("M %d %d L %d %d A %d %d 0 %d 0 %d %d Z", cx, cy, start.X, start.Y, radius, radius, largeArc, end.X, end.Y)
}

type point struct {
	X int
	Y int
}

func polarToCartesian(cx int, cy int, radius int, angleDeg float64) point {
	angleRad := (angleDeg - 90) * math.Pi / 180.0
	x := float64(cx) + float64(radius)*math.Cos(angleRad)
	y := float64(cy) + float64(radius)*math.Sin(angleRad)
	return point{X: int(math.Round(x)), Y: int(math.Round(y))}
}

func exitWithError(message string, err error) {
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", message, err)
	os.Exit(1)
}
