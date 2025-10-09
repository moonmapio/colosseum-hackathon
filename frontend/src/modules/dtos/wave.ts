import z from 'zod';
import { MintRecordDto } from './mint';

export type WaveParticipant = z.infer<typeof WaveParticipantSchema>;
export const WaveParticipantSchema = z.object({
	_id: z.string(),
	fullName: z.string(),
	avatarUrl: z.string(),
	walletAddress: z.string(),
});

export type WaveSession = {
	sphereId?: string;
	token?: string;
	joined: boolean;
	participants: WaveParticipant[];
	startedAt?: string;

	mint?: string;
	symbol?: string;
	name?: string;
	uriResolved?: MintRecordDto['uri_resolved'];
};

export type WaveStaats = z.infer<typeof WaveStaatsSchema>;
export const WaveStaatsSchema = z.object({
	sphereId: z.string(),
	active: z.boolean(),
	count: z.number(),
	participants: WaveParticipantSchema.array().min(0),
});
