use anchor_lang::prelude::*;

#[account]
#[derive(InitSpace)]
pub struct Global {
    pub authority: Pubkey,
    pub total_applications: u64,
    pub total_approved: u64,
    pub total_rejected: u64,
    pub total_banned: u64,
}
