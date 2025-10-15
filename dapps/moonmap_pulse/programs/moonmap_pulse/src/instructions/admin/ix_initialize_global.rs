use crate::{consts::PDA_GLOBAL, errors::MMError, state::*};
use anchor_lang::prelude::*;

#[derive(Accounts)]
pub struct InitializeGlobal<'info> {
    #[account(mut)]
    pub authority: Signer<'info>,
    #[account(init, payer = authority, space = 8 + Global::INIT_SPACE, seeds=[PDA_GLOBAL], bump)]
    pub global: Account<'info, Global>,
    pub system_program: Program<'info, System>,
}

pub fn initialize_global(ctx: Context<InitializeGlobal>) -> Result<()> {
    // wallet = "~/.config/solana/moonmap_authority.json"
    let deployer = Pubkey::from_str_const("9Bjg6RJjFpwCPMAEhWYnaLSvS3H2euidsgyw2V9g5hMR");
    require_keys_eq!(
        ctx.accounts.authority.key(),
        deployer,
        MMError::Unauthorized
    );

    let g = &mut ctx.accounts.global;
    g.authority = ctx.accounts.authority.key();
    g.moonmap_fee_bps = 0;
    g.validator_fee_bps = 0;
    g.creator_fee_bps = 0;
    Ok(())
}
