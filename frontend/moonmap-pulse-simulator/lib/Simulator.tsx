"use client";

import React, { useState, useMemo } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  ColumnDef,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  flexRender,
} from "@tanstack/react-table";
import {
  PieChart,
  Pie,
  Cell,
  Tooltip as ReTooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  LineChart,
  Line,
} from "recharts";
import { nanoid } from "nanoid";
import { CheckCircle, XCircle, ArrowUpDown, Info } from "lucide-react";

type User = {
  id: string;
  choice: number;
  bet: number;
  price: number;
  isWinner?: boolean;
  payout?: number;
  profit?: number;
};

type Simulation = {
  users: User[];
  totalPool: number;
  q: number[];
};

type WinnerResult = {
  option: number;
  moonmapFee: number;
  validatorFee: number;
  creatorFee: number;
  results: User[];
};

function formatMoney(value: number): string {
  return value.toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

function randomBet(min: number, max: number): number {
  const r = Math.random();
  const skewed = Math.pow(r, 3); // cuanto mayor el exponente, más sesgo hacia min
  return min + skewed * (max - min);
}

function randomBiasedChoice(numOptions: number): number {
  // más peso a las opciones del centro
  const r = Math.random();
  const bias = Math.pow(r, 1.5); // cuanto mayor el exponente, más sesgo
  return Math.floor(bias * numOptions);
}

// function randomWeightedChoice(numOptions: number): number {
//   const weights = Array.from({ length: numOptions }, () => Math.random());
//   const total = weights.reduce((a, b) => a + b, 0);
//   const rnd = Math.random() * total;
//   let cum = 0;
//   for (let i = 0; i < numOptions; i++) {
//     cum += weights[i];
//     if (rnd < cum) return i;
//   }
//   return numOptions - 1;
// }

function lmsrPrice(q: number[], b: number): number[] {
  const sumExp = q.reduce((acc, qi) => acc + Math.exp(qi / b), 0);
  return q.map((qi) => Math.exp(qi / b) / sumExp);
}

function randomNormal(mean: number, stdDev: number): number {
  const u1 = Math.random();
  const u2 = Math.random();
  const z = Math.sqrt(-2.0 * Math.log(u1)) * Math.cos(2.0 * Math.PI * u2);
  return mean + z * stdDev;
}

function checkPayoutIntegrity(
  totalPool: number,
  moonmapFee: number,
  validatorFee: number,
  creatorFee: number,
  results: User[]
) {
  const totalPayout = results.reduce((sum, u) => sum + (u.payout ?? 0), 0);
  const distributed = moonmapFee + validatorFee + creatorFee + totalPayout;
  const diff = totalPool - distributed;

  return {
    totalPayout,
    distributed,
    diff,
    fullyPaid: Math.abs(diff) < 0.01, // tolerancia de centavos
  };
}

export default function TruthPoolsSimulator() {
  const [numUsers, setNumUsers] = useState<number>(1000);
  const [numOptions, setNumOptions] = useState<number>(4);
  const [minBet, setMinBet] = useState<number>(10);
  const [maxBet, setMaxBet] = useState<number>(100);
  const [bParam, setBParam] = useState<number>(100);
  const [simulation, setSimulation] = useState<Simulation | null>(null);
  const [winner, setWinner] = useState<WinnerResult | null>(null);
  const [integrity, setIntegrity] = useState<{
    totalPayout: number;
    distributed: number;
    diff: number;
    fullyPaid: boolean;
  } | null>(null);
  const COLORS = ["#0088FE", "#00C49F", "#FFBB28", "#FF8042", "#A020F0"];
  const [moonmapFeePercent, setMoonmapFeePercent] = useState<number>(5);
  const [validatorFeePercent, setValidatorFeePercent] = useState<number>(5);
  const [creatorFeePercent, setCreatorFeePercent] = useState<number>(0.1);

  const simulate = (): void => {
    const q = Array(numOptions).fill(0);
    const users: User[] = [];

    const mean = (minBet + maxBet) / 2;
    const stdDev = (maxBet - minBet) / 3; // controla la dispersión

    for (let i = 0; i < numUsers; i++) {
      const choice = randomBiasedChoice(numOptions);
      //   const bet = Math.random() * (maxBet - minBet) + minBet;
      let bet = randomNormal(mean, stdDev);
      bet = Math.max(minBet, Math.min(maxBet, bet)); // recorta a los límites
      const prices = lmsrPrice(q, bParam);
      q[choice] += bet;
      users.push({ id: nanoid(6), choice, bet, price: prices[choice] });
    }

    const totalPool = users.reduce((sum, u) => sum + u.bet, 0);
    setSimulation({ users, totalPool, q });
    setWinner(null);
  };

  const validate = (): void => {
    if (!simulation) return;
    const chosen = Math.floor(Math.random() * numOptions);
    const totalPool = simulation.totalPool;
    const winners = simulation.users.filter((u) => u.choice === chosen);
    const winTotal = winners.reduce((sum, w) => sum + w.bet, 0);

    const moonmapFee = totalPool * (moonmapFeePercent / 100);
    const validatorFee = totalPool * (validatorFeePercent / 100);
    const creatorFee = totalPool * (creatorFeePercent / 100);
    const prizePool = totalPool - moonmapFee - validatorFee - creatorFee;

    const results: User[] = simulation.users.map((u) => {
      const isWinner = u.choice === chosen;
      const payout = isWinner ? (u.bet / winTotal) * prizePool : 0;
      const profit = payout - u.bet;
      return { ...u, isWinner, payout, profit };
    });

    setWinner({
      option: chosen,
      moonmapFee,
      validatorFee,
      creatorFee,
      results,
    });

    const integrity = checkPayoutIntegrity(
      totalPool,
      moonmapFee,
      validatorFee,
      creatorFee,
      results
    );
    console.log("Integrity Check:", integrity);
    setIntegrity(integrity);
  };

  const pieData = winner
    ? [
        { name: "MoonMap", value: winner.moonmapFee },
        { name: "Validators", value: winner.validatorFee },
        { name: "Creator", value: winner.creatorFee },
        {
          name: "Players",
          value: winner.results.reduce((sum, u) => sum + (u.payout ?? 0), 0),
        },
      ]
    : [];

  const betDistribution =
    simulation &&
    Array.from({ length: numOptions }, (_, i) => {
      const usersByOption = simulation.users.filter((u) => u.choice === i);
      const total = usersByOption.reduce((s, u) => s + u.bet, 0);
      return {
        option: `Option ${i + 1}`,
        total,
        count: usersByOption.length,
      };
    });

  const metrics = useMemo(() => {
    if (!winner) return null;
    const totalUsers = winner.results.length;
    const winnersCount = winner.results.filter((u) => u.isWinner).length;
    const losersCount = totalUsers - winnersCount;
    const avgROI =
      winner.results.reduce(
        (acc, u) => acc + ((u.payout ?? 0) / u.bet - 1),
        0
      ) / totalUsers;
    const avgProfit =
      winner.results.reduce((acc, u) => acc + (u.profit ?? 0), 0) / totalUsers;
    return { totalUsers, winnersCount, losersCount, avgROI, avgProfit };
  }, [winner]);

  const columns = useMemo<ColumnDef<User>[]>(
    () => [
      { accessorKey: "id", header: "ksid" },
      {
        accessorKey: "Option",
        header: "Choice",
        cell: ({ row }) => `${row.original.choice + 1}`,
      },
      {
        accessorKey: "bet",
        header: ({ column }) => (
          <button
            className="flex items-center gap-1"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            Bet <ArrowUpDown className="h-4 w-4 text-muted-foreground" />
          </button>
        ),
        cell: ({ row }) => `$${formatMoney(row.original.bet)}`,
      },
      {
        accessorKey: "payout",
        header: ({ column }) => (
          <button
            className="flex items-center gap-1"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            Payout <ArrowUpDown className="h-4 w-4 text-muted-foreground" />
          </button>
        ),
        cell: ({ row }) => `$${formatMoney(row.original.payout ?? 0)}`,
      },
      {
        accessorKey: "profit",
        header: ({ column }) => (
          <button
            className="flex items-center gap-1"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            Profit <ArrowUpDown className="h-4 w-4 text-muted-foreground" />
          </button>
        ),
        cell: ({ row }) => `$${formatMoney(row.original.profit ?? 0)}`,
      },
      {
        accessorKey: "isWinner",
        header: "Result",
        cell: ({ row }) =>
          row.original.isWinner ? (
            <CheckCircle className="text-primary h-5 w-5" />
          ) : (
            <XCircle className="text-muted-foreground h-5 w-5" />
          ),
      },
    ],
    []
  );

  const table = useReactTable({
    data: winner?.results ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
  });

  return (
    <TooltipProvider>
      <div className="p-6 space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Truth Pools Simulator</CardTitle>
          </CardHeader>
          <CardContent className="grid sm:grid-cols-5 gap-4">
            {[
              {
                label: "Users",
                value: numUsers,
                set: setNumUsers,
                info: "Número total de usuarios que apuestan.",
              },
              {
                label: "Options",
                value: numOptions,
                set: setNumOptions,
                info: "Cantidad de opciones del mercado.",
              },
              {
                label: "Min Bet",
                value: minBet,
                set: setMinBet,
                info: "Apuesta mínima por usuario.",
              },
              {
                label: "Max Bet",
                value: maxBet,
                set: setMaxBet,
                info: "Apuesta máxima por usuario.",
              },
              {
                label: "B (LMSR)",
                value: bParam,
                set: setBParam,
                info: "Parámetro de liquidez del modelo LMSR.",
              },
              {
                label: "MoonMap Fee (%)",
                value: moonmapFeePercent,
                set: setMoonmapFeePercent,
                info: "Percentage of the total amount allocated to MoonMap.",
              },
              {
                label: "Validator Fee (%)",
                value: validatorFeePercent,
                set: setValidatorFeePercent,
                info: "Percentage of the total amount allocated to validators.",
              },
              {
                label: "Creator Fee (%)",
                value: creatorFeePercent,
                set: setCreatorFeePercent,
                info: "Percentage of the total amount allocated to the market creator.",
              },
            ].map((p) => (
              <div key={p.label}>
                <div className="flex items-center gap-1 mb-1">
                  <span className="text-sm font-medium">{p.label}</span>
                  <Tooltip>
                    <TooltipTrigger>
                      <Info className="h-4 w-4 text-muted-foreground" />
                    </TooltipTrigger>
                    <TooltipContent>{p.info}</TooltipContent>
                  </Tooltip>
                </div>
                <Input
                  type="number"
                  value={p.value}
                  onChange={(e) => p.set(+e.target.value)}
                />
              </div>
            ))}
          </CardContent>
          <Button className="mx-4" onClick={simulate}>
            Simulate Market
          </Button>
        </Card>

        {simulation && (
          <div className="space-y-4">
            <p className="text-lg font-semibold">
              Total Pool Accumulated: ${formatMoney(simulation.totalPool)}
            </p>
            <Button onClick={validate}>Validate (choose winner)</Button>
          </div>
        )}

        {winner && (
          <Card>
            <CardHeader>
              <CardTitle
                className={winner ? "text-primary" : "text-muted-foreground"}
              >
                Results{" "}
                {winner
                  ? `(Winning Option: ${winner.option + 1})`
                  : "(No winner yet)"}
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid grid-cols-3 gap-6">
                <div>
                  <p>MoonMap Fee: ${formatMoney(winner.moonmapFee)}</p>
                  <p>Validator Fee: ${formatMoney(winner.validatorFee)}</p>
                  <p>Creator Fee: ${formatMoney(winner.creatorFee)}</p>
                  {metrics && (
                    <div className="mt-4 text-sm space-y-1">
                      <p>Total Users: {metrics.totalUsers.toLocaleString()}</p>
                      <p>Winners: {metrics.winnersCount.toLocaleString()}</p>
                      <p>Losers: {metrics.losersCount.toLocaleString()}</p>
                      <p>Avg ROI: {(metrics.avgROI * 100).toFixed(2)}%</p>
                      <p>Avg Profit: ${formatMoney(metrics.avgProfit)}</p>
                    </div>
                  )}

                  {integrity && (
                    <div className="mt-4 text-sm space-y-1 border-t pt-2">
                      <p className="font-semibold">Integrity Check:</p>
                      <p>Total Payout: ${formatMoney(integrity.totalPayout)}</p>
                      <p>Distributed: ${formatMoney(integrity.distributed)}</p>
                      <p>
                        Difference:{" "}
                        <span
                          className={
                            Math.abs(integrity.diff) < 0.01
                              ? "text-green-600 font-semibold"
                              : "text-red-600 font-semibold"
                          }
                        >
                          ${formatMoney(integrity.diff)}
                        </span>
                      </p>
                      <p>
                        Fully Paid:{" "}
                        {integrity.fullyPaid ? (
                          <CheckCircle className="text-primary inline h-4 w-4" />
                        ) : (
                          <XCircle className="text-muted-foreground inline h-4 w-4" />
                        )}
                      </p>
                    </div>
                  )}
                </div>
                <div className="col-span-2 h-64">
                  <ResponsiveContainer>
                    <PieChart>
                      <Pie dataKey="value" data={pieData} label>
                        {pieData.map((entry, index) => (
                          <Cell
                            key={`cell-${index}`}
                            fill={COLORS[index % COLORS.length]}
                          />
                        ))}
                      </Pie>
                      <ReTooltip />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              </div>

              {/* More Charts */}
              {betDistribution && (
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={betDistribution}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="option" />
                    <YAxis />
                    <ReTooltip
                      content={({ active, payload, label }) => {
                        if (active && payload && payload.length) {
                          const data = payload[0].payload;
                          return (
                            <div className="bg-background/90 border rounded-md shadow p-2 text-sm">
                              <p className="font-semibold">{label}</p>
                              <p>Users: {data.count.toLocaleString()}</p>
                              <p>Total Bet: ${formatMoney(data.total)}</p>
                            </div>
                          );
                        }
                        return null;
                      }}
                    />
                    <Bar
                      dataKey="count"
                      fill="#82ca9d"
                      name="Users"
                      //   barSize={40}
                    />
                  </BarChart>
                </ResponsiveContainer>
              )}

              <ResponsiveContainer width="100%" height={250}>
                <LineChart
                  data={winner.results.slice(0, 200)}
                  margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
                >
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="id" hide />
                  <YAxis />
                  <Tooltip />
                  <Line
                    type="monotone"
                    dataKey="profit"
                    stroke="#82ca9d"
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>

              {/* Data Table */}
              <div>
                <Table>
                  <TableHeader>
                    {table.getHeaderGroups().map((headerGroup) => (
                      <TableRow key={headerGroup.id}>
                        {headerGroup.headers.map((header) => (
                          <TableHead key={header.id}>
                            {flexRender(
                              header.column.columnDef.header,
                              header.getContext()
                            )}
                          </TableHead>
                        ))}
                      </TableRow>
                    ))}
                  </TableHeader>
                  <TableBody>
                    {table.getRowModel().rows.map((row) => (
                      <TableRow key={row.id}>
                        {row.getVisibleCells().map((cell) => (
                          <TableCell key={cell.id}>
                            {flexRender(
                              cell.column.columnDef.cell,
                              cell.getContext()
                            )}
                          </TableCell>
                        ))}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
                <div className="flex items-center justify-between mt-4">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => table.setPageIndex(0)}
                    disabled={!table.getCanPreviousPage()}
                  >
                    First
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => table.previousPage()}
                    disabled={!table.getCanPreviousPage()}
                  >
                    Previous
                  </Button>
                  <span className="text-sm">
                    Page {table.getState().pagination.pageIndex + 1} of{" "}
                    {table.getPageCount()}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => table.nextPage()}
                    disabled={!table.getCanNextPage()}
                  >
                    Next
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => table.setPageIndex(table.getPageCount() - 1)}
                    disabled={!table.getCanNextPage()}
                  >
                    Last
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </TooltipProvider>
  );
}
