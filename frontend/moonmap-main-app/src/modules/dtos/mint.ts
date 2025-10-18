import { z } from 'zod';

export const MintRecordDtoSchema = z.object({
	mint: z.string(),
	metadata_pda: z.string(),
	name: z.string(),
	symbol: z.string(),
	uri: z.string(),
	uri_resolved: z
		.looseObject({
			createdOn: z.string().optional(),
			createdOnName: z.string().optional(),
			description: z.string().optional(),
			image: z.string().optional(),
			name: z.string().optional(),
			symbol: z.string().optional(),

			twitter: z.string().optional(),
			telegram: z.string().optional(),
			discord: z.string().optional(),
			website: z.string().optional(),
		})
		.catchall(z.string().optional())
		.nullable()
		.optional(),
	seller_fee_bps: z.number(),
	update_authority: z.string(),
	decimals: z.number().nullable(),
	first_seen_sig: z.string(),
	last_seen_sig: z.string(),
	first_slot: z.number(),
	last_slot: z.number(),
	created_at: z.coerce.date(),
	recived_at: z.string().optional().nullable(),
	holders: z.number(),
	volume_1m: z.number(),
	volume_5m: z.number(),
	volume_10m: z.number(),
	interactions: z.number(),
	last_updated: z.coerce.date(),
});

export type MintRecordDto = z.infer<typeof MintRecordDtoSchema>;
