use anchor_lang::prelude::*;
use crate::contexts::InitializeGlobal;

pub fn ix_initialize_global(ctx: Context<InitializeGlobal>) -> Result<()> {
    let global = &mut ctx.accounts.global;
    global.authority = ctx.accounts.authority.key();
    global.total_applications = 0;
    global.total_approved = 0;
    global.total_rejected = 0;
    global.total_banned = 0;
    Ok(())
}
