use anchor_lang::prelude::*;
use crate::state::*;
use crate::consts::MOONMAP_AUTHORITY_PUBKEY;
use crate::errors::ErrorCode;

#[derive(Accounts)]
pub struct InitializeGlobal<'info> {
    #[account(
        mut,
        constraint = authority.key() == MOONMAP_AUTHORITY_PUBKEY @ ErrorCode::Unauthorized
    )]
    pub authority: Signer<'info>,
    #[account(
        init,
        payer = authority,
        space = 8 + Global::INIT_SPACE,
        seeds = [b"global"],
        bump
    )]
    pub global: Account<'info, Global>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct InitializeValidatorState<'info> {
    #[account(mut)]
    pub validator: Signer<'info>,
    #[account(
        init,
        payer = validator,
        space = 8 + ValidatorState::INIT_SPACE,
        seeds = [b"validator_state", validator.key().as_ref()],
        bump
    )]
    pub validator_state: Account<'info, ValidatorState>,
    pub system_program: Program<'info, System>,
}
