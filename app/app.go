package app

import (
	"encoding/json"
	"io"
	stdlog "log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"

	"github.com/Finschia/ostracon/libs/log"
	ostos "github.com/Finschia/ostracon/libs/os"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/Finschia/finschia-sdk/baseapp"
	"github.com/Finschia/finschia-sdk/client"
	nodeservice "github.com/Finschia/finschia-sdk/client/grpc/node"
	"github.com/Finschia/finschia-sdk/client/grpc/tmservice"
	"github.com/Finschia/finschia-sdk/codec"
	"github.com/Finschia/finschia-sdk/codec/types"
	"github.com/Finschia/finschia-sdk/server/api"
	"github.com/Finschia/finschia-sdk/server/config"
	servertypes "github.com/Finschia/finschia-sdk/server/types"
	"github.com/Finschia/finschia-sdk/simapp"
	"github.com/Finschia/finschia-sdk/store/streaming"
	sdk "github.com/Finschia/finschia-sdk/types"
	"github.com/Finschia/finschia-sdk/types/module"
	"github.com/Finschia/finschia-sdk/version"
	"github.com/Finschia/finschia-sdk/x/auth"
	"github.com/Finschia/finschia-sdk/x/auth/ante"
	authkeeper "github.com/Finschia/finschia-sdk/x/auth/keeper"
	authsims "github.com/Finschia/finschia-sdk/x/auth/simulation"
	authtx "github.com/Finschia/finschia-sdk/x/auth/tx"
	authtx2 "github.com/Finschia/finschia-sdk/x/auth/tx2"
	authtypes "github.com/Finschia/finschia-sdk/x/auth/types"
	"github.com/Finschia/finschia-sdk/x/auth/vesting"
	vestingtypes "github.com/Finschia/finschia-sdk/x/auth/vesting/types"
	"github.com/Finschia/finschia-sdk/x/authz"
	authzkeeper "github.com/Finschia/finschia-sdk/x/authz/keeper"
	authzmodule "github.com/Finschia/finschia-sdk/x/authz/module"
	"github.com/Finschia/finschia-sdk/x/bank"
	banktypes "github.com/Finschia/finschia-sdk/x/bank/types"
	"github.com/Finschia/finschia-sdk/x/bankplus"
	bankpluskeeper "github.com/Finschia/finschia-sdk/x/bankplus/keeper"
	"github.com/Finschia/finschia-sdk/x/capability"
	capabilitykeeper "github.com/Finschia/finschia-sdk/x/capability/keeper"
	capabilitytypes "github.com/Finschia/finschia-sdk/x/capability/types"
	"github.com/Finschia/finschia-sdk/x/collection"
	collectionkeeper "github.com/Finschia/finschia-sdk/x/collection/keeper"
	collectionmodule "github.com/Finschia/finschia-sdk/x/collection/module"
	"github.com/Finschia/finschia-sdk/x/crisis"
	crisiskeeper "github.com/Finschia/finschia-sdk/x/crisis/keeper"
	crisistypes "github.com/Finschia/finschia-sdk/x/crisis/types"
	distr "github.com/Finschia/finschia-sdk/x/distribution"
	distrclient "github.com/Finschia/finschia-sdk/x/distribution/client"
	distrkeeper "github.com/Finschia/finschia-sdk/x/distribution/keeper"
	distrtypes "github.com/Finschia/finschia-sdk/x/distribution/types"
	"github.com/Finschia/finschia-sdk/x/evidence"
	evidencekeeper "github.com/Finschia/finschia-sdk/x/evidence/keeper"
	evidencetypes "github.com/Finschia/finschia-sdk/x/evidence/types"
	"github.com/Finschia/finschia-sdk/x/feegrant"
	feegrantkeeper "github.com/Finschia/finschia-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/Finschia/finschia-sdk/x/feegrant/module"
	"github.com/Finschia/finschia-sdk/x/foundation"
	foundationclient "github.com/Finschia/finschia-sdk/x/foundation/client"
	foundationkeeper "github.com/Finschia/finschia-sdk/x/foundation/keeper"
	foundationmodule "github.com/Finschia/finschia-sdk/x/foundation/module"
	"github.com/Finschia/finschia-sdk/x/genutil"
	genutiltypes "github.com/Finschia/finschia-sdk/x/genutil/types"
	"github.com/Finschia/finschia-sdk/x/gov"
	govkeeper "github.com/Finschia/finschia-sdk/x/gov/keeper"
	govtypes "github.com/Finschia/finschia-sdk/x/gov/types"
	"github.com/Finschia/finschia-sdk/x/mint"
	mintkeeper "github.com/Finschia/finschia-sdk/x/mint/keeper"
	minttypes "github.com/Finschia/finschia-sdk/x/mint/types"
	ordamodule "github.com/Finschia/finschia-sdk/x/or/da"
	ordakeeper "github.com/Finschia/finschia-sdk/x/or/da/keeper"
	ordatypes "github.com/Finschia/finschia-sdk/x/or/da/types"
	"github.com/Finschia/finschia-sdk/x/or/rollup"
	rollupkeeper "github.com/Finschia/finschia-sdk/x/or/rollup/keeper"
	rolluptypes "github.com/Finschia/finschia-sdk/x/or/rollup/types"
	"github.com/Finschia/finschia-sdk/x/or/settlement"
	settlementkeeper "github.com/Finschia/finschia-sdk/x/or/settlement/keeper"
	settlementtypes "github.com/Finschia/finschia-sdk/x/or/settlement/types"
	"github.com/Finschia/finschia-sdk/x/params"
	paramsclient "github.com/Finschia/finschia-sdk/x/params/client"
	paramskeeper "github.com/Finschia/finschia-sdk/x/params/keeper"
	paramstypes "github.com/Finschia/finschia-sdk/x/params/types"
	paramproposal "github.com/Finschia/finschia-sdk/x/params/types/proposal"
	"github.com/Finschia/finschia-sdk/x/slashing"
	slashingkeeper "github.com/Finschia/finschia-sdk/x/slashing/keeper"
	slashingtypes "github.com/Finschia/finschia-sdk/x/slashing/types"
	"github.com/Finschia/finschia-sdk/x/staking"
	stakingkeeper "github.com/Finschia/finschia-sdk/x/staking/keeper"
	stakingtypes "github.com/Finschia/finschia-sdk/x/staking/types"
	stakingplusmodule "github.com/Finschia/finschia-sdk/x/stakingplus/module"
	"github.com/Finschia/finschia-sdk/x/token"
	"github.com/Finschia/finschia-sdk/x/token/class"
	classkeeper "github.com/Finschia/finschia-sdk/x/token/class/keeper"
	tokenkeeper "github.com/Finschia/finschia-sdk/x/token/keeper"
	tokenmodule "github.com/Finschia/finschia-sdk/x/token/module"
	"github.com/Finschia/finschia-sdk/x/upgrade"
	upgradeclient "github.com/Finschia/finschia-sdk/x/upgrade/client"
	upgradekeeper "github.com/Finschia/finschia-sdk/x/upgrade/keeper"
	upgradetypes "github.com/Finschia/finschia-sdk/x/upgrade/types"

	appante "github.com/Finschia/finschia/ante"
	appparams "github.com/Finschia/finschia/app/params"

	// unnamed import of statik for swagger UI support
	_ "github.com/Finschia/finschia/client/docs/statik"
)

