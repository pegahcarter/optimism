package fault

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-challenger/config"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/trace/alphabet"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/trace/cannon"
	faultTypes "github.com/ethereum-optimism/optimism/op-challenger/game/fault/types"
	"github.com/ethereum-optimism/optimism/op-challenger/game/scheduler"
	"github.com/ethereum-optimism/optimism/op-challenger/game/types"
	"github.com/ethereum-optimism/optimism/op-challenger/metrics"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

var (
	cannonGameType   = uint8(0)
	alphabetGameType = uint8(255)
)

type Registry interface {
	RegisterGameType(gameType uint8, creator scheduler.PlayerCreator)
}

func RegisterGameTypes(
	registry Registry,
	ctx context.Context,
	logger log.Logger,
	m metrics.Metricer,
	cfg *config.Config,
	txMgr txmgr.TxManager,
	client bind.ContractCaller,
) {
	if cfg.TraceTypeEnabled(config.TraceTypeCannon) {
		registerFaultGameType(ctx, registry, cannonGameType, logger, m, cfg, txMgr, client, func(addr common.Address, gameDepth uint64, dir string) (faultTypes.TraceProvider, faultTypes.OracleUpdater, error) {
			provider, err := cannon.NewTraceProvider(ctx, logger, m, cfg, client, dir, addr, gameDepth)
			if err != nil {
				return nil, nil, fmt.Errorf("create cannon trace provider: %w", err)
			}
			updater, err := cannon.NewOracleUpdater(ctx, logger, txMgr, addr, client)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create the cannon updater: %w", err)
			}
			return provider, updater, nil
		})
	}

	if cfg.TraceTypeEnabled(config.TraceTypeAlphabet) {
		registerFaultGameType(ctx, registry, alphabetGameType, logger, m, cfg, txMgr, client, func(addr common.Address, gameDepth uint64, dir string) (faultTypes.TraceProvider, faultTypes.OracleUpdater, error) {
			provider := alphabet.NewTraceProvider(cfg.AlphabetTrace, gameDepth)
			updater := alphabet.NewOracleUpdater(logger)
			return provider, updater, nil
		})
	}
}

func registerFaultGameType(
	ctx context.Context,
	registry Registry,
	gameType uint8,
	logger log.Logger,
	m metrics.Metricer,
	cfg *config.Config,
	txMgr txmgr.TxManager,
	client bind.ContractCaller,
	creator resourceCreator,
) {
	registry.RegisterGameType(gameType, func(game types.GameMetadata, dir string) (scheduler.GamePlayer, error) {
		return NewGamePlayer(ctx, logger, m, cfg, dir, game.Proxy, txMgr, client, creator)
	})
}