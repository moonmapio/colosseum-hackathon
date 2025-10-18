import TruthPoolsSimulator from "@/lib/Simulator";

export default function Home() {
  return (
    <div className="min-h-screen p-6 bg-background text-foreground">
      <main className="max-w-7xl mx-auto">
        <TruthPoolsSimulator />
      </main>
    </div>
  );
}
