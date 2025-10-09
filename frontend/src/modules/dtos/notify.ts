import { z } from 'zod';

export const NotifyEventSchema = z.object({
	subject: z.string(),
	id: z.string(),
	data: z.string(),
	header: z.object({
		'Nats-Expected-Stream': z.string().array(),
		'Nats-Msg-Id': z.string().array(),
	}),
});

export type NotifyEvent = z.infer<typeof NotifyEventSchema>;
