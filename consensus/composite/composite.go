package composite

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/clique"
	"github.com/ethereum/go-ethereum/consensus/myclique"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
)

const ()

type blacklistDirection uint

// Composite is the proof-of-stake-authority consensus engine proposed to support the
// Ethereum testnet following the Ropsten attacks.
type Composite struct {
	PoCliqueEngine   consensus.Engine
	MyPoCliqueEngine consensus.Engine
}

// New creates a Composite proof-of-stake-authority consensus engine with the initial
// validators set to the ones provided by the user.
func New(chainConfig *params.ChainConfig, db ethdb.Database) *Composite {
	chainConfig.MyClique = &params.MyCliqueConfig{Period: 3, Epoch: 30000}
	chainConfig.Clique = &params.CliqueConfig{Period: 3, Epoch: 30000}
	return &Composite{
		PoCliqueEngine:   clique.New(chainConfig.Clique, db),
		MyPoCliqueEngine: myclique.New(chainConfig.MyClique, db),
	}
}

func (c *Composite) getEngine(number uint64, fnName string) consensus.Engine {
	if number > 1000 {
		log.Info("composite consensus engine myClique", "function", fnName, "headerNum", number)
		return c.MyPoCliqueEngine
	} else {
		log.Info("composite consensus engine clique", "function", fnName, "headerNum", number)
		return c.PoCliqueEngine
	}
}

// Author retrieves the Ethereum address of the account that minted the given
// block, which may be different from the header's coinbase if a consensus
// engine is based on signatures.
func (c *Composite) Author(header *types.Header) (common.Address, error) {
	return c.getEngine(header.Number.Uint64(), "Author").Author(header)
}

// VerifyHeader checks whether a header conforms to the consensus rules of a
// given engine. Verifying the seal may be done optionally here, or explicitly
// via the VerifySeal method.
func (c *Composite) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header, seal bool) error {
	return c.getEngine(header.Number.Uint64(), "VerifyHeader").VerifyHeader(chain, header, seal)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications (the order is that of
// the input slice).
func (c *Composite) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	return c.getEngine(headers[0].Number.Uint64(), "VerifyHeaders").VerifyHeaders(chain, headers, seals)
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of a given engine.
func (c *Composite) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	return c.getEngine(block.Header().Number.Uint64(), "VerifyUncles").VerifyUncles(chain, block)
}

// Prepare initializes the consensus fields of a block header according to the
// rules of a particular engine. The changes are executed inline.
func (c *Composite) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	return c.getEngine(header.Number.Uint64(), "Prepare").Prepare(chain, header)
}

// Finalize runs any post-transaction state modifications (e.g. block rewards)
// but does not assemble the block.
//
// Note: The block header and state database might be updated to reflect any
// consensus rules that happen at finalization (e.g. block rewards).
func (c *Composite) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs *[]*types.Transaction,
	uncles []*types.Header, receipts *[]*types.Receipt, systemTxs []*types.Transaction) error {
	return c.getEngine(header.Number.Uint64(), "Finalize").Finalize(chain, header, state, txs, uncles, receipts, systemTxs)
}

// FinalizeAndAssemble runs any post-transaction state modifications (e.g. block
// rewards) and assembles the final block.
//
// Note: The block header and state database might be updated to reflect any
// consensus rules that happen at finalization (e.g. block rewards).
func (c *Composite) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, []*types.Receipt, error) {
	return c.getEngine(header.Number.Uint64(), "FinalizeAndAssemble").FinalizeAndAssemble(chain, header, state, txs, uncles, receipts)
}

// Seal generates a new sealing request for the given input block and pushes
// the result into the given channel.
//
// Note, the method returns immediately and will send the result async. More
// than one result may also be returned depending on the consensus algorithm.
func (c *Composite) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	return c.getEngine(block.Header().Number.Uint64(), "Seal").Seal(chain, block, results, stop)
}

// SealHash returns the hash of a block prior to it being sealed.
func (c *Composite) SealHash(header *types.Header) common.Hash {
	return c.getEngine(header.Number.Uint64(), "SealHash").SealHash(header)
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have.
func (c *Composite) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	return c.getEngine(parent.Number.Uint64(), "CalcDifficulty").CalcDifficulty(chain, time, parent)
}

// APIs returns the RPC APIs this consensus engine provides.
func (c *Composite) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return c.MyPoCliqueEngine.APIs(chain)
}

// Close terminates any background threads maintained by the consensus engine.
func (c *Composite) Close() error {
	c.PoCliqueEngine.Close()
	return c.MyPoCliqueEngine.Close()
}
