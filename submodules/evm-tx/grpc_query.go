
package app

import (
	"golang.org/x/exp/maps"

	// cosmos modules
	"cosmossdk.io/x/feegrant"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	// cosmos SDK modules
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	genutil "github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"

	// IBC modules
	packetforward "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibcfee "github.com/cosmos/ibc-go/v8/modules/apps/29-fee"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	ibctransfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	solomachine "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	// initia IBC modules
	ibchooks "github.com/initia-labs/initia/x/ibc-hooks"
	ibchookstypes "github.com/initia-labs/initia/x/ibc-hooks/types"
	ibcnfttransfer "github.com/initia-labs/initia/x/ibc/nft-transfer"
	ibcnfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"
	icaauth "github.com/initia-labs/initia/x/intertx"
	icaauthtypes "github.com/initia-labs/initia/x/intertx/types"

	// initia modules
	authzmodule "github.com/initia-labs/initia/x/authz/module"
	"github.com/initia-labs/initia/x/bank"
	"github.com/initia-labs/initia/x/move"
	movetypes "github.com/initia-labs/initia/x/move/types"

	// opinit modules
	opchild "github.com/initia-labs/OPinit/x/opchild"
	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"

	// skip-mev modules
	"github.com/skip-mev/block-sdk/v2/x/auction"
	auctiontypes "github.com/skip-mev/block-sdk/v2/x/auction/types"
	marketmap "github.com/skip-mev/connect/v2/x/marketmap"
	marketmaptypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/skip-mev/connect/v2/x/oracle"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"

	// noble forwarding keeper
	forwarding "github.com/noble-assets/forwarding/v2"
	forwardingtypes "github.com/noble-assets/forwarding/v2/types"

	// New module: example module
	example "github.com/myorg/myapp/x/example"
	exampletypes "github.com/myorg/myapp/x/example/types"
)

// module account permissions
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:      nil,
	icatypes.ModuleName:             nil,
	ibcfeetypes.ModuleName:          nil,
	ibctransfertypes.ModuleName:     {authtypes.Minter, authtypes.Burner},
	movetypes.MoveStakingModuleName: nil,
	// The module account for x/auction must be instantiated at genesis
	auctiontypes.ModuleName: nil,
	opchildtypes.ModuleName: {authtypes.Minter, authtypes.Burner},

	// connect oracle permissions
	oracletypes.ModuleName: nil,

	// for testing only
	authtypes.Minter: {authtypes.Minter},

	// New module permissions (if required)
	exampletypes.ModuleName: nil,
}

func appModules(
	app *MinitiaApp,
	skipGenesisInvariants bool,
) []module.AppModule {
	return []module.AppModule{
		auth.NewAppModule(app.appCodec, *app.AccountKeeper, nil, nil),
		bank.NewAppModule(app.appCodec, *app.BankKeeper, app.AccountKeeper),
		opchild.NewAppModule(app.appCodec, app.OPChildKeeper),
		capability.NewAppModule(app.appCodec, *app.CapabilityKeeper, false),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, nil),
		feegrantmodule.NewAppModule(app.appCodec, app.AccountKeeper, app.BankKeeper, *app.FeeGrantKeeper, app.interfaceRegistry),
		upgrade.NewAppModule(app.UpgradeKeeper, app.ac),
		authzmodule.NewAppModule(app.appCodec, *app.AuthzKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(app.appCodec, *app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		consensus.NewAppModule(app.appCodec, *app.ConsensusParamsKeeper),
		move.NewAppModule(app.appCodec, *app.MoveKeeper, app.vc, maps.Keys(maccPerms)),
		auction.NewAppModule(app.appCodec, *app.AuctionKeeper),
		// IBC modules
		ibc.NewAppModule(app.IBCKeeper),
		ibctransfer.NewAppModule(*app.TransferKeeper),
		ibcnfttransfer.NewAppModule(app.appCodec, *app.NftTransferKeeper),
		ica.NewAppModule(app.ICAControllerKeeper, app.ICAHostKeeper),
		icaauth.NewAppModule(app.appCodec, *app.ICAAuthKeeper),
		ibcfee.NewAppModule(*app.IBCFeeKeeper),
		ibctm.NewAppModule(),
		solomachine.NewAppModule(),
		packetforward.NewAppModule(app.PacketForwardKeeper, nil),
		ibchooks.NewAppModule(app.appCodec, *app.IBCHooksKeeper),
		forwarding.NewAppModule(app.ForwardingKeeper),
		// connect modules
		oracle.NewAppModule(app.appCodec, *app.OracleKeeper),
		marketmap.NewAppModule(app.appCodec, app.MarketMapKeeper),
		// Newly added module: example module
		example.NewAppModule(app.appCodec, app.ExampleKeeper),
	}
}

// ModuleBasics defines the BasicManager responsible for setting up basic
// module elements such as codec registration and genesis verification.
func newBasicManagerFromManager(app *MinitiaApp) module.BasicManager {
	basicManager := module.NewBasicManagerFromManager(
		app.ModuleManager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			// Adding the example module to the BasicManager:
			exampletypes.ModuleName: example.NewAppModuleBasic(),
		})
	basicManager.RegisterLegacyAminoCodec(app.legacyAmino)
	basicManager.RegisterInterfaces(app.interfaceRegistry)
	return basicManager
}

/*
orderBeginBlockers sets the order of BeginBlockers that run at the beginning of every block.
You can insert the new module in the desired order based on its requirements.
*/
func orderBeginBlockers() []string {
	return []string{
		capabilitytypes.ModuleName,
		opchildtypes.ModuleName,
		authz.ModuleName,
		movetypes.ModuleName,
		ibcexported.ModuleName,
		oracletypes.ModuleName,
		marketmaptypes.ModuleName,
		// Newly added module
		exampletypes.ModuleName,
	}
}

/*
Interchain Security Requirements:
- The provider's EndBlock retrieves validator updates from the staking module;
  therefore, staking.EndBlock must execute before provider.EndBlock.
- When creating a new consumer chain, the order must be:
  CreateChildClient(), staking.EndBlock, provider.EndBlock;
  thus, gov.EndBlock must execute before staking.EndBlock.
*/
func orderEndBlockers() []string {
	return []string{
		crisistypes.ModuleName,
		opchildtypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		oracletypes.ModuleName,
		marketmaptypes.ModuleName,
		forwardingtypes.ModuleName,
		// Newly added module
		exampletypes.ModuleName,
	}
}

/*
NOTE: The genutil module must run after staking so that pools are properly
initialized with tokens from genesis accounts.
NOTE: The genutil module must also run after auth to access its parameters.
NOTE: The Capability module must run first to initialize any capabilities,
allowing other modules to safely create or claim capabilities during InitChain.
*/
func orderInitBlockers() []string {
	return []string{
		capabilitytypes.ModuleName, authtypes.ModuleName, movetypes.ModuleName, banktypes.ModuleName,
		opchildtypes.ModuleName, genutiltypes.ModuleName, authz.ModuleName, group.ModuleName, crisistypes.ModuleName,
		upgradetypes.ModuleName, feegrant.ModuleName, consensusparamtypes.ModuleName, ibcexported.ModuleName,
		ibctransfertypes.ModuleName, ibcnfttransfertypes.ModuleName, icatypes.ModuleName, icaauthtypes.ModuleName,
		ibcfeetypes.ModuleName, auctiontypes.ModuleName, oracletypes.ModuleName,
		marketmaptypes.ModuleName, packetforwardtypes.ModuleName, ibchookstypes.ModuleName, forwardingtypes.ModuleName,
		// Newly added module
		exampletypes.ModuleName,
	}
}