const appName = "Finschia"

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		stakingplusmodule.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		foundationmodule.AppModuleBasic{},
		gov.NewAppModuleBasic(
			paramsclient.ProposalHandler,
			distrclient.ProposalHandler,
			upgradeclient.ProposalHandler,
			upgradeclient.CancelProposalHandler,
			foundationclient.ProposalHandler,
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		vesting.AppModuleBasic{},
		tokenmodule.AppModuleBasic{},
		collectionmodule.AppModuleBasic{},
		rollup.AppModuleBasic{},
		ordamodule.AppModuleBasic{},
		settlement.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		foundation.ModuleName:          nil,
		foundation.TreasuryName:        nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		rolluptypes.ModuleName:         {authtypes.Burner},
	}

	// module accounts that are allowed to receive tokens
	allowedReceivingModAcc = map[string]bool{
		// govtypes.ModuleName: true, // TODO: uncomment it when authority is ready
	}
)

var (
	_ simapp.App              = (*LinkApp)(nil)
	_ servertypes.Application = (*LinkApp)(nil)
)

// LinkApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type LinkApp struct { // nolint: golint
	*baseapp.BaseApp
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*sdk.KVStoreKey
	memKeys map[string]*sdk.MemoryStoreKey

	// keepers
	AccountKeeper    authkeeper.AccountKeeper
	BankKeeper       bankpluskeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    stakingkeeper.Keeper
	SlashingKeeper   slashingkeeper.Keeper
	MintKeeper       mintkeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	FoundationKeeper foundationkeeper.Keeper
	GovKeeper        govkeeper.Keeper
	CrisisKeeper     crisiskeeper.Keeper
	UpgradeKeeper    upgradekeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	AuthzKeeper      authzkeeper.Keeper
	// IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper   evidencekeeper.Keeper
	FeeGrantKeeper   feegrantkeeper.Keeper
	ClassKeeper      classkeeper.Keeper
	TokenKeeper      tokenkeeper.Keeper
	CollectionKeeper collectionkeeper.Keeper
	RollupKeeper     rollupkeeper.Keeper
	Ordakeeper       ordakeeper.Keeper
	SettlementKeeper settlementkeeper.Keeper

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager

	// the configurator
	configurator module.Configurator
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		stdlog.Println("Failed to get home dir %2", err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".finschia")
}

