package imports

import (
	"context"
	"fmt"
	"os"
	"strings"

	"path/filepath"

	_ "embed"

	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/config/json"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/access/grpc"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:embed compiled_registry.json
var defaultRegistry []byte
var defaultRegistryPath = "./registry.json"

type (
	CLIConfig = *config.Config
	Importer  struct {
		Network string
		Address string
		Verbose bool
	}
	SourceResolver struct {
		*config.Loader
		RegistryConfig CLIConfig
		ContractMap    ContractByName
		Writer         config.ReaderWriter
		TargetConfig   CLIConfig
		TargetPath     string
		Verbose        bool
		justAdded      map[string][]string // justAdded tracks the contracts added in a single invocation
	}
	importReplacer struct {
		*SourceResolver
		SourceDirectory string
		currentSrc      string
		onChainAddress  string
		importName      string
		network         string
	}
	ContractByName    map[string]ContractByNetwork
	ContractByNetwork map[string]config.Contract
)

// Get is the expected entry point for the importer
func (i *Importer) Get(rw config.ReaderWriter, name string) (err error) {
	ctx := context.Background()

	defer func() {
		if err != nil || i.Verbose {
			return // do not recover from panic, print stack trace
		}

		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("🛑  %w", panicErr.(error))
		}
	}()

	r := NewResolver(rw, i.Verbose, getRegistry(), getTarget())
	if i.Address != "" {
		r.AddEntry(name, i.Network, i.Address)
	}
	// get the source files and populate our in memory version of flow.json
	r.GetSource(ctx, name, i.Network)
	r.MustSave()
	fmt.Printf("🟢  Successfully imported %v contracts!", len(r.justAdded))
	return
}

func NewResolver(rw config.ReaderWriter, verbose bool, registryPath []string, targetPath string) *SourceResolver {
	configLoader := config.NewLoader(rw)
	configLoader.AddConfigParser(json.NewParser())
	sr := &SourceResolver{Loader: configLoader, Writer: rw, Verbose: verbose}
	sr.justAdded = map[string][]string{}
	sr.MustLoadRegistry(registryPath)
	sr.MustLoadFlowConfig(targetPath)
	return sr
}

func (s *SourceResolver) MustLoadRegistry(paths []string) {
	cfg, loadErr := s.Load(paths)
	if loadErr != nil {
		panic(fmt.Errorf("error while loading configuration from %v: %w", paths, loadErr))
	}
	s.RegistryConfig = cfg
	s.ContractMap = getContractByName(cfg.Contracts)
}

func (s *SourceResolver) MustLoadFlowConfig(configPath string) {
	targetCfg, loadErr := s.Load([]string{configPath})
	if loadErr != nil {
		panic(fmt.Errorf("error while loading configuration from %v: %w", targetCfg, loadErr))
	}
	s.TargetConfig = targetCfg
	s.TargetPath = configPath
}

// getRegistry gets contract info from multiple flow.json as provided by env or defaults to local flow.json
func getRegistry() []string {
	registry, has := os.LookupEnv("REGISTRY")
	allFiles := []string{"./flow.json"}
	if has {
		allFiles = strings.Split(registry, ",")
	}
	existingFiles := []string{}

	for _, fileLocation := range allFiles {
		_, err := os.Open(fileLocation)
		if err == nil {
			existingFiles = append(existingFiles, fileLocation)
		}
	}
	existingFiles = append(existingFiles, defaultRegistryPath) // always should have this, even if embedded
	return existingFiles
}

// getTarget will retrive the path to the flow.json we should write contracts to
func getTarget() string {
	target, has := os.LookupEnv("TARGET")
	if has {
		return target
	}
	return "./flow.json"
}

func (s *SourceResolver) MustSave() {
	err := s.Loader.Save(s.TargetConfig, s.TargetPath)
	if err != nil {
		panic(fmt.Errorf("error while saving config to path %v: %w", s.TargetPath, err))
	}
}

func getContractByName(contracts config.Contracts) ContractByName {
	res := ContractByName{}
	for _, c := range contracts {
		byNetwork, has := res[c.Name]
		if !has {
			byNetwork = ContractByNetwork{}
		}
		byNetwork[c.Network] = c
		res[c.Name] = byNetwork
	}
	return res
}

// getImportContractDir returns the default directory to be used for imports
func getImportContractDir() string {
	val, has := os.LookupEnv("IMPORT_DIR")
	if has {
		return val
	}
	return "./imports"
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *SourceResolver) checkJustAdded(ctx context.Context, name string, address string) bool {
	addresses, has := s.justAdded[name]
	if !has {
		s.justAdded[name] = []string{address}
		return false
	}
	for _, addr := range addresses {
		if addr == address {
			return true
		}
	}
	// TODO - need to handle the case where contract names overlap by adjusting import path
	s.justAdded[name] = append(addresses, address)
	return false
}

