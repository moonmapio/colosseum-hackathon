use anchor_lang::prelude::*;
use arcium_anchor::prelude::*;

pub mod consts;
pub mod contexts;
pub mod errors;
pub mod instructions;
pub mod state;

use contexts::*;
use instructions::*;

declare_id!("9Kg3tLr32aM2S7rUixQ299x6f7YwsAvYYHeQHqZmoon");

#[arcium_program]
pub mod moonmap_validator {
    use super::*;

    pub fn ix_initialize_global(ctx: Context<InitializeGlobal>) -> Result<()> {
        instructions::admin::ix_initialize_global(ctx)
    }

    pub fn initialize_validator_state(ctx: Context<InitializeValidatorState>) -> Result<()> {
        let validator_state = &mut ctx.accounts.validator_state;
        validator_state.bump = ctx.bumps.validator_state;
        validator_state.validator = ctx.accounts.validator.key();
        validator_state.nonce = 0;
        validator_state.actions = Vec::new();
        Ok(())
    }
}
