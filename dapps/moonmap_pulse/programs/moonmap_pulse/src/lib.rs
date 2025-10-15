use anchor_lang::prelude::*;

pub mod consts;
pub mod errors;
pub mod instructions;
pub mod math;
pub mod state;

use instructions::*;

declare_id!("3R7mbvtHmbBwX9KwXmf72ZFqzkyGfFRM8VWELfKHEf4g");

#[program]
pub mod moonmap_pulse {
    use super::*;

    pub fn initialize_global(ctx: Context<InitializeGlobal>) -> Result<()> {
        instructions::initialize_global(ctx)
    }

    pub fn set_global_fees(
        ctx: Context<SetGlobalFees>,
        moonmap_bps: u16,
        validator_bps: u16,
        creator_bps: u16,
    ) -> Result<()> {
        instructions::set_global_fees(ctx, moonmap_bps, validator_bps, creator_bps)
    }
}
