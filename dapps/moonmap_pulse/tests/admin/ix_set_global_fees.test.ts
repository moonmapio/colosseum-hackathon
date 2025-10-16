import * as anchor from "@coral-xyz/anchor";
import { Program } from "@coral-xyz/anchor";
import { MoonmapPulse } from "@types";
import { expect } from "chai";

describe("ix_set_global_fees", () => {
  const provider = anchor.AnchorProvider.env();
  anchor.setProvider(provider);

  const program = anchor.workspace.MoonmapPulse as Program<MoonmapPulse>;
  const systemProgram = anchor.web3.SystemProgram.programId;
  const deployer = provider.wallet;
  const badActor = anchor.web3.Keypair.generate();

  // Derive the same PDA for global
  const [globalPda] = anchor.web3.PublicKey.findProgramAddressSync(
    [Buffer.from("global")],
    program.programId
  );

  before(async () => {
    // ensure global exists (init once)
    try {
      const connection = provider.connection;
      const airdropSig = await connection.requestAirdrop(
        badActor.publicKey,
        1_000_000_000
      );
      const latestBlockhash = await connection.getLatestBlockhash();

      await connection.confirmTransaction({
        signature: airdropSig,
        ...latestBlockhash,
      });

      await program.methods
        .initializeGlobal()
        .accounts({
          authority: deployer.publicKey,
          // @ts-ignore — IDL compiled, but TypeScript doesn't recognize the following argument
          global: globalPda,
          systemProgram,
        })
        .rpc();
    } catch (_) {
      // ignore if already initialized
    }
  });

  it("Updates global fees successfully (by authority)", async () => {
    const txSig = await program.methods
      .setGlobalFees(2000, 3000, 4000)
      .accounts({
        authority: deployer.publicKey,
        global: globalPda,
      })
      .rpc();

    const global = await program.account.global.fetch(globalPda);
    expect(global.moonmapFeeBps).to.equal(2000);
    expect(global.validatorFeeBps).to.equal(3000);
    expect(global.creatorFeeBps).to.equal(4000);
  });

  it("Fails if called by non-authority", async () => {
    try {
      await program.methods
        .setGlobalFees(100, 100, 100)
        .accounts({
          authority: badActor.publicKey,
          global: globalPda,
        })
        .signers([badActor])
        .rpc();

      throw new Error("Expected to fail but succeeded");
    } catch (err: any) {
      const code = err.error?.errorCode?.code;
      expect(code).to.equal("ConstraintHasOne");
    }
  });

  it("Fails when total fees exceed 10_000 (100%)", async () => {
    try {
      await program.methods
        .setGlobalFees(8000, 3000, 1000) // overflow 12_000 total
        .accounts({
          authority: deployer.publicKey,
          global: globalPda,
        })
        .rpc();

      throw new Error("Expected to fail but succeeded");
    } catch (err: any) {
      const code = err.error?.errorCode?.code;
      expect(code).to.equal("InvalidFees");
    }
  });
});