// NewLinkApp returns a reference to an initialized Link.
func NewLinkApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, skipUpgradeHeights map[int64]bool,
	homePath string, invCheckPeriod uint, encodingConfig appparams.EncodingConfig, appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp),
) *LinkApp {
	appCodec := encodingConfig.Marshaler
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(appName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey,
		authzkeeper.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		paramstypes.StoreKey,
		upgradetypes.StoreKey,
		evidencetypes.StoreKey,
		capabilitytypes.StoreKey,
		feegrant.StoreKey,
		foundation.StoreKey,
		class.StoreKey,
		token.StoreKey,
		collection.StoreKey,
		rolluptypes.StoreKey,
		ordatypes.StoreKey,
		settlementtypes.StoreKey,
	)

	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	// NOTE: The testingkey is just mounted for testing purposes. Actual applications should
	// not include this key.
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	// configure state listening capabilities using AppOptions
	// we are doing nothing with the returned streamingServices and waitGroup in this case
	if _, _, err := streaming.LoadStreamingServices(bApp, appOpts, appCodec, keys); err != nil {
		ostos.Exit(err.Error())
	}

	app := &LinkApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(appCodec, legacyAmino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])
	// set the BaseApp's parameter store
	bApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// Applications that wish to enforce statically created ScopedKeepers should call `Seal` after creating
	// their scoped modules in `NewApp` with `ScopeToModule`
	app.CapabilityKeeper.Seal()

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec, keys[authtypes.StoreKey], app.GetSubspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
	)
	app.BankKeeper = bankpluskeeper.NewBaseKeeper(
		appCodec, keys[banktypes.StoreKey], app.AccountKeeper, app.GetSubspace(banktypes.ModuleName), app.BlockedAddrs(), true,
	)
	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec, keys[stakingtypes.StoreKey], app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName),
	)
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec, keys[minttypes.StoreKey], app.GetSubspace(minttypes.ModuleName), &stakingKeeper,
		app.AccountKeeper, app.BankKeeper, authtypes.FeeCollectorName,
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, keys[distrtypes.StoreKey], app.GetSubspace(distrtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		&stakingKeeper, authtypes.FeeCollectorName, app.ModuleAccountAddrs(),
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, keys[slashingtypes.StoreKey], &stakingKeeper, app.GetSubspace(slashingtypes.ModuleName),
	)
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		app.GetSubspace(crisistypes.ModuleName), invCheckPeriod, app.BankKeeper, authtypes.FeeCollectorName,
	)
	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], app.AccountKeeper)
	app.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath, app.BaseApp)
	foundationConfig := foundation.DefaultConfig()
	app.FoundationKeeper = foundationkeeper.NewKeeper(appCodec, keys[foundation.StoreKey], app.BaseApp.MsgServiceRouter(), app.AccountKeeper, app.BankKeeper, authtypes.FeeCollectorName, foundationConfig, foundation.DefaultAuthority().String(), app.GetSubspace(foundation.ModuleName))

	app.ClassKeeper = classkeeper.NewKeeper(appCodec, keys[class.StoreKey])
	app.TokenKeeper = tokenkeeper.NewKeeper(appCodec, keys[token.StoreKey], app.ClassKeeper)
	app.CollectionKeeper = collectionkeeper.NewKeeper(appCodec, keys[collection.StoreKey], app.ClassKeeper)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper = *stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	// Create Authz Keeper
	app.AuthzKeeper = authzkeeper.NewKeeper(keys[authzkeeper.StoreKey], appCodec, app.BaseApp.MsgServiceRouter())

	// create evidence keeper with router
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, keys[evidencetypes.StoreKey], &app.StakingKeeper, app.SlashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	// register the proposal types
	govRouter := govtypes.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.DistrKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper)).
		AddRoute(foundation.RouterKey, foundationkeeper.NewFoundationProposalsHandler(app.FoundationKeeper))

	govKeeper := govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey], app.GetSubspace(govtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		&stakingKeeper, govRouter,
	)
	app.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	app.Ordakeeper = ordakeeper.NewKeeper(appCodec, keys[ordatypes.StoreKey], authtypes.NewModuleAddress(govtypes.ModuleName).String(), app.AccountKeeper, nil)
	app.SettlementKeeper = settlementkeeper.NewKeeper(appCodec, keys[settlementtypes.StoreKey], keys[settlementtypes.MemStoreKey])
	app.RollupKeeper = rollupkeeper.NewKeeper(appCodec, app.BankKeeper, app.AccountKeeper, keys[rolluptypes.StoreKey], keys[rolluptypes.MemStoreKey], app.GetSubspace(rolluptypes.ModuleName))

	/****  Module Options ****/

	/****  Module Options ****/
	var skipGenesisInvariants = false
	opt := appOpts.Get(crisis.FlagSkipGenesisInvariants)
	if opt, ok := opt.(bool); ok {
		skipGenesisInvariants = opt
	}

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bankplus.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		foundationmodule.NewAppModule(appCodec, app.FoundationKeeper),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		stakingplusmodule.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.FoundationKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		params.NewAppModule(app.ParamsKeeper),
		tokenmodule.NewAppModule(appCodec, app.TokenKeeper),
		collectionmodule.NewAppModule(appCodec, app.CollectionKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		rollup.NewAppModule(appCodec, app.RollupKeeper, app.AccountKeeper, app.BankKeeper),
		ordamodule.NewAppModule(appCodec, app.Ordakeeper, app.AccountKeeper),
		settlement.NewAppModule(appCodec, app.SettlementKeeper, app.AccountKeeper, app.BankKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		foundation.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		token.ModuleName,
		collection.ModuleName,
		rolluptypes.ModuleName,
		ordatypes.ModuleName,
		settlementtypes.ModuleName,
	)
	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		foundation.ModuleName,
		token.ModuleName,
		collection.ModuleName,
		rolluptypes.ModuleName,
		ordatypes.ModuleName,
		settlementtypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	// wasm module should be a the end as it can call other modules functionality direct or via message dispatching during
	// genesis phase. For example bank transfer, auth account check, staking, ...
	// NOTE: staking module is required if HistoricalEntries param > 0
	// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		foundation.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		token.ModuleName,
		collection.ModuleName,
		rolluptypes.ModuleName,
		ordatypes.ModuleName,
		settlementtypes.ModuleName,
	)

	app.mm.RegisterInvariants(&app.CrisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName:    auth.NewAppModule(app.appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts),
		stakingtypes.ModuleName: staking.NewAppModule(app.appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.mm.Modules, overrideModules)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)

	anteHandler, err := appante.NewAnteHandler(
		appante.HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				BankKeeper:      app.BankKeeper,
				SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
				FeegrantKeeper:  app.FeeGrantKeeper,
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			},
		},
	)
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			ostos.Exit(err.Error())
		}

		// Initialize pinned codes in wasmvm as they are not persisted there
		ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

		// Initialize the keeper of bankkeeper
		app.BankKeeper.InitializeBankPlus(ctx)
	}

	return app
}

