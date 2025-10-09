import z from 'zod';

export type StatusUnionType = z.infer<typeof StatusUnion>;
export const StatusUnion = z.union([
	z.literal('pending'),
	z.literal('uploaded'),
	z.literal('processing'),
	z.literal('ready'),
	z.literal('failed'),
]);

export type UploadLogoDto = z.infer<typeof UploadLogoSchema>;
export const UploadLogoSchema = z.object({
	key: z.string(),
	uploadUrl: z.url(),
	expiresAt: z.coerce.date(),
	mediaId: z.string(),
	urls: z.record(z.string(), z.url()),
	status: StatusUnion,
	pending_event: z.looseObject({
		stream: z.string(),
		subject: z.string(),
		msgId: z.string(),
		data: z.looseObject({
			key: z.string(),
			mime: z.string(),
			uploaderId: z.string(),
			status: StatusUnion,
			updatedAt: z.coerce.date(),
			transitionOk: z.boolean(),
		}),
	}),
});
