import { z } from 'zod';

export const SphereSchema = z.object({
	_id: z.string(),
	mint: z.string(),
	createdBy: z.string(),
	createdAt: z.coerce.date(),
	lastUpdated: z.coerce.date(),
});
export type Sphere = z.infer<typeof SphereSchema>;

const ContentBase = z.object({
	// userId: z.string(),
	user: z.object({
		_id: z.string(),
		fullName: z.string(),
		avatarUrl: z.url(),
		// walletAddress: z.string(),
	}),
	type: z.string(),
	text: z.string().optional(),
	// reactions: z.record(z.string(), z.array(z.string())).optional().default({}),
	// reactions: z.record(z.string(), z.number()).optional().default({}),
	reactions: z.record(z.string(), z.string().array()).optional().default({}),

	mediaUrls: z.array(z.string()).optional().default([]),
	parentId: z.string().nullable(),
	createdAt: z.coerce.date(),
	updatedAt: z.coerce.date(),
	deleted: z.boolean(),
});

export const SphereContentCreatedSchema = ContentBase.extend({
	_id: z.string(), // hex
	sphereId: z.string(), // hex
});
export type SphereContentCreated = z.infer<typeof SphereContentCreatedSchema>;

export const SphereContentUpdatedSchema = z.object({
	_id: z.string(),
	sphereId: z.string(),
	updates: z.looseObject({}), // { text?, mediaUrls?, updatedAt, ... }
	updatedAt: z.coerce.date(),
});
export type SphereContentUpdated = z.infer<typeof SphereContentUpdatedSchema>;

export const SphereContentDeletedSchema = z.object({
	_id: z.string(),
	parentId: z.string(),
	sphereId: z.string(),
	deleted: z.boolean(),
	updatedAt: z.coerce.date(),
});
export type SphereContentDeleted = z.infer<typeof SphereContentDeletedSchema>;

export const SphereContentReactionSchema = z.object({
	_id: z.string(),
	sphereId: z.string(),
	parentId: z.string().nullable(),
	symbol: z.string(),
	action: z.enum(['added', 'removed']),
	userId: z.string(),
});
export type SphereContentReaction = z.infer<typeof SphereContentReactionSchema>;

export const SphereContentSchema = z.object({
	_id: z.string(),
	sphereId: z.string(), //z.union([zObjectIdLike, z.string()]),
	// userId: z.string(),
	user: z.object({
		_id: z.string(),
		fullName: z.string(),
		avatarUrl: z.url(),
		// walletAddress: z.string(),
	}),
	type: z.string(),
	text: z.string().optional(),
	mediaUrls: z.array(z.string()).optional().default([]),
	reactions: z.record(z.string(), z.string().array()).optional().default({}),
	parentId: z.string().nullable(),
	createdAt: z.coerce.date(),
	updatedAt: z.coerce.date(),
	deleted: z.boolean(),
});
export type SphereContent = z.infer<typeof SphereContentSchema>;

export const SpheresContentPaginationResult = z.object({
	parents: SphereContentSchema.array().min(0),
	childrens: z.record(z.string(), SphereContentSchema.array().min(0)),
});