// MakeCodecs constructs the *std.Codec and *codec.LegacyAmino instances used by
// Link. It is useful for tests and clients who do not want to construct the
// full fnsad application
func MakeCodecs() (codec.Codec, *codec.LegacyAmino) {
	cf := MakeEncodingConfig()
	return cf.Marshaler, cf.Amino
}

// Name returns the name of the App
func (app *LinkApp) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *LinkApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *LinkApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *LinkApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *LinkApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *LinkApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlockedAddrs returns all the app's module account addresses that are not
// allowed to receive external tokens.
func (app *LinkApp) BlockedAddrs() map[string]bool {
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
	}

	return blockedAddrs
}

// LegacyAmino returns LinkApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *LinkApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns Link's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *LinkApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns Link's InterfaceRegistry
func (app *LinkApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *LinkApp) GetKey(storeKey string) *sdk.KVStoreKey {
	return app.keys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *LinkApp) GetMemKey(storeKey string) *sdk.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *LinkApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface
func (app *LinkApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *LinkApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	authtx2.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(apiSvr.Router)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *LinkApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
	authtx2.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *LinkApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.interfaceRegistry)
}

func (app *LinkApp) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey sdk.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypes.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(foundation.ModuleName)
	paramsKeeper.Subspace(settlementtypes.ModuleName)
	paramsKeeper.Subspace(ordatypes.ModuleName)
	paramsKeeper.Subspace(rolluptypes.ModuleName)

	return paramsKeeper
}
