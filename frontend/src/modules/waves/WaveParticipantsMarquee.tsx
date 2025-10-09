'use client';

import { UserAvatar } from '@modules/spheres/UserAvatar';
import { useEphemeralAppStore } from '@modules/stores/main/ephemeral-hooks';

export function ParticipantsMarquee() {
	const participants = useEphemeralAppStore((s) => s.wave.participants);

	if (!participants?.length) return null;

	return (
		<div className="overflow-hidden w-40 sm:w-60 relative">
			<div className="flex gap-2 animate-marquee" style={{ animationDuration: `${participants.length * 3}s` }}>
				{participants.map((p) => (
					// <img key={p.id} src={p.avatarUrl} title={p.name} alt={p.name} className="w-6 h-6 rounded-full border border-white/30 shrink-0" />
					<UserAvatar key={p._id} imageUrl={p.avatarUrl} alt={p.fullName} height={24} width={24} />
				))}
				{/* repetir para loop infinito */}
				{participants.map((p) => (
					<UserAvatar key={p._id + '-clone'} imageUrl={p.avatarUrl} alt={p.fullName} height={24} width={24} />

					// <img
					// 	key={p.id + '-clone'}
					// 	src={p.avatarUrl}
					// 	title={p.name}
					// 	alt={p.name}
					// 	className="w-6 h-6 rounded-full border border-white/30 shrink-0"
					// />
				))}
			</div>
			<style jsx>{`
				@keyframes marquee {
					0% {
						transform: translateX(0%);
					}
					100% {
						transform: translateX(-50%);
					}
				}
				.animate-marquee {
					width: max-content;
					display: flex;
					animation: marquee linear infinite;
				}
			`}</style>
		</div>
	);
}
