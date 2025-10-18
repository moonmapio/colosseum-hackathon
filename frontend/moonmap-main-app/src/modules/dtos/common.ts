export enum ProjectChain {
	ETHEREUM = 'ETHEREUM',
	SOLANA = 'SOLANA',
	BASE = 'BASE',
	BNB = 'BNB',
	AVALANCHE = 'AVALANCHE',
	POLYGON = 'POLYGON',
	BLAST = 'BLAST',
	OTHER = 'OTHER',
}

export const normalizeChain = (raw: unknown): string | undefined => {
	if (typeof raw !== 'string') return undefined;

	const up = raw.trim().toUpperCase();

	const map: Record<string, ProjectChain> = {
		'BINANCE-SMART-CHAIN': ProjectChain.BNB,
		BNB: ProjectChain.BNB,
		BSC: ProjectChain.BNB,

		SOLANA: ProjectChain.SOLANA,
		ETHEREUM: ProjectChain.ETHEREUM,
		POLYGON: ProjectChain.POLYGON,
		MATIC: ProjectChain.POLYGON,
		BASE: ProjectChain.BASE,
		TRON: ProjectChain.OTHER,
		XDAI: ProjectChain.OTHER,
		XRP: ProjectChain.OTHER,
		BITTENSOR: ProjectChain.OTHER,
		ALEPHIUM: ProjectChain.OTHER,
		'THE-OPEN-NETWORK': ProjectChain.OTHER,
		UNKNOWN: ProjectChain.OTHER,
	};

	return map[up] ?? ProjectChain.OTHER; // fallback
};

export type WithID = {
	id?: string;
};
