use anchor_lang::prelude::*;

#[account]
#[derive(InitSpace, Debug)]
pub struct Global {
    pub authority: Pubkey,
    pub moonmap_fee_bps: u16,
    pub validator_fee_bps: u16,
    pub creator_fee_bps: u16,
}
