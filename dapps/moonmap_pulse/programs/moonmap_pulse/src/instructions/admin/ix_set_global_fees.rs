use crate::{consts::*, math::safe_sum_with_u16cap, state::*};
use anchor_lang::prelude::*;

#[derive(Accounts)]
pub struct SetGlobalFees<'info> {
    pub authority: Signer<'info>,
    #[account(mut, seeds=[PDA_GLOBAL], bump, has_one = authority)]
    pub global: Account<'info, Global>,
}

pub fn set_global_fees(
    ctx: Context<SetGlobalFees>,
    moonmap_bps: u16,
    validator_bps: u16,
    creator_bps: u16,
) -> Result<()> {
    safe_sum_with_u16cap(
        &[moonmap_bps, validator_bps, creator_bps],
        MAX_POSSIBLE_FEES,
    )?;

    let g = &mut ctx.accounts.global;
    g.moonmap_fee_bps = moonmap_bps;
    g.validator_fee_bps = validator_bps;
    g.creator_fee_bps = creator_bps;
    Ok(())
}
