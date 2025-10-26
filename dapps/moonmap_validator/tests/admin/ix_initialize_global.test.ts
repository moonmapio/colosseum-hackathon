import * as anchor from "@coral-xyz/anchor";
import { Program } from "@coral-xyz/anchor";
import { MoonmapValidator } from "@types";
import { expect } from "chai";

describe("ix_initialize_global", () => {
  const provider = anchor.AnchorProvider.env();
  anchor.setProvider(provider);

  const program = anchor.workspace
    .MoonmapValidator as Program<MoonmapValidator>;
  const systemProgram = anchor.web3.SystemProgram.programId;

  const [globalPda] = anchor.web3.PublicKey.findProgramAddressSync(
    [Buffer.from("global")],
    program.programId
  );

  it("Fails when a non-authority tries to initialize global", async () => {
    try {
      const badActor = anchor.web3.Keypair.generate();
      const connection = provider.connection;
      const airdropSig = await connection.requestAirdrop(
        badActor.publicKey,
        1_000_000_000
      );
      await connection.confirmTransaction(airdropSig);

      await program.methods
        .ixInitializeGlobal()
        .accounts({
          authority: badActor.publicKey,
          // @ts-ignore
          global: globalPda,
          systemProgram,
        })
        .signers([badActor])
        .rpc();

      throw new Error("Expected transaction to fail but it succeeded");
    } catch (err: any) {
      const code = err.error?.errorCode?.code;
      expect(code).to.equal("Unauthorized");
    }
  });

  it("Succeeds when the moonmap authority initializes global", async () => {
    const authority = provider.wallet;

    const txSig = await program.methods
      .ixInitializeGlobal()
      .accounts({
        authority: authority.publicKey,
        // @ts-ignore
        global: globalPda,
        systemProgram,
      })
      .rpc();

    const globalAccount = await program.account.global.fetch(globalPda);

    expect(globalAccount.authority.toBase58()).to.equal(
      authority.publicKey.toBase58()
    );
    expect(globalAccount.totalApplications.toNumber()).to.equal(0);
    expect(globalAccount.totalApproved.toNumber()).to.equal(0);
    expect(globalAccount.totalRejected.toNumber()).to.equal(0);
    expect(globalAccount.totalBanned.toNumber()).to.equal(0);
  });
});
