use anchor_lang::prelude::*;

#[error_code]
pub enum MMError {
    #[msg("unauthorized")]
    Unauthorized,

    #[msg("invalid fees")]
    InvalidFees,

    #[msg("invalid status")]
    InvalidStatus,

    #[msg("invalid outcome")]
    InvalidOutcome,

    #[msg("overflow")]
    Overflow,

    #[msg("insufficient shares")]
    InsufficientShares,

    #[msg("window ended")]
    WindowEnded,

    #[msg("window not ended")]
    WindowNotEnded,

    #[msg("no evidence")]
    NoEvidence,
}
