use anchor_lang::prelude::*;

#[account]
#[derive(InitSpace)]
pub struct ValidatorState {
    pub bump: u8,
    pub validator: Pubkey,
    pub nonce: u64,
    #[max_len(100)]
    pub actions: Vec<u64>,
}

#[derive(AnchorSerialize, AnchorDeserialize, Clone, Copy, PartialEq, Eq, InitSpace)]
pub enum ApplicationStatus {
    Pending,
    Approved,
    Rejected,
    Banned,
}

#[account]
#[derive(InitSpace)]
pub struct ValidatorApplication {
    pub bump: u8,
    pub applicant: Pubkey,
    pub action_id: u64,
    pub status: ApplicationStatus,
    #[max_len(2048)]
    pub encrypted_data: String,
    pub nonce_used: u64,
    pub created_at: i64,
    pub updated_at: i64,
    pub approved_by: Option<Pubkey>,
    pub rejected_by: Option<Pubkey>,
    pub banned_by: Option<Pubkey>,
}
