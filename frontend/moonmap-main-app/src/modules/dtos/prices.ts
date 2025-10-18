import z from 'zod';

export type PriceRelayCoin = z.infer<typeof PriceRelayCoinSchema>;
export const PriceRelayCoinSchema = z.object({
	symbol: z.string(),
	price: z.number(),
	source: z.string(),
	at: z.coerce.date(),
});

export type PriceRelayResponse = z.infer<typeof PriceRelayResponseSchema>;
export const PriceRelayResponseSchema = z.looseObject({
	BTC: PriceRelayCoinSchema,
	SOL: PriceRelayCoinSchema,
});
