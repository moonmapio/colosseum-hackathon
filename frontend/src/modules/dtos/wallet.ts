import { ObjectId } from 'mongodb';
import z from 'zod';

export type PublicWalletResDto = z.infer<typeof PublicWalletResDtoSchema>;
export const PublicWalletResDtoSchema = z.object({
	// _id: z.string().regex(objectIdRegex, { message: 'Invalid ObjectId' }),
	// _id: z.string().map(value => mongo.ObjectId.createFromHexString(value)),

	_id: z
		.custom<ObjectId>()
		.transform((s) => s.toHexString())
		.refine((s) => ObjectId.isValid(s)),

	walletAddress: z.string().min(1),
	// secret: z.string().max(0),
	origin: z.string(),
	ip: z.string(),
	host: z.string(),
	fullName: z.string(),
	xAccount: z.string().optional(),
	createdBy: z.literal('auto'),
	downloaded: z.boolean(),
	version: z.string(),
	development: z.boolean(),
	createdAt: z.coerce.date(),
	lastSeenAt: z.coerce.date(),
	avatarUrl: z.url(),
	// verified: z.boolean(),
});

export type PrivateWalletResDto = z.infer<typeof PrivateWalletResDtoSchema>;
export const PrivateWalletResDtoSchema = PublicWalletResDtoSchema.extend({
	secret: z.string(),
});

export type PrivateWalletAuditResDtoKeys = 'fullName' | 'xAccount';
export const PrivateWalletAuditResDtoKeysSchema = z.union([z.literal('fullName'), z.literal('xAccount')]);
export type PrivateWalletAuditResDto = {
	_id: ObjectId;
	walletAddress: string;
	field: PrivateWalletAuditResDtoKeys;
	old: string | null;
	new: string | null;
	at: Date;
	actor: 'moonmap_team' | 'user';
	source: 'web' | 'api' | 'batch';
};
