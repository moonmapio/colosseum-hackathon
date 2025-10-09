import z from 'zod';
import { normalizeChain, ProjectChain } from './common';

export type ProjectResDto = z.infer<typeof ProjectResDtoSchema>;

const IMAGE_FALLBACK = 'https://github.com/shadcn.png';

export const ProjectResDtoSchema = z.object({
	id: z.string(),
	name: z.string().min(1),
	symbol: z.string().min(1),
	chain: z.preprocess(normalizeChain, z.enum(ProjectChain)),
	contractAddress: z.string().optional().nullable(),
	narrative: z.string().optional().nullable(),
	launchDate: z.iso.datetime().optional().nullable(),
	twitter: z.url().or(z.literal('')).optional().nullable(),
	telegram: z.url().or(z.literal('')).optional().nullable(),
	discord: z.url().or(z.literal('')).optional().nullable(),
	website: z.url().or(z.literal('')).optional().nullable(),
	imageUrl: z.url().catch(IMAGE_FALLBACK),
	landingVideoUrl: z.string(),
	devXAccount: z.string(),
	devWallet: z.string().optional().nullable(),
	isVerified: z.boolean().default(false),

	createdAt: z.string().datetime(),
	updatedAt: z.string().datetime(),

	positiveVotes: z.number().default(0),
	negativeVotes: z.number().default(0),

	// landingVideoUrl: z.url().optional().nullable(),
	// launchType: z.enum(['stealth', 'airdrop', 'presale', 'fairlaunch', 'unknown']).default('unknown'),
	// status: z.enum(['draft', 'listed', 'launched', 'archived']).default('draft'),
});

export const CreateProjectSchema = z.object({
	id: z.string().optional(),
	name: z.string().min(1, 'Required'),
	symbol: z.string().min(1, 'Required'),
	chain: z.enum(ProjectChain, { error: 'Select a chain' }),
	contractAddress: z.string().optional(),
	narrative: z.string().optional(),
	launchDate: z.string().optional(), // datetime-local → ISO en submit
	twitter: z.union([z.url(), z.literal('')]).optional(),
	telegram: z.union([z.url(), z.literal('')]).optional(),
	discord: z.union([z.url(), z.literal('')]).optional(),
	website: z.union([z.url(), z.literal('')]).optional(),
	imageUrl: z.url(),
	devWallet: z.string().min(3),
});
