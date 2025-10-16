use anchor_lang::constant;

pub const PDA_GLOBAL: &[u8] = b"global";
pub const PDA_MARKET: &[u8] = b"market";
pub const PDA_ESCROW: &[u8] = b"escrow";
pub const PDA_POS: &[u8] = b"position";
pub const PDA_VALIDATOR: &[u8] = b"validator";
pub const PDA_EVIDENCE: &[u8] = b"evidence";
pub const PDA_VREQ: &[u8] = b"vreq";

#[constant]
pub const MAX_OUTCOMES: u8 = 8;

#[constant]
pub const MAX_NAME: u8 = 64;

#[constant]
pub const MAX_URI: u8 = 255;

// 0.5 % = 50, 0.05 % = 5, 100 % = 10 000
#[constant]
pub const MAX_POSSIBLE_FEES: u16 = 10_000;
