import * as anchor from "@coral-xyz/anchor";
import { Program } from "@coral-xyz/anchor";
import { MoonmapPulse } from "@types";
import { expect } from "chai";

describe("ix_initialize_global", () => {
  // Configure the client to use the local cluster.
  const provider = anchor.AnchorProvider.env();
  anchor.setProvider(provider);

  const program = anchor.workspace.MoonmapPulse as Program<MoonmapPulse>;
  const systemProgram = anchor.web3.SystemProgram.programId;

  // 🧠 Derive PDA for the global account
  const [globalPda] = anchor.web3.PublicKey.findProgramAddressSync(
    [Buffer.from("global")],
    program.programId
  );

  it("Fails when a non-deployer tries to initialize global", async () => {
    try {
      // 👤 wrong wallet (not authorized)
      const badActor = anchor.web3.Keypair.generate();
      const connection = provider.connection;
      const airdropSig = await connection.requestAirdrop(
        badActor.publicKey,
        1_000_000_000 // 1 SOL
      );
      await connection.confirmTransaction(airdropSig);

      await program.methods
        .initializeGlobal()
        .accounts({
          authority: badActor.publicKey,
          // @ts-ignore — IDL compiled, but TypeScript doesn't recognize the following argument
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

  it("Succeeds when the deployer initializes global", async () => {
    // 👤 Wallet that deployed the program (authority)

    // 👤 real wallet (deployer)
    const deployer = provider.wallet;

    // 🧩 Call instruction
    const txSig = await program.methods
      .initializeGlobal()
      .accounts({
        authority: deployer.publicKey,
        // @ts-ignore — IDL compiled, but TypeScript doesn't recognize the following argument
        global: globalPda,
        systemProgram,
      })
      .rpc();

    // 🔍 Fetch account data from chain
    const globalAccount = await program.account.global.fetch(globalPda);

    // ✅ Assertions
    expect(globalAccount.authority.toBase58()).to.equal(
      deployer.publicKey.toBase58()
    );
    expect(globalAccount.moonmapFeeBps).to.equal(0);
    expect(globalAccount.validatorFeeBps).to.equal(0);
    expect(globalAccount.creatorFeeBps).to.equal(0);
  });
});