func (s *SourceResolver) GetSource(ctx context.Context, name string, network string) {

	for _, c := range s.ContractMap[name] {
		s.TargetConfig.Contracts.AddOrUpdate(name, c)
	}
	n, err := s.RegistryConfig.Networks.ByName(network)
	handleErr(err)

	con := s.ContractMap[name][network]
	if s.Verbose {
		fmt.Printf("🔍  name: %v, network: %v, map: %v", name, network, con)
	}
	if con.Alias == "" {
		handleErr(fmt.Errorf("no contract address defined for %v on network %v", name, network))
	}
	if s.checkJustAdded(ctx, name, con.Alias) {
		return
	}

	fmt.Printf("📜  Importing source for %v from network %v\n", name, network)
	fc, err := grpc.NewClient(n.Host, gogrpc.WithTransportCredentials(insecure.NewCredentials()))
	handleErr(err)

	a, err := fc.GetAccount(ctx, flow.HexToAddress(con.Alias))

	handleErr(err)

	importsReplaced, err := s.checkImports(ctx, con, a.Contracts[name], network)
	handleErr(err)

	err = os.MkdirAll(filepath.Dir(con.Source), 0700)
	handleErr(err)
	err = s.Writer.WriteFile(con.Source, []byte(importsReplaced), 0777)
	handleErr(err)
}

func (s *SourceResolver) AddEntry(name string, network string, address string) ContractByNetwork {
	n := s.shimByNetwork(name, network, address)
	s.ContractMap[name] = n
	return n
}

// populateRegistry will add an entry to the registry
// this can be used if the registry did not have a contract that is a dependency of a well known contract
func (s *SourceResolver) shimByNetwork(name string, network string, address string) ContractByNetwork {
	importDir := getImportContractDir()
	src := importDir + "/" + name + ".cdc"
	return ContractByNetwork{
		network: config.Contract{Name: name, Source: src, Network: network, Alias: address},
	}
}

// checkImports will check the imports of a contract and replace the source with appropriate local values before
// writing the file locally. We also recursively fetch the contract.
func (s *SourceResolver) checkImports(ctx context.Context, contract config.Contract, src []byte, network string) (string, error) {
	contractSrc := string(src)
	if s.Verbose {
		fmt.Println("Full contract source: \n" + contractSrc)
	}
	copy := contractSrc
	p, err := parser.ParseProgram(contractSrc, nil)
	if err != nil {
		return "", err
	}
	for _, imp := range p.ImportDeclarations() {
		loc := imp.Location.String()
		for _, id := range imp.Identifiers {
			// the name of the import (e.g. FlowToken)
			importName := id.String()
			ir := importReplacer{
				SourceResolver:  s,
				SourceDirectory: filepath.Dir(contract.Source),
				currentSrc:      copy,
				onChainAddress:  loc,
				importName:      importName,
				network:         network,
			}
			copy = ir.replaceImport(ctx)
		}
	}
	return copy, nil
}

func (i *importReplacer) replaceImport(ctx context.Context) string {
	byNetwork, has := i.ContractMap[i.importName]

	if !has {
		fmt.Println("🔍  No import defined for dependency " + i.importName)
		byNetwork = i.AddEntry(i.importName, i.network, strings.Replace(i.onChainAddress, "0x", "", 1))
	}
	val, has := byNetwork[i.network]
	if !has {
		fmt.Println("no import specific to environment for " + i.importName)
		byNetwork = i.AddEntry(i.importName, i.network, strings.Replace(i.onChainAddress, "0x", "", 1))
		val = byNetwork[i.network]
	} else if val.Alias != i.onChainAddress {
		fmt.Printf("unexpected location for import %s in network %s, got %s expecting %s \n", i.importName, i.network, i.onChainAddress, val.Alias)
	}
	// we want to replace the address on the network net with val.Source for our local contract
	// is there a good way to do this with the AST program itself?
	// naive string replacement for now - there is likely a better way
	relPath, err := filepath.Rel(i.SourceDirectory, val.Source)
	handleErr(err)
	formatted := fmt.Sprintf(`"%v"`, relPath)
	fmt.Printf("⛓   Replacing %v with %v\n", "0x"+i.onChainAddress, formatted)
	copy := strings.Replace(i.currentSrc, "0x"+i.onChainAddress, formatted, 1)
	i.GetSource(ctx, i.importName, i.network)
	return copy
}
